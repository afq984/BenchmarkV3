package main

import (
	"flag"
	"log"
	"os"
	"time"
)

func main() {
	log.SetFlags(log.Lshortfile)

	var (
		config string
		detect bool
	)
	flag.StringVar(&config, "c", "", "config to use")
	flag.BoolVar(&detect, "detect", false, "detect the system only; don't run any benchmarks")
	flag.Parse()

	var (
		track string
		dt    time.Duration
		err   error
	)

	if detect {
		dt = 0
		config = "detect"
		track = "detect"
	} else {
		if quick {
			track = "quick"
		} else {
			track = "standard"
		}

		cfg, ok := configs[config]
		if !ok {
			log.Printf("unknown config: %q", config)
			os.Exit(1)
		}
		dt, err = Build(cfg)

		if err != nil {
			log.Println("benchmark failed")
			os.Exit(1)
		}
	}

	r := &Result{
		Track:  track,
		Config: config,
		Time:   dt,
	}
	populateSystem(r)

	r.PrettyPrint()
}
