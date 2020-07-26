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
	"os/exec"
	"path"

	"github.com/spf13/cobra"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/globals"
)

func runInteractiveCmd(s string) error {
	// #nosec G204
	cmd := exec.Command(s)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func useSandbox(cmd *cobra.Command, args []string) error {
	sandboxHome, _ := cmd.Flags().GetString(globals.SandboxHomeLabel)
	sandbox := ""
	sandboxList, err := common.GetSandboxesByDate(sandboxHome)
	if len(args) > 0 {
		sandbox = args[0]
	} else {
		if err != nil {
			return err
		}
		if len(sandboxList) == 0 {
			return fmt.Errorf("nothing to use. No sandboxes were found")
		}
		sandbox = sandboxList[len(sandboxList)-1].SandboxName
	}
	for _, sb := range sandboxList {
		if sb.SandboxName == sandbox {

			sandboxDir := path.Join(sandboxHome, sandbox)
			fmt.Printf("using %s\n", sandboxDir)
			useSingle := path.Join(sandboxDir, "use")
			useMulti := path.Join(sandboxDir, "n1")
			if common.ExecExists(useSingle) {
				fmt.Printf("%s\n", useSingle)
				err = runInteractiveCmd(useSingle)
				return err
			} else if common.ExecExists(useMulti) {
				err = runInteractiveCmd(useMulti)
				return err
			} else {
				return fmt.Errorf("no executable found for %s", sandbox)
			}
		}
	}
	return fmt.Errorf("sandbox %s not found", sandbox)
}

var useCmd = &cobra.Command{
	Use:   "use [sandbox_name]",
	Short: "uses a sandbox",
	Long: `Uses a given sandbox.
If a sandbox is indicated, it will be used.
Otherwise, it will use the latest deployed sandbox`,
	RunE: useSandbox,
}

func init() {
	rootCmd.AddCommand(useCmd)
}
