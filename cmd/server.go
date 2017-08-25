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

package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bep/s3tlsproxy/lib"
)

func (c *Commandeer) runServer() error {
	if err := c.init(); err != nil {
		return err
	}

	server, err := lib.NewServer(c.cfg, c.logger)
	if err != nil {
		return err
	}

	go func() {
		if err := server.Serve(); err != nil {
			c.logger.Error("serve", err)
			os.Exit(-1)
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	for {
		s := <-signalChan
		switch s {
		case syscall.SIGHUP:
			// TODO(bep) reload
		default:
			log.Printf("Captured %v. Exiting...", s)
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			server.Shutdown(shutdownCtx)

			<-shutdownCtx.Done()
			err := shutdownCtx.Err()
			if err != nil {
				c.logger.Error(err)
			}
			break
		}
	}

	return nil
}
