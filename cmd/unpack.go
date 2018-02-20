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
	"strings"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/unpack"
	"github.com/spf13/cobra"
)

func UnpackTarball(cmd *cobra.Command, args []string) {
	flags := cmd.Flags()
	Basedir, _ := flags.GetString("sandbox-binary")
	if !common.DirExists(Basedir) {
		fmt.Printf("Directory %s does not exist.\n", Basedir)
		fmt.Println("You should create it or provide an alternate base directory using --sandbox-binary")
		os.Exit(1)
	}

	Version, _ := flags.GetString("unpack-version")
	if Version == "" {
		fmt.Println("unpack: flag --unpack-version is mandatory")
		os.Exit(1)
	}
	// This call used to ensure that the port provided is in the right format
	common.VersionToPort(Version)
	Prefix, _ := flags.GetString("prefix")
	tarball := args[0]

	destination := Basedir + "/" + Prefix + Version
	if common.DirExists(destination) {
		fmt.Printf("Destination directory %s exists already\n", destination)
		os.Exit(1)
	}
	var extension string = ".tar.gz"
	extracted := path.Base(tarball)
	var barename string
	if strings.HasSuffix(tarball, extension) {
		barename = extracted[0 : len(extracted)-len(extension)]
	} else {
		fmt.Println("Tarball extension must be .tar.gz")
		os.Exit(1)
	}

	fmt.Printf("Unpacking tarball %s to %s\n", tarball, destination)
	verbosity_level := unpack.VERBOSE
	err := unpack.UnpackTar(tarball, Basedir, verbosity_level)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	final_name := Basedir + "/" + barename
	if final_name != destination {
		err = os.Rename(final_name, destination)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
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
you can call 'single', 'multiple', and 'replication' commands with only the MySQL version
for that tarball.`,
	Run: UnpackTarball,
	Example: `
    $ dbdeployer --unpack-version=8.0.4 unpack mysql-8.0.4-rc-linux-glibc2.12-x86_64.tar.gz
    Unpacking tarball mysql-8.0.4-rc-linux-glibc2.12-x86_64.tar.gz to $HOME/opt/mysql/8.0.4
    .........100.........200.........292
	`,
}

func init() {
	rootCmd.AddCommand(unpackCmd)

	unpackCmd.PersistentFlags().String("unpack-version", "", "which version is contained in the tarball (mandatory)")
	unpackCmd.PersistentFlags().String("prefix", "", "Prefix for the final expanded directory")
}
