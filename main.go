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

package main

import (
	"log"
	"os"

	"github.com/bep/s3tlsproxy/cmd"
)

var (
	// Vars set by Gorelaser
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	logger := log.New(os.Stderr, "", log.LstdFlags)

	c := cmd.New(logger)
	if err := c.Execute(); err != nil {
		logger.Fatal("error: ", err)
	}
}
