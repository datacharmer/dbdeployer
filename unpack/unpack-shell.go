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
	"github.com/datacharmer/dbdeployer/defaults"
	"io/ioutil"
	"os"
	"path"
)

func MergeShell(tarball, extension, basedir, destination, bareName string, verbosity int) error {
	// fmt.Printf("<%s> <%s> <%s> <%s> %d\n",tarball, basedir, destination, bareName, verbosity)
	if !common.DirExists(basedir) {
		return fmt.Errorf(defaults.ErrNamedDirectoryNotFound, "unpack directory", destination)
	}
	if !common.DirExists(destination) {
		return fmt.Errorf(defaults.ErrNamedDirectoryNotFound, "target server directory", destination)
	}
	extracted := path.Join(basedir, bareName)
	if common.DirExists(extracted) {
		return fmt.Errorf(defaults.ErrNamedDirectoryAlreadyExists, "unpacked shell directory", extracted)
	}

	var dirs = []string{"bin", "lib", "share"}
	for _, dir := range dirs {
		destPath := path.Join(destination, dir)
		if !common.DirExists(destPath) {
			common.Exitf(1, "destination server directory %s does not exist in %s\n", dir, destination)
		}
		destPath = path.Join(destination, dir, "mysqlsh")
		if dir != "bin" && common.DirExists(destPath) {
			return fmt.Errorf("destination shell directory %s/mysqlsh already exists in %s\n", dir, destination)
		}
	}

	var err error
	switch extension {
	case defaults.TarGzExt:
		err = UnpackTar(tarball, basedir, verbosity)
	case defaults.TarXzExt:
		err = UnpackXzTar(tarball, basedir, verbosity)
	default:
		return fmt.Errorf("unrecognized extension %s\n", extension)
	}
	if err != nil {
		return err
	}

	defer os.RemoveAll(extracted)
	common.AddToCleanupStack(common.RmdirAll, "RmdirAll", extracted)
	for _, dir := range dirs {
		fullPath := path.Join(extracted, dir)
		if !common.DirExists(fullPath) {
			return fmt.Errorf("source shell directory %s does not exist in %s\n", dir, extracted)
		}
	}
	bin := path.Join(extracted, "bin")
	files, err := ioutil.ReadDir(bin)
	if err != nil {
		return err
	}
	dirs = []string{"lib", "share"}
	for _, dir := range dirs {
		sourceDir := path.Join(extracted, dir, "mysqlsh")
		destDir := path.Join(destination, dir, "mysqlsh")
		if !common.DirExists(sourceDir) {
			return fmt.Errorf("source shell directory %s/mysqlsh does not exist in %s\n", dir, extracted)
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
