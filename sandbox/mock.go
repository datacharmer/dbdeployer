// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2020 Giuseppe Maxia
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
	"os"
	"path"
	"runtime"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/pkg/errors"
)

// Code in this module creates a fake directory structure that allows
// the testing of sandboxes without having MySQL packages.

const (
	DefaultMockDir       = "mock_dir"
	noOpMockTemplateName = globals.TmplNoOpMock
)

var (
	saveHome          string
	saveSandboxHome   string
	saveSandboxBinary string
	mockSandboxBinary string
	mockSandboxHome   string
)

func SetMockEnvironment(mockUpperDir string) error {
	if mockUpperDir == "" {
		mockUpperDir = DefaultMockDir
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
	_ = os.Setenv("HOME", home)
	_ = os.Setenv("SANDBOX_HOME", mockSandboxHome)
	_ = os.Setenv("SANDBOX_BINARY", mockSandboxBinary)
	_ = os.Setenv("HOME", home)
	_ = os.Setenv("SLEEP_TIME", "0")
	_ = os.Setenv("SB_MOCKING", "1")
	defaults.ResetDefaults()
	defaults.ConfigurationDir = path.Join(home, defaults.ConfigurationDirName)
	defaults.ConfigurationFile = path.Join(home, defaults.ConfigurationDirName, defaults.ConfigurationFileName)
	defaults.SandboxRegistry = path.Join(home, defaults.ConfigurationDirName, defaults.SandboxRegistryName)
	return nil
}

func RemoveMockEnvironment(mockUpperDir string) error {
	if !common.DirExists(mockUpperDir) {
		return fmt.Errorf("mock directory %s doesn't exist. Aborting", mockUpperDir)
	}
	err := os.RemoveAll(mockUpperDir)
	if err != nil {
		return err
	}
	_ = os.Setenv("HOME", saveHome)
	_ = os.Setenv("SANDBOX_HOME", saveSandboxHome)
	_ = os.Setenv("SANDBOX_BINARY", saveSandboxBinary)
	_ = os.Setenv("SLEEP_TIME", "")
	defaults.ResetDefaults()
	return nil
}

type MockFileSet struct {
	dir     string
	fileSet []ScriptDef
}

func CreateCustomMockVersion(version string, fileSet []MockFileSet) error {
	if mockSandboxBinary == "" {
		return fmt.Errorf("mock directory not set yet. - Call SetMockEnvironment() first")
	}
	logger, _, err := defaults.NewLogger(common.LogDirName(), "mock")
	if err != nil {
		return err
	}
	if !globals.MockTemplatesFilled {
		err = FillMockTemplates()
		if err != nil {
			return fmt.Errorf("error filling mock templates: %s", err)
		}
		globals.MockTemplatesFilled = true
	}

	var emptyData = make(common.StringMap)
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
			{globals.FnMysqldSafe, globals.TmplMysqldSafeMock, true},
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

func CreateMockVersion(version string) error {
	var fileSet = MySQLMockSet(false)
	return CreateCustomMockVersion(version, fileSet)
}

func init() {
	saveSandboxBinary = os.Getenv("SANDBOX_BINARY")
	saveSandboxHome = os.Getenv("SANDBOX_HOME")
	saveHome = os.Getenv("HOME")
}
