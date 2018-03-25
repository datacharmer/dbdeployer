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

	"github.com/spf13/cobra"
)

func ShowUsage(cmd *cobra.Command, args []string) {
	const basic_usage string = `
	USING A SANDBOX

Change directory to the newly created one (default: $SANDBOX_HOME/msb_VERSION 
for single sandboxes)
[ $SANDBOX_HOME = $HOME/sandboxes unless modified with flag --sandbox-home ]

The sandbox directory of the instance you just created contains some handy 
scripts to manage your server easily and in isolation.

"./start", "./status", "./restart", and "./stop" do what their name suggests. 
start and restart accept parameters that are eventually passed to the server. 
e.g.:

  ./start --server-id=1001

  ./restart --event-scheduler=disabled

"./use" calls the command line client with the appropriate parameters,
Example:

    ./use -BN -e "select @@server_id"
    ./use -u root

"./clear" stops the server and removes everything from the data directory, 
letting you ready to start from scratch. (Warning! It's irreversible!)

"./my" is a prefix script to invoke any command named "my*" from the 
MySQL /bin directory. It is important to use it rather than the 
corresponding globally installed tool, because this guarantees 
that you will be using the tool for the version you have deployed.
Examples:

    ./my sqldump db_name
    ./my sqlbinlog somefile
`
	const multiple_usage string = ` USING MULTIPLE SERVER SANDBOX
On a replication sandbox, you have the same commands (run "dbdeployer usage single"), 
with an "_all" suffix, meaning that you propagate the command to all the members. 
Then you have "./m" as a shortcut to use the master, "./s1" and "./s2" to access 
the slaves (and "s3", "s4" ... if you define more).

In group sandboxes without a master slave relationship (group replication and 
multiple sandboxes) the nodes can be accessed by ./n1, ./n2, ./n3, and so on.

start_all
status_all
restart_all
stop_all
use_all
clear_all
m
s1, s2, n1, n2

The scripts "check_slaves" or "check_nodes" give the status of replication in the sandbox.
`
	request := ""
	if len(args) > 0 {
		request = args[0]
	}
	switch request {
	case "single":
		fmt.Println(basic_usage)
	case "multiple":
		fmt.Println(multiple_usage)
	default:
		fmt.Println(basic_usage)
		fmt.Println(multiple_usage)
	}
}

var usageCmd = &cobra.Command{
	Use:   "usage [single|multiple]",
	Short: "Shows usage of installed sandboxes",
	Long:  `Shows syntax and examples of tools installed in database sandboxes.`,
	Run:   ShowUsage,
}

func init() {
	rootCmd.AddCommand(usageCmd)

	// usageCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
