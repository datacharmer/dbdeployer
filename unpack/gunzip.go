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

package unpack

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/globals"
)

func gUnzipData(data []byte) (resData []byte, err error) {
	b := bytes.NewBuffer(data)

	var r io.Reader
	r, err = gzip.NewReader(b)
	if err != nil {
		return
	}

	var resB bytes.Buffer
	_, err = resB.ReadFrom(r)
	if err != nil {
		return
	}

	resData = resB.Bytes()

	return
}

func GunzipFile(inputFileName, outputFileName string, overwrite bool) error {
	if !common.FileExists(inputFileName) {
		return fmt.Errorf(globals.ErrFileNotFound, inputFileName)
	}
	if !overwrite && common.FileExists(outputFileName) {
		return fmt.Errorf(globals.ErrFileAlreadyExists, outputFileName)
	}
	buf, err := common.SlurpAsBytes(inputFileName)
	if err != nil {
		return err
	}
	unpackData, err := gUnzipData(buf)
	if err != nil {
		return err
	}
	return common.WriteString(string(unpackData), outputFileName)
}
