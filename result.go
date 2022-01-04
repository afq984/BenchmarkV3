package main

import (
	"fmt"
	"os"
	"time"

	"github.com/afq984/BenchmarkV3/systemdetect"
)

type Result struct {
	Time     time.Duration
	Track    string
	Config   string
	Hostname string
	CPU      string
	Memory   int64
	Misc     string
}

const unknown = "<unknown>"

func populateSystem(r *Result) {
	var err error
	r.Hostname, err = os.Hostname()
	if err != nil {
		r.Hostname = unknown
	}

	r.CPU, err = systemdetect.Cpu()
	if err != nil {
		r.CPU = unknown
	}

	r.Memory, err = systemdetect.Memory()
	if err != nil {
		r.Memory = -1
	}

	r.Misc, err = systemdetect.Misc()
	if err != nil {
		r.Misc = unknown
	}
}

func (r *Result) PrettyPrint() {
	fmt.Printf("%f\t%s\t%s\t%s\t%s\t%d\t%s\n",
		float64(r.Time)/float64(time.Second),
		r.Track, r.Config, r.Hostname, r.CPU, r.Memory, r.Misc,
	)
}
