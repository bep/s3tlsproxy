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
	"io"
	"net/http"
	"time"

	"github.com/asdine/storm"
)

type FileHandler interface {
	OpenFile(filename string) (io.ReadCloser, error)
	DeleteFile(filename string) error
}

// A Header represents the key-value pairs in an HTTP header.
type Header map[string][]string

type File struct {
	// Path relative to cache dir: <host>/<bucket>/<bucketPath>/<filename>
	Filename string `storm:"id`

	// Size in bytes.
	Size int64

	// We store a sub-set of the headers received from S3 and replay when
	// reading from cache.
	Header Header

	CreatedAt time.Time `storm:"index"`
}

type DB struct {
	filename string
}

func (DB) open() (*storm.DB, error) {
	return nil, nil
}

// Plans:
//
// Request file:
// If db.File => if os.File => OK
// If db.File => if !os.File => delete db.File => stream and save db.File and os.File
// If !db.File  => stream and save db.File and os.File
func handleHTTPRequest(w http.ResponseWriter, r *http.Request) (status int, err error) {
	// If cached, serve file:
	// http.ServeContent
	return 0, nil
}

// On server start:
// For each os.File:
// If not in db.File => delete os.File
// For each db.File:
// If not in os.File => delete db.File
// Else: Set some in memory size counter.

// On free space:
// Order by CreatedAt asc

// On Garbage Collect:
//

func (d DB) deleteFile(filename string) error {
	db, err := d.open()
	if err != nil {
		return err
	}

	f := File{Filename: filename}
	return db.DeleteStruct(&f)
}
