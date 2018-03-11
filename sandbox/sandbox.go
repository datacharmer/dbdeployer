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

package sandbox

import (
	"fmt"
	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/concurrent"
	"os"
	"time"
)

type SandboxDef struct {
	DirName           string
	SBType            string
	Multi             bool
	NodeNum           int
	Version           string
	Basedir           string
	SandboxDir        string
	LoadGrants        bool
	InstalledPorts    []int
	Port              int
	UserPort          int
	BasePort          int
	MorePorts         []int
	Prompt            string
	DbUser            string
	RplUser           string
	DbPassword        string
	RplPassword       string
	RemoteAccess      string
	BindAddress       string
	CustomMysqld      string
	ServerId          int
	ReplOptions       string
	GtidOptions       string
	InitOptions       []string
	MyCnfOptions      []string
	PreGrantsSql      []string
	PreGrantsSqlFile  string
	PostGrantsSql     []string
	PostGrantsSqlFile string
	MyCnfFile         string
	NativeAuthPlugin  bool
	KeepUuid          bool
	SinglePrimary     bool
	Force             bool
	ExposeDdTables    bool
	RunConcurrently   bool
}

func GetOptionsFromFile(filename string) (options []string) {
	skip_options := map[string]bool{
		"user":         true,
		"port":         true,
		"socket":       true,
		"datadir":      true,
		"basedir":      true,
		"tmpdir":       true,
		"pid-file":     true,
		"server-id":    true,
		"bind-address": true,
		"log-error":    true,
	}
	config := common.ParseConfigFile(filename)
	for _, kv := range config["mysqld"] {

		if skip_options[kv.Key] {
			continue
		}
		options = append(options, fmt.Sprintf("%s = %s", kv.Key, kv.Value))
		//fmt.Printf("%d %s : %s \n", N, kv.key, kv.value)
	}
	return options
}

func CheckDirectory(sdef SandboxDef) SandboxDef {
	sandbox_dir := sdef.SandboxDir
	if common.DirExists(sandbox_dir) {
		if sdef.Force {
			fmt.Printf("Overwriting directory %s\n", sandbox_dir)
			stop_command := sandbox_dir + "/stop"
			if !common.ExecExists(stop_command) {
				stop_command = sandbox_dir + "/stop_all"
			}
			if !common.ExecExists(stop_command) {
				fmt.Printf("Neither 'stop' or 'stop_all' found in %s\n", sandbox_dir)
			}

			used_ports_list := common.GetInstalledPorts(sandbox_dir)
			my_used_ports := make(map[int]bool)
			for _, p := range used_ports_list {
				my_used_ports[p] = true
			}

			common.Run_cmd(stop_command)
			err, _ := common.Run_cmd_with_args("rm", []string{"-rf", sandbox_dir})
			if err != nil {
				fmt.Printf("Error while deleting sandbox %s\n", sandbox_dir)
				os.Exit(1)
			}
			var new_installed_ports []int
			for _, port := range sdef.InstalledPorts {
				if !my_used_ports[port] {
					new_installed_ports = append(new_installed_ports, port)
				}
			}
			sdef.InstalledPorts = new_installed_ports
		} else {
			fmt.Printf("Directory %s already exists. Use --force to override.\n", sandbox_dir)
			os.Exit(1)
		}
	}
	return sdef
}

func FindFreePort(base_port int, installed_ports []int,  how_many int) int {
	used_ports := make(map[int]bool)
	for _, p := range installed_ports {
		used_ports[p] = true
	}
	free_port := 0
	check_port := base_port
	for free_port == 0 {
		is_free := true
		candidate_port := check_port
		for N := check_port ; N < (check_port + how_many + 1) ; N++ {
			_, exists := used_ports[N]
			if exists {
				is_free = false
			}
			// fmt.Printf("+%d(%v) ", N, is_free )
		}
		if is_free {
			free_port = candidate_port
		} else {
			check_port += how_many
		}
		if check_port > 60000 {
			fmt.Printf("Could not find a free range for %d\n", base_port)
			os.Exit(1)
		}
	}
	// fmt.Printf("%v, %d\n",installed_ports, check_port)
	if check_port != base_port {
		if os.Getenv("SHOW_CHANGED_PORTS") != "" {
			fmt.Printf("#port %d changed to %d\n",base_port, check_port)
		}
	}
	return check_port
}

