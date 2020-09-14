package main

import (
	"flag"

	log "github.com/sirupsen/logrus"

	"github.com/brennerspear/go-getting-started/helpers"
)

var (
	optConfig = flag.String("c", "", "config file to be used, defaults to settings.toml in the same directory")
)

func main() {
	flag.Parse()

	var conf config.Config
	if *optConfig != "" {
		conf = config.LoadFile(*optConfig)
	} else {
		conf = config.Load()
	}

	web.SetServerLocation(&conf)
	//results.Initialize(&conf)
	//database.SetDBInfo(&conf)
	log.Fatal(web.ListenAndServe(&conf))
}
