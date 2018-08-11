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
	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/sandbox"
	"github.com/spf13/cobra"
	//"fmt"
)

func ReplicationSandbox(cmd *cobra.Command, args []string) {
	var sd sandbox.SandboxDef
	var semisync bool
	common.CheckOrigin(args)
	sd = FillSdef(cmd, args)
	sd.ReplOptions = sandbox.SingleTemplates["replication_options"].Contents
	flags := cmd.Flags()
	semisync, _ = flags.GetBool(defaults.SemiSyncLabel)
	nodes, _ := flags.GetInt(defaults.NodesLabel)
	topology, _ := flags.GetString(defaults.TopologyLabel)
	master_ip, _ := flags.GetString(defaults.MasterIpLabel)
	master_list, _ := flags.GetString(defaults.MasterListLabel)
	slave_list, _ := flags.GetString(defaults.SlaveListLabel)
	sd.SinglePrimary, _ = flags.GetBool(defaults.SinglePrimaryLabel)
	repl_history_dir, _ := flags.GetBool(defaults.ReplHistoryDirLabel)
	if repl_history_dir {
		sd.HistoryDir = "REPL_DIR"
	}
	if topology != defaults.FanInLabel && topology != defaults.AllMastersLabel {
		master_list = ""
		slave_list = ""
	}
	if semisync {
		if topology != defaults.MasterSlaveLabel {
			common.Exit(1, "--semi-sync is only available with master/slave topology")
		}
		if common.GreaterOrEqualVersion(sd.Version, []int{5, 5, 1}) {
			sd.SemiSyncOptions = sandbox.SingleTemplates["semisync_master_options"].Contents
		} else {
			common.Exit(1, "--semi-sync requires version 5.5.1+")
		}
	}
	if sd.SinglePrimary && topology != defaults.GroupLabel {
		common.Exit(1, "Option 'single-primary' can only be used with 'group' topology ")
	}
	origin := args[0]
	if args[0] != sd.BasedirName {
		origin = sd.BasedirName
	}
	//fmt.Printf("%#v\n",sd)
	sandbox.CreateReplicationSandbox(sd, origin, topology, nodes, master_ip, master_list, slave_list)
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
		$ dbdeployer deploy replication 5.7    # deploys highest revision for 5.7
		$ dbdeployer deploy replication 5.7.21 # deploys a specific revision
		$ dbdeployer deploy replication /path/to/5.7.21 # deploys a specific revision in a given path
		# (implies topology = master-slave)

		$ dbdeployer deploy --topology=master-slave replication 5.7
		# (explicitly setting topology)

		$ dbdeployer deploy --topology=group replication 5.7
		$ dbdeployer deploy --topology=group replication 8.0 --single-primary
		$ dbdeployer deploy --topology=all-masters replication 5.7
		$ dbdeployer deploy --topology=fan-in replication 5.7
	`,
}

func init() {
	// rootCmd.AddCommand(replicationCmd)
	deployCmd.AddCommand(replicationCmd)
	//replicationCmd.PersistentFlags().StringSliceP("master-options", "", "", "Extra options for the master")
	//replicationCmd.PersistentFlags().StringSliceP("slave-options", "", "", "Extra options for the slaves")
	//replicationCmd.PersistentFlags().StringSliceP("node-options", "", "", "Extra options for all nodes")
	//replicationCmd.PersistentFlags().StringSliceP("one-node-options", "", "", "Extra options for one node (format #:option)")
	replicationCmd.PersistentFlags().StringP(defaults.MasterListLabel, "", defaults.MasterListValue, "Which nodes are masters in a multi-source deployment")
	replicationCmd.PersistentFlags().StringP(defaults.SlaveListLabel, "", defaults.SlaveListValue, "Which nodes are slaves in a multi-source deployment")
	replicationCmd.PersistentFlags().StringP(defaults.MasterIpLabel, "", defaults.MasterIpValue, "Which IP the slaves will connect to")
	replicationCmd.PersistentFlags().StringP(defaults.TopologyLabel, "t", defaults.TopologyValue, "Which topology will be installed")
	replicationCmd.PersistentFlags().IntP(defaults.NodesLabel, "n", defaults.NodesValue, "How many nodes will be installed")
	replicationCmd.PersistentFlags().BoolP(defaults.SinglePrimaryLabel, "", false, "Using single primary for group replication")
	replicationCmd.PersistentFlags().BoolP(defaults.SemiSyncLabel, "", false, "Use semi-synchronous plugin")
	replicationCmd.PersistentFlags().Bool(defaults.ReplHistoryDirLabel, false, "uses the replication directory to store mysql client history")
}
