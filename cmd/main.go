package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/struckchure/udl/sites"
	"github.com/struckchure/udl/types"
)

func main() {
	versionFlag := flag.Bool("version", false, "Check CLI Version")

	flag.Parse()
	if *versionFlag {
		fmt.Printf("Version: %s\nCommit: %s\nBuild Date: %s\n", version, commit, buildDate)
		return
	}

	site := sites.NewMobiletvshowsSite()
	err := site.Run(types.RunOption{})
	if err != nil {
		log.Fatal(err)
	}
}
