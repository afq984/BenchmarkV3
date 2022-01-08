package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type System struct {
	Name string
}

var _ Package = &System{}

func (s *System) SetUp(buildDir string) error {
	p, err := exec.LookPath(s.Name)
	if err != nil {
		log.Printf("cannot find %q in system path", s.Name)
		return err
	}
	log.Printf("found %q at %q", s.Name, p)

	err = os.Symlink(p, filepath.Join(buildDir, s.Name))
	if err != nil {
		log.Printf("cannot symlink %q", s.Name)
		return err
	}

	return nil
}
