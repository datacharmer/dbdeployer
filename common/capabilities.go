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
package common

import "github.com/datacharmer/dbdeployer/globals"

type MinimumVersion []int
type Capability struct {
	Description string         `json:"description"`
	Since       MinimumVersion `json:"since"`
	Until       MinimumVersion `json:"until"`
}
type FeatureList map[string]Capability

type Capabilities struct {
	Flavor   string      `json:"flavor"`
	Features FeatureList `json:"features"`
}

const (
	// Tarball flavors
	MySQLFlavor         = "mysql"
	PerconaServerFlavor = "percona"
	MariaDbFlavor       = "mariadb"
	NDBFlavor           = "ndb"
	TiDbFlavor          = "tidb"

	// Feature names
	InstallDb        = "installdb"
	DynVariables     = "dynVars"
	SemiSynch        = "semiSync"
	CrashSafe        = "crashSafe"
	GTID             = "GTID"
	EnhancedGTID     = "enhancedGTID"
	Initialize       = "initialize"
	CreateUser       = "createUser"
	SuperReadOnly    = "superReadOnly"
	MySQLX           = "mysqlx"
	MySQLXDefault    = "mysqlxDefault"
	MultiSource      = "multiSource"
	GroupReplication = "groupReplication"
	SetPersist       = "setPersist"
	Roles            = "roles"
	NativeAuth       = "nativeAuth"
	DataDict         = "datadict"
)

var MySQLCapabilities = Capabilities{
	Flavor: MySQLFlavor,
	Features: FeatureList{
		InstallDb: {
			Description: "uses mysql_install_db",
			Since:       globals.MinimumMySQLInstallDb,
			Until:       globals.MaximumMySQLInstallDb,
		},
		DynVariables: {
			Description: "dynamic variables",
			Since:       globals.MinimumDynVariablesVersion,
		},
		SemiSynch: {
			Description: "semi-synchronous replication",
			Since:       globals.MinimumSemiSyncVersion,
		},
		CrashSafe: {
			Description: "crash-safe replication",
			Since:       globals.MinimumCrashSafeVersion,
		},
		GTID: {
			Description: "Global transaction identifiers",
			Since:       globals.MinimumGtidVersion,
		},
		EnhancedGTID: {
			Description: "Enhanced Global transaction identifiers",
			Since:       globals.MinimumEnhancedGtidVersion,
		},
		Initialize: {
			Description: "mysqld --initialize as default",
			Since:       globals.MinimumDefaultInitializeVersion,
		},
		CreateUser: {
			Description: "Create user mandatory",
			Since:       globals.MinimumCreateUserVersion,
		},
		SuperReadOnly: {
			Description: "super-read-only support",
			Since:       globals.MinimumSuperReadOnly,
		},
		MySQLX: {
			Description: "MySQLX supported",
			Since:       globals.MinimumMysqlxVersion,
		},
		MySQLXDefault: {
			Description: "MySQLX enabled by default",
			Since:       globals.MinimumMysqlxDefaultVersion,
		},
		MultiSource: {
			Description: "multi-source replication",
			Since:       globals.MinimumMultiSourceReplVersion,
		},
		GroupReplication: {
			Description: "group replication",
			Since:       globals.MinimumGroupReplVersion,
		},
		SetPersist: {
			Description: "Set persist supported",
			Since:       globals.MinimumPersistVersion,
		},
		Roles: {
			Description: "Roles supported",
			Since:       globals.MinimumRolesVersion,
		},
		NativeAuth: {
			Description: "Native Authentication plugin",
			Since:       globals.MinimumNativeAuthPluginVersion,
		},
		DataDict: {
			Description: "data dictionary",
			Since:       globals.MinimumDataDictionaryVersion,
		},
	},
}

var PerconaCapabilities = Capabilities{
	Flavor:   PerconaServerFlavor,
	Features: MySQLCapabilities.Features,
}

var TiDBCapabilities = Capabilities{
	// No capabilities so far
}
var NDBCapabilities = Capabilities{
	// No capabilities so far
}

// NOTE: We only list the capabilities
// for which dbdeployer needs to take action
var MariadbCapabilities = Capabilities{
	Flavor: MariaDbFlavor,
	Features: FeatureList{
		InstallDb: {
			Description: "uses mysql_install_db",
			Since:       globals.MinimumMySQLInstallDb,
			Until:       nil,
		},
		DynVariables: MySQLCapabilities.Features[DynVariables],
		SemiSynch:    MySQLCapabilities.Features[SemiSynch],
	},
}

var AllCapabilities = map[string]Capabilities{
	MySQLFlavor:         MySQLCapabilities,
	PerconaServerFlavor: PerconaCapabilities,
	MariaDbFlavor:       MariadbCapabilities,
	TiDbFlavor:          TiDBCapabilities,
	NDBFlavor:           NDBCapabilities,
}

func HasCapability(flavor, feature, version string) (bool, error) {
	versionList, err := VersionToList(version)
	if err != nil {
		return false, err
	}
	for flavorName, capabilities := range AllCapabilities {
		if flavorName == flavor {
			featureDefinition, ok := capabilities.Features[feature]
			if ok {
				overMinimum, err := GreaterOrEqualVersionList(versionList, featureDefinition.Since)
				if err != nil {
					return false, err
				}
				withinMaximum := true
				if featureDefinition.Until != nil {
					withinMaximum, err = GreaterOrEqualVersionList(featureDefinition.Until, versionList)
					if err != nil {
						return false, err
					}
				}
				return overMinimum && withinMaximum, nil
			}
		}
	}
	return false, nil
}
