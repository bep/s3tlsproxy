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
	"net/http"

	"crypto/tls"

	"golang.org/x/crypto/acme/autocert"
)

// Server represents the caching HTTP server.
type Server struct {
	cfg        Config
	tlsEnabled bool

	logger *Logger

	server *http.Server
}

type httpHandlers struct {
	cfg    Config
	logger *Logger
}

func NewServer(cfg Config, logger *Logger) (*Server, error) {
	// TODO(bep) validate config

	h := http.NewServeMux()
	mw := &httpHandlers{cfg: cfg, logger: logger}

	h.Handle("/", mw.secure(mw.serveFile()))

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
	c := newCacheHandler(m.cfg, m.logger)
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO containsDotDot https://github.com/golang/go/blob/f9cf8e5ab11c7ea3f1b9fde302c0a325df020b1a/src/net/http/fs.go#L665
		err := c.handleRequest(w, r)
		if err != nil {
			m.logger.Error("handleRequest", err)
			// TODO(bep) status code/err handling
		}
	}
}