func CheckPort(sandbox_type string, installed_ports []int, port int) {
	conflict := 0
	for _, p := range installed_ports {
		if p == port {
			conflict = p
		}
		/*
		if sandbox_type == "group-node" {
			if p == (port + defaults.Defaults().GroupPortDelta) {
				conflict = p
			}
		}
		*/
	}
	if conflict > 0 {
		fmt.Printf("Port conflict detected. Port %d is already used\n", conflict)
		os.Exit(1)
	}
}

func getmatch(key string, names []string, matches []string) string {
	if len(matches) < len(names) {
		return ""
	}
	for n, _ := range names {
		if names[n] == key {
			return matches[n]
		}
	}
	return ""
}

func FixServerUuid(sdef SandboxDef) (uuid_file, new_uuid string)  {
	if !common.GreaterOrEqualVersion(sdef.Version, []int{5, 6, 9}) {
		return
	}
	new_uuid = fmt.Sprintf("server-uuid=%s", common.MakeCustomizedUuid(sdef.Port, sdef.NodeNum))
	operation_dir := sdef.SandboxDir + "/data"
	uuid_file = operation_dir + "/auto.cnf"
	//if !common.DirExists(operation_dir) {
	//	fmt.Printf("Directory %s does not exist\n", operation_dir)
	//	os.Exit(1)
	//}
	//uuid_string = []string{"[auto]", new_uuid}
	return
	//err := common.WriteStrings(uuid_string, uuid_file, "")
	//if err != nil {
	//	fmt.Printf("%s\n", err)
	//	os.Exit(1)
	//}
	//check_uuid := common.SlurpAsString(uuid_file)
	//fmt.Printf("UUID file (%s) updated : %s\n", uuid_file, new_uuid)
	//fmt.Printf("new UUID : %s\n", check_uuid)
}

func slice_to_text(s_array []string) string {
	var text string = ""
	for _, v := range s_array {
		if len(v) > 0 {
			text += fmt.Sprintf("%s\n", v)
		}
	}
	return text
}

