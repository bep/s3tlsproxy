// Copyright © 2017 Bjørn Erik Pedersen <bjorn.erik.pedersen@gmail.com>.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lib

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/boltdb/bolt"

	"github.com/asdine/storm"
)

// A header represents the key-value pairs in a HTTP header.
type header map[string][]string

func (h header) String() string {
	var s string
	for k, v := range h {
		s += fmt.Sprintf("%s: %v ", k, v)
	}

	return s
}

type fileMeta struct {
	// Path relative to cache dir: <host>/<bucket>/<bucketPath>/<filename>
	// TODO(bep) consider bucket per host
	Filename string `storm:"id"`

	// Size in bytes.
	Size int64

	// File mod time
	ModTime time.Time

	// HTTP status code. Note that only HTTP 200's are stored on the file system.
	StatusCode int

	// We store a sub-set of the headers received from S3 and replay when
	// reading from cache.
	Header header

	CreatedAt time.Time `storm:"index"`
}

type readSeekCloser interface {
	io.ReadSeeker
	io.Closer
}

type cache struct {
	cfg     Config
	logger  *Logger
	storage s3Client
}

func newCache(cfg Config, logger *Logger) *cache {
	return &cache{cfg: cfg, logger: logger, storage: s3Client{cfg: cfg, logger: logger}}
}

func (c *cache) handleRequest(rw http.ResponseWriter, req *http.Request) error {
	urlPath := strings.TrimLeft(req.URL.Path, " /")

	if urlPath == "" || strings.HasSuffix(urlPath, "/") {
		// TODO(bep)
		urlPath = path.Join(urlPath, "index.html")
	}

	host, found := c.cfg.host(req.Host)
	if !found {
		return fmt.Errorf("host %s not found", req.Host)
	}

	relPath := host.hostPath(urlPath)

	meta, err := c.getFileMeta(relPath)
	if err != nil {
		return err
	}

	if meta != nil {
		c.logger.Debug("area", "cache", "filename", meta.Filename, "status", meta.StatusCode, "header", meta.Header)

		f, err := c.getFile(relPath)
		if err != nil {
			return err
		}

		if f == nil {
			// File has been deleted by some other process.
			// TODO(bep)
			fmt.Println("TODO")
			return nil
		}

		defer f.Close()
		for k, v := range meta.Header {
			for _, vv := range v {
				rw.Header().Add(k, vv)
			}
		}
		http.ServeContent(rw, req, urlPath, meta.ModTime, f)
		return nil

	}

	meta, err = c.getAndWriteFile(urlPath, host, rw, req)
	if err != nil {
		return err
	}

	err = c.doWithDB(func(db *storm.DB) error {
		return db.Save(meta)
	})

	return err
}

func (c *cache) getFileMeta(relPath string) (*fileMeta, error) {
	db, err := c.openDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var fm fileMeta

	err = db.One("Filename", relPath, &fm)
	if err != nil {
		if err == storm.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &fm, nil
}

func (c *cache) getAndWriteFile(
	urlPath string, host Host,
	rw http.ResponseWriter, req *http.Request) (*fileMeta, error) {

	// TODO(bep) temp file and rename
	filename := filepath.Join(c.cfg.CacheDir, host.hostPath(urlPath))
	dir := filepath.Dir(filename)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	f, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Stream to both file and client at the same time.
	mw := io.MultiWriter(rw, f)

	return c.storage.getAndWrite(urlPath, host, mw, rw, req)

}

func (c *cache) getFile(relPath string) (readSeekCloser, error) {
	filename := filepath.Join(c.cfg.CacheDir, filepath.FromSlash(relPath))
	f, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return f, nil
}

// TODO(bep) transactions
func (c *cache) doWithDB(f func(db *storm.DB) error) error {
	db, err := c.openDB()
	if err != nil {
		return err
	}
	defer db.Close()
	return f(db)
}

func (c *cache) openDB() (*storm.DB, error) {
	return storm.Open(c.cfg.DBFilename, storm.BoltOptions(0600, &bolt.Options{Timeout: 10 * time.Second}))
}
