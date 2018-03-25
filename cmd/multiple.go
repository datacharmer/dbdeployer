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
	//"fmt"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/sandbox"
	"github.com/spf13/cobra"
)

func MultipleSandbox(cmd *cobra.Command, args []string) {
	var sd sandbox.SandboxDef
	common.CheckOrigin(args)
	flags := cmd.Flags()
	sd = FillSdef(cmd, args)
	nodes, _ := flags.GetInt("nodes")
	sd.SBType = "multiple"
	sandbox.CreateMultipleSandbox(sd, args[0], nodes)
}

var multipleCmd = &cobra.Command{
	Use:   "multiple MySQL-Version",
	Args:  cobra.ExactArgs(1),
	Short: "create multiple sandbox",
	Long: `Creates several sandboxes of the same version,
without any replication relationship.
For this command to work, there must be a directory $HOME/opt/mysql/5.7.21, containing
the binary files from mysql-5.7.21-$YOUR_OS-x86_64.tar.gz
Use the "unpack" command to get the tarball into the right directory.
`,
	Run: MultipleSandbox,
	Example: `
	$ dbdeployer deploy multiple 5.7.21
	`,
}

func init() {
	deployCmd.AddCommand(multipleCmd)
	multipleCmd.PersistentFlags().IntP("nodes", "n", 3, "How many nodes will be installed")
}
