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
	"net/http"

	"github.com/bep/s3tlsproxy/lib/sig"
	"github.com/unrolled/secure"
)

func (m *httpHandlers) validateSig(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO(bep) check if the logic below holds water (re. proxies etc.)
		url := r.URL.String()

		scheme := r.URL.Scheme
		if scheme == "" {
			scheme = "https://"
		}

		fullURL := scheme + r.Host + url

		sig := sig.New(m.c.cfg.SecretKey)

		verified, err := sig.VerifyURL(fullURL, r.Method)
		m.c.logger.Debug("area", "sig", "url", fullURL, "verified", verified, "err", err)
		// TODO(bep) status codes handling for the below.
		if !verified {
			return
		}

		if err != nil {
			m.c.logger.Error("area", "sig", "error", err)
			return
		}

		h.ServeHTTP(w, r)
	})
}

func (m *httpHandlers) secure(h http.Handler) http.Handler {
	// TODO(bep) => config
	return secure.New(secure.Options{
		AllowedHosts:         m.c.cfg.hostNames(),
		HostsProxyHeaders:    []string{"X-Forwarded-Host"},
		SSLRedirect:          true,
		SSLHost:              "",
		SSLProxyHeaders:      map[string]string{"X-Forwarded-Proto": "https"},
		STSSeconds:           315360000,
		STSIncludeSubdomains: false,
		STSPreload:           true,
		FrameDeny:            true,
		ContentTypeNosniff:   true,
		BrowserXssFilter:     true,
		//ContentSecurityPolicy: "default-src 'self'",
		//PublicKey: `pin-sha256="base64+primary=="; pin-sha256="base64+backup=="; max-age=5184000;"`,

		IsDevelopment: false,
	}).Handler(h)

}
