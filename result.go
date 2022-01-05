package main

import (
	"net/url"
	"os"
	"strconv"

	"github.com/afq984/BenchmarkV3/systemdetect"
)

type Result struct {
	Time     float64 `json:"time"`
	Track    string  `json:"track"`
	Config   string  `json:"config"`
	Hostname string  `json:"hostname"`
	CPU      string  `json:"cpu"`
	Memory   int64   `json:"memory"`
	Misc     string  `json:"misc"`
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

func submissionURL(r *Result) string {
	// Prefer uppercase in URLs to save QR code space
	u, err := url.Parse("HTTPS://AFQ984.GITHUB.IO/BenchmarkV3/")
	if err != nil {
		panic(err)
	}

	q := u.Query()
	q.Add("T", strconv.FormatFloat(r.Time, 'g', -1, 64))
	q.Add("K", r.Track)
	q.Add("G", r.Config)
	q.Add("H", r.Hostname)
	q.Add("C", r.CPU)
	q.Add("R", strconv.FormatInt(r.Memory, 10))
	q.Add("M", r.Misc)
	u.RawQuery = q.Encode()

	return u.String()
}
