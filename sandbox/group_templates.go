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

//go:build go1.16
// +build go1.16

package sandbox

import (
	_ "embed"

	"github.com/datacharmer/dbdeployer/globals"
)

// Templates for group replication

var (
	//go:embed templates/group/init_nodes.gotxt
	initNodesTemplate string

	//go:embed templates/group/check_nodes.gotxt
	checkNodesTemplate string

	//go:embed templates/group/group_repl_options.gotxt
	groupReplOptionsTemplate string

	GroupTemplates = TemplateCollection{
		globals.TmplInitNodes: TemplateDesc{
			Description: "Initialize group replication after deployment",
			Notes:       "",
			Contents:    initNodesTemplate,
		},
		globals.TmplCheckNodes: TemplateDesc{
			Description: "Checks the status of group replication",
			Notes:       "",
			Contents:    checkNodesTemplate,
		},
		globals.TmplGroupReplOptions: TemplateDesc{
			Description: "replication options for Group replication node",
			Notes:       "",
			Contents:    groupReplOptionsTemplate,
		},
	}
)
