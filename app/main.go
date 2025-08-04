package main

import (
	"fyne.io/fyne/v2/app"

	"github.com/struckchure/udl/pages"
)

func main() {
	a := app.New()
	w := a.NewWindow("Universal Downloader")

	w.SetContent(pages.Hello())
	w.ShowAndRun()
}
