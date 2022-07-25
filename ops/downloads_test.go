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
			if err != nil {
				t.Fatalf("error getting tarball for %s", tb.Name)
			}
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
				if err != nil {
					t.Fatalf("error validating version of %s", tb.Name)
				}
				isGreater, err := common.GreaterOrEqualVersion(tb.Version, latestVersionList)
				if err != nil {
					t.Fatalf("error comparing version (%s) of %s with %v", tb.Version, tb.Name, latestVersionList)
				}
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
			if err != nil {
				t.Fatalf("error  retrieving tarball for latest %s: %s", v, err)
			}
			downloaded := path.Join(curDir, tb.Name)
			if common.FileExists(downloaded) {
				_ = os.Remove(downloaded)
			} else {
				t.Fatalf("downloaded tarball %s not found", downloaded)
			}
		})
		t.Run("by-name-"+tb.Name, func(t *testing.T) {
			err := GetRemoteTarball(DownloadsOptions{
				TarballName: tb.Name,
				TarballUrl:  "",
				TarballOS:   tb.OperatingSystem,
			})
			if err != nil {
				t.Fatalf("error  retrieving tarball for %s: %s", tb.Name, err)
			}
			downloaded := path.Join(curDir, tb.Name)
			if common.FileExists(downloaded) {
				_ = os.Remove(downloaded)
			} else {
				t.Fatalf("downloaded tarball %s not found", downloaded)
			}
		})
		t.Run("by-URL-"+tb.Name, func(t *testing.T) {
			err := GetRemoteTarball(DownloadsOptions{
				TarballName: "",
				TarballUrl:  tb.Url,
			})
			if err != nil {
				t.Fatalf("error  retrieving tarball (by URL) for %s: %s", tb.Name, err)
			}
			downloaded := path.Join(curDir, tb.Name)
			if common.FileExists(downloaded) {
				_ = os.Remove(downloaded)
			} else {
				t.Fatalf("downloaded tarball %s not found", downloaded)
			}
		})
	}
}
