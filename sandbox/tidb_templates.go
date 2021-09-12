// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2021 Giuseppe Maxia
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

//go:build go1.16
// +build go1.16

package sandbox

import (
	_ "embed"
	"fmt"
	"os"
	"regexp"

	"github.com/datacharmer/dbdeployer/globals"
)

var (
	//go:embed templates/tidb/init_db.gotxt
	tidbInitTemplate string
	//go:embed templates/tidb/my_cnf.gotxt
	tidbMyCnfTemplate string

	//go:embed templates/tidb/start.gotxt
	tidbStartTemplate string

	//go:embed templates/tidb/stop.gotxt
	tidbStopTemplate string

	//go:embed templates/tidb/send_kill.gotxt
	tidbSendKillTemplate string

	//go:embed templates/tidb/grants_5x.gotxt
	tidbGrantsTemplate string

	//go:embed templates/tidb/after_start.gotxt
	tidbAfterStartTemplate string
)

const tidbPrefix = "tidb_"

// Every template in this collection will replace the corresponding one in SingleTemplates
// when the flavor is "tidb"
var TidbTemplates = TemplateCollection{
	globals.TmplTidbInitDb: TemplateDesc{
		Description: "Initialization template for the TiDB server",
		Notes:       "This should normally run only once",
		Contents:    tidbInitTemplate,
	},
	globals.TmplTidbMyCnf: TemplateDesc{
		Description: "Default options file for a TiDB sandbox",
		Notes:       "",
		Contents:    tidbMyCnfTemplate,
	},
	globals.TmplTidbStart: TemplateDesc{
		Description: "Stops a database in a single TiDB sandbox",
		Notes:       "",
		Contents:    tidbStartTemplate,
	},
	globals.TmplTidbStop: TemplateDesc{
		Description: "Stops a database in a single TiDB sandbox",
		Notes:       "",
		Contents:    tidbStopTemplate,
	},
	globals.TmplTidbSendKill: TemplateDesc{
		Description: "Sends a kill signal to the TiDB database",
		Notes:       "",
		Contents:    tidbSendKillTemplate,
	},
	globals.TmplTidbGrants5x: TemplateDesc{
		Description: "Grants for TiDB sandboxes",
		Notes:       "",
		Contents:    tidbGrantsTemplate,
	},
	globals.TmplTidbAfterStart: TemplateDesc{
		Description: "commands to run after the database started",
		Notes:       "This script does nothing. You can change it and reuse through --use-template",
		Contents:    tidbAfterStartTemplate,
	},
}

func init() {
	// Makes sure that all template names in TidbTemplates start with 'tidb_'
	// This is an important assumption that will be used in sandbox.go
	// to replace templates for "tidb" flavor
	re := regexp.MustCompile(`^` + tidbPrefix)
	for name := range TidbTemplates {
		if !re.MatchString(name) {
			fmt.Printf("found template name '%s' that does not start with '%s'\n", name, tidbPrefix)
			os.Exit(1)
		}
	}
}
