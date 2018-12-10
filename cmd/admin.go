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
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/datacharmer/dbdeployer/sandbox"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"os"
	"path"
)

func unPreserveSandbox(sandboxDir, sandboxName string) {
	fullPath := path.Join(sandboxDir, sandboxName)
	if !common.DirExists(fullPath) {
		common.Exitf(1, globals.ErrDirectoryNotFound, fullPath)
	}
	preserve := path.Join(fullPath, globals.ScriptNoClearAll)
	if !common.ExecExists(preserve) {
		preserve = path.Join(fullPath, globals.ScriptNoClear)
	}
	if !common.ExecExists(preserve) {
		common.CondPrintf("Sandbox %s is not locked\n", sandboxName)
		return
	}
	isMultiple := true
	clear := path.Join(fullPath, globals.ScriptClearAll)
	if !common.ExecExists(clear) {
		clear = path.Join(fullPath, globals.ScriptClear)
		isMultiple = false
	}
	if !common.ExecExists(clear) {
		common.Exitf(1, globals.ErrExecutableNotFound, clear)
	}
	noClear := path.Join(fullPath, globals.ScriptNoClear)
	if isMultiple {
		noClear = path.Join(fullPath, globals.ScriptNoClearAll)
	}
	err := os.Remove(clear)
	common.ErrCheckExitf(err, 1, globals.ErrWhileRemoving, clear, err)
	err = os.Rename(noClear, clear)
	common.ErrCheckExitf(err, 1, globals.ErrWhileRenamingScript, err)
	common.CondPrintf("Sandbox %s unlocked\n", sandboxName)
}

func preserveSandbox(sandboxDir, sandboxName string) {
	fullPath := path.Join(sandboxDir, sandboxName)
	if !common.DirExists(fullPath) {
		common.Exitf(1, globals.ErrDirectoryNotFound, fullPath)
	}
	preserve := path.Join(fullPath, globals.ScriptNoClearAll)
	if !common.ExecExists(preserve) {
		preserve = path.Join(fullPath, globals.ScriptNoClear)
	}
	if common.ExecExists(preserve) {
		common.CondPrintf("Sandbox %s is already locked\n", sandboxName)
		return
	}
	isMultiple := true
	clear := path.Join(fullPath, globals.ScriptClearAll)
	if !common.ExecExists(clear) {
		clear = path.Join(fullPath, globals.ScriptClear)
		isMultiple = false
	}
	if !common.ExecExists(clear) {
		common.Exitf(1, globals.ErrExecutableNotFound, clear)
	}
	noClear := path.Join(fullPath, globals.ScriptNoClear)
	clearCmd := globals.ScriptClear
	noClearCmd := globals.ScriptNoClear
	if isMultiple {
		noClear = path.Join(fullPath, globals.ScriptNoClearAll)
		clearCmd = globals.ScriptClearAll
		noClearCmd = globals.ScriptNoClearAll
	}
	err := os.Rename(clear, noClear)
	common.ErrCheckExitf(err, 1, globals.ErrWhileRenamingScript, err)
	template := sandbox.SingleTemplates["sb_locked_template"].Contents
	var data = common.StringMap{
		"TemplateName": "sb_locked_template",
		"SandboxDir":   sandboxName,
		"AppVersion":   common.VersionDef,
		"Copyright":    sandbox.Copyright,
		"ClearCmd":     clearCmd,
		"NoClearCmd":   noClearCmd,
	}
	template = common.TrimmedLines(template)
	newClearMessage := common.TemplateFill(template, data)
	err = common.WriteString(newClearMessage, clear)
	if err != nil {
		common.Exitf(1, "%+v", err)
	}
	err = os.Chmod(clear, 0744)
	if err != nil {
		common.Exitf(1, "%+v", err)
	}
	common.CondPrintf("Sandbox %s locked\n", sandboxName)
}

func lockSandbox(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		common.Exit(1,
			"'lock' requires the name of a sandbox (or ALL)",
			"Example: dbdeployer admin lock msb_5_7_21")
	}
	candidateSandbox := args[0]
	sandboxDir, err := getAbsolutePathFromFlag(cmd, "sandbox-home")
	if err != nil {
		common.Exitf(1, "%+v", err)
	}
	lockList := []string{candidateSandbox}
	if candidateSandbox == "ALL" || candidateSandbox == "all" {
		installedSandboxes, err := common.GetInstalledSandboxes(sandboxDir)
		common.ErrCheckExitf(err, 1, globals.ErrRetrievingSandboxList, err)
		lockList = common.SandboxInfoToFileNames(installedSandboxes)
	}
	if len(lockList) == 0 {
		common.CondPrintf("Nothing to lock in %s\n", sandboxDir)
		return
	}
	for _, sb := range lockList {
		preserveSandbox(sandboxDir, sb)
	}
}

