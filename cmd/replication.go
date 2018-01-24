// Copyright © 2017-2018 Giuseppe Maxia
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

	"dbdeployer/sandbox"
	"github.com/spf13/cobra"
)

func ReplicationSandbox(cmd *cobra.Command, args []string) {
	var sd sandbox.SandboxDef
	flags := cmd.Flags()
	sd.Port = sandbox.VersionToPort(args[0])
	sd.Version = args[0]
	sd.Basedir, _ = flags.GetString("sandbox-binary")
	sd.SandboxDir, _ = flags.GetString("sandbox-home")
	sd.DbUser, _ = flags.GetString("db-user")
	sd.DbPassword, _ = flags.GetString("db-password")
	sd.RplUser, _ = flags.GetString("rpl-user")
	sd.RplPassword, _ = flags.GetString("rpl-password")
	sd.RemoteAccess, _ = flags.GetString("remote-access")
	sd.BindAddress, _ = flags.GetString("bind-address")
	sd.InitOptions, _ = flags.GetString("init-options")
	sd.MyCnfOptions, _ = flags.GetString("my-cnf-options")
	nodes, _ := flags.GetInt("nodes")
	topology, _ := flags.GetString("topology")

	var gtid bool
	gtid, _ = flags.GetBool("gtid")
	if gtid {
		sd.GtidOptions = sandbox.GtidOptions
	}
	sd.ReplOptions = sandbox.ReplOptions
	sandbox.CreateReplicationSandbox(sd, args[0], topology, nodes)
}

// replicationCmd represents the replication command
var replicationCmd = &cobra.Command{
	Use:   "replication",
	Args:  cobra.MinimumNArgs(1),
	Short: "create replication sandbox",
	Long:  ``,
	Run:   ReplicationSandbox,
}

func init() {
	rootCmd.AddCommand(replicationCmd)
	replicationCmd.PersistentFlags().String("topology", "master-slave", "Which topology will be installed")
	replicationCmd.PersistentFlags().Int("nodes", 3, "How many nodes will be installed")
	//replicationCmd.PersistentFlags().Int("slaves",  2, "How many slaves will be installed")
}
