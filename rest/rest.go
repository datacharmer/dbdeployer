// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2019 Giuseppe Maxia
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

	"github.com/dustin/go-humanize"
	"github.com/pkg/errors"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
)

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

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("[DownloadFile] error getting %s: %s", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("[DownloadFile] received code %d ", resp.StatusCode)
	}

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("[DownloadFile] error creating file %s", filepath)
	}
	defer out.Close()

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
		return fmt.Errorf("[DownloadFile] error during data writing to file %s: %s", filepath, err)
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
