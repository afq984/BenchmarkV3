package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/pflag"
)

func main() {
	log.SetFlags(log.Lshortfile)

	var (
		config       string
		detect       bool
		outputURL    string
		downloadOnly bool
	)
	pflag.StringVarP(&config, "config", "c", "auto", "config to use")
	pflag.BoolVar(&detect, "detect", false, "detect the system only; don't run any benchmarks")
	pflag.StringVar(&outputURL, "output-url", "", "write submission URL to file")
	pflag.BoolVar(&downloadOnly, "download-only", false, "only download the files; don't run any benchmarks")
	pflag.Parse()

	if downloadOnly {
		downloadMain(config)
	} else {
		benchmarkMain(detect, config, outputURL)
	}
}

func getConfig(config string) *Config {
	if config == "auto" {
		config = autoselectConfig()
		log.Printf("auto selected config %q, if this is not what you want, change it with the -c flag", config)
	}

	cfg, ok := configs[config]
	if !ok {
		log.Fatalf("unknown config: %q", config)
	}

	return cfg
}

func downloadMain(config string) {
	cfg := getConfig(config)
	DownloadOnly(cfg)
}

func benchmarkMain(detect bool, config string, outputURL string) {
	var (
		track string
		dt    time.Duration
		err   error
	)

	if detect {
		dt = 99 * time.Second
		config = "detect"
		track = "detect"
	} else {
		if quick {
			track = "quick"
		} else {
			track = "standard"
		}

		cfg := getConfig(config)
		dt, err = Build(cfg)

		if err != nil {
			log.Println("benchmark failed")
			os.Exit(1)
		}
	}

	fmt.Println()
	fmt.Println("build completed in", dt)
	fmt.Println("builds per hour:", float64(time.Hour)/float64(dt))
	fmt.Println()

	r := &Result{
		Track:  track,
		Config: config,
		Time:   float64(dt) / float64(time.Second),
	}
	populateSystem(r)

	fmt.Println("Visit the following link to submit the results:")
	fmt.Println(submissionURL(r))

	if outputURL != "" {
		err := os.WriteFile(outputURL, []byte(submissionURL(r)), 0644)
		if err != nil {
			log.Println("failed to write output url:", err)
			os.Exit(1)
		}
	}
}
