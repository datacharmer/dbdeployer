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

	"github.com/datacharmer/dbdeployer/globals"
)

//go:embed templates/ndb/ndb_start_cluster.gotxt
var ndbStartTemplate string

//go:embed templates/ndb/ndb_config.gotxt
var ndbConfigTemplate string

//go:embed templates/ndb/ndb_stop_cluster.gotxt
var ndbStopTemplate string

//go:embed templates/ndb/ndb_mgm.gotxt
var ndbMgmTemplate string

//go:embed templates/ndb/ndb_check_status.gotxt
var ndbCheckStatusTemplate string

var NdbTemplates = TemplateCollection{
	globals.TmplNdbStartCluster: TemplateDesc{
		Description: "NDB start cluster",
		Notes:       "",
		Contents:    ndbStartTemplate,
	},
	globals.TmplNdbStopCluster: TemplateDesc{
		Description: "NDB stop cluster",
		Notes:       "",
		Contents:    ndbStopTemplate,
	},
	globals.TmplNdbConfig: TemplateDesc{
		Description: "NDB cluster configuration",
		Notes:       "",
		Contents:    ndbConfigTemplate,
	},
	globals.TmplNdbMgm: TemplateDesc{
		Description: "NDB cluster manager",
		Notes:       "",
		Contents:    ndbMgmTemplate,
	},
	globals.TmplNdbCheckStatus: TemplateDesc{
		Description: "NDB check cluster status",
		Notes:       "",
		Contents:    ndbCheckStatusTemplate,
	},
}
