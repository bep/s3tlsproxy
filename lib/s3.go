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
	"strings"
	"time"

	"github.com/kr/s3"
)

// We store and replay all S3 headers not in this list.
var s3HeadersBlacklist = map[string]bool{
	"Server":         true,
	"Content-Length": true,

	"Date":             true,
	"Accept-Ranges":    true,
	"X-Amz-Request-Id": true,
	"X-Amz-Id-2":       true,
}

type s3Client struct {
	cfg Config
}

func (s s3Client) Do(w http.ResponseWriter, r *http.Request) error {
	path := strings.TrimLeft(r.URL.Path, " /")
	if strings.HasSuffix(path, "/") {
		// TODO(bep)
		path = "index.html"
	}

	host, found := s.cfg.host(r.Host)
	if !found {
		return fmt.Errorf("host %s not found", r.Host)
	}

	// TODO(bep) https
	url := fmt.Sprintf("http://%s.s3.amazonaws.com/%s", host.Bucket, host.bucketPath(path))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	// We will store the Content-Encoding header and replay that later.
	req.Header.Add("Accept-Encoding", "gzip")
	req.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))

	s3.Sign(req, s3.Keys{
		AccessKey: host.AccessKey,
		SecretKey: host.SecretKey,
	})

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		fmt.Println("Not Found:", path)
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed for path %s: %d", path, resp.StatusCode)
	}

	for k, v := range resp.Header {
		if s3HeadersBlacklist[k] {
			continue
		}
		w.Header().Add(k, v[0])
	}

	_, err = io.Copy(w, resp.Body)

	return err
}
