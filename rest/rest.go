// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2020 Giuseppe Maxia
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

package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/pkg/errors"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/globals"
)

type ReleaseUser struct {
	Login string `json:"login"`
	Type  string `json:"type"`
	URL   string `json:"url"`
}
type DbdeployerRelease struct {
	Assets []struct {
		BrowserDownloadURL string      `json:"browser_download_url"`
		ContentType        string      `json:"content_type"`
		CreatedAt          string      `json:"created_at"`
		DownloadCount      int64       `json:"download_count"`
		Name               string      `json:"name"`
		Size               int64       `json:"size"`
		State              string      `json:"state"`
		Uploader           ReleaseUser `json:"uploader"`
		URL                string      `json:"url"`
	} `json:"assets"`
	Author          ReleaseUser `json:"author"`
	Body            string      `json:"body"`
	CreatedAt       string      `json:"created_at"`
	Draft           bool        `json:"draft"`
	HTMLURL         string      `json:"html_url"`
	Name            string      `json:"name"`
	Prerelease      bool        `json:"prerelease"`
	PublishedAt     string      `json:"published_at"`
	TagName         string      `json:"tag_name"`
	TarballURL      string      `json:"tarball_url"`
	TargetCommitish string      `json:"target_commitish"`
	ZipballURL      string      `json:"zipball_url"`
}

type RemoteFilesMap = map[string][]string

// var RemoteRepo string = "https://github.com/datacharmer/mysql-docker-minimal/blob/master/dbdata"
// var RemoteRepoRaw string = "https://raw.githubusercontent.com/datacharmer/mysql-docker-minimal/master/dbdata"
var FileUrlTemplate string = "{{.RemoteRepo}}/{{.FileName}}"
var IndexUrlTemplate string = "{{.RemoteRepo}}/{{.FileName}}"

func IndexUrl() string {
	var data = common.StringMap{
		"RemoteRepo": defaults.Defaults().RemoteRepository,
		"FileName":   defaults.Defaults().RemoteIndexFile,
	}
	iUrl, err := common.SafeTemplateFill("indexUrl func", IndexUrlTemplate, data)
	if err != nil {
		common.Exitf(1, "error creating Index Url: %s", err)
	}
	return iUrl
}

func FileUrl(fileName string) string {
	var data = common.StringMap{
		"RemoteRepo": defaults.Defaults().RemoteRepository,
		"FileName":   fileName,
	}
	fUrl, err := common.SafeTemplateFill("fileUrl func", FileUrlTemplate, data)
	if err != nil {
		common.Exitf(1, "error creating file URL: %s", err)
	}
	return fUrl
}

type PassThru struct {
	io.Reader
	total           int64 // Total # of bytes transferred
	maxBytesPerDot  int64 // How many bytes before we show a dot
	maxBytesPerMark int64 // How many bytes before we show the actual amount
	stepProgress    int64 // Bytes to calculate towards showing the dot
	markProgress    int64 // Bytes to calculate towards showing the amount
	showProgress    bool  // Shall we show the progress at all
}

// Read 'overrides' the underlying io.Reader's Read method.
// This is the one that will be called by io.Copy(). We simply
// use it to keep track of byte counts and then forward the call.
func (pt *PassThru) Read(p []byte) (int, error) {
	n, err := pt.Reader.Read(p)
	pt.total += int64(n)
	pt.stepProgress += int64(n)
	pt.markProgress += int64(n)

	if err == nil {
		if pt.showProgress && pt.stepProgress >= pt.maxBytesPerDot {
			if pt.markProgress >= pt.maxBytesPerMark {
				fmt.Print(humanize.Bytes(uint64(pt.total)))
				pt.markProgress -= pt.maxBytesPerMark
			} else {
				fmt.Print(".")
			}
			pt.stepProgress -= pt.maxBytesPerDot
		}
	}
	if pt.showProgress && err == io.EOF {
		fmt.Println(" ", humanize.Bytes(uint64(pt.total)))
	}
	return n, err
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath string, url string, progress bool, progressStep int64) error {
	return DownloadFileWithRetry(filepath, url, progress, progressStep, 0)
}

