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
	"github.com/datacharmer/dbdeployer/ops"
	"github.com/spf13/cobra"

	"github.com/datacharmer/dbdeployer/globals"
)

func initEnvironment(cmd *cobra.Command, args []string) error {

	flags := cmd.Flags()
	sandboxBinary, _ := flags.GetString(globals.SandboxBinaryLabel)
	sandboxHome, _ := flags.GetString(globals.SandboxHomeLabel)
	dryRun, _ := flags.GetBool(globals.DryRunLabel)
	skipDownloads, _ := flags.GetBool(globals.SkipAllDownloadsLabel)
	skipTarballDownload, _ := flags.GetBool(globals.SkipTarballDownloadLabel)
	skipCompletion, _ := flags.GetBool(globals.SkipShellCompletionLabel)

	return ops.InitEnvironment(ops.InitOptions{
		SandboxBinary:       sandboxBinary,
		SandboxHome:         sandboxHome,
		DryRun:              dryRun,
		SkipDownloads:       skipDownloads,
		SkipTarballDownload: skipTarballDownload,
		SkipCompletion:      skipCompletion,
	})
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "initializes dbdeployer environment",
	Long: `Initializes dbdeployer environment: 
* creates $SANDBOX_HOME and $SANDBOX_BINARY directories
* downloads and expands the latest MySQL tarball
* installs shell completion file`,
	RunE: initEnvironment,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.PersistentFlags().Bool(globals.SkipAllDownloadsLabel, false, "Do not download any file (skip both MySQL tarball and shell completion file)")
	initCmd.PersistentFlags().Bool(globals.SkipTarballDownloadLabel, false, "Do not download MySQL tarball")
	initCmd.PersistentFlags().Bool(globals.SkipShellCompletionLabel, false, "Do not download shell completion file")
	initCmd.PersistentFlags().Bool(globals.DryRunLabel, false, "Show operations but don't run them")
}
