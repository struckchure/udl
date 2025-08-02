package udl

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

type Query map[string]string

func (s Query) String() string {
	values := url.Values{}
	for key, val := range s {
		values.Set(key, val)
	}

	return values.Encode()
}

func DownloadWithProgress(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	size := resp.ContentLength
	if size <= 0 {
		fmt.Println("Warning: could not determine file size")
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Create and start the progress bar
	pb := NewProgressBar()
	pb.Start()

	var downloaded int64
	buf := make([]byte, 32*1024) // 32KB buffer
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			out.Write(buf[:n])
			downloaded += int64(n)
			if size > 0 {
				percent := float64(downloaded) / float64(size)
				pb.Update(percent)
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}

	pb.Update(1.0) // ensure full
	pb.Stop()

	fmt.Println("\nDownload complete:", filepath)
	return nil
}
