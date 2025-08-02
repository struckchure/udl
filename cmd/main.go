package main

import (
	"log"

	"github.com/struckchure/udl/sites"
	"github.com/struckchure/udl/types"
)

func main() {
	site := sites.NewMobiletvshowsSite()
	err := site.Run(types.RunOption{})
	if err != nil {
		log.Fatal(err)
	}
}
