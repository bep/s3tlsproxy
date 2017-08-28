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
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/bep/s3deploy/lib"
)

var (
	version = "DEV"
	commit  = "unknown"
	date    = "unknown"
)

const (
	appName = "s3proxydeployer" // TODO(bep) name
)

func main() {
	var (
		cachePurgeURL string
	)

	cfg, err := lib.FlagsToConfig()
	if err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&cachePurgeURL, "cachePurgeURL", "", "Base URL needed to purge cache (or set S3P_CACHE_PURGE_URL")

	flag.Parse()

	fmt.Printf("%s %s (commit %s, built at %v)\n", appName, version, commit, date)

	if cfg.PrintVersion {
		return
	}

	if cfg.Help {
		flag.Usage()
		return
	}

	if cachePurgeURL == "" {
		cachePurgeURL = os.Getenv("S3P_CACHE_PURGE_URL")
	}

	if cachePurgeURL == "" {
		flag.Usage()
		return
	}

	stats, err := lib.Deploy(cfg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(stats.Summary())

	if err := purgeCache(stats, cachePurgeURL); err != nil {
		log.Fatal(err)
	}

	// Testing
	userName := os.Getenv("CIRCLE_PROJECT_USERNAME")
	repoName := os.Getenv("CIRCLE_PROJECT_REPONAME")
	sha := os.Getenv("CIRCLE_SHA1")
	if userName != "" && repoName != "" && sha != "" {
		preview := "https://cdn1.bep.is/"
		project := githubProject{userName: userName, repoName: repoName}
		if err := postCommitStatus(project, sha, preview, true); err != nil {
			fmt.Println("GitHub preview failed:", err)
		}
	} else {
		fmt.Println("Missing GitHub info:", "=>", userName, "=>", repoName, "=>", sha)
	}

}

func purgeCache(stats lib.DeployStats, cachePurgeURL string) error {
	// TODO(bep) more fine grained purge logic, add prefix=filename
	// TODO(bep) cross domain purge

	if stats.FileCountChanged() == 0 {
		return nil
	}

	resp, err := http.Get(cachePurgeURL)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP-%d", resp.StatusCode)
	}
	return nil

}
