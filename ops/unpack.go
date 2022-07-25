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
	"regexp"
	"strings"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/datacharmer/dbdeployer/unpack"
)

type UnpackOptions struct {
	SandboxBinary string
	TarballName   string
	TargetServer  string
	Version       string
	Prefix        string
	Flavor        string
	Verbosity     int
	IsShell       bool
	Overwrite     bool
	DryRun        bool
}

func UnpackTarball(options UnpackOptions) error {
	Basedir := options.SandboxBinary
	verbosity := options.Verbosity
	if !common.DirExists(Basedir) {
		return fmt.Errorf(globals.ErrDirectoryNotFound, Basedir)
	}
	tarball := options.TarballName
	reVersion := regexp.MustCompile(`(\d+\.\d+\.\d+)`)
	verList := reVersion.FindAllStringSubmatch(tarball, -1)

	detectedVersion := ""
	if verList != nil {
		detectedVersion = verList[0][0]
	}
	// common.CondPrintf(">> %#v %s\n",verList, detected_version)

	isShell := options.IsShell
	target := options.TargetServer
	if !isShell && target != "" {
		return fmt.Errorf("unpack: Option --target-server can only be used with --shell")
	}

	overwrite := options.Overwrite
	flavor := options.Flavor
	dryRun := options.DryRun
	if flavor == "" {
		baseName := common.BaseName(tarball)
		flavor = common.DetectTarballFlavor(baseName)
		if flavor == "" {
			return fmt.Errorf("no flavor detected in %s. Please use --%s", tarball, globals.FlavorLabel)
		}
	}
	Version := options.Version
	if Version == "" {
		Version = detectedVersion
	}
	if Version == "" {
		return fmt.Errorf("unpack: No version was detected from tarball name. " +
			"Flag --unpack-version becomes mandatory")
	}
	// This call used to ensure that the port provided is in the right format
	_, err := common.VersionToPort(Version)
	if err != nil {
		return fmt.Errorf("version %s not in the required format", Version)
	}
	Prefix := options.Prefix
	if isShell {
		fmt.Printf("%s\n", Version)
		var canBeEmbedded bool
		canBeEmbedded, err = common.HasCapability(common.MySQLShellFlavor, common.EmbedMySQLShell, Version)
		if err != nil {
			return fmt.Errorf("error detecting shell capability: %s", err)
		}
		if !canBeEmbedded {
			return fmt.Errorf("MySQL shell version %s insufficient for embedding", Version)
		}
	}

	destination := path.Join(Basedir, Prefix+Version)
	if target != "" {
		destination = path.Join(Basedir, target)
	}
	if common.DirExists(destination) && !isShell {
		if overwrite {
			if dryRun {
				fmt.Printf("delete binaries %s %s\n", Basedir, Prefix+Version)
			} else {
				isDeleted, err := DeleteBinaries(Basedir, Prefix+Version, false)
				if !isDeleted {
					return fmt.Errorf("directory %s could not be removed", Prefix+Version)
				}
				if err != nil {
					return fmt.Errorf("error removing directory %s: %s", Prefix+Version, err)
				}
			}
		} else {
			return fmt.Errorf(globals.ErrNamedDirectoryAlreadyExists, "destination directory", destination)
		}
	}
	extracted := path.Base(tarball)
	var bareName string

	var extractFunc func(string, string, int) error
	var foundExtension string

	switch {
	case strings.HasSuffix(tarball, globals.TarGzExt):
		extractFunc = unpack.UnpackTar
		foundExtension = globals.TarGzExt
	case strings.HasSuffix(tarball, globals.TarXzExt):
		extractFunc = unpack.UnpackXzTar
		foundExtension = globals.TarXzExt
	default:
		return fmt.Errorf("tarball extension must be either '%s' or '%s'", globals.TarGzExt, globals.TarXzExt)
	}
	err = unpack.VerifyTarFile(tarball)
	if err != nil {
		return fmt.Errorf("validation for %s failed: %s", tarball, err)
	}
	bareName = extracted[0 : len(extracted)-len(globals.TarGzExt)]
	if isShell {
		common.CondPrintf("Merging shell tarball %s to %s\n", common.ReplaceLiteralHome(tarball), common.ReplaceLiteralHome(destination))
		if !dryRun {
			err := unpack.MergeShell(tarball, foundExtension, Basedir, destination, bareName, verbosity)
			if err != nil {
				return fmt.Errorf("error while unpacking mysql shell tarball : %s", err)
			}
		}
		return nil
	}

	common.CondPrintf("Unpacking tarball %s to %s\n", tarball, common.ReplaceLiteralHome(destination))
	if dryRun {
		return nil
	}

	err = extractFunc(tarball, Basedir, verbosity)
	if err != nil {
		return err
	}
	finalName := path.Join(Basedir, bareName)
	// If the directory was not created, it probably means that the tarball was not well organised
	// and either lacked the top directory or the top directory had a different name
	if !common.DirExists(finalName) {
		return fmt.Errorf("problem with tarball %s: directory %s was not created", tarball, finalName)
	}
	if finalName != destination {
		common.CondPrintf("Renaming directory %s to %s\n", finalName, destination)
		err = os.Rename(finalName, destination)
		if err != nil {
			return err
		}
	}
	err = common.WriteString(flavor, path.Join(destination, globals.FlavorFileName))
	if err != nil {
		return fmt.Errorf("error writing %s in %s: %s", globals.FlavorFileName, destination, err)
	}
	return nil
}
