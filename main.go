package main

import (
	"flag"
	"log"
	"os"
)

func main() {
	log.SetFlags(log.Lshortfile)

	c := flag.String("c", "", "config to use")
	flag.Parse()
	cfg, ok := configs[*c]
	if !ok {
		log.Printf("unknown config: %q", *c)
		os.Exit(1)
	}

	err := Build(cfg)
	if err != nil {
		os.Exit(1)
	}
}
