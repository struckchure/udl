package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/charmbracelet/huh"
	"github.com/samber/lo"

	"github.com/struckchure/udl"
	"github.com/struckchure/udl/sites"
)

var (
	// Series
	mobiletvshowsSite = sites.NewMobiletvshowsSite()
	animepaheRu       = sites.NewAnimepaheRu()

	// Movies
	fzmoviesNg = sites.NewFzMoviesNg()
)

var series []udl.ISite = []udl.ISite{mobiletvshowsSite, animepaheRu}
var movies []udl.ISite = []udl.ISite{fzmoviesNg}

func main() {
	versionFlag := flag.Bool("version", false, "Check CLI Version")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("Version: %s\nCommit: %s\nBuild Date: %s\n", version, commit, buildDate)
		return
	}

	var mode string
	err := huh.NewSelect[string]().
		Title("Mode").
		Options(
			huh.NewOption("Series", "series"),
			huh.NewOption("Movies", "movies"),
		).
		Value(&mode).Run()
	if err != nil {
		log.Fatalln(err)
	}

	var siteOptions []udl.ISite
	switch mode {
	case "series":
		siteOptions = series
	case "movies":
		siteOptions = movies
	default:
		log.Fatalf("Unsupported mode: %s", mode)
	}

	var site udl.ISite
	err = huh.NewSelect[udl.ISite]().
		Title(lo.Capitalize(mode) + " / Choose Site").
		Options(
			lo.Map(siteOptions, func(item udl.ISite, idx int) huh.Option[udl.ISite] {
				return huh.NewOption(item.Name(), item)
			})...,
		).
		Value(&site).Run()
	if err != nil {
		log.Fatalln(err)
	}

	err = site.Run(udl.RunOption{})
	if err != nil {
		log.Fatal(err)
	}
}