func unlockSandbox(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		common.Exit(1,
			"'unlock' requires the name of a sandbox (or ALL)",
			"Example: dbdeployer admin unlock msb_5_7_21")
	}
	candidateSandbox := args[0]
	sandboxDir, err := getAbsolutePathFromFlag(cmd, "sandbox-home")
	if err != nil {
		common.Exitf(1, "%+v", err)
	}
	lockList := []string{candidateSandbox}
	if candidateSandbox == "ALL" || candidateSandbox == "all" {
		installedSandboxes, err := common.GetInstalledSandboxes(sandboxDir)
		common.ErrCheckExitf(err, 1, globals.ErrRetrievingSandboxList, err)
		lockList = common.SandboxInfoToFileNames(installedSandboxes)
	}
	if len(lockList) == 0 {
		common.CondPrintf("Nothing to lock in %s\n", sandboxDir)
		return
	}
	for _, sb := range lockList {
		unPreserveSandbox(sandboxDir, sb)
	}
}

func upgradeSandbox(sandboxDir, oldSandbox, newSandbox string) error {
	var possibleUpgrades = map[string]string{
		"5.0": "5.1",
		"5.1": "5.5",
		"5.5": "5.6",
		"5.6": "5.7",
		"5.7": "8.0",
		"8.0": "8.0",
	}
	err := os.Chdir(sandboxDir)
	common.ErrCheckExitf(err, 1, "can't change directory to %s", sandboxDir)
	scripts := []string{globals.ScriptStart, globals.ScriptStop, globals.ScriptMy}
	for _, dir := range []string{oldSandbox, newSandbox} {
		if !common.DirExists(dir) {
			common.Exitf(1, globals.ErrDirectoryNotFoundInUpper, dir, sandboxDir)
		}
		for _, script := range scripts {
			if !common.ExecExists(path.Join(dir, script)) {
				common.Exit(1, fmt.Sprintf(globals.ErrScriptNotFoundInUpper, script, dir),
					"The upgrade only works between SINGLE deployments")
			}
		}
	}
	newSbdesc, err := common.ReadSandboxDescription(newSandbox)
	if err != nil {
		return errors.Wrapf(err, "error reading new sandbox description")
	}
	oldSbdesc, err := common.ReadSandboxDescription(oldSandbox)
	if err != nil {
		return errors.Wrapf(err, "error reading old sandbox description")
	}
	mysqlUpgrade := path.Join(newSbdesc.Basedir, "bin", "mysql_upgrade")
	if !common.ExecExists(mysqlUpgrade) {
		_ = common.WriteString("", path.Join(newSandbox, "no_upgrade"))
		return errors.Errorf("mysql_upgrade not found in %s. Upgrade is not possible", newSbdesc.Basedir)
	}
	newVersionList, err := common.VersionToList(newSbdesc.Version)
	if err != nil {
		return errors.Wrapf(err, "error converting new sandbox version to major/minor/rev")
	}
	newMajor := newVersionList[0]
	newMinor := newVersionList[1]
	newRev := newVersionList[2]
	oldVersionList, err := common.VersionToList(oldSbdesc.Version)
	if err != nil {
		return errors.Wrapf(err, "error converting old sandbox version to major/minor/rev")
	}
	oldMajor := oldVersionList[0]
	oldMinor := oldVersionList[1]
	oldRev := oldVersionList[2]
	newUpgradeVersion := fmt.Sprintf("%d.%d", newVersionList[0], newVersionList[1])
	oldUpgradeVersion := fmt.Sprintf("%d.%d", oldVersionList[0], oldVersionList[1])
	if oldMajor == 10 || newMajor == 10 {
		common.Exit(1, "upgrade from and to MariaDB is not supported")
	}
	greaterThanNewVersion, err := common.GreaterOrEqualVersion(oldSbdesc.Version, newVersionList)
	common.ErrCheckExitf(err, 1, globals.ErrWhileComparingVersions)
	if greaterThanNewVersion {
		common.Exitf(1, "version %s must be greater than %s", newUpgradeVersion, oldUpgradeVersion)
	}
	canBeUpgraded := false
	if oldMajor < newMajor {
		canBeUpgraded = true
	} else {
		if oldMajor == newMajor && oldMinor < newMinor {
			canBeUpgraded = true
		} else {
			if oldMajor == newMajor && oldMinor == newMinor && oldRev < newRev {
				canBeUpgraded = true
			}
		}
	}
	if !canBeUpgraded {
		return errors.Errorf("version '%s' can only be upgraded to '%s' or to the same version with a higher revision", oldUpgradeVersion, possibleUpgrades[oldUpgradeVersion])
	}
	newSandboxOldData := path.Join(newSandbox, globals.DataDirName+"-"+newSandbox)
	if common.DirExists(newSandboxOldData) {
		return errors.Errorf("sandbox '%s' is already the upgrade from an older version", newSandbox)
	}
	_, err = common.RunCmd(path.Join(oldSandbox, globals.ScriptStop))
	if err != nil {
		return errors.Wrapf(err, globals.ErrWhileStoppingSandbox, oldSandbox)
	}
	_, err = common.RunCmd(path.Join(newSandbox, globals.ScriptStop))
	if err != nil {
		return errors.Wrapf(err, globals.ErrWhileStoppingSandbox, newSandbox)
	}
	mvArgs := []string{path.Join(newSandbox, globals.DataDirName), newSandboxOldData}
	_, err = common.RunCmdWithArgs("mv", mvArgs)
	if err != nil {
		return errors.Wrapf(err, "error while moving data directory in sandbox %s", newSandbox)
	}

	mvArgs = []string{path.Join(oldSandbox, globals.DataDirName), path.Join(newSandbox, globals.DataDirName)}
	_, err = common.RunCmdWithArgs("mv", mvArgs)
	if err != nil {
		return errors.Wrapf(err, "error while moving data directory from sandbox %s to %s", oldSandbox, newSandbox)
	}
	common.CondPrintf("Data directory %s/data moved to %s/data \n", oldSandbox, newSandbox)

	_, err = common.RunCmd(path.Join(newSandbox, globals.ScriptStart))
	if err != nil {
		return errors.Wrapf(err, globals.ErrWhileStartingSandbox, newSandbox)
	}
	upgradeArgs := []string{"sql_upgrade"}
	_, err = common.RunCmdWithArgs(path.Join(newSandbox, globals.ScriptMy), upgradeArgs)
	if err != nil {

		return errors.Wrapf(err, "error while running mysql_upgrade in %s", newSandbox)
	}
	fmt.Println("")
	common.CondPrintf("The data directory from %s/data is preserved in %s\n", newSandbox, newSandboxOldData)
	common.CondPrintf("The data directory from %s/data is now used in %s/data\n", oldSandbox, newSandbox)
	common.CondPrintf("%s is not operational and can be deleted\n", oldSandbox)
	return nil
}

