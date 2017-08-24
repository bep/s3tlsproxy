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

func (c *cacheHandler) shrinkTo(bytes int64) error {
	return nil
}
