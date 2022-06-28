package ts

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestSandboxes(t *testing.T) {

	// Directories in testdata are created by the setup code in TestMain
	dirs, err := filepath.Glob("testdata/*")
	if err != nil {
		t.Skip("no directories found in testdata")
	}
	for _, dir := range dirs {
		t.Run(path.Base(dir), func(t *testing.T) {
			testscript.Run(t, testscript.Params{
				Dir:      dir,
				Setup:    setupEnv,
				TestWork: true,
				Cmds:     defineCommands(),
			})
		})
	}

}

func setupEnv(env *testscript.Env) error {

	return nil
}

func TestMain(m *testing.M) {

	// TODO: initialize the environment so that it doesn't depend on manual setup
	// This function assumes that the versions below are already installed
	// A proper implementation will use "dbdeployer init" to create a fresh environment
	// and download the needed versions
	// Furthermore, the function should detect the latest version available for each subversion
	// and use that list instead of the one provided here.

	versions := []string{"5.0.96", "5.1.73", "5.5.53", "5.6.41", "5.7.30", "8.0.29"}
	//versions := []string{"5.0.96"}
	/*
		fakeHome, err := filepath.Abs("test_home")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if !dirExists(fakeHome) {
			err = os.Mkdir(fakeHome, 0755)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
		err = os.Setenv("SANDBOX_HOME", path.Join(fakeHome, "sandboxes"))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = os.Setenv("SANDBOX_BINARY", path.Join(fakeHome, "opt", "mysql"))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

	*/

	for _, v := range versions {
		label := strings.Replace(v, ".", "_", -1)
		err := buildTests("templates", "testdata", label, map[string]string{
			"DbVersion": v,
			"DbPathVer": label,
			"Home":      os.Getenv("HOME"),
			"TmpDir":    "/tmp",
		})
		if err != nil {
			fmt.Printf("error creating the tests for %s :%s\n", label, err)
			os.Exit(1)
		}
	}
	exitCode := m.Run()
	if dirExists("testdata") {
		_ = os.RemoveAll("testdata")
	}
	os.Exit(exitCode)
}