// DownloadFileWithRetry will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFileWithRetry(filepath string, url string, progress bool, progressStep, retriesOnFailure int64) error {

	// Get the data
	var resp *http.Response
	var err error
	var attempts int64 = 1
	if retriesOnFailure > 10 {
		retriesOnFailure = 10
	}
	resp, err = http.Get(url) // #nosec G107
	for err != nil && attempts < retriesOnFailure {
		time.Sleep(time.Second)
		resp, err = http.Get(url) // #nosec G107
		attempts++
	}
	if err != nil {
		return fmt.Errorf("[DownloadFileWithRetry] error getting %s (attempts: %d): %s", url, attempts, err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("[DownloadFileWithRetry] error closing response body: %s", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("[DownloadFileWithRetry] received code %d ", resp.StatusCode)
	}

	// Create the file
	out, err := os.Create(filepath) // #nosec G304
	if err != nil {
		return fmt.Errorf("[DownloadFileWithRetry] error creating file %s", filepath)
	}
	defer out.Close() // #nosec G307

	progressMark := progressStep * 10
	if progressStep <= 0 {
		progress = false
	}
	newBody := &PassThru{
		Reader:          resp.Body,
		maxBytesPerDot:  progressStep,
		maxBytesPerMark: progressMark,
		showProgress:    progress,
	}

	// Write the body to file
	_, err = io.Copy(out, newBody)
	if err != nil {
		return fmt.Errorf("[DownloadFileWithRetry] error during data writing to file %s: %s", filepath, err)
	}

	return nil
}

func GetRemoteIndex() (index RemoteFilesMap, err error) {

	localFileName := "/tmp/available.json"

	err = DownloadFile(localFileName, IndexUrl(), false, 0)
	if err != nil {
		return index, errors.Wrapf(err, "error retrieving downloads index")
	}

	var availableText []byte
	availableText, err = common.SlurpAsBytes(localFileName)
	if err != nil {
		return
	}
	err = json.Unmarshal(availableText, &index)
	return

}

func getReleaseText(tag string) ([]byte, error) {
	endUrl := "/"

	if tag == "" {
		endUrl = ""
	}
	if tag != "" && tag != "latest" {
		tag = "tags/" + tag
	}
	releaseUrl := fmt.Sprintf("https://api.github.com/repos/datacharmer/dbdeployer/releases%s%s", endUrl, tag)
	if os.Getenv("SBDEBUG") != "" {
		fmt.Printf("%s\n", releaseUrl)
	}
	// #nosec G107
	resp, err := http.Get(releaseUrl)
	if err != nil {
		return globals.EmptyBytes, fmt.Errorf("[getReleaseText] error getting %s: %s", releaseUrl, err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("[getReleaseText] error closing response body: %s", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return globals.EmptyBytes, fmt.Errorf("[getReleaseText] received code %d ", resp.StatusCode)
	}

	htmlData, err := io.ReadAll(resp.Body)
	if err != nil {
		return globals.EmptyBytes, err
	}
	return htmlData, nil
}

func GetLatestRelease(tag string) (DbdeployerRelease, error) {
	var release DbdeployerRelease
	if tag == "" {
		tag = "latest"
	}
	htmlData, err := getReleaseText(tag)
	if err != nil {
		return DbdeployerRelease{}, err
	}
	err = json.Unmarshal(htmlData, &release)
	if err != nil {
		return DbdeployerRelease{}, err
	}
	return release, nil
}

func GetReleases() ([]DbdeployerRelease, error) {
	var releases []DbdeployerRelease
	htmlData, err := getReleaseText("")
	if err != nil {
		return []DbdeployerRelease{}, err
	}
	err = json.Unmarshal(htmlData, &releases)
	if err != nil {
		return []DbdeployerRelease{}, err
	}
	return releases, nil
}
