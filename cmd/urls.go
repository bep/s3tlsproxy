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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bep/s3tlsproxy/lib"
	"github.com/bep/s3tlsproxy/lib/sig"
	"github.com/spf13/cobra"
)

type urls struct {
	cfg    lib.Config
	logger *lib.Logger

	cmd *cobra.Command

	// Flags
	url      string
	method   string
	duration string
	exclude  string
}

func (c *Commandeer) newUrls() urls {
	u := urls{cfg: c.cfg, logger: c.logger}

	cmd := &cobra.Command{
		Use:   "urls",
		Short: "URL related utilities",
	}

	cmdSign := &cobra.Command{
		Use:   "sign",
		Short: "Signs the given URL for the given HTTP method",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.init(); err != nil {
				return err
			}

			if u.url == "" || u.method == "" || u.duration == "" {
				return errors.New("missing flag value")
			}

			duration, err := time.ParseDuration(u.duration)
			if err != nil {
				fmt.Println(5 * time.Hour)
				return fmt.Errorf("invalid value for 'duration': %s", err)
			}

			s := sig.New(c.cfg.SecretKey)

			var excludeParams []string
			if u.exclude != "" {
				excludeParams = strings.Split(u.exclude, ",")
			}

			signedURL, err := s.SignURL(u.url, u.method, duration, excludeParams...)
			if err != nil {
				return err
			}

			fmt.Println(signedURL)

			return nil
		},
	}

	cmdSign.Flags().StringVarP(&u.url, "url", "", "", "the URL to sign")
	cmdSign.Flags().StringVarP(&u.method, "method", "", "", "the HTTP method")
	cmdSign.Flags().StringVarP(&u.duration, "duration", "", "", "time to live")
	cmdSign.Flags().StringVarP(&u.exclude, "exclude", "", "", "optional comma separated list of HTTP paramaters to exclude when signing")

	cmd.AddCommand(cmdSign)
	u.cmd = cmd

	return u
}
