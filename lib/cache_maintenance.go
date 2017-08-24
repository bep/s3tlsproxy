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
	"os"
	"path/filepath"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
)

func (c *cacheHandler) purgePrefix(prefix string) error {
	db, err := c.openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	// TODO(bep) remove
	var all []fileMeta
	db.All(&all)
	for _, a := range all {
		c.logger.Debug("file", a.Filename)
	}

	// TODO(bep) a way to do this with an index.
	q := db.Select(q.Re("Filename", "^"+prefix))

	var files []fileMeta

	if err := q.Find(&files); err != nil && err != storm.ErrNotFound {
		return err
	}

	c.logger.Info("area", "cache", "tag", "purge", "prefix", prefix, "count", len(files))

	for _, file := range files {
		if err := db.DeleteStruct(&file); err != nil && err != storm.ErrNotFound {
			return err
		}
	}

	return nil
}

// This is a Least Recently Used (LRU) cache.
// So: Sort by created and delete until we reach the target.
func (c *cacheHandler) shrinkTo(target int64) error {
	db, err := c.openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	var files []fileMeta

	err = db.AllByIndex("CreatedAt", &files)
	if err != nil {
		if err == storm.ErrNotFound {
			return nil
		}
		return err
	}

	var (
		totalSize int64 = 0
		fileCount       = 0
	)

	for _, f := range files {
		totalSize += f.Size
		fileCount++
	}

	c.logger.Debug("area", "cache", "tag", "shrink",
		"target", target,
		"total", totalSize, "files", fileCount)

	if totalSize <= target {
		return nil
	}

	toDelete := totalSize - target

	for i, file := range files {
		if err := db.DeleteStruct(&file); err != nil && err != storm.ErrNotFound {
			return err
		}

		osFilename := filepath.Join(c.cfg.CacheDir, filepath.FromSlash(file.Filename))

		if err := os.Remove(osFilename); err != nil && !os.IsNotExist(err) {
			return err
		}

		toDelete -= file.Size

		if toDelete <= 0 {
			c.logger.Debug("area", "cache", "tag", "shrink",
				"deleted", (i + 1))
			break
		}
	}

	return nil
}

// TODO(bep) Handle leftover directories
