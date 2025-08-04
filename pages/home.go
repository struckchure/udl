package pages

import (
	"log"

	"fyne.io/fyne/v2/widget"
)

func Hello() *widget.Button {
	return widget.NewButton("click me", func() {
		log.Println("tapped")
	})
}
