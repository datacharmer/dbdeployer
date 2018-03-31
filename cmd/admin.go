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
	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/sandbox"
	"github.com/spf13/cobra"
	"os"
)

func UnpreserveSandbox(sandbox_dir, sandbox_name string) {
	full_path := sandbox_dir + "/" + sandbox_name
	if !common.DirExists(full_path) {
		fmt.Printf("Directory '%s' not found\n", full_path)
		os.Exit(1)
	}
	preserve := full_path + "/no_clear_all"
	if !common.ExecExists(preserve) {
		preserve = full_path + "/no_clear"
	}
	if !common.ExecExists(preserve) {
		fmt.Printf("Sandbox %s is not locked\n",sandbox_name)
		return	
	}
	is_multiple := true
	clear := full_path + "/clear_all"
	if !common.ExecExists(clear) {
		clear = full_path + "/clear"
		is_multiple = false
	}
	if !common.ExecExists(clear) {
		fmt.Printf("Executable '%s' not found\n", clear)
		os.Exit(1)
	}
	no_clear := full_path + "/no_clear"
	if is_multiple {
		no_clear = full_path + "/no_clear_all"
	}
	err := os.Remove(clear)
	if err != nil {
		fmt.Printf("Error while removing %s \n%s\n",clear, err)
		os.Exit(1)
	}
	err = os.Rename(no_clear, clear)
	if err != nil {
		fmt.Printf("Error while renaming  script\n%s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Sandbox %s unlocked\n",sandbox_name)
}



func PreserveSandbox(sandbox_dir, sandbox_name string) {
	full_path := sandbox_dir + "/" + sandbox_name
	if !common.DirExists(full_path) {
		fmt.Printf("Directory '%s' not found\n", full_path)
		os.Exit(1)
	}
	preserve := full_path + "/no_clear_all"
	if !common.ExecExists(preserve) {
		preserve = full_path + "/no_clear"
	}
	if common.ExecExists(preserve) {
		fmt.Printf("Sandbox %s is already locked\n",sandbox_name)
		return	
	}
	is_multiple := true
	clear := full_path + "/clear_all"
	if !common.ExecExists(clear) {
		clear = full_path + "/clear"
		is_multiple = false
	}
	if !common.ExecExists(clear) {
		fmt.Printf("Executable '%s' not found\n", clear)
		os.Exit(1)
	}
	no_clear := full_path + "/no_clear"
	clear_cmd := "clear"
	no_clear_cmd := "no_clear"
	if is_multiple {
		no_clear = full_path + "/no_clear_all"
		clear_cmd = "clear_all"
		no_clear_cmd = "no_clear_all"
	}
	err := os.Rename(clear, no_clear)
	if err != nil {
		fmt.Printf("Error while renaming script.\n%s\n",err)
		os.Exit(1)
	}
	template := sandbox.SingleTemplates["sb_locked_template"].Contents
	var data = common.Smap{
		"TemplateName" : "sb_locked_template",
		"SandboxDir" : sandbox_name,
		"AppVersion" : common.VersionDef,
		"Copyright" : sandbox.Copyright,
		"ClearCmd" : clear_cmd,
		"NoClearCmd" : no_clear_cmd,
	}
	template = common.TrimmedLines(template)
	new_clear_message := common.Tprintf(template, data)
	common.WriteString(new_clear_message, clear)
	os.Chmod(clear, 0744)
	fmt.Printf("Sandbox %s locked\n",sandbox_name)
}

func LockSandbox(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Printf("'lock' requires the name of a sandbox (or ALL)")
		fmt.Printf("Example: dbdeployer admin lock msb_5_7_21")
		os.Exit(1)
	}
	flags := cmd.Flags()
	sandbox := args[0]
	sandbox_dir, _ := flags.GetString("sandbox-home")
	lock_list := []string{sandbox}
	if sandbox == "ALL" || sandbox == "all" {
		lock_list = common.SandboxInfoToFileNames(common.GetInstalledSandboxes(sandbox_dir))
	}
	if len(lock_list) == 0 {
		fmt.Printf("Nothing to lock in %s\n", sandbox_dir)
		return
	}
	for _, sb := range lock_list {
		PreserveSandbox(sandbox_dir, sb)
	}
}

func UnlockSandbox(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Printf("'unlock' requires the name of a sandbox (or ALL)")
		fmt.Printf("Example: dbdeployer admin unlock msb_5_7_21")
		os.Exit(1)
	}
	flags := cmd.Flags()
	sandbox := args[0]
	sandbox_dir, _ := flags.GetString("sandbox-home")
	lock_list := []string{sandbox}
	if sandbox == "ALL" || sandbox == "all" {
		lock_list = common.SandboxInfoToFileNames(common.GetInstalledSandboxes(sandbox_dir))
	}
	if len(lock_list) == 0 {
		fmt.Printf("Nothing to lock in %s\n", sandbox_dir)
		return
	}
	for _, sb := range lock_list {
		UnpreserveSandbox(sandbox_dir, sb)
	}
}


var (
	adminCmd = &cobra.Command{
		Use:   "admin",
		Short: "sandbox management tasks",
		Aliases: []string{"manage"},
		Long: `Runs commands related to the administration of sandboxes.`,
	}

	adminLockCmd = &cobra.Command{
		Use:     "lock sandbox_name",
		Aliases: []string{"preserve"},
		Short:   "Locks a sandbox, preventing deletion",
		Long: `Prevents deletion for a given sandbox.
Note that the deletion being prevented is only the one occurring through dbdeployer. 
Users can still delete locked sandboxes manually.`,
		Run: LockSandbox,
	}

	adminUnlockCmd = &cobra.Command{
		Use:     "unlock sandbox_name",
		Aliases: []string{"unpreserve"},
		Short:   "Unlocks a sandbox",
		Long: `Removes lock, allowing deletion of a given sandbox`,
		Run: UnlockSandbox,
	}
)

func init() {
	rootCmd.AddCommand(adminCmd)
	adminCmd.AddCommand(adminLockCmd)
	adminCmd.AddCommand(adminUnlockCmd)
}
