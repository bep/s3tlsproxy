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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	// TODO(bep)
	gitHubStatusApi = "https://api.github.com/repos/%s/%s/statuses/%s"
)

type commitStatus struct {
	// success, error, or failure.
	State string `json:"state"`

	// The target URL to associate with this status.
	//This URL will be linked from the GitHub UI to allow users to easily see the 'source' of the Status.
	TargetURL string `json:"target_url"`

	// Must be less than 1024 bytes.
	Description string `json:"description"`

	// A string label to differentiate this status from the status of other systems.
	//  Default: "default"
	Context string `json:"context"`
}

type githubProject struct {
	userName string
	repoName string
}

func postCommitStatus(project githubProject,
	sha, url string,
	succcess bool) error {

	gitHubToken := os.Getenv("GITHUB_TOKEN")
	if gitHubToken == "" {
		return errors.New("Missing GITHUB_TOKEN in env")
	}

	state := "success"
	if succcess {
		state = "failure"
	}
	cs := commitStatus{State: state, TargetURL: url, Description: "Your deployment was successful!", Context: "ci/bep-deploy"}
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(&cs); err != nil {
		return err
	}

	api := fmt.Sprintf(gitHubStatusApi, project.userName, project.repoName, sha)

	req, err := http.NewRequest("POST", api, &b)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "token "+gitHubToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if isError(resp) {
		b, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("GitHub callback failed: %s", string(b))
	}

	return nil

}

func isError(resp *http.Response) bool {
	return resp.StatusCode < 200 || resp.StatusCode > 299
}
