package sites

import (
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/charmbracelet/huh"
	"github.com/gocolly/colly/v2"
	"github.com/samber/lo"

	"github.com/struckchure/udl"
)

type FzMoviesNg struct {
	udl.BaseSite
	BaseUrl string
}

func (m *FzMoviesNg) Name() string {
	return fmt.Sprintf("FzMovies - (%s)", m.BaseUrl)
}

func (m *FzMoviesNg) Run(option udl.RunOption) error {
	c := colly.NewCollector()

	var nextLink string
	var stage string
	results := []huh.Option[udl.Descriptor]{}

	// search results
	target := "div.magsoul-grid-post-inside div.magsoul-grid-post-details.magsoul-grid-post-block h3.magsoul-grid-post-title a"
	c.OnHTML(target, func(e *colly.HTMLElement) {
		results = append(
			results,
			huh.NewOption(
				e.Text,
				udl.Descriptor{Title: e.Text, Link: e.Attr("href")},
			),
		)
	})

	// download page - 1
	target = "div.magsoul-box-inside div.entry-content.magsoul-clearfix a[href]"
	c.OnHTML(target, func(e *colly.HTMLElement) {
		stage = "2"
		e.Request.Visit(e.Attr("href"))
	})

	// meetdownload page
	target = "div.bezende > script:first-child"
	c.OnHTML(target, func(e *colly.HTMLElement) {
		re := regexp.MustCompile(`document\.getElementById\(['"]downloadButton['"]\)\.onclick\s*=\s*function\s*\([^)]*\)\s*\{[^}]*location\.href\s*=\s*'([^']+)'`)
		matches := re.FindStringSubmatch(e.Text)
		if len(matches) < 1 {
			log.Fatal("No URL found")
		}

		m.download(strings.TrimSpace(matches[1]))
	})

	c.OnError(func(r *colly.Response, _ error) {
		// for some reason, the site throws a 404 even when the page exists, but it's parseable

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(r.Body)))
		if err != nil {
			log.Fatal(err)
		}

		downloadButton := doc.Find("#downloadButton")
		r.Request.Visit(lo.Must(downloadButton.Attr("href")))
		stage = "3"
	})

	c.OnScraped(func(r *colly.Response) {
		if lo.Contains([]string{"1", "2", "3"}, stage) {
			return
		}

		var series udl.Descriptor
		err := huh.NewSelect[udl.Descriptor]().
			Title("Choose Movie").
			Options(results...).
			Value(&series).Run()
		if err != nil {
			log.Fatal(err)
		}

		stage = "1"
		nextLink = series.Link
		r.Request.Visit(nextLink)
	})

	var search string
	err := huh.
		NewInput().
		Title("What do you want to watch?").
		Validate(huh.ValidateNotEmpty()).
		Value(&search).Run()
	if err != nil {
		log.Fatal(err)
	}

	query := udl.Query{"s": search}

	c.Visit(m.BaseUrl + "/?" + query.String())

	return nil
}

func (m *FzMoviesNg) download(link string) {
	if link != "" {
		splittedLink := strings.Split(link, "-")
		prevExt := splittedLink[len(splittedLink)-1]
		filename := strings.Replace(link, prevExt, "."+prevExt, 1)
		path := filepath.Base(filename)

		err := udl.DownloadWithProgress(link, path)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func NewFzMoviesNg() udl.ISite {
	return &FzMoviesNg{BaseUrl: "https://www.fzmovies.ng"}
}
