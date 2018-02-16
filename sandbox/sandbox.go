package sandbox

import (
	"fmt"
	"github.com/datacharmer/dbdeployer/common"
	"os"
	"time"
	"regexp"
	"strconv"
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
	CustomMysqld	  string
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
	KeepAuthPlugin    bool
	KeepUuid          bool
	SinglePrimary     bool
	Force             bool
	ExposeDdTables 	  bool
}

const (
	MasterSlaveBasePort        int    = 10000
	GroupReplicationBasePort   int    = 12000
	GroupReplicationSPBasePort int    = 13000
	CircReplicationBasePort    int    = 14000
	MultipleBasePort           int    = 16000
	GroupPortDelta             int    = 125
	SandboxPrefix              string = "msb_"
	MasterSlavePrefix          string = "rsandbox_"
	GroupPrefix                string = "group_msb_"
	GroupSPPrefix              string = "group_sp_msb_"
	MultiplePrefix             string = "multi_msb_"
	ReplOptions                string = `
relay-log-index=mysql-relay
relay-log=mysql-relay
log-bin=mysql-bin
log-error=msandbox.err
`
	GtidOptions string = `
master-info-repository=table
relay-log-info-repository=table
gtid_mode=ON
log-slave-updates
enforce-gtid-consistency
`
)
var Expose_dd_tables []string= []string{
	"set persist debug='+d,skip_dd_table_access_check'",
	"set @col_type=(select c.type from mysql.columns c inner join mysql.tables t where t.id=table_id and t.name='tables' and c.name='hidden')",
	"set @visible=(if(@col_type = 'MYSQL_TYPE_ENUM', 'Visible', '0'))",
	"set @hidden=(if(@col_type = 'MYSQL_TYPE_ENUM', 'System', '1'))",
	"create table sys.dd_hidden_tables (id bigint unsigned, name varchar(64), schema_id bigint unsigned)",
	"insert into sys.dd_hidden_tables select id, name, schema_id from mysql.tables where hidden=@hidden",
	"update mysql.tables set hidden=@visible where hidden=@hidden and schema_id = 1",
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

func CheckPort(sandbox_type string, installed_ports []int, port int) {
	conflict := 0
	for _, p := range installed_ports {
		if p == port {
			conflict = p
		}
		if sandbox_type == "group-node" {
			if p == (port + GroupPortDelta) {
				conflict = p
			}
		}
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

func MakeNewServerUuid(sdef SandboxDef) string {
	node_num := sdef.NodeNum
	re_digit := regexp.MustCompile(`\d`)
	group1 := fmt.Sprintf("%08d", sdef.Port)
	group2 := fmt.Sprintf("%04d-%04d-%04d", 0, 0, 0)
	group3 := fmt.Sprintf("%012d", sdef.Port)
	//              12345678 1234 1234 1234 123456789012
	//    new_uuid="00000000-0000-0000-0000-000000000000"
	if node_num > 0 {
		group2 = re_digit.ReplaceAllString(group2, fmt.Sprintf("%d", node_num))
		group3 = re_digit.ReplaceAllString(group3, fmt.Sprintf("%d", node_num))
	}
	return fmt.Sprintf("server-uuid=%s-%s-%s", group1, group2, group3)
}

func FixServerUuid(sdef SandboxDef) {
	if !GreaterOrEqualVersion(sdef.Version, []int{5, 6, 9}) {
		return
	}
	new_uuid := MakeNewServerUuid(sdef)
	operation_dir := sdef.SandboxDir + "/data"
	uuid_file := operation_dir + "/auto.cnf"
	if !common.DirExists(operation_dir) {
		fmt.Printf("Directory %s does not exist\n", operation_dir)
		os.Exit(1)
	}
	new_contents := []string{"[auto]", new_uuid}
	err := common.WriteStrings(new_contents, uuid_file, "")
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	//check_uuid := common.SlurpAsString(uuid_file)
	//fmt.Printf("UUID file (%s) updated : %s\n", uuid_file, new_uuid)
	//fmt.Printf("new UUID : %s\n", check_uuid)
}

func VersionToList(version string) []int {
	// A valid version must be made of 3 integers
	re1 := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)$`)
	// Also valid version is 3 numbers with a prefix
	re2 := regexp.MustCompile(`^[^.0-9-]+(\d+)\.(\d+)\.(\d+)$`)
	verList1 := re1.FindAllStringSubmatch(version, -1)
	verList2 := re2.FindAllStringSubmatch(version, -1)
	verList := verList1
	//fmt.Printf("%#v\n", verList)
	if verList == nil {
		verList = verList2
	}
	if verList == nil {
		fmt.Println("Required version format: x.x.xx")
		return []int{-1}
		//os.Exit(1)
	}

	major, err1 := strconv.Atoi(verList[0][1])
	minor, err2 := strconv.Atoi(verList[0][2])
	rev, err3 := strconv.Atoi(verList[0][3])
	if err1 != nil || err2 != nil || err3 != nil {
		return []int{-1}
	}
	return []int{major, minor, rev}
}

func VersionToName(version string) string {
	re := regexp.MustCompile(`\.`)
	name := re.ReplaceAllString(version, "_")
	return name
}

func VersionToPort(version string) int {
	verList := VersionToList(version)
	major := verList[0]
	if major < 0 {
		return -1
	}
	minor := verList[1]
	rev := verList[2]
	//if major < 0 || minor < 0 || rev < 0 {
	//	return -1
	//}
	completeVersion := fmt.Sprintf("%d%d%02d", major, minor, rev)
	// fmt.Println(completeVersion)
	i, err := strconv.Atoi(completeVersion)
	if err == nil {
		return i
	}
	return -1
}

func GreaterOrEqualVersion(version string, compared_to []int) bool {
	var cmajor, cminor, crev int = compared_to[0], compared_to[1], compared_to[2]
	verList := VersionToList(version)
	major := verList[0]
	if major < 0 {
		return false
	}
	minor := verList[1]
	rev := verList[2]

	if major == 10 {
		return false
	}
	sversion := fmt.Sprintf("%02d%02d%02d", major, minor, rev)
	scompare := fmt.Sprintf("%02d%02d%02d", cmajor, cminor, crev)
	// fmt.Printf("<%s><%s>\n", sversion, scompare)
	return sversion >= scompare
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

func CreateSingleSandbox(sdef SandboxDef, origin string) {

	var sandbox_dir string
	sdef.Basedir = sdef.Basedir + "/" + sdef.Version
	if !common.DirExists(sdef.Basedir) {
		fmt.Printf("Base directory %s does not exist\n", sdef.Basedir)
		os.Exit(1)
	}

	//fmt.Printf("origin: %s\n", origin)
	//fmt.Printf("def: %#v\n", sdef)
	// port = VersionToPort(sdef.Version)
	version_fname := VersionToName(sdef.Version)
	if sdef.Prompt == "" {
		sdef.Prompt = "mysql"
	}
	if sdef.DirName == "" {
		sdef.DirName = SandboxPrefix + version_fname
	}
	sandbox_dir = sdef.SandboxDir + "/" + sdef.DirName
	//sandbox_home := sdef.SandboxDir
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
		if !GreaterOrEqualVersion(sdef.Version, []int{8, 0, 0}) {
			fmt.Printf("--expose-dd-tables requires MySQL 8.0.0+\n")
			os.Exit(1)
		}
		for _, line := range Expose_dd_tables {
			sdef.PostGrantsSql = append(sdef.PostGrantsSql, line)
		}
		if sdef.CustomMysqld !="" && sdef.CustomMysqld != "mysqld-debug" {
			fmt.Printf("--expose-dd-tables requires mysqld-debug. A different file was indicated (--custom-mysqld=%s)\n", sdef.CustomMysqld)
			fmt.Println("Either use \"mysqld-debug\" or remove --custom-mysqld")
			os.Exit(1)
		}
		sdef.CustomMysqld="mysqld-debug"
	}
	if sdef.CustomMysqld != "" {
		custom_mysqld := sdef.Basedir + "/" + sdef.CustomMysqld
		if !common.ExecExists(custom_mysqld) {
			fmt.Printf("File %s not found or not executable\n",custom_mysqld)
			fmt.Printf("The file \"%s\" (defined with --custom-mysqld) must be in the same directory as the regular mysqld\n", sdef.CustomMysqld)
			os.Exit(1)
		}
	}
	if GreaterOrEqualVersion(sdef.Version, []int{8, 0, 4}) {
		if sdef.KeepAuthPlugin == false {
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
	timestamp := time.Now()
	var data common.Smap = common.Smap{"Basedir": sdef.Basedir,
		"Copyright":    Copyright,
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
	err := os.Mkdir(sandbox_dir, 0755)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// fmt.Printf("creating: %s\n", datadir)
	err = os.Mkdir(datadir, 0755)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// fmt.Printf("creating: %s\n", tmpdir)
	err = os.Mkdir(tmpdir, 0755)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	script := sdef.Basedir + "/scripts/mysql_install_db"
	var cmd_list []string
	cmd_list = append(cmd_list, "--no-defaults")
	cmd_list = append(cmd_list, "--user="+os.Getenv("USER"))
	cmd_list = append(cmd_list, "--basedir="+sdef.Basedir)
	cmd_list = append(cmd_list, "--datadir="+datadir)
	cmd_list = append(cmd_list, "--tmpdir="+sandbox_dir+"/tmp")
	if GreaterOrEqualVersion(sdef.Version, []int{5, 7, 0}) {
		script = sdef.Basedir + "/bin/mysqld"
		cmd_list = append(cmd_list, "--initialize-insecure")
	}
	// fmt.Printf("Script: %s\n", script)
	if !common.ExecExists(script) {
		fmt.Printf("Script '%s' not found\n", script)
		os.Exit(1)
	}
	if len(sdef.InitOptions) > 0 {
		for _, op := range sdef.InitOptions {
			cmd_list = append(cmd_list, op)
		}
	}
	script_text := script
	for _, item := range cmd_list {
		script_text += " \\\n\t" + item
	}
	// fmt.Printf("using basedir: %s\n", sdef.Basedir)
	// fmt.Printf("%v\n", cmd_list)
	data["InitScript"] = script_text
	write_script(SingleTemplates, "init_db", "init_db_template", sandbox_dir, data, true)
	//cmd := exec.Command(script, cmd_list...)
	//var out bytes.Buffer
	//var stderr bytes.Buffer
	//cmd.Stdout = &out
	//cmd.Stderr = &stderr
	//err = cmd.Run()
	err, _ = common.Run_cmd_ctrl(sandbox_dir+"/init_db", true)
	if err == nil {
		fmt.Printf("Database installed in %s\n", sandbox_dir)
		if !sdef.Multi {
			fmt.Printf("run 'dbdeployer usage single' for basic instructions'\n")
		}
	} else {
		fmt.Printf("err: %s\n", err)
		// fmt.Printf("cmd: %#v\n", cmd)
		// fmt.Printf("stdout: %s\n", out.String())
		// fmt.Printf("stderr: %s\n", stderr.String())
	}

	if sdef.SBType == "" {
		sdef.SBType = "single"
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
		}
	}
	common.WriteSandboxDescription(sandbox_dir, sb_desc)
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
	case GreaterOrEqualVersion(sdef.Version, []int{8, 0, 0}):
		write_script(SingleTemplates, "grants.mysql", "grants_template8x", sandbox_dir, data, false)
	case GreaterOrEqualVersion(sdef.Version, []int{5, 7, 6}):
		write_script(SingleTemplates, "grants.mysql", "grants_template57", sandbox_dir, data, false)
	default:
		write_script(SingleTemplates, "grants.mysql", "grants_template5x", sandbox_dir, data, false)
	}
	write_script(SingleTemplates, "sb_include", "sb_include_template", sandbox_dir, data, false)

	if !sdef.KeepUuid {
		FixServerUuid(sdef)
	}

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
	common.Run_cmd(sandbox_dir + "/start")
	if sdef.LoadGrants {
		common.Run_cmd_with_args(sandbox_dir+"/load_grants", []string{"pre_grants.sql"})
		common.Run_cmd(sandbox_dir + "/load_grants")
		common.Run_cmd_with_args(sandbox_dir+"/load_grants", []string{"post_grants.sql"})
	}
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
