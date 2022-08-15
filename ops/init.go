// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2022 Giuseppe Maxia
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

package ops

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/downloads"
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/datacharmer/dbdeployer/rest"
)

type InitOptions struct {
	SandboxBinary       string
	SandboxHome         string
	DryRun              bool
	SkipDownloads       bool
	SkipTarballDownload bool
	SkipCompletion      bool
}

func verifyInitOptions(options InitOptions) error {
	if options.SandboxBinary == "" {
		return fmt.Errorf("initOptions.SandboxBinary not set")
	}
	if options.SandboxHome == "" {
		return fmt.Errorf("initOptions.SandboxHome not set")
	}
	return nil
}

func InitEnvironment(options InitOptions) error {
	err := verifyInitOptions(options)
	if err != nil {
		return err
	}
	sandboxBinary := options.SandboxBinary
	sandboxHome := options.SandboxHome
	dryRun := options.DryRun
	skipDownloads := options.SkipDownloads
	skipCompletion := options.SkipCompletion

	sandboxHome, err = common.AbsolutePath(sandboxHome)
	if err != nil {
		return fmt.Errorf("error getting absolute path of %s: %s", sandboxHome, err)
	}
	sandboxBinary, err = common.AbsolutePath(sandboxBinary)
	if err != nil {
		return fmt.Errorf("error getting absolute path of %s: %s", sandboxBinary, err)
	}
	if sandboxBinary == sandboxHome {
		return fmt.Errorf("sandbox-binary and sandbox-home cannot be the same directory (%s)", sandboxHome)
	}
	fmt.Printf("SANDBOX_BINARY %s\n", sandboxBinary)
	fmt.Printf("SANDBOX_HOME   %s\n", sandboxHome)

	for _, ext := range []string{globals.TarExt, globals.TarGzExt, globals.TarXzExt, globals.ZipExt} {
		for _, dir := range []string{sandboxBinary, sandboxHome} {
			if common.Matches(dir, `\`+ext+`$`) {
				return fmt.Errorf("`SANDBOX_HOME and SANDBOX_BINARY cannot have extension %s", ext)
			}
		}
	}
	if common.FileExists(sandboxBinary) && !common.DirExists(sandboxBinary) {
		return fmt.Errorf("a file with the name %s exists already. Can't create a directory with such name", sandboxBinary)
	}
	if common.FileExists(sandboxHome) && !common.DirExists(sandboxHome) {
		return fmt.Errorf("a file with the name %s exists already. Can't create a directory with such name", sandboxHome)
	}

	creationLabel := "was created"
	updateLabel := "Updating"
	if dryRun {
		creationLabel = "would be created"
		updateLabel = "Would update"
	}
	needDownload := false
	fmt.Printf("\n%s\n", globals.DashLine)
	if common.DirExists(sandboxBinary) {
		fmt.Printf("Directory %s ($SANDBOX_BINARY) already exists\n", sandboxBinary)
		files, err := os.ReadDir(sandboxBinary)
		if err != nil {
			return fmt.Errorf("error reading sandbox binary directory %s: %s", sandboxBinary, err)
		}
		// Sandbox binary directory exists.
		// We now check whether there is any expanded tarball directory
		numSandboxes := 0
		for _, f := range files {
			if f.IsDir() {
				bin := path.Join(sandboxBinary, f.Name(), "bin")
				if common.DirExists(bin) {
					numSandboxes++
				}
			}
		}
		if numSandboxes == 0 {
			needDownload = true
		}

	} else {

		if !dryRun {
			err = os.MkdirAll(sandboxBinary, globals.PublicDirectoryAttr)
			if err != nil {
				return fmt.Errorf("error creating sandbox binary directory %s: %s", sandboxBinary, err)
			}
			needDownload = true
		}
		fmt.Printf("Directory %s ($SANDBOX_BINARY) %s\n", sandboxBinary, creationLabel)
	}
	fmt.Println("This directory is the destination for expanded tarballs")

	fmt.Printf("\n%s\n", globals.DashLine)
	if common.DirExists(sandboxHome) {
		fmt.Printf("Directory %s ($SANDBOX_HOME) already exists \n", sandboxHome)
	} else {
		if !dryRun {

			err = os.MkdirAll(sandboxHome, globals.PublicDirectoryAttr)
			if err != nil {
				return fmt.Errorf("error creating sandbox home directory %s: %s", sandboxHome, err)
			}
		}
		fmt.Printf("Directory %s ($SANDBOX_HOME) %s\n", sandboxHome, creationLabel)
	}
	fmt.Println("This directory is the destination for deployed sandboxes")

	if sandboxBinary != defaults.Defaults().SandboxBinary {
		fmt.Printf("\n%s\n", globals.DashLine)
		fmt.Printf("%s defaults for directory %s ($SANDBOX_BINARY)\n", updateLabel, sandboxBinary)
		fmt.Printf("# dbdeployer defaults update %s %s \n", globals.SandboxBinaryLabel, sandboxBinary)
		if !dryRun {
			defaults.UpdateDefaults(globals.SandboxBinaryLabel, sandboxBinary, true)
		}
	}

	if sandboxHome != defaults.Defaults().SandboxHome {
		fmt.Printf("\n%s\n", globals.DashLine)
		fmt.Printf("%s defaults for directory %s ($SANDBOX_HOME)\n", updateLabel, sandboxHome)
		fmt.Printf("# dbdeployer defaults update %s %s \n", globals.SandboxHomeLabel, sandboxHome)
		if !dryRun {
			defaults.UpdateDefaults(globals.SandboxHomeLabel, sandboxHome, true)
		}
	}
	fmt.Println()

	if needDownload {
		err = initDownloadTarball(options)
		if err != nil {
			return err
		}
	}
	if !common.DirExists(defaults.ConfigurationDir) {
		err = os.Mkdir(defaults.ConfigurationDir, globals.PublicDirectoryAttr)
		if err != nil {
			return err
		}
	}
	if skipCompletion {
		return nil
	}
	fmt.Println(globals.DashLine)
	fmt.Println("# dbdeployer defaults enable-bash-completion --run-it --remote")
	if needDownload || !(skipDownloads || dryRun) {
		err = ProcessBashCompletionEnabling(true, true, defaults.Defaults().RemoteCompletionUrl, "")
	}
	return err
}

