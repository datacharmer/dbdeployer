// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2019 Giuseppe Maxia
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
package rest

import (
	"github.com/datacharmer/dbdeployer/compare"
	"os"
	"testing"
)

func TestDownloadFile(t *testing.T) {
	compare.SkipOnDemand("SKIP_REST_TEST", t)

	fileName := "mysql-4.1.22.tar.xz"
	url := FileUrl(fileName)
	err := DownloadFile(fileName, url)
	if err == nil {
		t.Logf("OK\n")
		_ = os.Remove(fileName)
	} else {
		t.Logf("### ERR %s\n", err)
		t.Fail()
	}

}

func TestGetRemoteIndex(t *testing.T) {
	compare.SkipOnDemand("SKIP_REST_TEST", t)
	index, err := GetRemoteIndex()
	if err == nil {
		t.Logf(" OK - %+v", index)
	} else {
		t.Logf("### ERR %s\n", err)
		t.Fail()
	}
}
