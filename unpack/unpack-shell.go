// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2018 Giuseppe Maxia
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
	"fmt"
	"github.com/datacharmer/dbdeployer/common"
	"io/ioutil"
	"os"
)

func MergeShell(tarball, basedir, destination, barename string, verbosity int) error {
	// fmt.Printf("<%s> <%s> <%s> <%s> %d\n",tarball, basedir, destination, barename, verbosity)
	if !common.DirExists(basedir) {
		common.Exitf(1, "Unpack directory %s does not exist\n", destination)
	}
	if !common.DirExists(destination) {
		common.Exitf(1, "Target server directory %s does not exist\n", destination)
	}
	extracted := basedir + "/" + barename
	if common.DirExists(extracted) {
		common.Exitf(1, "Unpacked shell directory %s already exists", extracted)
	}

	var dirs = []string{"bin", "lib", "share"}
	for _, dir := range dirs {
		destPath := destination + "/" + dir
		if !common.DirExists(destPath) {
			common.Exitf(1, "Destination server directory %s does not exist in %s\n", dir, destination)
		}
		destPath = destination + "/" + dir + "/mysqlsh"
		if dir != "bin" && common.DirExists(destPath) {
			common.Exitf(1, "Destination shell directory %s/mysqlsh already exists in %s\n", dir, destination)
		}
	}

	err := UnpackTar(tarball, basedir, verbosity)
	if err != nil {
		return err
	}

	defer os.RemoveAll(extracted)
	common.AddToCleanupStack(common.RmdirAll, "RmdirAll", extracted)
	for _, dir := range dirs {
		fullPath := extracted + "/" + dir
		if !common.DirExists(fullPath) {
			common.Exitf(1, "Source shell directory %s does not exist in %s\n", dir, extracted)
		}
	}
	bin := extracted + "/bin"
	files, err := ioutil.ReadDir(bin)
	if err != nil {
		return err
	}
	dirs = []string{"lib", "share"}
	for _, dir := range dirs {
		sourceDir := extracted + "/" + dir + "/mysqlsh"
		destDir := destination + "/" + dir + "/mysqlsh"
		if !common.DirExists(sourceDir) {
			common.Exitf(1, "Source shell directory %s/mysqlsh does not exist in %s\n", dir, extracted)
		}
		if verbosity >= VERBOSE {
			fmt.Printf("Move %s %s\n", sourceDir, destDir)
		}
		err = os.Rename(sourceDir, destDir)
		if err != nil {
			return err
		}
	}

	for _, f := range files {
		sourceFile := fmt.Sprintf("%s/%s", bin, f.Name())
		destFile := fmt.Sprintf("%s/bin/%s", destination, f.Name())
		if verbosity >= VERBOSE {
			fmt.Printf("Copy %s %s \n", sourceFile, destFile)
		}
		common.CopyFile(sourceFile, destFile)
	}
	return nil
}