func initDownloadTarball(options InitOptions) error {

	version := "8.0"
	OS := runtime.GOOS
	minimal := false
	if strings.EqualFold(OS, "linux") {
		minimal = true
	}
	arch := runtime.GOARCH
	tarball, err := downloads.FindOrGuessTarballByVersionFlavorOS(version, common.MySQLFlavor,
		OS, arch, minimal, true, false)
	if err != nil {
		return fmt.Errorf("error getting version %s (%s-%s-%s)[minimal: %v - newest: %v - guess: %v]: %s",
			version, common.MySQLFlavor, OS, arch, minimal, true, false, err)
	}
	fmt.Println(globals.DashLine)
	fmt.Printf("# dbdeployer downloads get %s\n", tarball.Name)
	if !(options.DryRun || options.SkipDownloads || options.SkipTarballDownload) {

		err = rest.DownloadFile(tarball.Name, tarball.Url, true, globals.TenMB)
		if err != nil {
			return fmt.Errorf("error downloading file %s", tarball.Name)
		}
	}
	fmt.Println(globals.DashLine)
	fmt.Printf("# dbdeployer unpack %s\n", tarball.Name)
	if !(options.DryRun || options.SkipDownloads || options.SkipTarballDownload) {
		err = UnpackTarball(UnpackOptions{
			SandboxBinary: options.SandboxBinary,
			TarballName:   tarball.Name,
		})
		if err != nil {
			return err
		}
		fmt.Println(globals.DashLine)
		fmt.Println("dbdeployer versions")
		err = ShowVersions(VersionOptions{
			SandboxBinary: options.SandboxBinary,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
