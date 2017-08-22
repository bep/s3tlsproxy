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
	"log"

	"github.com/spf13/cobra"
)

// Commandeer holds the available CLI commands and their flag values.
type Commandeer struct {
	logger  *log.Logger
	cfgFile string

	rootCmd *cobra.Command
}

func New(logger *log.Logger) Commandeer {

	c := Commandeer{logger: logger}

	c.rootCmd = &cobra.Command{
		Use:   "s3tlsproxy",
		Short: "A caching proxy for Amazon S3 with automatic TLS.",
		//	Run: func(cmd *cobra.Command, args []string) { },
	}

	c.rootCmd.PersistentFlags().StringVar(&c.cfgFile, "config", "", "config file (default is ./.s3tlsproxy.toml)")

	return c

}

func (c Commandeer) Execute() error {
	return c.rootCmd.Execute()
}
