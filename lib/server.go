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
	"context"
	"fmt"
	"net/http"
	"path"

	"crypto/tls"

	"golang.org/x/crypto/acme/autocert"
)

const (
	// Make it unlikely that this clashes with some other static resource.
	appNS = "__s3p"
)

// Server represents the caching HTTP server.
type Server struct {
	cfg        Config
	tlsEnabled bool

	logger *Logger

	server *http.Server
}

type httpHandlers struct {
	c *cache
}

func NewServer(cfg Config, logger *Logger) (*Server, error) {
	// TODO(bep) validate config

	var (
		h  = http.NewServeMux()
		c  = newCache(cfg, logger)
		mw = &httpHandlers{c: c}
	)

	var purger http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		prefix := r.FormValue("prefix")
		host, found := cfg.host(r.Host)
		if !found {
			c.logger.Error("area", "cache", "tag", "purge", "host_not_found", r.Host)
			return
		}

		prefix = path.Join(host.Name, prefix)

		if err := c.purgePrefix(prefix); err != nil {
			c.logger.Error("area", "cache", "tag", "purge", "prefix", prefix, "error", err)
		}

	}

	var shrinker http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		target := 50 << 10 // TODO(bep) Take value from conf
		if err := c.shrinkTo(int64(target)); err != nil {
			c.logger.Error("area", "cache", "tag", "shrink", "error", err)
		}
	}

	var (
		// Make the handler chaining a little bit more fluid.
		secure      = mw.secure
		validateSig = mw.validateSig
	)

	h.Handle(fmt.Sprintf("/%s/purge", appNS), secure(validateSig(purger)))
	h.Handle(fmt.Sprintf("/%s/shrink", appNS), secure(validateSig(shrinker)))
	h.Handle("/", secure(mw.serveFile()))

	tlsEnabled, err := cfg.isTLSConfigured()
	if err != nil {
		return nil, err
	}

	var s *http.Server

	if tlsEnabled {
		m := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(cfg.hostNames()...),
			Cache:      autocert.DirCache(cfg.TLSCertsDir),
		}
		s = &http.Server{
			Addr:      cfg.ServerAddr,
			TLSConfig: &tls.Config{GetCertificate: m.GetCertificate},
			Handler:   h,
		}
	} else {
		s = &http.Server{
			Addr:    cfg.ServerAddr,
			Handler: h,
		}
	}

	return &Server{cfg: cfg, logger: logger, server: s, tlsEnabled: tlsEnabled}, nil
}

func (s *Server) Serve() error {
	s.logger.Info("Listener", s.cfg.ServerAddr)
	if s.tlsEnabled {
		return s.server.ListenAndServeTLS("", "")
	}
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (m *httpHandlers) serveFile() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		// TODO containsDotDot https://github.com/golang/go/blob/f9cf8e5ab11c7ea3f1b9fde302c0a325df020b1a/src/net/http/fs.go#L665
		err := m.c.handleRequest(w, r)
		if err != nil {
			m.c.logger.Error("handleRequest", err)
			// TODO(bep) status code/err handling
		}
	}
}
