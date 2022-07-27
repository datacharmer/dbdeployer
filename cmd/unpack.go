// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2022 Giuseppe Maxia
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

	"github.com/datacharmer/dbdeployer/ops"
	"github.com/spf13/cobra"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/globals"
)

func unpackTarball(cmd *cobra.Command, args []string) {
	flags := cmd.Flags()
	Basedir, err := getAbsolutePathFromFlag(cmd, "sandbox-binary")
	common.ErrCheckExitf(err, 1, "error getting absolute path for 'sandbox-binary'")
	verbosity, _ := flags.GetInt(globals.VerbosityLabel)
	Prefix, _ := flags.GetString(globals.PrefixLabel)
	isShell, _ := flags.GetBool(globals.ShellLabel)
	target, _ := flags.GetString(globals.TargetServerLabel)
	overwrite, _ := flags.GetBool(globals.OverwriteLabel)
	flavor, _ := flags.GetString(globals.FlavorLabel)
	dryRun, _ := flags.GetBool(globals.DryRunLabel)
	Version, _ := flags.GetString(globals.UnpackVersionLabel)
	if !common.DirExists(Basedir) {
		common.Exit(1,
			fmt.Sprintf(globals.ErrDirectoryNotFound, Basedir),
			"You should create it or provide an alternate base directory using --sandbox-binary")
	}

	err = ops.UnpackTarball(ops.UnpackOptions{
		SandboxBinary: Basedir,
		TarballName:   args[0],
		TargetServer:  target,
		Version:       Version,
		Prefix:        Prefix,
		Flavor:        flavor,
		Verbosity:     verbosity,
		IsShell:       isShell,
		Overwrite:     overwrite,
		DryRun:        dryRun,
	})

	if err != nil {
		common.Exitf(1, "error unpacking tarball %s: %s", args[0], err)
	}
}

// unpackCmd represents the unpack command
var unpackCmd = &cobra.Command{
	Use:     "unpack MySQL-tarball",
	Args:    cobra.ExactArgs(1),
	Aliases: []string{"extract", "untar", "unzip", "inflate", "expand"},
	Short:   "unpack a tarball into the binary directory",
	Long: `If you want to create a sandbox from a tarball (.tar.gz or .tar.xz), you first need to unpack it
into the sandbox-binary directory. This command carries out that task, so that afterwards 
you can call 'deploy single', 'deploy multiple', and 'deploy replication' commands with only 
the MySQL version for that tarball.
If the version is not contained in the tarball name, it should be supplied using --unpack-version.
If there is already an expanded tarball with the same version, a new one can be differentiated with --prefix.
`,
	Run: unpackTarball,
	Example: `
    $ dbdeployer unpack mysql-8.0.4-rc-linux-glibc2.12-x86_64.tar.gz
    Unpacking tarball mysql-8.0.4-rc-linux-glibc2.12-x86_64.tar.gz to $HOME/opt/mysql/8.0.4

    $ dbdeployer unpack --prefix=ps Percona-Server-5.7.21-linux.tar.gz
    Unpacking tarball Percona-Server-5.7.21-linux.tar.gz to $HOME/opt/mysql/ps5.7.21

    $ dbdeployer unpack --unpack-version=8.0.18 --prefix=bld mysql-mybuild.tar.gz
    Unpacking tarball mysql-mybuild.tar.gz to $HOME/opt/mysql/bld8.0.18
	`,
	Annotations: map[string]string{"export": ExportAnnotationToJson(StringExport)},
}

func init() {
	rootCmd.AddCommand(unpackCmd)

	unpackCmd.PersistentFlags().Int(globals.VerbosityLabel, 1, "Level of verbosity during unpack (0=none, 2=maximum)")
	unpackCmd.PersistentFlags().String(globals.UnpackVersionLabel, "", "which version is contained in the tarball")
	unpackCmd.PersistentFlags().String(globals.PrefixLabel, "", "Prefix for the final expanded directory")
	unpackCmd.PersistentFlags().Bool(globals.ShellLabel, false, "Unpack a shell tarball into the corresponding server directory")
	unpackCmd.PersistentFlags().Bool(globals.OverwriteLabel, false, "Overwrite the destination directory if already exists")
	unpackCmd.PersistentFlags().Bool(globals.DryRunLabel, false, "Show unpack operations, but do not run them")
	unpackCmd.PersistentFlags().String(globals.TargetServerLabel, "", "Uses a different server to unpack a shell tarball")
	unpackCmd.PersistentFlags().String(globals.FlavorLabel, "", "Defines the tarball flavor (MySQL, NDB, Percona Server, etc)")
}
