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

func (s s3Client) getAndWrite(path string, host Host,
	w io.Writer, rw http.ResponseWriter, req *http.Request) (*fileMeta, error) {

	// TODO(bep) https
	url := fmt.Sprintf("http://%s.s3.amazonaws.com/%s", host.Bucket, host.bucketPath(path))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	defer resp.Body.Close()

	if !s.cacheableStatusCode(resp.StatusCode) {
		return nil, fmt.Errorf("Failed for path %s: %d", path, resp.StatusCode)
	}

	statusOK := resp.StatusCode == http.StatusOK

	h := make(header)

	keyContentType := "Content-Type"
	textPlain := "text/plain"

	for k, v := range resp.Header {
		if s3HeadersBlacklist[k] {
			continue
		}
		if !statusOK && strings.EqualFold(k, keyContentType) {
			// Prevent XML errors from S3
			h[k] = append(h[k], textPlain)
			continue
		}
		for _, vv := range v {
			h[k] = append(h[k], vv)
		}
	}

	now := time.Now()

	fm := &fileMeta{
		Filename:   host.hostPath(path),
		Size:       resp.ContentLength,
		ModTime:    now, // TODO(bep)
		StatusCode: resp.StatusCode,
		Header:     h,
		CreatedAt:  now,
	}

	for k, v := range fm.Header {
		for _, vv := range v {
			rw.Header().Add(k, vv)
		}
	}

	rw.WriteHeader(resp.StatusCode)

	fmt.Println("Stream from S3", path, "Status:", resp.StatusCode)

	var content io.Reader

	// TODO(bep) consider the error page cache
	if statusOK {
		content = resp.Body
	} else {
		content = strings.NewReader(resp.Status)
	}

	_, err = io.Copy(w, content)
	if err != nil {
		return nil, err
	}

	return fm, nil
}

func (s s3Client) cacheableStatusCode(status int) bool {
	return status == http.StatusOK || status == http.StatusNotFound
}
