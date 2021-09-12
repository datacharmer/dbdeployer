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

// Templates for group replication

var (

	//go:embed templates/pxc/pxc_start.gotxt
	pxcStartTemplate string

	//go:embed templates/pxc/check_pxc_nodes.gotxt
	checkPxcNodesTemplate string

	//go:embed templates/pxc/pxc_replication.gotxt
	pxcReplicationTemplate string

	PxcTemplates = TemplateCollection{
		globals.TmplPxcCheckNodes: TemplateDesc{
			Description: "Checks the status of PXC replication",
			Notes:       "",
			Contents:    checkPxcNodesTemplate,
		},
		globals.TmplPxcReplication: TemplateDesc{
			Description: "Replication options for PXC",
			Notes:       "",
			Contents:    pxcReplicationTemplate,
		},
		globals.TmplPxcStart: TemplateDesc{
			Description: "start all nodes in a PXC",
			Notes:       "",
			Contents:    pxcStartTemplate,
		},
	}
)
