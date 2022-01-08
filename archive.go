package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver/v3"
)

// a version of archiver.Unarchive that allows overwritting existing files
func unarchive(source, destination string) error {
	var u archiver.Unarchiver
	switch {
	case strings.HasSuffix(source, ".zip"):
		u = archiver.NewZip()
		u.(*archiver.Zip).OverwriteExisting = true
	case strings.HasSuffix(source, ".tar.gz"):
		u = archiver.NewTarGz()
		u.(*archiver.TarGz).OverwriteExisting = true
	case strings.HasSuffix(source, ".tar.xz"):
		u = archiver.NewTarXz()
		u.(*archiver.TarXz).OverwriteExisting = true
	default:
		return fmt.Errorf("unknown file extension: %s", source)
	}
	return u.Unarchive(source, destination)
}

type mismatchedSha256 struct {
	want string
	got  string
}

var _ error = &mismatchedSha256{}

func (err *mismatchedSha256) Error() string {
	return fmt.Sprintf("sha256sum mismatch: want %s; got %s", err.want, err.got)
}

type Archive struct {
	URL       string
	Sha256    string
	ExtractTo string
}

var _ Package = &Archive{}

// returns the download name of the archive
func (a *Archive) savePath() string {
	u, err := url.Parse(a.URL)
	if err != nil {
		panic(err)
	}
	return filepath.Join(downloadDir, path.Base(u.Path))
}

// check that the downloaded archive matches the specified checksum
func (a *Archive) check() error {
	f, err := os.Open(a.savePath())
	if err != nil {
		return err
	}

	log.Println("checking sha256:", a.savePath())
	h := sha256.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return err
	}

	s := hex.EncodeToString(h.Sum(nil))
	if a.Sha256 != s {
		return &mismatchedSha256{
			want: a.Sha256,
			got:  s,
		}
	}

	return nil
}

// downloadWithoutChecks the archive without any checks
func (a *Archive) downloadWithoutChecks() error {
	err := os.MkdirAll(downloadDir, 0755)
	if err != nil {
		return err
	}

	r, err := http.Get(a.URL)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	f, err := os.Create(a.savePath())
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, r.Body)
	if err != nil {
		return err
	}

	return nil
}

func (a *Archive) downloadWithChecks() error {
	err := a.check()
	if err == nil {
		return nil
	}

	if os.IsNotExist(err) {
		log.Printf("downloading %s", a.savePath())

	} else {
		log.Printf("redownloading %s (%s)", a.savePath(), err)
	}

	err = a.downloadWithoutChecks()
	if err != nil {
		log.Println("download failed:", err)
		return err
	}

	err = a.check()
	if err != nil {
		log.Println("download failed:", err)
		return err
	}

	return nil
}

func (a *Archive) DownloadAndExtract(buildDir string) error {
	err := a.downloadWithChecks()
	if err != nil {
		return err
	}

	log.Println("extracting:", a.savePath())
	extractTo := buildDir
	if a.ExtractTo != "" {
		extractTo = filepath.Join(extractTo, a.ExtractTo)
	}
	err = unarchive(a.savePath(), extractTo)
	if err != nil {
		log.Println("extract failed:", err)
		return err
	}

	return nil
}

func (a *Archive) SetUp(buildDir string) error {
	return a.DownloadAndExtract(buildDir)
}
