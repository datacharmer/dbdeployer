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

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/sandbox"
	"github.com/spf13/cobra"
)

func ReplicationSandbox(cmd *cobra.Command, args []string) {
	var sd sandbox.SandboxDef
	var semisync bool
	common.CheckOrigin(args)
	sd = FillSdef(cmd, args)
	sd.ReplOptions = sandbox.SingleTemplates["replication_options"].Contents
	flags := cmd.Flags()
	semisync, _ = flags.GetBool("semi-sync")
	nodes, _ := flags.GetInt("nodes")
	topology, _ := flags.GetString("topology")
	master_ip, _ := flags.GetString("master-ip")
	master_list, _ := flags.GetString("master-list")
	slave_list, _ := flags.GetString("slave-list")
	sd.SinglePrimary, _ = flags.GetBool("single-primary")
	if topology != "fan-in" && topology != "all-masters" {
		master_list = ""
		slave_list = ""
	}
	if semisync {
		if topology != "master-slave" {
			fmt.Println("--semi-sync is only available with master/slave topology")
			os.Exit(1)
		}
		if common.GreaterOrEqualVersion(sd.Version, []int{5, 5, 1}) {
			sd.SemiSyncOptions = sandbox.SingleTemplates["semisync_master_options"].Contents
		} else {
			fmt.Println("--semi-sync requires version 5.5.1+")
			os.Exit(1)
		}
	}
	if sd.SinglePrimary && topology != "group" {
		fmt.Println("Option 'single-primary' can only be used with 'group' topology ")
		os.Exit(1)
	}
	sandbox.CreateReplicationSandbox(sd, args[0], topology, nodes, master_ip, master_list, slave_list)
}

// replicationCmd represents the replication command
var replicationCmd = &cobra.Command{
	Use: "replication MySQL-Version",
	//Args:  cobra.ExactArgs(1),
	Short: "create replication sandbox",
	Long: `The replication command allows you to deploy several nodes in replication.
Allowed topologies are "master-slave" for all versions, and  "group", "all-masters", "fan-in"
for  5.7.17+.
For this command to work, there must be a directory $HOME/opt/mysql/5.7.21, containing
the binary files from mysql-5.7.21-$YOUR_OS-x86_64.tar.gz
Use the "unpack" command to get the tarball into the right directory.
`,
	//Allowed topologies are "master-slave", "group" (requires 5.7.17+),
	//"fan-in" and "all-msters" (require 5.7.9+)
	Run: ReplicationSandbox,
	Example: `
		$ dbdeployer deploy replication 5.7.21
		# (implies topology = master-slave)

		$ dbdeployer deploy --topology=master-slave replication 5.7.21
		# (explicitly setting topology)

		$ dbdeployer deploy --topology=group replication 5.7.21
		$ dbdeployer deploy --topology=group replication 8.0.4 --single-primary
		$ dbdeployer deploy --topology=all-masters replication 5.7.21
		$ dbdeployer deploy --topology=fan-in replication 5.7.21
	`,
}

func init() {
	// rootCmd.AddCommand(replicationCmd)
	deployCmd.AddCommand(replicationCmd)
	//replicationCmd.PersistentFlags().StringSliceP("master-options", "", "", "Extra options for the master")
	//replicationCmd.PersistentFlags().StringSliceP("slave-options", "", "", "Extra options for the slaves")
	//replicationCmd.PersistentFlags().StringSliceP("node-options", "", "", "Extra options for all nodes")
	//replicationCmd.PersistentFlags().StringSliceP("one-node-options", "", "", "Extra options for one node (format #:option)")
	replicationCmd.PersistentFlags().StringP("master-list", "", "1,2", "Which nodes are masters in a multi-source deployment")
	replicationCmd.PersistentFlags().StringP("slave-list", "", "3", "Which nodes are slaves in a multi-source deployment")
	replicationCmd.PersistentFlags().StringP("master-ip", "", "127.0.0.1", "Which IP the slaves will connect to")
	replicationCmd.PersistentFlags().StringP("topology", "t", "master-slave", "Which topology will be installed")
	replicationCmd.PersistentFlags().IntP("nodes", "n", 3, "How many nodes will be installed")
	replicationCmd.PersistentFlags().BoolP("single-primary", "", false, "Using single primary for group replication")
	replicationCmd.PersistentFlags().BoolP("semi-sync", "", false, "Use semi-synchronous plugin")
}