func runUpgradeSandbox(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		common.Exit(1,
			"'upgrade' requires the name of two sandboxes ",
			"Example: dbdeployer admin upgrade msb_5_7_23 msb_8_0_12")
	}
	oldSandbox := args[0]
	newSandbox := args[1]
	sandboxDir, err := getAbsolutePathFromFlag(cmd, "sandbox-home")
	if err != nil {
		common.Exitf(1, "%+v", err)
	}
	err = upgradeSandbox(sandboxDir, oldSandbox, newSandbox)
	if err != nil {
		common.Exitf(1, "%+v", err)
	}
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
		Run: lockSandbox,
	}

	adminUnlockCmd = &cobra.Command{
		Use:     "unlock sandbox_name",
		Aliases: []string{"unpreserve"},
		Short:   "Unlocks a sandbox",
		Long:    `Removes lock, allowing deletion of a given sandbox`,
		Run:     unlockSandbox,
	}
	adminUpgradeCmd = &cobra.Command{
		Use:   "upgrade sandbox_name newer_sandbox",
		Short: "Upgrades a sandbox to a newer version",
		Long: `Upgrades a sandbox to a newer version.
The sandbox with the new version must exist already.
The data directory of the old sandbox will be moved to the new one.`,
		Example: "dbdeployer admin upgrade msb_8_0_11 msb_8_0_12",
		Run:     runUpgradeSandbox,
	}
)

func init() {
	rootCmd.AddCommand(adminCmd)
	adminCmd.AddCommand(adminLockCmd)
	adminCmd.AddCommand(adminUnlockCmd)
	adminCmd.AddCommand(adminUpgradeCmd)
}
