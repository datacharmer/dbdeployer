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
	"path/filepath"
	"runtime"
	"strings"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/downloads"
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/datacharmer/dbdeployer/rest"
	"github.com/dustin/go-humanize"
)

type DownloadsOptions struct {
	SandboxBinary     string
	TarballName       string
	TarballUrl        string
	TarballOS         string
	TarballArch       string
	Flavor            string
	TargetServer      string
	Prefix            string
	Version           string
	Newest            bool
	Minimal           bool
	GuessLatest       bool
	Unpack            bool
	Overwrite         bool
	DeleteAfterUnpack bool
	DryRun            bool
	IsShell           bool
	Quiet             bool
	Retries           int64
	VerbosityLevel    int
	ProgressStep      int64
}

func GetRemoteTarball(options DownloadsOptions) error {
	var tarball downloads.TarballDescription
	var downloadedTarball string
	var err error
	found := false

	if options.TarballUrl != "" {
		tarball, err = findRemoteTarballByNameOrUrl(options.TarballUrl, options.TarballOS)
		if err != nil {
			return err
		}
		found = true
	}
	if !found && options.TarballName != "" {
		tarball, err = findRemoteTarballByNameOrUrl(options.TarballName, options.TarballOS)
		if err != nil {
			return err
		}
		found = true
	}
	if !found && options.Version != "" {
		tarball, err = findRemoteTarballByVersion(options.Version, options.Flavor, options.TarballOS, options.TarballArch, options.Minimal, options.Newest, options.GuessLatest)
		if err != nil {
			return err
		}
		found = true
	}
	if !found {
		return fmt.Errorf("couldn't find tarball - no name, URL, or version was provided")
	}

	if options.DryRun {
		fmt.Printf("tarball found: %s\n", tarball.Name)
		fmt.Printf("would download: %s\n", tarball.Url)
	} else {
		fileName := tarball.Name
		absPath, err := common.AbsolutePath(fileName)
		if err != nil {
			return err
		}
		if options.DryRun {
			DisplayTarball(tarball)
			return nil
		}
		if common.FileExists(absPath) {
			return fmt.Errorf(globals.ErrFileAlreadyExists, absPath)
		}
		if !options.Quiet {
			fmt.Printf("Downloading %s\n", tarball.Name)
		}
		err = rest.DownloadFileWithRetry(absPath, tarball.Url, options.Quiet, options.ProgressStep, options.Retries)
		if err != nil {
			return fmt.Errorf("error getting remote file %s - %s", fileName, err)
		}
		downloadedTarball = absPath
		err = postDownloadOps(tarball, fileName, absPath)
		if err != nil {
			return err
		}
	}
	if options.Unpack {
		target := path.Join(options.SandboxBinary) // add target here
		if options.DryRun {
			fmt.Printf("would unpack tarball into: %s\n", target)
		} else {
			err = UnpackTarball(UnpackOptions{
				SandboxBinary: options.SandboxBinary,
				TarballName:   tarball.Name,
				TargetServer:  options.TargetServer,
				Version:       tarball.Version,
				Prefix:        options.Prefix,
				Flavor:        tarball.Flavor,
				Verbosity:     options.VerbosityLevel,
				IsShell:       options.IsShell,
				Overwrite:     options.Overwrite,
				DryRun:        options.DryRun,
			})
			if err != nil {
				return err
			}
		}
		if options.DeleteAfterUnpack {
			if options.DryRun {
				fmt.Printf("deleting tarball: %s\n", tarball.Name)
			}
			if downloadedTarball == "" {
				return fmt.Errorf("unhandled error. After unpack, the tarball to be deleted was not found")
			}
			err = os.Remove(downloadedTarball)
			if err != nil {
				return fmt.Errorf("error removing downloaded file %s - %s", downloadedTarball, err)
			}
		}
	}
	return nil
}

func postDownloadOps(tarball downloads.TarballDescription, fileName, absPath string) error {
	fmt.Printf("File %s downloaded\n", absPath)

	if tarball.Checksum == "" {
		fmt.Println("No checksum to compare")
	} else {
		err := downloads.CompareTarballChecksum(tarball, absPath)
		if err != nil {
			return fmt.Errorf("error comparing checksum for tarball %s - %s", fileName, err)
		}
		fmt.Println("Checksum matches")
	}
	warningMsg := getOSWarning(tarball)
	if warningMsg != "" {
		fmt.Println(globals.HashLine)
		fmt.Println(warningMsg)
		fmt.Println(globals.HashLine)
	}
	return nil
}

