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

package sig

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	urls "net/url"
	"strconv"
	"strings"
	"time"
)

type Sig struct {
	secret string
	now    func() time.Time
}

func New(secret string) Sig {
	return Sig{secret: secret, now: time.Now}
}

func (s Sig) SignURL(url, httpMethod string, ttl time.Duration, excludedQueryParams ...string) (string, error) {
	if url == "" || httpMethod == "" || ttl <= 0 {
		return "", errors.New("invalid argument(s)")
	}

	parsedURL, err := urls.Parse(url)
	if err != nil {
		return "", err
	}

	query := parsedURL.Query()

	query.Set("secret", s.secret)
	query.Set("method", httpMethod)
	query.Set("expires", strconv.FormatInt(s.now().Add(ttl).Unix(), 10))

	var exlcudedKeyVals = make(map[string]string)

	if len(excludedQueryParams) > 0 {
		query.Set("exclude", strings.Join(excludedQueryParams, ","))
		for _, excluded := range excludedQueryParams {
			// Not included in signing, will re-add them to the URL later.
			excludedVal := query.Get(excluded)
			if excludedVal != "" {
				exlcudedKeyVals[excluded] = excludedVal
				query.Del(excluded)
			}
		}
	}

	parsedURL.RawQuery = query.Encode()

	sig := s.sum(parsedURL.String())

	query = parsedURL.Query()
	query.Del("method")
	query.Del("secret")

	for k, v := range exlcudedKeyVals {
		query.Set(k, v)
	}
	query.Set("sig", sig)

	parsedURL.RawQuery = query.Encode()

	return parsedURL.String(), nil

}

func (s Sig) VerifyURL(url, httpMethod string) (bool, error) {
	parsedURL, err := urls.Parse(url)
	if err != nil {
		return false, err
	}

	query := parsedURL.Query()
	sig := query.Get("sig")
	expiresStr := query.Get("expires")

	if sig == "" || expiresStr == "" {
		return false, errors.New("invalid URL")
	}

	exludedParamsStr := query.Get("exclude")

	if exludedParamsStr != "" {
		excludedQueryParams := strings.Split(exludedParamsStr, ",")
		for _, key := range excludedQueryParams {
			query.Del(key)
		}

	}

	expires, err := strconv.ParseInt(expiresStr, 10, 64)
	if err != nil {
		return false, errors.New("invalid URL")
	}

	if expires < s.now().Unix() {
		return false, nil
	}

	query.Del("sig")
	query.Set("secret", s.secret)
	query.Set("method", httpMethod)

	parsedURL.RawQuery = query.Encode()

	valid := s.sum(parsedURL.String()) == sig

	return valid, nil
}

func (s Sig) sum(url string) string {
	checksum := sha1.Sum([]byte(url))
	return base64.URLEncoding.EncodeToString(checksum[:])
}
