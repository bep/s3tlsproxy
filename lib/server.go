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
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
)

// Server represents the caching HTTP server.
type Server struct {
	cfg    Config
	logger *log.Logger

	server *http.Server
}

func NewServer(cfg Config, logger *log.Logger) (*Server, error) {
	// TODO(bep) validate config

	h := http.NewServeMux()

	h.Handle("/", handlers.LoggingHandler(os.Stdout, handler(cfg, logger)))

	s := &http.Server{Addr: cfg.ServerAddr, Handler: h}

	return &Server{cfg: cfg, logger: logger, server: s}, nil
}

func (s *Server) Serve() error {
	s.logger.Printf("Listening on %s ...\n", s.cfg.ServerAddr)
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func handler(cfg Config, logger *log.Logger) http.HandlerFunc {
	c := newCacheHandler(cfg, logger)
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO containsDotDot https://github.com/golang/go/blob/f9cf8e5ab11c7ea3f1b9fde302c0a325df020b1a/src/net/http/fs.go#L665
		err := c.handleRequest(w, r)
		if err != nil {
			logger.Println("error:", err)
			// TODO(bep) status code/err handling
		}
	}
}
