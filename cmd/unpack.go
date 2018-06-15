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

package cmd

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/unpack"
	"github.com/spf13/cobra"
)

func UnpackTarball(cmd *cobra.Command, args []string) {
	flags := cmd.Flags()
	Basedir := GetAbsolutePathFromFlag(cmd, "sandbox-binary")
	verbosity, _ := flags.GetInt("verbosity")
	if !common.DirExists(Basedir) {
		common.Exit(1,
			fmt.Sprintf("Directory %s does not exist.", Basedir),
			"You should create it or provide an alternate base directory using --sandbox-binary")
	}
	tarball := args[0]
	re_version := regexp.MustCompile(`(\d+\.\d+\.\d+)`)
	verList := re_version.FindAllStringSubmatch(tarball, -1)
	detected_version := verList[0][0]
	// fmt.Printf(">> %#v %s\n",verList, detected_version)
	
	Version, _ := flags.GetString("unpack-version")
	if Version == "" {
		Version = detected_version
	}
	if Version == "" {
		common.Exit(1, 
			"unpack: No version was detected from tarball name. ",
			"Flag --unpack-version becomes mandatory")
	}
	// This call used to ensure that the port provided is in the right format
	common.VersionToPort(Version)
	Prefix, _ := flags.GetString("prefix")

	destination := Basedir + "/" + Prefix + Version
	if common.DirExists(destination) {
		common.Exit(1, fmt.Sprintf("Destination directory %s exists already\n", destination))
	}
	var extension string = ".tar.gz"
	extracted := path.Base(tarball)
	var barename string
	if strings.HasSuffix(tarball, extension) {
		barename = extracted[0 : len(extracted)-len(extension)]
	} else {
		common.Exit(1, "Tarball extension must be .tar.gz")
	}

	fmt.Printf("Unpacking tarball %s to %s\n", tarball, common.ReplaceLiteralHome(destination))
	//verbosity_level := unpack.VERBOSE
	err := unpack.UnpackTar(tarball, Basedir, verbosity)
	if err != nil {
		common.Exit(1, fmt.Sprintf("%s", err))
	}
	final_name := Basedir + "/" + barename
	if final_name != destination {
        fmt.Printf("Renaming directory %s to %s\n",final_name, destination)
		err = os.Rename(final_name, destination)
		if err != nil {
			common.Exit(1, fmt.Sprintf("%s", err))
		}
	}
}

// unpackCmd represents the unpack command
var unpackCmd = &cobra.Command{
	Use:   "unpack MySQL-tarball",
	Args:  cobra.ExactArgs(1),
	Aliases: []string{"extract", "untar", "unzip", "inflate", "expand"},
	Short: "unpack a tarball into the binary directory",
	Long: `If you want to create a sandbox from a tarball, you first need to unpack it
into the sandbox-binary directory. This command carries out that task, so that afterwards 
you can call 'deploy single', 'deploy multiple', and 'deploy replication' commands with only 
the MySQL version for that tarball.
If the version is not contained in the tarball name, it should be supplied using --unpack-version.
If there is already an expanded tarball with the same version, a new one can be differentiated with --prefix.
`,
	Run: UnpackTarball,
	Example: `
    $ dbdeployer unpack mysql-8.0.4-rc-linux-glibc2.12-x86_64.tar.gz
    Unpacking tarball mysql-8.0.4-rc-linux-glibc2.12-x86_64.tar.gz to $HOME/opt/mysql/8.0.4

    $ dbdeployer unpack --prefix=ps Percona-Server-5.7.21-linux.tar.gz
    Unpacking tarball Percona-Server-5.7.21-linux.tar.gz to $HOME/opt/mysql/ps5.7.21

    $ dbdeployer unpack --unpack-version=8.0.18 --prefix=bld mysql-mybuild.tar.gz
    Unpacking tarball mysql-mybuild.tar.gz to $HOME/opt/mysql/bld8.0.18
	`,
}

func init() {
	rootCmd.AddCommand(unpackCmd)

	unpackCmd.PersistentFlags().Int("verbosity", 1, "Level of verbosity during unpack (0=none, 2=maximum)")
	unpackCmd.PersistentFlags().String("unpack-version", "", "which version is contained in the tarball")
	unpackCmd.PersistentFlags().String("prefix", "", "Prefix for the final expanded directory")
}
