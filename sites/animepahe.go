package sites

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/gocolly/colly/v2"
	"github.com/samber/lo"
	"github.com/struckchure/udl"
)

type Animepahe struct {
	udl.BaseSite
	APIBase string
}

func (a *Animepahe) Name() string {
	return "Animepahe"
}

func (a *Animepahe) Run(option udl.RunOption) error {
	a.APIBase = "https://animepahe.ru/api"

	// Step 1: Prompt user for search
	var search string
	err := huh.
		NewInput().
		Title("What anime do you want to watch?").
		Validate(huh.ValidateNotEmpty()).
		Value(&search).Run()
	if err != nil {
		log.Fatal(err)
	}

	// Step 2: Search anime
	searchResults, err := a.searchAnime(search)
	if err != nil {
		log.Fatal(err)
	}
	if len(searchResults) == 0 {
		log.Fatal("No results found.")
	}
	log.Printf("🔗 SEARCHRESULT %s", searchResults)
	// Step 3: Prompt user to select series
	options := lo.Map(searchResults, func(item AnimeSearchResult, _ int) huh.Option[AnimeSearchResult] {
		return huh.NewOption(item.Title, item)
	})

	var selectedAnime AnimeSearchResult
	if err := huh.NewSelect[AnimeSearchResult]().
		Title("Choose Anime").
		Options(options...).
		Value(&selectedAnime).Run(); err != nil {
		log.Fatal(err)
	}

	// Step 4: List episodes
	episodes, err := a.fetchEpisodes(selectedAnime.ID)
	if err != nil {
		log.Fatal(err)
	}
	if len(episodes) == 0 {
		log.Fatal("No episodes found.")
	}

	// Step 5: Multi-select episodes (like MobileTvShows does)
	epOptions := lo.Map(episodes, func(ep AnimeEpisode, _ int) huh.Option[AnimeEpisode] {
		title := fmt.Sprintf("Episode %d [%s]", ep.Episode, ep.Fansub)
		return huh.NewOption(title, ep)
	})

	var selectedEpisodes []AnimeEpisode
	if err := huh.NewMultiSelect[AnimeEpisode]().
		Title("Choose Episodes").
		Options(epOptions...).
		Value(&selectedEpisodes).Run(); err != nil {
		log.Fatal(err)
	}

	// Step 6: Bulk download (like MobileTvShows)
	a.bulkDownload(selectedEpisodes)

	return nil
}

// ========== INTERNAL TYPES AND METHODS ==========

type AnimeSearchResponse struct {
	Data []AnimeSearchResult `json:"data"`
}

type AnimeSearchResult struct {
	Title string `json:"title"`
	ID    int    `json:"id"`
}

type AnimeEpisodesResponse struct {
	Data []AnimeEpisode `json:"data"`
}

type AnimeEpisode struct {
	Episode int    `json:"episode"`
	Session string `json:"session"`
	Fansub  string `json:"fansub"`
}

func (a *Animepahe) searchAnime(query string) ([]AnimeSearchResult, error) {
	endpoint := fmt.Sprintf("%s?m=search&q=%s", a.APIBase, url.QueryEscape(query))

	client := &http.Client{}
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; udl-bot/1.0)")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	
	var response AnimeSearchResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %v", err)
	}

	return response.Data, nil
}

func (a *Animepahe) fetchEpisodes(id int) ([]AnimeEpisode, error) {
	endpoint := fmt.Sprintf("%s?m=release&id=%d&sort=episode_asc", a.APIBase, id)

	client := &http.Client{}
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; udl-bot/1.0)")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	
	var response AnimeEpisodesResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, fmt.Errorf("failed to parse episodes response: %v", err)
	}

	return response.Data, nil
}

// Bulk download function similar to MobileTvShows
func (a *Animepahe) bulkDownload(episodes []AnimeEpisode) {
	start := time.Now()

	var wg sync.WaitGroup

	for _, episode := range episodes {
		wg.Add(1)
		go func(ep AnimeEpisode) {
			defer wg.Done()
			a.downloadEpisode(ep)
		}(episode) // pass as arg to avoid closure capture issue
	}

	wg.Wait() // Wait for all goroutines to finish

	elapsed := time.Since(start)
	fmt.Printf("Took %.2f minute(s) to download %d episode(s)!\n", elapsed.Minutes(), len(episodes))
}

func (a *Animepahe) downloadEpisode(episode AnimeEpisode) {
	// Step 6: Scrape download URL from episode page
	downloadURL, err := a.resolveStreamURL(episode.Session)
	if err != nil {
		log.Printf("Failed to resolve stream URL for episode %d: %v", episode.Episode, err)
		return
	}

	// Step 7: Download
	fmt.Printf("Downloading Episode %d...\n", episode.Episode)
	err = a.download(downloadURL, episode.Episode)
	if err != nil {
		log.Printf("Failed to download episode %d: %v", episode.Episode, err)
	}
}

func (a *Animepahe) resolveStreamURL(session string) (string, error) {
	streamPage := fmt.Sprintf("https://animepahe.ru/play/%s", session)

	var resolved string
	c := colly.NewCollector()
	
	// Set user agent like other scrapers
	c.UserAgent = "Mozilla/5.0 (compatible; udl-bot/1.0)"

	// Look for various patterns that might contain the stream URL
	c.OnHTML("script", func(e *colly.HTMLElement) {
		text := e.Text
		
		// Try multiple extraction patterns
		patterns := []string{
			`https://[^"'\s]+\.m3u8[^"'\s]*`,
			`https://[^"'\s]+\.mp4[^"'\s]*`,
			`https://kwik\.cx/[^"'\s]+`,
		}
		
		for _, pattern := range patterns {
			re := regexp.MustCompile(pattern)
			matches := re.FindStringSubmatch(text)
			if len(matches) > 0 {
				resolved = matches[0]
				return
			}
		}
	})

	// Also check for direct links in the page
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		href := e.Attr("href")
		if strings.Contains(href, ".mp4") || strings.Contains(href, ".m3u8") {
			resolved = href
		}
	})

	err := c.Visit(streamPage)
	if err != nil {
		return "", err
	}
	
	if resolved == "" {
		return "", fmt.Errorf("failed to extract stream link from session: %s", session)
	}
	
	return resolved, nil
}

func (a *Animepahe) download(link string, episodeNum int) error {
	if link == "" {
		return fmt.Errorf("empty link")
	}
	
	// Create a more descriptive filename
	filename := fmt.Sprintf("Episode_%d_%s", episodeNum, filepath.Base(link))
	
	// Handle cases where the base might not have an extension
	if !strings.Contains(filename, ".") {
		filename += ".mp4"
	}
	
	return udl.DownloadWithProgress(link, filename)
}

func NewAnimepahe() udl.ISite {
	return &Animepahe{}
}