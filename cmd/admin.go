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
		common.Exitf(1, "Directory '%s' not found", full_path)
	}
	preserve := full_path + "/no_clear_all"
	if !common.ExecExists(preserve) {
		preserve = full_path + "/no_clear"
	}
	if !common.ExecExists(preserve) {
		fmt.Printf("Sandbox %s is not locked\n", sandbox_name)
		return
	}
	is_multiple := true
	clear := full_path + "/clear_all"
	if !common.ExecExists(clear) {
		clear = full_path + "/clear"
		is_multiple = false
	}
	if !common.ExecExists(clear) {
		common.Exitf(1, "Executable '%s' not found", clear)
	}
	no_clear := full_path + "/no_clear"
	if is_multiple {
		no_clear = full_path + "/no_clear_all"
	}
	err := os.Remove(clear)
	common.ErrCheckExitf(err, 1, "Error while removing %s \n%s", clear, err)
	err = os.Rename(no_clear, clear)
	common.ErrCheckExitf(err, 1, "Error while renaming  script\n%s", err)
	fmt.Printf("Sandbox %s unlocked\n", sandbox_name)
}

func PreserveSandbox(sandbox_dir, sandbox_name string) {
	full_path := sandbox_dir + "/" + sandbox_name
	if !common.DirExists(full_path) {
		common.Exitf(1, "Directory '%s' not found", full_path)
	}
	preserve := full_path + "/no_clear_all"
	if !common.ExecExists(preserve) {
		preserve = full_path + "/no_clear"
	}
	if common.ExecExists(preserve) {
		fmt.Printf("Sandbox %s is already locked\n", sandbox_name)
		return
	}
	is_multiple := true
	clear := full_path + "/clear_all"
	if !common.ExecExists(clear) {
		clear = full_path + "/clear"
		is_multiple = false
	}
	if !common.ExecExists(clear) {
		common.Exitf(1, "Executable '%s' not found", clear)
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
	common.ErrCheckExitf(err, 1, "Error while renaming script.\n%s", err)
	template := sandbox.SingleTemplates["sb_locked_template"].Contents
	var data = common.StringMap{
		"TemplateName": "sb_locked_template",
		"SandboxDir":   sandbox_name,
		"AppVersion":   common.VersionDef,
		"Copyright":    sandbox.Copyright,
		"ClearCmd":     clear_cmd,
		"NoClearCmd":   no_clear_cmd,
	}
	template = common.TrimmedLines(template)
	new_clear_message := common.TemplateFill(template, data)
	common.WriteString(new_clear_message, clear)
	os.Chmod(clear, 0744)
	fmt.Printf("Sandbox %s locked\n", sandbox_name)
}

func LockSandbox(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		common.Exit(1,
			"'lock' requires the name of a sandbox (or ALL)",
			"Example: dbdeployer admin lock msb_5_7_21")
	}
	sandbox := args[0]
	sandbox_dir := GetAbsolutePathFromFlag(cmd, "sandbox-home")
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
		common.Exit(1,
			"'unlock' requires the name of a sandbox (or ALL)",
			"Example: dbdeployer admin unlock msb_5_7_21")
	}
	sandbox := args[0]
	sandbox_dir := GetAbsolutePathFromFlag(cmd, "sandbox-home")
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

