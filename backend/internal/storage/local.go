package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
)

// Local stores objects on disk under a root directory. All file access goes through os.Root,
// so a key can never escape the directory.
type Local struct {
	root    *os.Root
	baseURL string
}

// NewLocal creates dir if needed. baseURL is the public URL prefix that serves the files (e.g. http://localhost:8081/uploads).
func NewLocal(dir, baseURL string) (*Local, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create upload dir: %w", err)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, fmt.Errorf("open upload dir: %w", err)
	}
	return &Local{root: root, baseURL: strings.TrimRight(baseURL, "/")}, nil
}

func (l *Local) Put(_ context.Context, key string, r io.Reader, _ string) error {
	if dir := path.Dir(key); dir != "." {
		if err := l.root.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	// Write to a temp name and rename, so readers never see a partial file.
	tmp := key + ".tmp"
	f, err := l.root.Create(tmp)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, r); err != nil {
		f.Close()
		_ = l.root.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		_ = l.root.Remove(tmp)
		return err
	}
	return l.root.Rename(tmp, key)
}

func (l *Local) Delete(_ context.Context, key string) error {
	err := l.root.Remove(key)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}

func (l *Local) URL(key string) string {
	return l.baseURL + "/" + key
}

// Handler serves stored files (no directory listings). Mount it with http.StripPrefix.
func (l *Local) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		f, err := l.root.Open(name)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer f.Close()
		info, err := f.Stat()
		if err != nil || info.IsDir() || strings.HasSuffix(name, ".tmp") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		http.ServeContent(w, r, name, info.ModTime(), f)
	})
}
