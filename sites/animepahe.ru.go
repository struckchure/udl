package sites

import (
	"encoding/json"
	"fmt"
	"log"
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

type AnimepaheRu struct {
	udl.BaseSite
	BaseUrl string
	APIBase string
}

func (a *AnimepaheRu) Name() string {
	return fmt.Sprintf("Animepahe - (%s)", a.BaseUrl)
}

func (a *AnimepaheRu) Run(option udl.RunOption) error {
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
	if len(searchResults.Data) == 0 {
		log.Fatal("No results found.")
	}

	// Step 3: Prompt user to select series
	options := lo.Map(searchResults.Data, func(item AnimeData, _ int) huh.Option[AnimeData] {
		return huh.NewOption(item.Title, item)
	})

	var selectedAnime AnimeData
	if err := huh.NewSelect[AnimeData]().
		Title("Choose Anime").
		Options(options...).
		Value(&selectedAnime).Run(); err != nil {
		log.Fatal(err)
	}

	// Step 4: List episodes
	episodes, err := a.fetchEpisodes(selectedAnime.Session)
	if err != nil {
		log.Fatal(err)
	}
	if len(episodes) == 0 {
		log.Fatal("No episodes found.")
	}

	// Step 5: Multi-select episodes
	epOptions := lo.Map(episodes, func(ep AnimeEpisode, _ int) huh.Option[AnimeEpisode] {
		title := fmt.Sprintf("%s - Episode %d", selectedAnime.Title, ep.Episode)
		return huh.NewOption(title, ep)
	})

	var selectedEpisodes []AnimeEpisode
	err = huh.NewMultiSelect[AnimeEpisode]().
		Title("Choose Episodes").
		Options(epOptions...).
		Value(&selectedEpisodes).
		Run()
	if err != nil {
		log.Fatal(err)
	}

	// Step 6: Bulk download
	a.bulkDownload(selectedAnime.Session, selectedEpisodes)

	return nil
}

type AnimeSearchResponse struct {
	Total       int         `json:"total"`
	PerPage     int         `json:"per_page"`
	CurrentPage int         `json:"current_page"`
	LastPage    int         `json:"last_page"`
	From        int         `json:"from"`
	To          int         `json:"to"`
	Data        []AnimeData `json:"data"`
}

type AnimeData struct {
	ID       int     `json:"id"`
	Title    string  `json:"title"`
	Type     string  `json:"type"`
	Episodes int     `json:"episodes"`
	Status   string  `json:"status"`
	Season   string  `json:"season"`
	Year     int     `json:"year"`
	Score    float64 `json:"score"`
	Poster   string  `json:"poster"`
	Session  string  `json:"session"`
}

func (a *AnimepaheRu) searchAnime(query string) (*AnimeSearchResponse, error) {
	endpoint := fmt.Sprintf("%s?m=search&q=%s", a.APIBase, url.QueryEscape(query))

	c := colly.NewCollector()
	c.UserAgent = "Mozilla/5.0 (compatible; udl-bot/1.0)"

	var results AnimeSearchResponse

	c.OnResponse(func(r *colly.Response) {
		if err := json.Unmarshal(r.Body, &results); err != nil {
			log.Printf("Failed to parse search response: %v", err)
		}
	})

	if err := c.Visit(endpoint); err != nil {
		return nil, err
	}

	return &results, nil
}

type AnimeEpisodesResponse struct {
	Total       int            `json:"total"`
	PerPage     int            `json:"per_page"`
	CurrentPage int            `json:"current_page"`
	LastPage    int            `json:"last_page"`
	NextPageURL *string        `json:"next_page_url"`
	PrevPageURL *string        `json:"prev_page_url"`
	From        int            `json:"from"`
	To          int            `json:"to"`
	Data        []AnimeEpisode `json:"data"`
}

type AnimeEpisode struct {
	ID        int    `json:"id"`
	AnimeID   int    `json:"anime_id"`
	Episode   int    `json:"episode"`
	Episode2  int    `json:"episode2"`
	Edition   string `json:"edition"`
	Title     string `json:"title"`
	Snapshot  string `json:"snapshot"`
	Disc      string `json:"disc"`
	Audio     string `json:"audio"`
	Duration  string `json:"duration"`
	Session   string `json:"session"`
	Filler    int    `json:"filler"`
	CreatedAt string `json:"created_at"`
}

func (a *AnimepaheRu) fetchEpisodes(id string) ([]AnimeEpisode, error) {
	c := colly.NewCollector()
	c.UserAgent = "Mozilla/5.0 (compatible; udl-bot/1.0)"

	var results AnimeEpisodesResponse
	c.OnResponse(func(r *colly.Response) {
		if err := json.Unmarshal(r.Body, &results); err != nil {
			log.Printf("Failed to parse episodes response: %v", err)
		}
	})

	endpoint := fmt.Sprintf("%s?m=release&id=%s&sort=episode_asc", a.APIBase, id)
	if err := c.Visit(endpoint); err != nil {
		return nil, err
	}

	return results.Data, nil
}

func (a *AnimepaheRu) bulkDownload(season string, episodes []AnimeEpisode) {
	start := time.Now()

	var wg sync.WaitGroup

	for _, episode := range episodes {
		wg.Add(1)
		go func(ep AnimeEpisode) {
			defer wg.Done()
			a.downloadEpisode(season, ep)
		}(episode)
	}

	wg.Wait()

	elapsed := time.Since(start)
	fmt.Printf("Took %.2f minute(s) to download %d episode(s)!\n", elapsed.Minutes(), len(episodes))
}

func (a *AnimepaheRu) downloadEpisode(season string, episode AnimeEpisode) {
	downloadURL, err := a.resolveStreamURL(season, episode.Session)
	if err != nil {
		log.Printf("Failed to resolve stream URL for episode %d: %v", episode.Episode, err)
		return
	}

	fmt.Printf("Downloading Episode %d...\n", episode.Episode)
	err = a.download(downloadURL, episode.Episode)
	if err != nil {
		log.Printf("Failed to download episode %d: %v", episode.Episode, err)
	}
}

func (a *AnimepaheRu) resolveStreamURL(season string, episode string) (string, error) {
	streamPage := fmt.Sprintf("https://animepahe.ru/play/%s/%s", season, episode)

	c := colly.NewCollector()
	c.UserAgent = "Mozilla/5.0 (compatible; udl-bot/1.0)"

	var resolved string

	c.OnHTML("#pickDownload > a[href]", func(e *colly.HTMLElement) {
		text := e.Text
		fmt.Println(text)

		patterns := []string{
			`https://[^"'\s]+\.m3u8[^"'\s]*`,
			`https://[^"'\s]+\.mp4[^"'\s]*`,
			`https://kwik\.cx/[^"'\s]+`,
		}

		for _, pattern := range patterns {
			re := regexp.MustCompile(pattern)
			if matches := re.FindStringSubmatch(text); len(matches) > 0 {
				resolved = matches[0]
				return
			}
		}
	})

	err := c.Visit(streamPage)
	if err != nil {
		return "", err
	}

	if resolved == "" {
		return "", fmt.Errorf("failed to extract stream link from session: %s", episode)
	}

	return resolved, nil
}

func (a *AnimepaheRu) download(link string, episodeNum int) error {
	if link == "" {
		return fmt.Errorf("empty link")
	}

	filename := fmt.Sprintf("Episode_%d_%s", episodeNum, filepath.Base(link))
	if !strings.Contains(filename, ".") {
		filename += ".mp4"
	}

	return udl.DownloadWithProgress(link, filename)
}

func NewAnimepaheRu() udl.ISite {
	return &AnimepaheRu{BaseUrl: "https://animepahe.ru"}
}
