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
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/spf13/cobra"
	"os"
)

func ShowDefaults(cmd *cobra.Command, args []string) {
	defaults.ShowDefaults(defaults.Defaults())
}

func WriteDefaults(cmd *cobra.Command, args []string) {
	defaults.WriteDefaultsFile(defaults.ConfigurationFile, defaults.Defaults())
	fmt.Printf("# Default values exported to %s\n",defaults.ConfigurationFile)
}

func RemoveDefaults(cmd *cobra.Command, args []string) {
	defaults.RemoveDefaultsFile()
}

func LoadDefaults(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Printf("'load' requires a file name\n")
		os.Exit(1)
	}
	filename := args[0]
	new_defaults := defaults.ReadDefaultsFile(filename)
	if defaults.ValidateDefaults(new_defaults) {
		defaults.WriteDefaultsFile(defaults.ConfigurationFile, new_defaults)
	}
}

func ExportDefaults(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Printf("'export' requires a file name\n")
		os.Exit(1)
	}
	filename := args[0]
	if common.FileExists(filename) {
		fmt.Printf("File %s already exists. Will not overwrite\n", filename)
		os.Exit(1)
	}
	defaults.WriteDefaultsFile(filename, defaults.Defaults())
	fmt.Printf("# Defaults exported to file %s\n", filename)
}

func UpdateDefaults(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		fmt.Printf("'update' requires a label and a value\n")
		fmt.Printf("Example: dbdeployer admin update master-slave-base-port 17500")
		os.Exit(1)
	}
	label := args[0]
	value := args[1]
	defaults.UpdateDefaults(label, value)
	defaults.ShowDefaults(defaults.Defaults())
}

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
		fmt.Println("Executable '%s' not found\n", clear)
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
		fmt.Println("Executable '%s' not found\n", clear)
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
	if sandbox == "ALL" {
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
	if sandbox == "ALL" {
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
		Short: "administrative tasks",
		Aliases: []string{"defaults"},
		Long: `Runs commands related to the administration of dbdeployer,
such as showing the defaults and saving new ones.`,
	}

	adminShowCmd = &cobra.Command{
		Use:     "show",
		Short:   "shows defaults",
		Aliases: []string{"list"},
		Long:    `Shows currently defined defaults`,
		Run:     ShowDefaults,
	}

	adminLoadCmd = &cobra.Command{
		Use:   "load file_name",
		Short: "Load defaults from file",
		Long:  fmt.Sprintf(`Reads defaults from file and saves them to dbdeployer configuration file (%s)`, defaults.ConfigurationFile),
		Run:   LoadDefaults,
	}

	adminUpdateCmd = &cobra.Command{
		Use:   "update label value",
		Short: "Load defaults from file",
		Example: `
	$ dbdeployer admin update master-slave-base-port 17500		
`,
		Long: `Updates one field of the defaults. Stores the result in the dbdeployer configuration file.
Use "dbdeployer admin show" to see which values are available`,
		Run: UpdateDefaults,
	}

	adminExportCmd = &cobra.Command{
		Use:   "export filename",
		Short: "Export current defaults to a given file",
		Long:  `Saves current defaults to a fileer-defined file`,
		Run:   ExportDefaults,
	}

	adminStoreCmd = &cobra.Command{
		Use:   "store",
		Short: "Store current defaults",
		Long:  fmt.Sprintf(`Saves current defaults to dbdeployer configuration file (%s)`, defaults.ConfigurationFile),
		Run:   WriteDefaults,
	}

	adminRemoveCmd = &cobra.Command{
		Use:     "reset",
		Aliases: []string{"remove"},
		Short:   "Remove current defaults file",
		Long: fmt.Sprintf(`Removes current dbdeployer configuration file (%s)`, defaults.ConfigurationFile) + `
Afterwards, dbdeployer will use the internally stored defaults.
`,
		Run: RemoveDefaults,
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
		Long: `iRemoves lock, allowing deletion of a given sandbox`,
		Run: UnlockSandbox,
	}

)

func init() {
	rootCmd.AddCommand(adminCmd)
	adminCmd.AddCommand(adminStoreCmd)
	adminCmd.AddCommand(adminShowCmd)
	adminCmd.AddCommand(adminRemoveCmd)
	adminCmd.AddCommand(adminLoadCmd)
	adminCmd.AddCommand(adminUpdateCmd)
	adminCmd.AddCommand(adminExportCmd)
	adminCmd.AddCommand(adminLockCmd)
	adminCmd.AddCommand(adminUnlockCmd)
}
