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
	importPrefix = "import_"

	//go:embed templates/import/import_no_op.gotxt
	noOpTemplate string

	//go:embed templates/import/import_use.gotxt
	importUseTemplate string

	//go:embed templates/import/import_mysqlsh.gotxt
	importMysqlshTemplate string

	//go:embed templates/import/import_status.gotxt
	importStatusTemplate string

	//go:embed templates/import/import_sb_include.gotxt
	importSbIncludeTemplate string

	//go:embed templates/import/import_my_cnf.gotxt
	ImportMyCnfTemplate string

	//go:embed templates/import/import_metadata.gotxt
	importMetadataTemplate string

	//go:embed templates/import/import_test_sb.gotxt
	importTestSbTemplate string
)

var ImportTemplates = TemplateCollection{
	globals.TmplImportInitDb: TemplateDesc{
		Description: "Initialization template for the imported server",
		Notes:       "no op",
		Contents:    noOpTemplate,
	},
	globals.TmplImportUse: TemplateDesc{
		Description: "use template for the imported server",
		Notes:       "",
		Contents:    importUseTemplate,
	},
	globals.TmplImportStart: TemplateDesc{
		Description: "start template for the imported server",
		Notes:       "no op",
		Contents:    noOpTemplate,
	},
	globals.TmplImportRestart: TemplateDesc{
		Description: "restart template for the imported server",
		Notes:       "no op",
		Contents:    noOpTemplate,
	},
	globals.TmplImportStatus: TemplateDesc{
		Description: "status template for the imported server",
		Notes:       "",
		Contents:    importStatusTemplate,
	},
	globals.TmplImportSendKill: TemplateDesc{
		Description: "send_kill template for the imported server",
		Notes:       "no op",
		Contents:    noOpTemplate,
	},
	globals.TmplImportStop: TemplateDesc{
		Description: "stop template for the imported server",
		Notes:       "no op",
		Contents:    noOpTemplate,
	},
	globals.TmplImportLoadGrants: TemplateDesc{
		Description: "load_grants template for the imported server",
		Notes:       "no op",
		Contents:    noOpTemplate,
	},
	globals.TmplImportClear: TemplateDesc{
		Description: "clear template for the imported server",
		Notes:       "no op",
		Contents:    noOpTemplate,
	},
	globals.TmplImportAddOption: TemplateDesc{
		Description: "add_option template for the imported server",
		Notes:       "no op",
		Contents:    noOpTemplate,
	},
	globals.TmplImportShowBinlog: TemplateDesc{
		Description: "show_binlog template for the imported server",
		Notes:       "no op",
		Contents:    noOpTemplate,
	},
	globals.TmplImportShowLog: TemplateDesc{
		Description: "show_log template for the imported server",
		Notes:       "no op",
		Contents:    noOpTemplate,
	},
	globals.TmplImportShowRelaylog: TemplateDesc{
		Description: "show_relaylog template for the imported server",
		Notes:       "no op",
		Contents:    noOpTemplate,
	},
	globals.TmplImportMysqlsh: TemplateDesc{
		Description: "Invokes the MySQL shell with an appropriate URI for imported server",
		Notes:       "",
		Contents:    importMysqlshTemplate,
	},
	globals.TmplImportMyCnf: TemplateDesc{
		Description: "configuration file for imported mysql client",
		Notes:       "",
		Contents:    ImportMyCnfTemplate,
	},
	globals.TmplImportTestSb: TemplateDesc{
		Description: "Tests basic imported sandbox functionality",
		Notes:       "",
		Contents:    importTestSbTemplate,
	},
	globals.TmplImportSbInclude: TemplateDesc{
		Description: "Common variables and routines for imported sandboxes scripts",
		Notes:       "",
		Contents:    importSbIncludeTemplate,
	},
	globals.TmplImportMetadata: TemplateDesc{
		Description: "Show data about the sandbox",
		Notes:       "",
		Contents:    importMetadataTemplate,
	},
}

func init() {
	// Makes sure that all template names in ImportTemplates start with 'import_'
	// This is an important assumption that will be used in sandbox.go
	// to replace templates for imported sandboxes
	re := regexp.MustCompile(`^` + importPrefix)
	for name := range ImportTemplates {
		if !re.MatchString(name) {
			fmt.Printf("found template name '%s' that does not start with '%s'\n", name, importPrefix)
			os.Exit(1)
		}
		if len(ImportTemplates[name].Contents) == 0 {
			// The template is empty
			fmt.Printf("Template '%s' is empty\n", name)
			os.Exit(1)
		}
	}
}
