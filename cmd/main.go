package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/charmbracelet/huh"
	"github.com/samber/lo"
	"github.com/struckchure/udl/sites"
	"github.com/struckchure/udl/types"
)

var (
	// series
	mobiletvshowsSite = sites.NewMobiletvshowsSite()

	// movies
	fzmoviesNg = sites.NewFzMoviesNg()
)

var series []types.ISite = []types.ISite{mobiletvshowsSite}
var movies []types.ISite = []types.ISite{fzmoviesNg}

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

	var site types.ISite
	err = huh.NewSelect[types.ISite]().
		Title(lo.Capitalize(mode) + " / Choose Site").
		Options(
			lo.Map(
				lo.Ternary(mode == "series", series, movies),
				func(item types.ISite, idx int) huh.Option[types.ISite] {
					return huh.NewOption(item.Name(), item)
				},
			)...,
		).
		Value(&site).Run()
	if err != nil {
		log.Fatalln(err)
	}

	err = site.Run(types.RunOption{})
	if err != nil {
		log.Fatal(err)
	}
}
