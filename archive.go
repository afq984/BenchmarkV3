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
	"runtime"
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

// topRelPath returns name with its first path component (the archive's single
// top-level directory) removed, e.g. "LLVM-22.1.8-Linux-X64/bin/clang" ->
// "bin/clang". The top-level entry itself maps to "".
func topRelPath(name string) string {
	if i := strings.IndexByte(name, '/'); i >= 0 {
		return name[i+1:]
	}
	return ""
}

// pendingLink is a symlink deferred during extraction so it can be materialized
// after all real files exist (see materializeLinks).
type pendingLink struct {
	path   string
	target string
}

func makeFileHandler(destination string, keep func(relPath string) bool, links *[]pendingLink) archiver.FileHandler {
	return func(ctx context.Context, f archiver.File) error {
		if keep != nil && !keep(topRelPath(f.NameInArchive)) {
			return nil
		}

		path := filepath.Join(destination, f.NameInArchive)

		err := os.MkdirAll(filepath.Dir(path), 0755)
		if err != nil {
			return err
		}

		switch {
		case f.FileInfo.IsDir():
			// MkdirAll (not Mkdir): the directory may already exist because a
			// kept file created it, and entry order is not guaranteed.
			return os.MkdirAll(path, f.Mode())
		case f.FileInfo.Mode().IsRegular():
			return writeFile(ctx, path, f)
		case f.FileInfo.Mode()&fs.ModeSymlink != 0:
			if runtime.GOOS == "windows" {
				// Windows symlinks need a privilege the process may not have;
				// defer them and materialize as hardlinks/copies once every real
				// file has been extracted.
				*links = append(*links, pendingLink{path: path, target: f.LinkTarget})
				return nil
			}
			return writeSymlink(ctx, path, f)
		default:
			return fmt.Errorf("cannot handle file mode: %v", f.FileInfo.Mode())
		}
	}
}

// materializeLinks recreates deferred symlinks (Windows only) by dereferencing
// them: a symlink to a file becomes a hardlink (falling back to a copy), and a
// symlink to a directory becomes a recursive copy. Chains are followed to the
// real target. The sysroot's symlinks are all relative, so they resolve within
// the extracted tree.
func materializeLinks(links []pendingLink) error {
	targetOf := make(map[string]string, len(links))
	for _, l := range links {
		t := filepath.FromSlash(l.target)
		if !filepath.IsAbs(t) {
			t = filepath.Join(filepath.Dir(l.path), t)
		}
		targetOf[filepath.Clean(l.path)] = filepath.Clean(t)
	}
	for _, l := range links {
		cur := targetOf[filepath.Clean(l.path)]
		for i := 0; i < 40; i++ {
			next, ok := targetOf[cur]
			if !ok {
				break
			}
			cur = next
		}
		info, err := os.Stat(cur)
		if err != nil {
			continue // dangling target; the build does not use it
		}
		os.RemoveAll(l.path)
		if info.IsDir() {
			if err := copyTree(cur, l.path); err != nil {
				return err
			}
		} else if os.Link(cur, l.path) != nil {
			if err := copyFile(cur, l.path, info.Mode()); err != nil {
				return err
			}
		}
	}
	return nil
}

func copyFile(src, dst string, mode fs.FileMode) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, mode)
}

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if os.Link(p, target) != nil {
			return copyFile(p, target, info.Mode())
		}
		return nil
	})
}

func unarchive(ctx context.Context, source, destination string, keep func(relPath string) bool) error {
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

	var links []pendingLink
	if err := u.Extract(ctx, r, nil, makeFileHandler(destination, keep, &links)); err != nil {
		return err
	}
	return materializeLinks(links)
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
	// Keep, if non-nil, restricts extraction to entries whose path relative to
	// the archive's top-level directory it accepts (parent directories of kept
	// files are created automatically). Used to skip the many gigabytes of the
	// LLVM toolchain and source tree that building llc never touches.
	Keep func(relPath string) bool
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
	err = unarchive(ctx, a.savePath(), extractTo, a.Keep)
	if err != nil {
		log.Println("extract failed:", err)
		return err
	}

	return nil
}

func (a *Archive) SetUp(ctx context.Context, buildDir string) error {
	return a.DownloadAndExtract(ctx, buildDir)
}
