package main

import (
	"flag"
	"log"
	"os"
)

var (
	configFile string
)

func init() {
	flag.StringVar(&configFile, "c", "/etc/goget-url.conf", "Path to configuration file")
}

func main() {
	flag.Parse()

	c, err := parseConfigFile(configFile)
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}

	log.Fatal(listenAndServe(c))
}
