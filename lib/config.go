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
	"os"
	"path"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {

	// Location of the actual files.
	CacheDir string

	// TLS will be enabled if set.
	TLSCertsDir string

	// Path and name of the metadata database.
	DBFilename string

	ServerAddr string

	Hosts map[string]Host

	DefaultHostAccessKey string
	DefaultHostSecretKey string
	SecretKey            string
}

type Host struct {
	Name   string
	Bucket string
	Path   string

	AccessKey string
	SecretKey string
}

func (h Host) hostPath(in string) string {
	return path.Join(h.Name, h.Bucket, h.Path, in)
}

func (h Host) bucketPath(in string) string {
	return path.Join(h.Path, in)
}

func readConfig(r io.Reader) (Config, error) {
	var c Config
	if _, err := toml.DecodeReader(r, &c); err != nil {
		return c, fmt.Errorf("failed to read config: %s", err)
	}

	for name, host := range c.Hosts {
		host.Name = name
		if host.AccessKey == "" {
			host.AccessKey = c.DefaultHostAccessKey
		}
		if host.SecretKey == "" {
			host.SecretKey = c.DefaultHostSecretKey
		}
		c.Hosts[name] = host
	}

	return c, nil
}

func LoadConfig(filename string) (Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return Config{}, err
	}
	defer f.Close()
	return readConfig(f)
}

func (c Config) hostNames() []string {
	var names []string
	for k, _ := range c.Hosts {
		names = append(names, k)
	}

	sort.Strings(names)

	return names
}

func (c Config) host(hostName string) (Host, bool) {
	hostName = strings.Split(hostName, ":")[0]
	h, found := c.Hosts[hostName]
	return h, found
}

func (c Config) isTLSConfigured() (bool, error) {
	if c.TLSCertsDir == "" {
		return false, nil
	}

	fi, err := os.Stat(c.TLSCertsDir)
	if err != nil || !fi.IsDir() {
		return false, fmt.Errorf("dir %s not valid as certificate dir", c.TLSCertsDir)
	}

	return true, nil

}
