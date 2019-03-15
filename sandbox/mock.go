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
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/pkg/errors"
	"os"
	"path"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"runtime"
)

// Code in this module creates a fake directory structure that allows
// the testing of sandboxes without having MySQL packages.

const (
	defaultMockDir       string = "mock_dir"
	noOpMockTemplateName        = "no_op_mock_template"
)

var (
	saveHome          string
	saveSandboxHome   string
	saveSandboxBinary string
	saveSleepTime     string
	mockSandboxBinary string
	mockSandboxHome   string
)

func setMockEnvironment(mockUpperDir string) error {
	if mockUpperDir == "" {
		mockUpperDir = defaultMockDir
	}
	if common.DirExists(mockUpperDir) {
		return fmt.Errorf("mock directory %s already exists. Aborting", mockUpperDir)
	}
	PWD := os.Getenv("PWD")
	home := fmt.Sprintf("%s/%s/home", PWD, mockUpperDir)
	sandboxBinaryUpper := fmt.Sprintf("%s/opt", home)
	mockSandboxBinary = fmt.Sprintf("%s/opt/mysql", home)
	mockSandboxHome = fmt.Sprintf("%s/sandboxes", home)
	for _, dir := range []string{mockUpperDir, home, sandboxBinaryUpper, mockSandboxBinary, mockSandboxHome} {
		err := os.Mkdir(dir, globals.PublicDirectoryAttr)
		if err != nil {
			return errors.Wrapf(err, "error creating directory %s", dir)
		}
	}
	saveHome = os.Getenv("HOME")
	saveSandboxBinary = os.Getenv("SANDBOX_BINARY")
	saveSandboxHome = os.Getenv("SANDBOX_HOME")
	os.Setenv("HOME", home)
	os.Setenv("SANDBOX_HOME", mockSandboxHome)
	os.Setenv("SANDBOX_BINARY", mockSandboxBinary)
	os.Setenv("HOME", home)
	os.Setenv("SLEEP_TIME", "0")
	defaults.ConfigurationDir = path.Join(home, defaults.ConfigurationDirName)
	defaults.ConfigurationFile = path.Join(home, defaults.ConfigurationDirName, defaults.ConfigurationFileName)
	defaults.SandboxRegistry = path.Join(home, defaults.ConfigurationDirName, defaults.SandboxRegistryName)
	defaults.SandboxRegistryLock = path.Join(home, defaults.ConfigurationDirName, defaults.SandboxRegistryLockName)
	return nil
}

func removeMockEnvironment(mockUpperDir string) error {
	if !common.DirExists(mockUpperDir) {
		return fmt.Errorf("mock directory %s doesn't exist. Aborting", mockUpperDir)
	}
	err := os.RemoveAll(mockUpperDir)
	if err != nil {
		return err
	}
	os.Setenv("HOME", saveHome)
	os.Setenv("SANDBOX_HOME", saveSandboxHome)
	os.Setenv("SANDBOX_BINARY", saveSandboxBinary)
	os.Setenv("SLEEP_TIME", "")
	return nil
}

type MockFileSet struct {
	dir     string
	fileSet []ScriptDef
}

func createCustomMockVersion(version string, fileSet []MockFileSet) error {
	if mockSandboxBinary == "" {
		return fmt.Errorf("mock directory not set yet. - Call setMockEnvironment() first")
	}
	logger, _, err := defaults.NewLogger(common.LogDirName(), "mock")
	if err != nil {
		return err
	}

	var emptyData = common.StringMap{}
	versionDir := path.Join(mockSandboxBinary, version)
	err = os.Mkdir(versionDir, globals.PublicDirectoryAttr)
	if err != nil {
		return errors.Wrapf(err, globals.ErrCreatingDirectory, versionDir, err)
	}
	for _, fs := range fileSet {
		dir := path.Join(versionDir, fs.dir)
		err := os.Mkdir(dir, globals.PublicDirectoryAttr)
		if err != nil {
			return errors.Wrapf(err, globals.ErrCreatingDirectory, dir, err)
		}
		err = writeScripts(ScriptBatch{
			tc:         MockTemplates,
			data:       emptyData,
			sandboxDir: dir,
			logger:     logger,
			scripts:    fs.fileSet,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func MySQLMockSet(debug bool) []MockFileSet {
	currentOs := runtime.GOOS
	extension := ""
	switch currentOs {
	case "linux":
		extension = "so"
	case "darwin":
		extension = "dylib"
	default:
		common.Exitf(1, "unhandled operating system %s", currentOs)
	}
	var fileSet []MockFileSet
	libmysqlclientFileName := fmt.Sprintf("libmysqlclient.%s", extension)

	libFileSet := MockFileSet{
		"lib",
		[]ScriptDef{
			{libmysqlclientFileName, noOpMockTemplateName, false},
		},
	}
	mysqld := globals.FnMysqld
	if debug {
		mysqld = globals.FnMysqldDebug
	}
	binFileSet := MockFileSet{
		"bin",
		[]ScriptDef{
			{mysqld, noOpMockTemplateName, true},
			{globals.FnMysql, noOpMockTemplateName, true},
			{globals.FnMysqldSafe, "mysqld_safe_mock_template", true},
		},
	}
	scriptsFileSet := MockFileSet{
		"scripts",
		[]ScriptDef{
			{globals.FnMysqlInstallDb, noOpMockTemplateName, true},
		},
	}
	fileSet = append(fileSet, binFileSet)
	fileSet = append(fileSet, libFileSet)
	fileSet = append(fileSet, scriptsFileSet)
	return fileSet
}

func createMockVersion(version string) error {
	var fileSet = MySQLMockSet(false)
	return createCustomMockVersion(version, fileSet)
}

func init() {
	saveSandboxBinary = os.Getenv("SANDBOX_BINARY")
	saveSandboxHome = os.Getenv("SANDBOX_HOME")
	saveHome = os.Getenv("HOME")
	saveSleepTime = os.Getenv("SLEEP_TIME")
}