func getOSWarning(tarball downloads.TarballDescription) string {
	currentOS := strings.ToLower(runtime.GOOS)
	currentArch := strings.ToLower(runtime.GOARCH)
	tarballOS := strings.ToLower(tarball.OperatingSystem)
	tarballArch := strings.ToLower(tarball.Arch)
	if currentOS != tarballOS {
		return fmt.Sprintf("WARNING: Current OS is %s, but the tarball's OS is %s", currentOS, tarballOS)
	}
	if tarballArch != "" && currentArch != tarballArch {
		return fmt.Sprintf("WARNING: current architecture is %s, but the tarball architecture is %s", currentArch, tarballArch)
	}
	return ""
}

func findRemoteTarballByNameOrUrl(wanted, wantedOs string) (downloads.TarballDescription, error) {

	var tarball downloads.TarballDescription
	var err error

	if common.IsUrl(wanted) {
		tarball, err = downloads.FindTarballByUrl(wanted)
		if err == nil {
			return tarball, nil
		} else {
			name := filepath.Base(wanted)
			lowerName := strings.ToLower(name)
			flavor, version, shortVersion, err := common.FindTarballInfo(name)
			if err != nil {
				return downloads.TarballDescription{}, err
			}
			tarballOs := runtime.GOOS
			if wantedOs != "" {
				tarballOs = wantedOs
			} else {
				switch {
				case strings.Contains(lowerName, "linux"):
					tarballOs = "linux"
				case strings.Contains(lowerName, "macos") || strings.Contains(lowerName, "darwin"):
					tarballOs = "darwin"
				default:
					return downloads.TarballDescription{}, fmt.Errorf("unable to determine the operating system of tarball '%s'", name)
				}
			}
			tarball = downloads.TarballDescription{
				Name:            name,
				Url:             wanted,
				Flavor:          flavor,
				ShortVersion:    shortVersion,
				Version:         version,
				OperatingSystem: tarballOs,
			}
		}
	} else {
		tarball, err = downloads.FindTarballByName(wanted)
		if err != nil {
			return downloads.TarballDescription{}, err
		}
	}
	return tarball, nil
}

func findRemoteTarballByVersion(version, flavor, OS, arch string, minimal, newest, guessLatest bool) (downloads.TarballDescription, error) {

	if arch == "" {
		arch = runtime.GOARCH
	}
	if OS == "" {
		OS = runtime.GOOS
	}
	if flavor == "" {
		flavor = common.MySQLFlavor
	}
	var tarball downloads.TarballDescription
	var err error

	tarball, err = downloads.FindOrGuessTarballByVersionFlavorOS(version, flavor, OS, arch, minimal, newest, guessLatest)
	if err != nil {
		return downloads.TarballDescription{}, fmt.Errorf(fmt.Sprintf("Error getting version %s (%s-%s)[minimal: %v - newest: %v - guess: %v]: %s",
			version, flavor, OS, minimal, newest, guessLatest, err))
	}
	return tarball, nil
}

func DisplayTarball(tarball downloads.TarballDescription) {
	fmt.Printf("Name:          %s\n", tarball.Name)
	fmt.Printf("Short version: %s\n", tarball.ShortVersion)
	fmt.Printf("Version:       %s\n", tarball.Version)
	fmt.Printf("Flavor:        %s\n", tarball.Flavor)
	fmt.Printf("OS:            %s-%s\n", tarball.OperatingSystem, tarball.Arch)
	fmt.Printf("URL:           %s\n", tarball.Url)
	fmt.Printf("Checksum:      %s\n", tarball.Checksum)
	fmt.Printf("Size:          %s\n", humanize.Bytes(uint64(tarball.Size)))
	if tarball.Notes != "" {
		fmt.Printf("Notes:         %s\n", tarball.Notes)
	}
	if tarball.DateAdded != "" {
		fmt.Printf("Added on:      %s\n", tarball.DateAdded)
	}
}
