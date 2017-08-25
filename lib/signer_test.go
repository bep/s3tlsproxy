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
	"net/url"
	"testing"

	"time"

	"github.com/stretchr/testify/require"
)

func TestSignURL(t *testing.T) {
	assert := require.New(t)

	someRandomDate := time.Date(1972, time.November, 2, 28, 0, 0, 0, time.UTC)
	backThen := func() time.Time { return someRandomDate }
	sig := Sig{secret: "topsecret", now: backThen}

	signed, err := sig.SignURL("https://example.org:1234/path/?a=b", "GET", 10*time.Hour)
	assert.NoError(err)
	verified, err := sig.VerifyURL(signed, "GET")
	assert.NoError(err)
	assert.True(verified)

	// Invalid secret
	sig2 := Sig{secret: "qwerty", now: backThen}
	verified, err = sig2.VerifyURL(signed, "GET")
	assert.NoError(err)
	assert.False(verified)

	// Method mismatch
	verified, _ = sig.VerifyURL(signed, "DELETE")
	assert.False(verified)

	// Invalid signature
	parsedSigned, _ := url.Parse(signed)
	q := parsedSigned.Query()
	q.Set("sig", "invalid")
	parsedSigned.RawQuery = q.Encode()
	verified, err = sig.VerifyURL(parsedSigned.String(), "GET")
	assert.NoError(err)
	assert.False(verified)

	// Missing sig/expires
	_, err = sig.VerifyURL("https://example.com?expires=12345", "GET")
	assert.Error(err)
	_, err = sig.VerifyURL("https://example.com?sig=abc", "GET")
	assert.Error(err)

	// Expired
	sig.now = func() time.Time { return someRandomDate.Add(11 * time.Hour) }
	verified, _ = sig.VerifyURL(signed, "GET")
	assert.False(verified)

	// General error cases
	_, err = sig.SignURL("", "GET", 1*time.Hour)
	assert.Error(err)
	_, err = sig.SignURL("https://example.org", "GET", -1*time.Hour)
	assert.Error(err)
	_, err = sig.SignURL("https://example.org", "", 1*time.Hour)
	assert.Error(err)
	_, err = sig.VerifyURL("https://example.com?sig=abc&expires=abc", "GET")
	assert.Error(err)

}
