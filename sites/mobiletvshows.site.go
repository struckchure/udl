package sites

import (
	"fmt"
	"log"
	"net/url"
	"path/filepath"
	"sync"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/gocolly/colly/v2"
	"github.com/samber/lo"

	"github.com/struckchure/udl"
)

type IMobileTvShowsSite interface {
	udl.ISite
}

type MobileTvShowsSite struct {
	BaseUrl string
}

func (m *MobileTvShowsSite) Name() string {
	return fmt.Sprintf("MobileTvShows - (%s)", m.BaseUrl)
}

func (m *MobileTvShowsSite) Run(option udl.RunOption) error {
	c := colly.NewCollector()

	results := []huh.Option[udl.Descriptor]{}

	target := "div.mainbox3 > table:first-child > tbody:first-child > tr > td:nth-child(2) span"
	c.OnHTML(target, func(e *colly.HTMLElement) {
		desc := e.ChildText("small:nth-child(4)")
		results = append(results, huh.NewOption(
			e.ChildText("a[href]"),
			udl.Descriptor{
				Title: e.ChildText("a[href]") + lo.Ternary(desc != "", " / "+desc, ""),
				Link:  lo.Must(url.JoinPath(m.BaseUrl, e.ChildAttr("a[href]", "href"))),
			}),
		)
	})

	c.OnScraped(func(r *colly.Response) {
		var series udl.Descriptor
		err := huh.NewSelect[udl.Descriptor]().
			Title("Choose Series").
			Options(results...).
			Value(&series).Run()
		if err != nil {
			log.Fatalln(err)
		}

		m.ListSeasons(series)
	})

	var search string
	err := huh.
		NewInput().
		Title("What do you want to watch?").
		Validate(huh.ValidateNotEmpty()).
		Value(&search).Run()
	if err != nil {
		log.Fatalln(err)
	}

	query := udl.Query{"search": search}

	c.Visit(lo.Must(url.JoinPath(m.BaseUrl, "search.php")) + "?" + query.String())

	return nil
}

func (m *MobileTvShowsSite) ListSeasons(series udl.Descriptor) {
	c := colly.NewCollector()

	results := []huh.Option[udl.Descriptor]{}

	target := `div[itemprop="containsSeason"] > div.mainbox2`
	c.OnHTML(target, func(e *colly.HTMLElement) {
		results = append(
			results,
			huh.NewOption(
				e.ChildText("a[href]"),
				udl.Descriptor{
					Title: e.ChildText("a[href]"),
					Link:  lo.Must(url.JoinPath(m.BaseUrl, e.ChildAttr("a[href]", "href"))),
				},
			),
		)
	})

	c.OnScraped(func(r *colly.Response) {
		var season udl.Descriptor
		err := huh.NewSelect[udl.Descriptor]().
			Title("Choose Series").
			Options(results...).
			Value(&season).Run()
		if err != nil {
			log.Fatalln(err)
		}

		m.ListEpisodes(season)
	})

	c.Visit(series.Link)
}

func (m *MobileTvShowsSite) ListEpisodes(season udl.Descriptor) {
	c := colly.NewCollector()

	results := []huh.Option[udl.Descriptor]{}

	target := `div.mainbox > table:first-child > tbody:first-child > tr:first-child > td:nth-child(2) > span:first-child`
	c.OnHTML(target, func(e *colly.HTMLElement) {
		results = append(results, huh.NewOption(
			e.ChildText("small:first-child"),
			udl.Descriptor{
				Title: e.ChildText("small:first-child") + "/ High MP4",
				Link:  m.BaseUrl + "/" + e.ChildAttr("a[href]:nth-child(2)", "href"),
			},
		),
		)
	})

	c.OnScraped(func(r *colly.Response) {
		var episodes []udl.Descriptor
		err := huh.NewMultiSelect[udl.Descriptor]().
			Title("Choose Series").
			Options(results...).
			Value(&episodes).Run()
		if err != nil {
			log.Fatalln(err)
		}

		m.BulkDownload(episodes)
	})

	c.Visit(season.Link)
}

func (m *MobileTvShowsSite) BulkDownload(episodes []udl.Descriptor) {
	start := time.Now()

	var wg sync.WaitGroup

	for _, episode := range episodes {
		wg.Add(1)
		go func(ep udl.Descriptor) {
			defer wg.Done()
			m.Download(ep)
		}(episode) // pass as arg to avoid closure capture issue
	}

	wg.Wait() // Wait for all goroutines to finish

	elapsed := time.Since(start)
	fmt.Printf("Took %.2f minute(s) to download %d episode(s)!", elapsed.Minutes(), len(episodes))
}

func (m *MobileTvShowsSite) Download(episode udl.Descriptor) {
	c := colly.NewCollector()

	target := "div.mainbox2:nth-child(31) > table:nth-child(1) > tbody:nth-child(1) > tr:nth-child(1) > td:nth-child(2) > span:nth-child(1) > a:nth-child(1)"
	c.OnHTML(target, func(e *colly.HTMLElement) {
		episode.Link = m.BaseUrl + "/" + e.Attr("href")
		e.Request.Visit(episode.Link)
	})

	target = "div.downloadlinks2:nth-child(12) > p:nth-child(2) > input:nth-child(1)"
	c.OnHTML(target, func(e *colly.HTMLElement) {
		episode.Link = e.Attr("value")

		path := filepath.Base(episode.Link)
		err := udl.DownloadWithProgress(episode.Link, path)
		if err != nil {
			log.Fatal(err)
		}
	})

	c.Visit(episode.Link)
}

func NewMobiletvshowsSite() udl.ISite {
	return &MobileTvShowsSite{BaseUrl: "https://mobiletvshows.site"}
}
