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
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetConfig(t *testing.T) {
	assert := require.New(t)

	basic := `
cacheDir = "cache"
TLSCertsDir = "certs"
DBFilename = "db/s3p.db"
serverAddr = ":8080"
defaultHostAccessKey = "yourHostSecretAccessKey"
defaultHostSecretKey = "yourHostSecretKey"
secretKey = "yourSecret"

[hosts]
[hosts."example.org"]
bucket = "bucket1"
path = "path1"
accessKey = "ac1"
secretKey = "as1"
[hosts."example.com"]
bucket = "bucket2"
path = "path2"

`
	c, err := readConfig(strings.NewReader(basic))

	assert.NoError(err)
	assert.Equal("cache", c.CacheDir)
	assert.Len(c.Hosts, 2)
	assert.Equal([]string{"example.com", "example.org"}, c.hostNames())

	h := c.Hosts["example.org"]

	assert.Equal("ac1", h.AccessKey)
	assert.Equal("as1", h.SecretKey)

	h = c.Hosts["example.com"]

	assert.Equal("yourHostSecretAccessKey", h.AccessKey)
	assert.Equal("yourHostSecretKey", h.SecretKey)

	// TODO(bep) env overrides

}
