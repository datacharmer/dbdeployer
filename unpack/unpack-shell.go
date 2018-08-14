// Copyright © 2011-12 Qtrac Ltd.
//
// This program or package and any associated files are licensed under the
// Apache License, Version 2.0 (the "License"); you may not use these files
// except in compliance with the License. You can get a copy of the License
// at: http://www.apache.org/licenses/LICENSE-2.0.
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

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
		common.Exit(1, fmt.Sprintf("Unpack directory %s does not exist\n", destination))
	}
	if !common.DirExists(destination) {
		common.Exit(1, fmt.Sprintf("Target server directory %s does not exist\n", destination))
	}
	extracted := basedir + "/" + barename
	if common.DirExists(extracted) {
		common.Exit(1, fmt.Sprintf("Unpacked shell directory %s already exists", extracted))
	}

	var dirs = []string{"bin", "lib", "share"}
	for _, dir := range dirs {
		dest_path := destination + "/" + dir
		if !common.DirExists(dest_path) {
			common.Exit(1, fmt.Sprintf("Destination server directory %s does not exist in %s\n", dir, destination))
		}
		dest_path = destination + "/" + dir + "/mysqlsh"
		if dir != "bin" && common.DirExists(dest_path) {
			common.Exit(1, fmt.Sprintf("Destination shell directory %s/mysqlsh already exists in %s\n", dir, destination))
		}
	}

	err := UnpackTar(tarball, basedir, verbosity)
	if err != nil {
		return err
	}

	defer os.RemoveAll(extracted)
	common.AddToCleanupStack(common.RmdirAll, "RmdirAll", extracted)
	for _, dir := range dirs {
		full_path := extracted + "/" + dir
		if !common.DirExists(full_path) {
			common.Exit(1, fmt.Sprintf("Source shell directory %s does not exist in %s\n", dir, extracted))
		}
	}
	bin := extracted + "/bin"
	files, err := ioutil.ReadDir(bin)
	if err != nil {
		return err
	}
	dirs = []string{"lib", "share"}
	for _, dir := range dirs {
		source_dir := extracted + "/" + dir + "/mysqlsh"
		dest_dir := destination + "/" + dir + "/mysqlsh"
		if !common.DirExists(source_dir) {
			common.Exit(1, fmt.Sprintf("Source shell directory %s/mysqlsh does not exist in %s\n", dir, extracted))
		}
		if verbosity >= VERBOSE {
			fmt.Printf("Move %s %s\n", source_dir, dest_dir)
		}
		err = os.Rename(source_dir, dest_dir)
		if err != nil {
			return err
		}
	}

	for _, f := range files {
		source_file := fmt.Sprintf("%s/%s", bin, f.Name())
		dest_file := fmt.Sprintf("%s/bin/%s", destination, f.Name())
		if verbosity >= VERBOSE {
			fmt.Printf("Copy %s %s \n", source_file, dest_file)
		}
		common.CopyFile(source_file, dest_file)
	}
	return nil
}
