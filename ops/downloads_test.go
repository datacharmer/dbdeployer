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
	"os"
	"path"
	"runtime"
	"strings"
	"testing"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/downloads"
	qt "github.com/frankban/quicktest"
)

func TestGetTarball(t *testing.T) {

	sbEnv := os.Getenv("SANDBOX_BINARY")
	sandboxBinary := defaults.Defaults().SandboxBinary

	if sbEnv != "" {
		sandboxBinary = sbEnv
	}

	var newestList = make(map[string]downloads.TarballDescription)

	for _, tb := range downloads.DefaultTarballRegistry.Tarballs {
		t.Run(tb.Name, func(t *testing.T) {
			c := qt.New(t)
			retrieved, err := findRemoteTarballByNameOrUrl(tb.Name, tb.OperatingSystem)
			c.Assert(err, qt.IsNil)
			c.Assert(retrieved, qt.DeepEquals, tb)
			if !strings.EqualFold(tb.OperatingSystem, runtime.GOOS) {
				return
			}
			if !strings.EqualFold(tb.Flavor, common.MySQLFlavor) {
				return
			}
			if strings.EqualFold(tb.OperatingSystem, "linux") && !tb.Minimal {
				return
			}
			latestTb, ok := newestList[tb.ShortVersion]
			if ok {
				latestVersionList, err := common.VersionToList(latestTb.Version)
				c.Assert(err, qt.IsNil)
				isGreater, err := common.GreaterOrEqualVersion(tb.Version, latestVersionList)
				c.Assert(err, qt.IsNil)
				if isGreater {
					newestList[tb.ShortVersion] = tb
				}
			} else {
				newestList[tb.ShortVersion] = tb
			}
		})
	}
	curDir := os.Getenv("PWD")
	for v, tb := range newestList {
		c := qt.New(t)
		if tb.Flavor != common.MySQLFlavor {
			continue
		}
		if !strings.EqualFold(tb.OperatingSystem, runtime.GOOS) {
			continue
		}
		t.Run("latest "+v, func(t *testing.T) {
			err := GetRemoteTarball(DownloadsOptions{
				SandboxBinary: sandboxBinary,
				TarballOS:     tb.OperatingSystem,
				Flavor:        tb.Flavor,
				Version:       v,
				Newest:        true,
				Minimal:       strings.EqualFold(tb.OperatingSystem, "linux"),
			})

			c.Assert(err, qt.IsNil)
			downloaded := path.Join(curDir, tb.Name)
			if common.FileExists(downloaded) {
				_ = os.Remove(downloaded)
			} else {
				t.Fatalf("downloaded tarball %s not found", downloaded)
			}
		})
		t.Run("by-name-"+tb.Name, func(t *testing.T) {
			c := qt.New(t)
			err := GetRemoteTarball(DownloadsOptions{
				TarballName: tb.Name,
				TarballUrl:  "",
				TarballOS:   tb.OperatingSystem,
			})
			c.Assert(err, qt.IsNil)
			downloaded := path.Join(curDir, tb.Name)
			if common.FileExists(downloaded) {
				_ = os.Remove(downloaded)
			} else {
				t.Fatalf("downloaded tarball %s not found", downloaded)
			}
		})
		t.Run("by-URL-"+tb.Name, func(t *testing.T) {
			c := qt.New(t)
			err := GetRemoteTarball(DownloadsOptions{
				TarballName: "",
				TarballUrl:  tb.Url,
			})
			c.Assert(err, qt.IsNil)
			downloaded := path.Join(curDir, tb.Name)
			if common.FileExists(downloaded) {
				_ = os.Remove(downloaded)
			} else {
				t.Fatalf("downloaded tarball %s not found", downloaded)
			}
		})
	}
}
