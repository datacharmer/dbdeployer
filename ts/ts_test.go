package ts

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestSandboxes(t *testing.T) {

	cmds := map[string]func(ts *testscript.TestScript, neg bool, args []string){
		// find_errors will check that the error log in a sandbox contains the string ERROR
		// invoke as "find_errors /path/to/sandbox"
		// The command can be negated, i.e. it will succeed if the log does not contain the string ERROR
		// "! find_errors /path/to/sandbox"
		"find_errors": findErrorsInLogFile,

		// check_file will check that a given list of files exists
		// invoke as "check_file /path/to/sandbox file1 [file2 [file3 [file4]]]"
		// The command can be negated, i.e. it will succeed if the given files do not exist
		// "! check_file /path/to/sandbox file1 [file2 [file3 [file4]]]"
		"check_file": checkFile,

		// sleep will pause execution for the required number of seconds
		// Invoke as "sleep 3"
		// If no number is passed, it pauses for 1 second
		"sleep": sleep,
	}
	t.Run("single", func(t *testing.T) {
		testscript.Run(t, testscript.Params{
			Dir:      "testdata/single",
			TestWork: true,
			Cmds:     cmds,
		})
	})
	t.Run("replication", func(t *testing.T) {
		testscript.Run(t, testscript.Params{
			Dir:      "testdata/replication",
			TestWork: true,
			Cmds:     cmds,
		})
	})
	t.Run("group", func(t *testing.T) {
		testscript.Run(t, testscript.Params{
			Dir:      "testdata/group",
			TestWork: true,
			Cmds:     cmds,
		})
	})
	t.Run("group_sp", func(t *testing.T) {
		testscript.Run(t, testscript.Params{
			Dir:      "testdata/group_sp",
			TestWork: true,
			Cmds:     cmds,
		})
	})
}

func TestMain(m *testing.M) {

	// TODO: initialize the environment so that it doesn't depend on manual setup
	// This function assumes that the versions below are already installed
	// A proper implementation will use "dbdeployer init" to create a fresh environment
	// and download the needed versions
	// Furthermore, the function should detect the latest version available for each subversion
	// and use that list instead of the one provided here.

	versions := []string{"5.0.96", "5.1.73", "5.5.53", "5.6.41", "5.7.30", "8.0.29"}
	for _, v := range versions {
		label := strings.Replace(v, ".", "_", -1)
		err := buildTests("templates", "testdata", label, map[string]string{
			"DbVersion": v,
			"DbPathVer": label,
			"Home":      os.Getenv("HOME"),
			"TmpDir":    "/tmp",
		})
		if err != nil {
			fmt.Printf("%s\n", err)
			os.Exit(1)
		}
	}
	exitCode := m.Run()
	if dirExists("testdata") {
		_ = os.RemoveAll("testdata")
	}
	os.Exit(exitCode)
}