func CreateSingleSandbox(sdef SandboxDef, origin string) (exec_list []concurrent.ExecutionList) {

	var sandbox_dir string

	sdef.Basedir = sdef.Basedir + "/" + sdef.Version
	if !common.DirExists(sdef.Basedir) {
		fmt.Printf("Base directory %s does not exist\n", sdef.Basedir)
		os.Exit(1)
	}

	//fmt.Printf("origin: %s\n", origin)
	//fmt.Printf("def: %#v\n", sdef)
	// port = VersionToPort(sdef.Version)
	version_fname := common.VersionToName(sdef.Version)
	if sdef.Prompt == "" {
		sdef.Prompt = "mysql"
	}
	if sdef.DirName == "" {
		sdef.DirName = defaults.Defaults().SandboxPrefix + version_fname
	}
	sandbox_dir = sdef.SandboxDir + "/" + sdef.DirName
	sdef.SandboxDir = sandbox_dir
	datadir := sandbox_dir + "/data"
	tmpdir := sandbox_dir + "/tmp"
	global_tmp_dir := os.Getenv("TMPDIR")
	if global_tmp_dir == "" {
		global_tmp_dir = "/tmp"
	}
	if !common.DirExists(global_tmp_dir) {
		fmt.Printf("TMP directory %s does not exist\n", global_tmp_dir)
		os.Exit(1)
	}
	if sdef.ExposeDdTables {
		if !common.GreaterOrEqualVersion(sdef.Version, []int{8, 0, 0}) {
			fmt.Printf("--expose-dd-tables requires MySQL 8.0.0+\n")
			os.Exit(1)
		}
		sdef.PostGrantsSql = append(sdef.PostGrantsSql, SingleTemplates["expose_dd_tables"].Contents)
		if sdef.CustomMysqld != "" && sdef.CustomMysqld != "mysqld-debug" {
			fmt.Printf("--expose-dd-tables requires mysqld-debug. A different file was indicated (--custom-mysqld=%s)\n", sdef.CustomMysqld)
			fmt.Println("Either use \"mysqld-debug\" or remove --custom-mysqld")
			os.Exit(1)
		}
		sdef.CustomMysqld = "mysqld-debug"
	}
	if sdef.CustomMysqld != "" {
		custom_mysqld := sdef.Basedir + "/bin/" + sdef.CustomMysqld
		if !common.ExecExists(custom_mysqld) {
			fmt.Printf("File %s not found or not executable\n", custom_mysqld)
			fmt.Printf("The file \"%s\" (defined with --custom-mysqld) must be in the same directory as the regular mysqld\n", sdef.CustomMysqld)
			os.Exit(1)
		}
	}
	if common.GreaterOrEqualVersion(sdef.Version, []int{8, 0, 4}) {
		if sdef.NativeAuthPlugin == true {
			sdef.InitOptions = append(sdef.InitOptions, "--default_authentication_plugin=mysql_native_password")
			sdef.MyCnfOptions = append(sdef.MyCnfOptions, "default_authentication_plugin=mysql_native_password")
		}
	}
	if sdef.MyCnfFile != "" {
		options := GetOptionsFromFile(sdef.MyCnfFile)
		if len(options) > 0 {
			sdef.MyCnfOptions = append(sdef.MyCnfOptions, fmt.Sprintf("# options retrieved from %s", sdef.MyCnfFile))
		}
		for _, option := range options {
			// fmt.Printf("[%s]\n", option)
			sdef.MyCnfOptions = append(sdef.MyCnfOptions, option)
		}
	}
	//fmt.Printf("%#v\n", sdef)
	if sdef.NodeNum == 0 {
		sdef.Port = FindFreePort(sdef.Port, sdef.InstalledPorts, 1)
	}
	timestamp := time.Now()
	var data common.Smap = common.Smap{"Basedir": sdef.Basedir,
		"Copyright":    SingleTemplates["Copyright"].Contents,
		"AppVersion":   common.VersionDef,
		"DateTime":     timestamp.Format(time.UnixDate),
		"SandboxDir":   sandbox_dir,
		"CustomMysqld": sdef.CustomMysqld,
		"Port":         sdef.Port,
		"BasePort":     sdef.BasePort,
		"Prompt":       sdef.Prompt,
		"Version":      sdef.Version,
		"Datadir":      datadir,
		"Tmpdir":       tmpdir,
		"GlobalTmpDir": global_tmp_dir,
		"DbUser":       sdef.DbUser,
		"DbPassword":   sdef.DbPassword,
		"RplUser":      sdef.RplUser,
		"RplPassword":  sdef.RplPassword,
		"RemoteAccess": sdef.RemoteAccess,
		"BindAddress":  sdef.BindAddress,
		"OsUser":       os.Getenv("USER"),
		"ReplOptions":  sdef.ReplOptions,
		"GtidOptions":  sdef.GtidOptions,
		"ExtraOptions": slice_to_text(sdef.MyCnfOptions),
	}
	if sdef.ServerId > 0 {
		data["ServerId"] = fmt.Sprintf("server-id=%d", sdef.ServerId)
	} else {
		data["ServerId"] = ""
	}
	if common.DirExists(sandbox_dir) {
		sdef = CheckDirectory(sdef)
	}
	CheckPort(sdef.SBType, sdef.InstalledPorts, sdef.Port)

	//fmt.Printf("creating: %s\n", sandbox_dir)
	common.Mkdir(sandbox_dir)
	// fmt.Printf("creating: %s\n", datadir)
	common.Mkdir(datadir)
	// fmt.Printf("creating: %s\n", tmpdir)
	common.Mkdir(tmpdir)
	script := sdef.Basedir + "/scripts/mysql_install_db"
	init_script_flags := ""
	if common.GreaterOrEqualVersion(sdef.Version, []int{5, 7, 0}) {
		script = sdef.Basedir + "/bin/mysqld"
		init_script_flags = "--initialize-insecure"
	}
	// fmt.Printf("Script: %s\n", script)
	if !common.ExecExists(script) {
		fmt.Printf("Script '%s' not found\n", script)
		os.Exit(1)
	}
	if len(sdef.InitOptions) > 0 {
		for _, op := range sdef.InitOptions {
			init_script_flags += " " + op
		}
	}
	data["InitScript"] = script
	data["InitDefaults"] = "--no-defaults"
	if init_script_flags != "" {
		init_script_flags = fmt.Sprintf("\\\n    %s", init_script_flags)
	}
	data["ExtraInitFlags"] = init_script_flags
	data["FixUuidFile1"] = ""
	data["FixUuidFile2"] = ""

	if !sdef.KeepUuid {
		uuid_fname, new_uuid := FixServerUuid(sdef)
		if uuid_fname != "" {
			data["FixUuidFile1"] = fmt.Sprintf(`echo "[data]" > %s`, uuid_fname)
			data["FixUuidFile2"] = fmt.Sprintf(`echo "%s" >> %s`, new_uuid, uuid_fname)
		}
	}

	write_script(SingleTemplates, "init_db", "init_db_template", sandbox_dir, data, true)
	if sdef.RunConcurrently {
		var eCommand = concurrent.ExecCommand{
			Cmd : sandbox_dir+"/init_db",
			Args : []string{},
		}
		exec_list = append(exec_list, concurrent.ExecutionList{0, eCommand})
	} else {
		err, _ := common.Run_cmd_ctrl(sandbox_dir+"/init_db", true)
		if err == nil {
			// fmt.Printf("Database installed in %s\n", sandbox_dir)
			if !sdef.Multi {
				fmt.Printf("run 'dbdeployer usage single' for basic instructions'\n")
			}
		} else {
			fmt.Printf("err: %s\n", err)
		}
	}

	if sdef.SBType == "" {
		sdef.SBType = "single"
	}
	sb_item := defaults.SandboxItem{
		Origin : sdef.Basedir,
		SBType : sdef.SBType,
		Version: sdef.Version,
		Port:    []int{sdef.Port},
		Nodes:   []string{},
		Destination: sandbox_dir,
	}
	sb_desc := common.SandboxDescription{
		Basedir: sdef.Basedir,
		SBType:  sdef.SBType,
		Version: sdef.Version,
		Port:    []int{sdef.Port},
		Nodes:   0,
		NodeNum: sdef.NodeNum,
	}
	if len(sdef.MorePorts) > 0 {
		for _, port := range sdef.MorePorts {
			sb_desc.Port = append(sb_desc.Port, port)
			sb_item.Port = append(sb_item.Port, port)
		}
	}
	common.WriteSandboxDescription(sandbox_dir, sb_desc)
	if sdef.SBType == "single" {
		defaults.UpdateCatalog(sandbox_dir, sb_item)
	}
	write_script(SingleTemplates, "start", "start_template", sandbox_dir, data, true)
	write_script(SingleTemplates, "status", "status_template", sandbox_dir, data, true)
	write_script(SingleTemplates, "stop", "stop_template", sandbox_dir, data, true)
	write_script(SingleTemplates, "clear", "clear_template", sandbox_dir, data, true)
	write_script(SingleTemplates, "use", "use_template", sandbox_dir, data, true)
	write_script(SingleTemplates, "send_kill", "send_kill_template", sandbox_dir, data, true)
	write_script(SingleTemplates, "restart", "restart_template", sandbox_dir, data, true)
	write_script(SingleTemplates, "load_grants", "load_grants_template", sandbox_dir, data, true)
	write_script(SingleTemplates, "add_option", "add_option_template", sandbox_dir, data, true)
	write_script(SingleTemplates, "my", "my_template", sandbox_dir, data, true)
	write_script(SingleTemplates, "show_binlog", "show_binlog_template", sandbox_dir, data, true)
	write_script(SingleTemplates, "show_relaylog", "show_relaylog_template", sandbox_dir, data, true)
	write_script(SingleTemplates, "test_sb", "test_sb_template", sandbox_dir, data, true)

	write_script(SingleTemplates, "my.sandbox.cnf", "my_cnf_template", sandbox_dir, data, false)
	switch {
	case common.GreaterOrEqualVersion(sdef.Version, []int{8, 0, 0}):
		write_script(SingleTemplates, "grants.mysql", "grants_template8x", sandbox_dir, data, false)
	case common.GreaterOrEqualVersion(sdef.Version, []int{5, 7, 6}):
		write_script(SingleTemplates, "grants.mysql", "grants_template57", sandbox_dir, data, false)
	default:
		write_script(SingleTemplates, "grants.mysql", "grants_template5x", sandbox_dir, data, false)
	}
	write_script(SingleTemplates, "sb_include", "sb_include_template", sandbox_dir, data, false)


	pre_grant_sql_file := sandbox_dir + "/pre_grants.sql"
	post_grant_sql_file := sandbox_dir + "/post_grants.sql"
	if sdef.PreGrantsSqlFile != "" {
		common.CopyFile(sdef.PreGrantsSqlFile, pre_grant_sql_file)
	}
	if sdef.PostGrantsSqlFile != "" {
		common.CopyFile(sdef.PostGrantsSqlFile, post_grant_sql_file)
	}

	if len(sdef.PreGrantsSql) > 0 {
		if common.FileExists(pre_grant_sql_file) {
			common.AppendStrings(sdef.PreGrantsSql, pre_grant_sql_file, ";")
		} else {
			common.WriteStrings(sdef.PreGrantsSql, pre_grant_sql_file, ";")
		}
	}
	if len(sdef.PostGrantsSql) > 0 {
		if common.FileExists(post_grant_sql_file) {
			common.AppendStrings(sdef.PostGrantsSql, post_grant_sql_file, ";")
		} else {
			common.WriteStrings(sdef.PostGrantsSql, post_grant_sql_file, ";")
		}
	}
	//common.Run_cmd(sandbox_dir + "/start", []string{})
	if sdef.RunConcurrently {
		var eCommand2 = concurrent.ExecCommand{
			Cmd : sandbox_dir+"/start",
			Args : []string{},
		}
		exec_list = append(exec_list, concurrent.ExecutionList{2, eCommand2})
		if sdef.LoadGrants {
			var eCommand3 = concurrent.ExecCommand{
				Cmd : sandbox_dir+"/load_grants",
				Args : []string{"pre_grants.sql"},
			}
			var eCommand4 = concurrent.ExecCommand{
				Cmd : sandbox_dir+"/load_grants",
				Args : []string{},
			}
			var eCommand5 = concurrent.ExecCommand{
				Cmd : sandbox_dir+"/load_grants",
				Args : []string{"post_grants.sql"},
			}
			exec_list = append(exec_list, concurrent.ExecutionList{3, eCommand3})
			exec_list = append(exec_list, concurrent.ExecutionList{4, eCommand4})
			exec_list = append(exec_list, concurrent.ExecutionList{5, eCommand5})
		}

	} else {
		common.Run_cmd(sandbox_dir + "/start")
		if sdef.LoadGrants {
			common.Run_cmd_with_args(sandbox_dir+"/load_grants", []string{"pre_grants.sql"})
			common.Run_cmd(sandbox_dir + "/load_grants")
			common.Run_cmd_with_args(sandbox_dir+"/load_grants", []string{"post_grants.sql"})
		}
	}
	return
}

func write_script(temp_var TemplateCollection, name, template_name, directory string, data common.Smap, make_executable bool) {
	template := temp_var[template_name].Contents
	template = common.TrimmedLines(template)
	data["TemplateName"] = template_name
	text := common.Tprintf(template, data)
	if make_executable {
		write_exec(name, text, directory)
	} else {
		write_regular_file(name, text, directory)
	}
}

func write_exec(filename, text, directory string) {
	fname := write_regular_file(filename, text, directory)
	os.Chmod(fname, 0744)
}

func write_regular_file(filename, text, directory string) string {
	fname := directory + "/" + filename
	common.WriteString(text, fname)
	return fname
}
