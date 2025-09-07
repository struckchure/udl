package sites

// import (
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"net/url"
// 	"path/filepath"
// 	"sync"
// 	"time"

// 	"github.com/charmbracelet/huh"
// 	"github.com/gocolly/colly/v2"
// 	"github.com/samber/lo"
// 	"resty.dev/v3"

// 	"github.com/struckchure/udl"
// )

// type AnimepaheRu struct {
// 	udl.BaseSite
// 	client  *resty.Client
// 	BaseUrl string
// }

// func (m *AnimepaheRu) Name() string {
// 	return fmt.Sprintf("Animepahe - (%s)", m.BaseUrl)
// }

// type AnimeSearchResponse struct {
// 	Total       int         `json:"total"`
// 	PerPage     int         `json:"per_page"`
// 	CurrentPage int         `json:"current_page"`
// 	LastPage    int         `json:"last_page"`
// 	From        int         `json:"from"`
// 	To          int         `json:"to"`
// 	Data        []AnimeData `json:"data"`
// }

// type AnimeData struct {
// 	ID       int     `json:"id"`
// 	Title    string  `json:"title"`
// 	Type     string  `json:"type"`
// 	Episodes int     `json:"episodes"`
// 	Status   string  `json:"status"`
// 	Season   string  `json:"season"`
// 	Year     int     `json:"year"`
// 	Score    float64 `json:"score"`
// 	Poster   string  `json:"poster"`
// 	Session  string  `json:"session"`
// }

// func (m *AnimepaheRu) Run(option udl.RunOption) error {
// 	var search string
// 	err := huh.
// 		NewInput().
// 		Title("What do you want to watch?").
// 		Validate(huh.ValidateNotEmpty()).
// 		Value(&search).Run()
// 	if err != nil {
// 		log.Fatalln(err)
// 	}

// 	// https://animepahe.ru/api?m=search&q=demon%20slayer

// 	c := colly.NewCollector()
// 	c.UserAgent = "Mozilla/5.0 (compatible; udl-bot/1.0)"

// 	var results AnimeSearchResponse
// 	c.OnResponse(func(r *colly.Response) {
// 		if err := json.Unmarshal(r.Body, &results); err != nil {
// 			log.Panicln(err)
// 		}
// 	})

// 	c.Visit(fmt.Sprintf("https://animepahe.ru/api?m=search&q=%s", search))

// 	// res, err := m.client.
// 	// 	R().
// 	// 	SetResult(&results).
// 	// 	SetQueryParam("m", "search").
// 	// 	SetQueryParam("q", search).
// 	// 	SetHeader("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:142.0) Gecko/20100101 Firefox/142.0").
// 	// 	Get("/")
// 	// fmt.Println(res.Request.Header)
// 	// if err != nil {
// 	// 	log.Fatalln(err)
// 	// }

// 	// fmt.Println(res.RawResponse.Body)

// 	// m.listSeasons(udl.Descriptor{Title: results.Data[0].Title, Link: results.Data[0].Session})

// 	return nil
// }

// func (m *AnimepaheRu) listSeasons(series udl.Descriptor) {
// 	c := colly.NewCollector()

// 	results := []huh.Option[udl.Descriptor]{}

// 	target := `div[itemprop="containsSeason"] > div.mainbox2`
// 	c.OnHTML(target, func(e *colly.HTMLElement) {
// 		results = append(
// 			results,
// 			huh.NewOption(
// 				e.ChildText("a[href]"),
// 				udl.Descriptor{
// 					Title: e.ChildText("a[href]"),
// 					Link:  lo.Must(url.JoinPath(m.BaseUrl, e.ChildAttr("a[href]", "href"))),
// 				},
// 			),
// 		)
// 	})

// 	c.OnScraped(func(r *colly.Response) {
// 		var season udl.Descriptor
// 		err := huh.NewSelect[udl.Descriptor]().
// 			Title("Choose Series").
// 			Options(results...).
// 			Value(&season).Run()
// 		if err != nil {
// 			log.Fatalln(err)
// 		}

// 		m.listEpisodes(season)
// 	})

// 	c.Visit(series.Link)
// }

// func (m *AnimepaheRu) listEpisodes(season udl.Descriptor) {
// 	c := colly.NewCollector()

// 	results := []huh.Option[udl.Descriptor]{}

// 	target := `div.mainbox > table:first-child > tbody:first-child > tr:first-child > td:nth-child(2) > span:first-child`
// 	c.OnHTML(target, func(e *colly.HTMLElement) {
// 		results = append(results, huh.NewOption(
// 			e.ChildText("small:first-child"),
// 			udl.Descriptor{
// 				Title: e.ChildText("small:first-child") + "/ High MP4",
// 				Link:  m.BaseUrl + "/" + e.ChildAttr("a[href]:nth-child(2)", "href"),
// 			},
// 		),
// 		)
// 	})

// 	c.OnScraped(func(r *colly.Response) {
// 		var episodes []udl.Descriptor
// 		err := huh.NewMultiSelect[udl.Descriptor]().
// 			Title("Choose Series").
// 			Options(results...).
// 			Value(&episodes).Run()
// 		if err != nil {
// 			log.Fatalln(err)
// 		}

// 		m.bulkDownload(episodes)
// 	})

// 	c.Visit(season.Link)
// }

// func (m *AnimepaheRu) bulkDownload(episodes []udl.Descriptor) {
// 	start := time.Now()

// 	var wg sync.WaitGroup

// 	for _, episode := range episodes {
// 		wg.Add(1)
// 		go func(ep udl.Descriptor) {
// 			defer wg.Done()
// 			m.download(ep)
// 		}(episode) // pass as arg to avoid closure capture issue
// 	}

// 	wg.Wait() // Wait for all goroutines to finish

// 	elapsed := time.Since(start)
// 	fmt.Printf("Took %.2f minute(s) to download %d episode(s)!", elapsed.Minutes(), len(episodes))
// }

// func (m *AnimepaheRu) download(episode udl.Descriptor) {
// 	c := colly.NewCollector()

// 	target := "div.mainbox2:nth-child(31) > table:nth-child(1) > tbody:nth-child(1) > tr:nth-child(1) > td:nth-child(2) > span:nth-child(1) > a:nth-child(1)"
// 	c.OnHTML(target, func(e *colly.HTMLElement) {
// 		episode.Link = m.BaseUrl + "/" + e.Attr("href")
// 		e.Request.Visit(episode.Link)
// 	})

// 	target = "div.downloadlinks2:nth-child(12) > p:nth-child(2) > input:nth-child(1)"
// 	c.OnHTML(target, func(e *colly.HTMLElement) {
// 		episode.Link = e.Attr("value")

// 		path := filepath.Base(episode.Link)
// 		err := udl.DownloadWithProgress(episode.Link, path)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 	})

// 	c.Visit(episode.Link)
// }

// func NewAnimepaheRu() udl.ISite {
// 	client := resty.New().SetBaseURL("https://animepahe.ru/api")
// 	defer client.Close()

// 	return &AnimepaheRu{
// 		client:  client,
// 		BaseUrl: "https://animepahe.ru",
// 	}
// }
