package sites

import (
	"fmt"
	"log"
	"net/url"
	"path/filepath"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gocolly/colly/v2"
	"github.com/samber/lo"

	"github.com/struckchure/udl"
	"github.com/struckchure/udl/types"
)

type IMobileTvShowsSite interface {
	types.ISite
}

type MobileTvShowsSite struct {
	BaseUrl string
}

func (m *MobileTvShowsSite) Name() string {
	return fmt.Sprintf("MobileTvShows - (%s)", m.BaseUrl)
}

func (m *MobileTvShowsSite) Run(option types.RunOption) error {
	c := colly.NewCollector()

	results := []list.Item{}

	target := "div.mainbox3 > table:first-child > tbody:first-child > tr > td:nth-child(2) span"
	c.OnHTML(target, func(e *colly.HTMLElement) {
		results = append(results, udl.SelectModel{
			Title: e.ChildText("a[href]"),
			Value: lo.Must(url.JoinPath(m.BaseUrl, e.ChildAttr("a[href]", "href"))),
			Desc:  e.ChildText("small:nth-child(4)"),
		})
	})

	c.OnScraped(func(r *colly.Response) {
		form := udl.SelectModelForm{Model: list.New(results, udl.ItemDelegate{}, 0, 0)}
		form.Model.Title = "Select Series"

		p := tea.NewProgram(form, tea.WithAltScreen())
		payload, err := p.Run()
		if err != nil {
			log.Fatal(err)
		}

		season, ok := payload.(udl.SelectModelForm).Model.SelectedItem().(udl.SelectModel)
		if !ok {
			return
		}

		m.ListSeasons(types.Descriptor{
			Title: season.Title,
			Link:  season.Value,
		})
	})

	form := udl.InputModelForm()
	form.Input.Placeholder = "Search for Series"

	p := tea.NewProgram(form)
	payload, err := p.Run()
	if err != nil {
		log.Fatal(err)
	}

	query := udl.Query{"search": payload.(udl.InputModel).Input.Value()}

	c.Visit(lo.Must(url.JoinPath(m.BaseUrl, "search.php")) + "?" + query.String())

	return nil
}

func (m *MobileTvShowsSite) ListSeasons(series types.Descriptor) {
	c := colly.NewCollector()

	results := []list.Item{}

	target := `div[itemprop="containsSeason"] > div.mainbox2`
	c.OnHTML(target, func(e *colly.HTMLElement) {
		results = append(results, udl.SelectModel{
			Title: e.ChildText("a[href]"),
			Value: lo.Must(url.JoinPath(m.BaseUrl, e.ChildAttr("a[href]", "href"))),
			Desc:  "",
		})
	})

	c.OnScraped(func(r *colly.Response) {
		form := udl.SelectModelForm{Model: list.New(results, udl.ItemDelegate{}, 0, 0)}
		form.Model.Title = "Select Season"

		p := tea.NewProgram(form, tea.WithAltScreen())
		payload, err := p.Run()
		if err != nil {
			log.Fatal(err)
		}

		season, ok := payload.(udl.SelectModelForm).Model.SelectedItem().(udl.SelectModel)
		if !ok {
			return
		}

		m.ListEpisodes(types.Descriptor{
			Title: season.Title,
			Link:  season.Value,
		})
	})

	c.Visit(series.Link)
}

func (m *MobileTvShowsSite) ListEpisodes(season types.Descriptor) {
	c := colly.NewCollector()

	results := []list.Item{}

	target := `div.mainbox > table:first-child > tbody:first-child > tr:first-child > td:nth-child(2) > span:first-child`
	c.OnHTML(target, func(e *colly.HTMLElement) {
		results = append(results, udl.SelectModel{
			Title: e.ChildText("small:first-child"),
			Value: m.BaseUrl + "/" + e.ChildAttr("a[href]:nth-child(2)", "href"),
			Desc:  "High MP4",
		})
	})

	c.OnScraped(func(r *colly.Response) {
		form := udl.SelectModelForm{Model: list.New(results, udl.ItemDelegate{}, 0, 0)}
		form.Model.Title = "Select Episode"

		p := tea.NewProgram(form, tea.WithAltScreen())
		payload, err := p.Run()
		if err != nil {
			log.Fatal(err)
		}

		season, ok := payload.(udl.SelectModelForm).Model.SelectedItem().(udl.SelectModel)
		if !ok {
			return
		}

		m.Download(types.Descriptor{
			Title: season.Title,
			Link:  season.Value,
		})
	})

	c.Visit(season.Link)
}

func (m *MobileTvShowsSite) Download(episode types.Descriptor) {
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

func NewMobiletvshowsSite() types.ISite {
	return &MobileTvShowsSite{BaseUrl: "https://mobiletvshows.site"}
}
