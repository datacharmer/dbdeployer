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

package ts_static

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/datacharmer/dbdeployer/cmd"
	"github.com/datacharmer/dbdeployer/common"
	"github.com/rogpeppe/go-internal/testscript"
)

func TestStaticScripts(t *testing.T) {
	t.Parallel()
	testscript.Run(t, testscript.Params{
		Dir:                 "testdata",
		RequireExplicitExec: true,
	})
}

func TestMain(m *testing.M) {
	var srv *http.Server
	fileServerDone := &sync.WaitGroup{}

	identity := common.BaseName(os.Args[0])
	inMain := strings.Contains(identity, "ts_static.test")
	if inMain {
		fileServerDone.Add(1)
		srv = startFileServer(fileServerDone, "./downloads-arch", 9000)
	}
	exitCode := testscript.RunMain(m, map[string]func() int{
		"dbdeployer": cmd.Execute,
	})
	if inMain && srv != nil && srv.Addr != "" {
		if err := srv.Shutdown(context.Background()); err != nil {
			fmt.Printf("error stopping file server: %s\n", err)
			os.Exit(1)
		}
		fileServerDone.Wait()
	}
	os.Exit(exitCode)
}

func startFileServer(wg *sync.WaitGroup, directory string, port int) *http.Server {
	fs := http.FileServer(http.Dir(directory))
	srv := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: fs}

	go func() {
		defer wg.Done()
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	return srv
}
