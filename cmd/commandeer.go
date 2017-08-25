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
	"os"

	"github.com/bep/s3tlsproxy/lib"
	"github.com/go-kit/kit/log"
	"github.com/spf13/cobra"
)

// Commandeer is the entry point for the different commands.
type Commandeer struct {
	cfgFile string

	cfg    lib.Config
	logger *lib.Logger

	rootCmd *cobra.Command
}

func (c *Commandeer) init() error {
	cfg, err := c.loadConfig()
	if err != nil {
		return err
	}
	c.cfg = cfg
	c.logger = cfg.CreateLogger()

	return nil
}

func New() *Commandeer {
	c := &Commandeer{logger: lib.NewLogger(log.NewLogfmtLogger(os.Stdout))}

	c.rootCmd = &cobra.Command{
		Use:   "s3tlsproxy",
		Short: "A caching proxy for Amazon S3 with automatic TLS.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runServer()
		},
	}

	c.rootCmd.PersistentFlags().StringVar(&c.cfgFile, "config", "", "config file (default is ./config.toml)")
	c.rootCmd.AddCommand(c.newUrls().cmd)

	return c
}

func (c Commandeer) loadConfig() (lib.Config, error) {
	filename := c.cfgFile
	if filename == "" {
		filename = "./config.toml"
	}
	return lib.LoadConfig(filename)
}

func (c *Commandeer) Execute() error {
	if err := c.rootCmd.Execute(); err != nil {
		c.logger.Error(err)
		return err
	}
	return nil
}