func UpgradeSandbox(sandbox_dir, old_sandbox, new_sandbox string) {
	var possible_upgrades = map[string]string{
		"5.0": "5.1",
		"5.1": "5.5",
		"5.5": "5.6",
		"5.6": "5.7",
		"5.7": "8.0",
		"8.0": "8.0",
	}
	err := os.Chdir(sandbox_dir)
	common.ErrCheckExitf(err, 1, "Error: can't change directory to %s", sandbox_dir)
	scripts := []string{"start", "stop", "my"}
	for _, dir := range []string{old_sandbox, new_sandbox} {
		if !common.DirExists(dir) {
			common.Exitf(1, "Error: Directory %s not found in %s", dir, sandbox_dir)
		}
		for _, script := range scripts {
			if !common.ExecExists(dir + "/" + script) {
				common.Exit(1, fmt.Sprintf("Error: script %s not found in %s", script, dir),
					"The upgrade only works between SINGLE deployments")
			}
		}
	}
	new_sbdesc := common.ReadSandboxDescription(new_sandbox)
	old_sbdesc := common.ReadSandboxDescription(old_sandbox)
	mysql_upgrade := new_sbdesc.Basedir + "/bin/mysql_upgrade"
	if !common.ExecExists(mysql_upgrade) {
		common.WriteString("", new_sandbox+"/no_upgrade")
		common.Exit(0, "mysql_upgrade not found in %s. Upgrade is not possible", new_sbdesc.Basedir)
	}
	new_version_list := common.VersionToList(new_sbdesc.Version)
	new_major := new_version_list[0]
	new_minor := new_version_list[1]
	new_rev := new_version_list[2]
	old_version_list := common.VersionToList(old_sbdesc.Version)
	old_major := old_version_list[0]
	old_minor := old_version_list[1]
	old_rev := old_version_list[2]
	new_upgrade_version := fmt.Sprintf("%d.%d", new_version_list[0], new_version_list[1])
	old_upgrade_version := fmt.Sprintf("%d.%d", old_version_list[0], old_version_list[1])
	if old_major == 10 || new_major == 10 {
		common.Exit(1, "Upgrade from and to MariaDB is not supported")
	}
	if common.GreaterOrEqualVersion(old_sbdesc.Version, new_version_list) {
		common.Exitf(1, "Version %s must be greater than %s", new_upgrade_version, old_upgrade_version)
	}
	can_be_upgraded := false
	if old_major < new_major {
		can_be_upgraded = true
	} else {
		if old_major == new_major && old_minor < new_minor {
			can_be_upgraded = true
		} else {
			if old_major == new_major && old_minor == new_minor && old_rev < new_rev {
				can_be_upgraded = true
			}
		}
	}
	if !can_be_upgraded {
		common.Exitf(1, "Version %s can only be upgraded to %s or to the same version with a higher revision", old_upgrade_version, possible_upgrades[old_upgrade_version])
	}
	new_sandbox_old_data := new_sandbox + "/data-" + new_sandbox
	if common.DirExists(new_sandbox_old_data) {
		common.Exitf(1, "Sandbox %s is already the upgrade from an older version", new_sandbox)
	}
	err, _ = common.Run_cmd(old_sandbox + "/stop")
	common.ErrCheckExitf(err, 1, "Error while stopping sandbox %s", old_sandbox)
	err, _ = common.Run_cmd(new_sandbox + "/stop")
	common.ErrCheckExitf(err, 1, "Error while stopping sandbox %s", new_sandbox)
	mv_args := []string{new_sandbox + "/data", new_sandbox_old_data}
	err, _ = common.Run_cmd_with_args("mv", mv_args)
	common.ErrCheckExitf(err, 1, "Error while moving data directory in sandbox %s", new_sandbox)

	mv_args = []string{old_sandbox + "/data", new_sandbox + "/data"}
	err, _ = common.Run_cmd_with_args("mv", mv_args)
	common.ErrCheckExitf(err, 1, "Error while moving data directory from sandbox %s to %s", old_sandbox, new_sandbox)
	fmt.Printf("Data directory %s/data moved to %s/data \n", old_sandbox, new_sandbox)

	err, _ = common.Run_cmd(new_sandbox + "/start")
	common.ErrCheckExitf(err, 1, "Error while starting sandbox %s", new_sandbox)
	upgrade_args := []string{"sql_upgrade"}
	err, _ = common.Run_cmd_with_args(new_sandbox+"/my", upgrade_args)
	common.ErrCheckExitf(err, 1, "Error while running mysql_upgrade in %s", new_sandbox)
	fmt.Println("")
	fmt.Printf("The data directory from %s/data is preserved in %s\n", new_sandbox, new_sandbox_old_data)
	fmt.Printf("The data directory from %s/data is now used in %s/data\n", old_sandbox, new_sandbox)
	fmt.Printf("%s is not operational and can be deleted\n", old_sandbox)
}

func RunUpgradeSandbox(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		common.Exit(1,
			"'upgrade' requires the name of two sandboxes ",
			"Example: dbdeployer admin upgrade msb_5_7_23 msb_8_0_12")
	}
	old_sandbox := args[0]
	new_sandbox := args[1]
	sandbox_dir := GetAbsolutePathFromFlag(cmd, "sandbox-home")
	UpgradeSandbox(sandbox_dir, old_sandbox, new_sandbox)
}

var (
	adminCmd = &cobra.Command{
		Use:     "admin",
		Short:   "sandbox management tasks",
		Aliases: []string{"manage"},
		Long:    `Runs commands related to the administration of sandboxes.`,
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
		Long:    `Removes lock, allowing deletion of a given sandbox`,
		Run:     UnlockSandbox,
	}
	adminUpgradeCmd = &cobra.Command{
		Use:   "upgrade sandbox_name newer_sandbox",
		Short: "Upgrades a sandbox to a newer version",
		Long: `Upgrades a sandbox to a newer version.
The sandbox with the new version must exist already.
The data directory of the old sandbox will be moved to the new one.`,
		Example: "dbdeployer admin upgrade msb_8_0_11 msb_8_0_12",
		Run:     RunUpgradeSandbox,
	}
)

func init() {
	rootCmd.AddCommand(adminCmd)
	adminCmd.AddCommand(adminLockCmd)
	adminCmd.AddCommand(adminUnlockCmd)
	adminCmd.AddCommand(adminUpgradeCmd)
}
