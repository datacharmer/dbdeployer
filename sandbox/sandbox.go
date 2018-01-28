package sandbox

import (
	"bytes"
	"github.com/datacharmer/dbdeployer/common"
	"fmt"
	//_ "github.com/spf13/cobra"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type SandboxDef struct {
	DirName      string
	Variant      string
	Version      string
	Basedir      string
	SandboxDir   string
	LoadGrants   bool
	Port         int
	DbUser       string
	RplUser      string
	DbPassword   string
	RplPassword  string
	RemoteAccess string
	BindAddress  string
	ServerId     int
	ReplOptions  string
	GtidOptions  string
	InitOptions  []string
	MyCnfOptions []string
	KeepAuthPlugin bool
}

const (
	MasterSlaveBasePort      int    = 10000
	GroupReplicationBasePort int    = 12000
	CircReplicationBasePort  int    = 14000
	MultipleBasePort      	 int    = 16000
	SandboxPrefix		 	 string = "msb_"
	MasterSlavePrefix		 string = "rsandbox_"
	GroupPrefix		 		 string = "group_msb_"
	MultiplePrefix		 	 string = "multi_msb_"
	ReplOptions              string = `
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

var Origins = [...]string{
	"Tarball",
	"NumberedTarball",
	"BareVersion",
	"FullDir",
	"NoSuchOrigin",
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
		return []int{-1}
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
		options_list := strings.Split(v, " ")
		for _, op := range (options_list) {
			if len(op) > 0 {
				text += fmt.Sprintf("%s\n",op)
			}
		}
	}
	return text
}

func CreateSingleSandbox(sdef SandboxDef, origin string) {

	if common.FileExists(origin) && strings.HasSuffix(origin, ".tar.gz") {
		fmt.Println("If you want to use a tarball to create a sandbox,\n")
		fmt.Println("you should first use the 'unpack' command\n")
		os.Exit(1)
	}

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
	if sdef.DirName == "" {
		sdef.DirName = SandboxPrefix + version_fname
	}
	sandbox_dir = sdef.SandboxDir + "/" + sdef.DirName
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
	if GreaterOrEqualVersion(sdef.Version, []int{8, 0, 4}) {
		if sdef.KeepAuthPlugin == false {
			sdef.InitOptions = append(sdef.InitOptions, "--default_authentication_plugin=mysql_native_password")
			sdef.MyCnfOptions = append(sdef.MyCnfOptions, "default_authentication_plugin=mysql_native_password")
		}
	}
	//fmt.Printf("%#v\n", sdef)
	var data common.Smap = common.Smap{"Basedir": sdef.Basedir,
		"Copyright":    Copyright,
		"SandboxDir":   sandbox_dir,
		"Port":         sdef.Port,
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
		"ExtraOptions":  slice_to_text(sdef.MyCnfOptions),
	}
	if sdef.ServerId > 0 {
		data["ServerId"] = fmt.Sprintf("server-id=%d", sdef.ServerId)
	} else {
		data["ServerId"] = ""
	}
	if common.DirExists(sandbox_dir) {
		fmt.Printf("Directory %s already exists\n", sandbox_dir)
		os.Exit(1)
	}
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
	// fmt.Printf("using basedir: %s\n", sdef.Basedir)
	// fmt.Printf("%v\n", cmd_list)
	cmd := exec.Command(script, cmd_list...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err == nil {
		fmt.Printf("Database installed in %s\n", sandbox_dir)
	} else {
		fmt.Printf("err: %s\n", err)
		fmt.Printf("cmd: %#v\n", cmd)
		fmt.Printf("stdout: %s\n", out.String())
		fmt.Printf("stderr: %s\n", stderr.String())
	}

	write_script("start", start_template, sandbox_dir, data, true)
	write_script("status", status_template, sandbox_dir, data, true)
	write_script("stop", stop_template, sandbox_dir, data, true)
	write_script("clear", clear_template, sandbox_dir, data, true)
	write_script("use", use_template, sandbox_dir, data, true)
	write_script("send_kill", send_kill_template, sandbox_dir, data, true)
	write_script("restart", restart_template, sandbox_dir, data, true)
	write_script("load_grants", load_grants_template, sandbox_dir, data, true)
	write_script("add_option", add_option_template, sandbox_dir, data, true)
	write_script("my", my_template, sandbox_dir, data, true)
	write_script("show_binlog", show_binlog_template, sandbox_dir, data, true)
	write_script("show_relaylog", show_relaylog_template, sandbox_dir, data, true)

	write_script("my.sandbox.cnf", my_cnf_template, sandbox_dir, data, false)
	if GreaterOrEqualVersion(sdef.Version, []int{5, 7, 6}) {
		write_script("grants.mysql", grants_template57, sandbox_dir, data, false)
	} else {
		write_script("grants.mysql", grants_template5x, sandbox_dir, data, false)
	}
	write_script("sb_include", "", sandbox_dir, data, false)

	//run_cmd(sandbox_dir + "/start", []string{})
	run_cmd(sandbox_dir + "/start")
	if sdef.LoadGrants {
		run_cmd(sandbox_dir + "/load_grants")
	}
}

//func run_cmd( c string, args []string) {
func run_cmd(c string) {
	//cmd := exec.Command(c, args...)
	cmd := exec.Command(c, "")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("err: %s\n", err)
		fmt.Printf("cmd: %#v\n", cmd)
		fmt.Printf("stdout: %s\n", out.String())
		fmt.Printf("stderr: %s\n", stderr.String())
		os.Exit(1)
	} else {
		fmt.Printf("%s", out.String())
	}
}

func write_script(name, template, directory string, data common.Smap, make_executable bool) {
	template = common.TrimmedLines(template)
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
