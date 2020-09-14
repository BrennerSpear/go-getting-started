package main

import (
	"flag"

	log "github.com/sirupsen/logrus"

	"github.com/BrennerSpear/go-getting-started/helpers"
)

var (
	optConfig = flag.String("c", "", "config file to be used, defaults to settings.toml in the same directory")
)

func main() {
	flag.Parse()

	var conf helpers.Config
	if *optConfig != "" {
		conf = helpers.LoadFile(*optConfig)
	} else {
		conf = helpers.Load()
	}

	helpers.SetServerLocation(&conf)
	//results.Initialize(&conf)
	//database.SetDBInfo(&conf)
	log.Fatal(helpers.ListenAndServe(&conf))
}