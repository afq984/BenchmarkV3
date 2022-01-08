package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver/v4"
)

func writeFile(ctx context.Context, path string, f archiver.File) error {
	r, err := f.Open()
	if err != nil {
		return err
	}
	defer r.Close()

	w, err := os.Create(path)
	if err != nil {
		return err
	}
	defer w.Close()

	err = w.Chmod(f.Mode())
	if err != nil {
		return err
	}

	_, err = io.Copy(w, readerContext(ctx, r))
	return err
}

func writeSymlink(ctx context.Context, path string, f archiver.File) error {
	if f.LinkTarget == "" {
		panic("empty LinkTarget")
	}

	_, err := os.Lstat(path)
	if err == nil {
		err = os.Remove(path)
		if err != nil {
			return err
		}
	}

	return os.Symlink(f.LinkTarget, path)
}

func makeFileHandler(destination string) archiver.FileHandler {
	return func(ctx context.Context, f archiver.File) error {
		path := filepath.Join(destination, f.NameInArchive)

		err := os.MkdirAll(filepath.Dir(path), 0755)
		if err != nil {
			return err
		}

		switch {
		case f.FileInfo.IsDir():
			return os.Mkdir(path, f.Mode())
		case f.FileInfo.Mode().IsRegular():
			return writeFile(ctx, path, f)
		case f.FileInfo.Mode()&fs.ModeSymlink != 0:
			return writeSymlink(ctx, path, f)
		default:
			return fmt.Errorf("cannot handle file mode: %v", f.FileInfo.Mode())
		}
	}
}

func unarchive(ctx context.Context, source, destination string) error {
	var u archiver.Extractor
	var d archiver.Decompressor
	switch {
	case strings.HasSuffix(source, ".zip"):
		u = archiver.Zip{}
	case strings.HasSuffix(source, ".tar.gz"):
		u = archiver.Tar{}
		d = archiver.Gz{}
	case strings.HasSuffix(source, ".tar.xz"):
		u = archiver.Tar{}
		d = archiver.Xz{}
	default:
		return fmt.Errorf("unknown file extension: %s", source)
	}

	var r io.ReadCloser
	f, err := os.Open(source)
	if err != nil {
		return err
	}
	defer f.Close()

	if d == nil {
		r = f
	} else {
		r, err = d.OpenReader(f)
		if err != nil {
			return err
		}
	}

	return u.Extract(ctx, r, nil, makeFileHandler(destination))
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
func (a *Archive) check(ctx context.Context) error {
	f, err := os.Open(a.savePath())
	if err != nil {
		return err
	}

	log.Println("checking sha256:", a.savePath())
	h := sha256.New()
	_, err = io.Copy(h, readerContext(ctx, f))
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
func (a *Archive) downloadWithoutChecks(ctx context.Context) error {
	err := os.MkdirAll(downloadDir, 0755)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.URL, nil)
	if err != nil {
		panic(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad response status %d, %q", resp.StatusCode, a.URL)
	}

	f, err := os.Create(a.savePath())
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func (a *Archive) downloadWithChecks(ctx context.Context) error {
	err := a.check(ctx)
	if err == nil {
		return nil
	}

	if os.IsNotExist(err) {
		log.Printf("downloading %s", a.savePath())

	} else {
		log.Printf("redownloading %s (%s)", a.savePath(), err)
	}

	err = a.downloadWithoutChecks(ctx)
	if err != nil {
		log.Println("download failed:", err)
		return err
	}

	err = a.check(ctx)
	if err != nil {
		log.Println("download failed:", err)
		return err
	}

	return nil
}

func (a *Archive) DownloadAndExtract(ctx context.Context, buildDir string) error {
	err := a.downloadWithChecks(ctx)
	if err != nil {
		return err
	}

	log.Println("extracting:", a.savePath())
	extractTo := buildDir
	if a.ExtractTo != "" {
		extractTo = filepath.Join(extractTo, a.ExtractTo)
	}
	err = unarchive(ctx, a.savePath(), extractTo)
	if err != nil {
		log.Println("extract failed:", err)
		return err
	}

	return nil
}

func (a *Archive) SetUp(ctx context.Context, buildDir string) error {
	return a.DownloadAndExtract(ctx, buildDir)
}
