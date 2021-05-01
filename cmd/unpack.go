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

package cmd

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/datacharmer/dbdeployer/unpack"
)

func unpackTarball(cmd *cobra.Command, args []string) {
	flags := cmd.Flags()
	Basedir, err := getAbsolutePathFromFlag(cmd, "sandbox-binary")
	common.ErrCheckExitf(err, 1, "error getting absolute path for 'sandbox-binary'")
	verbosity, _ := flags.GetInt(globals.VerbosityLabel)
	if !common.DirExists(Basedir) {
		common.Exit(1,
			fmt.Sprintf(globals.ErrDirectoryNotFound, Basedir),
			"You should create it or provide an alternate base directory using --sandbox-binary")
	}
	tarball := args[0]
	reVersion := regexp.MustCompile(`(\d+\.\d+\.\d+)`)
	verList := reVersion.FindAllStringSubmatch(tarball, -1)

	detectedVersion := ""
	if verList != nil {
		detectedVersion = verList[0][0]
	}
	// common.CondPrintf(">> %#v %s\n",verList, detected_version)

	isShell, _ := flags.GetBool(globals.ShellLabel)
	target, _ := flags.GetString(globals.TargetServerLabel)
	if !isShell && target != "" {
		common.Exit(1,
			"unpack: Option --target-server can only be used with --shell")
	}

	overwrite, _ := flags.GetBool(globals.OverwriteLabel)
	flavor, _ := flags.GetString(globals.FlavorLabel)
	dryRun, _ := flags.GetBool(globals.DryRunLabel)
	if flavor == "" {
		baseName := common.BaseName(tarball)
		flavor = common.DetectTarballFlavor(baseName)
		if flavor == "" {
			common.Exitf(1, "No flavor detected in %s. Please use --%s", tarball, globals.FlavorLabel)
		}
	}
	Version, _ := flags.GetString(globals.UnpackVersionLabel)
	if Version == "" {
		Version = detectedVersion
	}
	if Version == "" {
		common.Exit(1,
			"unpack: No version was detected from tarball name. ",
			"Flag --unpack-version becomes mandatory")
	}
	// This call used to ensure that the port provided is in the right format
	_, err = common.VersionToPort(Version)
	if err != nil {
		common.Exitf(1, "version %s not in the required format", Version)
	}
	Prefix, _ := flags.GetString(globals.PrefixLabel)
	if isShell {
		fmt.Printf("%s\n", Version)
		var canBeEmbedded bool
		canBeEmbedded, err = common.HasCapability(common.MySQLShellFlavor, common.EmbedMySQLShell, Version)
		if err != nil {
			common.Exitf(1, "error detecting shell capability: %s", err)
		}
		if !canBeEmbedded {
			common.Exitf(1, "MySQL shell version %s insufficient for embedding", Version)
		}
	}

	destination := path.Join(Basedir, Prefix+Version)
	if target != "" {
		destination = path.Join(Basedir, target)
	}
	if common.DirExists(destination) && !isShell {
		if overwrite {
			if dryRun {
				fmt.Printf("delete binaries %s %s\n", Basedir, Prefix+Version)
			} else {
				isDeleted, err := deleteBinaries(Basedir, Prefix+Version, false)
				if !isDeleted {
					common.Exitf(1, "directory %s could not be removed", Prefix+Version)
				}
				if err != nil {
					common.Exitf(1, "error removing directory %s: %s", Prefix+Version, err)
				}
			}
		} else {
			common.Exitf(1, globals.ErrNamedDirectoryAlreadyExists, "destination directory", destination)
		}
	}
	extracted := path.Base(tarball)
	var bareName string

	var extractFunc func(string, string, int) error
	var foundExtension string

	switch {
	case strings.HasSuffix(tarball, globals.TarGzExt):
		extractFunc = unpack.UnpackTar
		foundExtension = globals.TarGzExt
	case strings.HasSuffix(tarball, globals.TarXzExt):
		extractFunc = unpack.UnpackXzTar
		foundExtension = globals.TarXzExt
	default:
		common.Exitf(1, "tarball extension must be either '%s' or '%s'", globals.TarGzExt, globals.TarXzExt)
	}
	err = unpack.VerifyTarFile(tarball)
	if err != nil {
		common.Exitf(1, "validation for %s failed: %s", tarball, err)
	}
	bareName = extracted[0 : len(extracted)-len(globals.TarGzExt)]
	if isShell {
		common.CondPrintf("Merging shell tarball %s to %s\n", common.ReplaceLiteralHome(tarball), common.ReplaceLiteralHome(destination))
		if !dryRun {
			err := unpack.MergeShell(tarball, foundExtension, Basedir, destination, bareName, verbosity)
			common.ErrCheckExitf(err, 1, "error while unpacking mysql shell tarball : %s", err)
		}
		return
	}

	common.CondPrintf("Unpacking tarball %s to %s\n", tarball, common.ReplaceLiteralHome(destination))
	if dryRun {
		return
	}
	//verbosity_level := unpack.VERBOSE
	// err := unpack.UnpackTar(tarball, Basedir, verbosity)
	err = extractFunc(tarball, Basedir, verbosity)
	common.ErrCheckExitf(err, 1, "%s", err)
	finalName := path.Join(Basedir, bareName)
	// If the directory was not created, it probably means that the tarball was not well organised
	// and either lacked the top directory or the top directory had a different name
	if !common.DirExists(finalName) {
		common.Exitf(1, "problem with tarball %s: directory %s was not created", tarball, finalName)
	}
	if finalName != destination {
		common.CondPrintf("Renaming directory %s to %s\n", finalName, destination)
		err = os.Rename(finalName, destination)
		common.ErrCheckExitf(err, 1, "%s", err)
	}
	err = common.WriteString(flavor, path.Join(destination, globals.FlavorFileName))
	common.ErrCheckExitf(err, 1, "error writing %s in %s", globals.FlavorFileName, destination)
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
