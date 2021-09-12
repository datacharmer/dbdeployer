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

// Templates for multiple sandboxes

var (
	//go:embed templates/multiple/start_multi.gotxt
	startMultiTemplate string

	//go:embed templates/multiple/restart_multi.gotxt
	restartMultiTemplate string

	//go:embed templates/multiple/use_multi.gotxt
	useMultiTemplate string

	//go:embed templates/multiple/exec_multi.gotxt
	execMultiTemplate string

	//go:embed templates/multiple/metadata_multi.gotxt
	metadataMultiTemplate string

	//go:embed templates/multiple/use_multi_admin.gotxt
	useMultiAdminTemplate string

	//go:embed templates/multiple/stop_multi.gotxt
	stopMultiTemplate string

	//go:embed templates/multiple/send_kill_multi.gotxt
	sendKillMultiTemplate string

	//go:embed templates/multiple/clear_multi.gotxt
	clearMultiTemplate string

	//go:embed templates/multiple/status_multi.gotxt
	statusMultiTemplate string

	//go:embed templates/multiple/test_sb_multi.gotxt
	testSbMultiTemplate string

	//go:embed templates/multiple/node.gotxt
	nodeTemplate string

	//go:embed templates/multiple/node_admin.gotxt
	nodeAdminTemplate string

	//go:embed templates/multiple/replicate_from_multi.gotxt
	replicateFromMultiTemplate string

	//go:embed templates/multiple/sysbench_multi.gotxt
	sysbenchMultiTemplate string

	//go:embed templates/multiple/sysbench_ready_multi.gotxt
	sysbenchReadyMultiTemplate string

	MultipleTemplates = TemplateCollection{
		globals.TmplStartMulti: TemplateDesc{
			Description: "Starts all nodes (with optional mysqld arguments)",
			Notes:       "",
			Contents:    startMultiTemplate,
		},
		globals.TmplRestartMulti: TemplateDesc{
			Description: "Restarts all nodes (with optional mysqld arguments)",
			Notes:       "",
			Contents:    restartMultiTemplate,
		},
		globals.TmplUseMulti: TemplateDesc{
			Description: "Runs the same SQL query in all nodes",
			Notes:       "",
			Contents:    useMultiTemplate,
		},
		globals.TmplExecMulti: TemplateDesc{
			Description: "Runs the same command in all nodes",
			Notes:       "",
			Contents:    execMultiTemplate,
		},
		globals.TmplMetadataMulti: TemplateDesc{
			Description: "Runs a metadata query in all nodes",
			Notes:       "",
			Contents:    metadataMultiTemplate,
		},
		globals.TmplUseMultiAdmin: TemplateDesc{
			Description: "Runs the same SQL query (as admin user) in all nodes",
			Notes:       "",
			Contents:    useMultiAdminTemplate,
		},
		globals.TmplStopMulti: TemplateDesc{
			Description: "Stops all nodes",
			Notes:       "",
			Contents:    stopMultiTemplate,
		},
		globals.TmplSendKillMulti: TemplateDesc{
			Description: "Sends kill signal to all nodes",
			Notes:       "",
			Contents:    sendKillMultiTemplate,
		},
		globals.TmplClearMulti: TemplateDesc{
			Description: "Removes data from all nodes",
			Notes:       "",
			Contents:    clearMultiTemplate,
		},
		globals.TmplStatusMulti: TemplateDesc{
			Description: "Shows status for all nodes",
			Notes:       "",
			Contents:    statusMultiTemplate,
		},
		globals.TmplTestSbMulti: TemplateDesc{
			Description: "Run sb test on all nodes",
			Notes:       "",
			Contents:    testSbMultiTemplate,
		},
		globals.TmplNode: TemplateDesc{
			Description: "Runs the MySQL client for a given node",
			Notes:       "",
			Contents:    nodeTemplate,
		},
		globals.TmplNodeAdmin: TemplateDesc{
			Description: "Runs the MySQL client for a given node as admin user",
			Notes:       "",
			Contents:    nodeAdminTemplate,
		},
		globals.TmplReplicateFromMulti: TemplateDesc{
			Description: "calls script replicate_from from node #1",
			Notes:       "",
			Contents:    replicateFromMultiTemplate,
		},
		globals.TmplSysbenchMulti: TemplateDesc{
			Description: "calls script sysbench from node #1",
			Notes:       "",
			Contents:    sysbenchMultiTemplate,
		},
		globals.TmplSysbenchReadyMulti: TemplateDesc{
			Description: "calls script sysbench_ready from node #1",
			Notes:       "",
			Contents:    sysbenchReadyMultiTemplate,
		},
	}
)
