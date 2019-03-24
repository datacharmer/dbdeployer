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
	Flavor      string      `json:"flavor"`
	Description string      `json:"description"`
	Features    FeatureList `json:"features"`
}

type ElementPath struct {
	dir      string
	fileName string
}
type FlavorIndicator struct {
	elements  []ElementPath
	flavor    string
	AllNeeded bool
}

const (
	// Tarball flavors
	MySQLFlavor         = "mysql"
	PerconaServerFlavor = "percona"
	MariaDbFlavor       = "mariadb"
	NdbFlavor           = "ndb"
	PxcFlavor           = "pxc"
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
	XtradbCluster    = "xtradbCluster"
	NdbCluster       = "ndbCluster"
	RootAuth         = "rootAuth"
	AdminAddress     = "adminAddress"
)

var MySQLCapabilities = Capabilities{
	Flavor:      MySQLFlavor,
	Description: "MySQL server",
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
		AdminAddress: {
			Description: "Connection through admin address",
			Since:       globals.MinimumAdminAddressVersion,
		},
	},
}

// Flavor indicators must be listed from the most complex ones to the
// simplest ones, because we want to catch the flavors that require
// multiple elements to be identified. If we put the simpler ones on top,
// we would miss the complex ones.
var FlavorCompositionList = []FlavorIndicator{
	{
		AllNeeded: true,
		elements: []ElementPath{
			{"bin", globals.FnGarbd},
			{"lib", globals.FnLibGaleraSmmSo},
			{"lib", globals.FnLibPerconaServerClientSo},
		},
		flavor: PxcFlavor,
	},
	{
		AllNeeded: true,
		elements: []ElementPath{
			{"bin", globals.FnGarbd},
			{"lib", globals.FnLibGaleraSmmA},
			{"lib", globals.FnLibPerconaServerClientA},
		},
		flavor: PxcFlavor,
	},
	{
		AllNeeded: true,
		elements: []ElementPath{
			{"bin", globals.FnGarbd},
			{"lib", globals.FnLibGaleraSmmDylib},
			{"lib", globals.FnLibPerconaServerClientDylib},
		},
		flavor: PxcFlavor,
	},
	{
		AllNeeded: true,
		elements: []ElementPath{
			{"bin", globals.FnNdbd},
			{"bin", globals.FnNdbdMgm},
			{"bin", globals.FnNdbdMgmd},
			{"bin", globals.FnNdbdMtd},
			{"lib", globals.FnNdbdEngineSo},
		},
		flavor: NdbFlavor,
	},
	{
		AllNeeded: false,
		elements: []ElementPath{
			{"bin", globals.FnAriaChk},
			{"lib", globals.FnLibMariadbClientA},
			{"lib", globals.FnLibMariadbClientDylib},
			{"lib", globals.FnLibMariadbA},
			{"lib", globals.FnLibMariadbDylib},
		},
		flavor: MariaDbFlavor,
	},

	{
		AllNeeded: false,
		elements: []ElementPath{
			{"lib", globals.FnLibPerconaServerClientA},
			{"lib", globals.FnLibPerconaServerClientSo},
			{"lib", globals.FnLibPerconaServerClientDylib},
		},
		flavor: PerconaServerFlavor,
	},

	{
		AllNeeded: false,
		elements: []ElementPath{
			{"bin", globals.FnTiDbServer},
		},
		flavor: TiDbFlavor,
	},
	{
		AllNeeded: false,
		elements: []ElementPath{
			{"bin", globals.FnMysqld},
			{"bin", globals.FnMysqldDebug},
			{"lib", globals.FnLibMySQLClientA},
		},
		flavor: MySQLFlavor,
	},
}

var PerconaCapabilities = Capabilities{
	Flavor:      PerconaServerFlavor,
	Description: "Percona Server",
	Features:    MySQLCapabilities.Features,
}

var TiDBCapabilities = Capabilities{
	Flavor:      TiDbFlavor,
	Description: "TiDB isolated server",
	Features:    FeatureList{
		// No capabilities so far
	},
}
var NdbCapabilities = Capabilities{
	Flavor:      NdbFlavor,
	Description: "MySQL NDB Cluster",
	Features: FeatureList{
		CreateUser:    MySQLCapabilities.Features[CreateUser],
		DataDict:      MySQLCapabilities.Features[DataDict],
		DynVariables:  MySQLCapabilities.Features[DynVariables],
		Initialize:    MySQLCapabilities.Features[Initialize],
		MySQLXDefault: MySQLCapabilities.Features[MySQLXDefault],
		Roles:         MySQLCapabilities.Features[Roles],
		SetPersist:    MySQLCapabilities.Features[SetPersist],
		NdbCluster: {
			Description: "MySQL NDB Cluster",
			Since:       globals.MinimumNdbClusterVersion,
		},
	},
}

var PxcCapabilities = Capabilities{
	Flavor:      PxcFlavor,
	Description: "Percona XtraDB Cluster",
	Features: addCapabilities(PerconaCapabilities.Features,
		FeatureList{
			XtradbCluster: {
				Description: "Xtradb Cluster creation",
				Since:       globals.MinimumXtradbClusterVersion,
			},
		}),
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
		RootAuth: {
			Description: "Root Authentication during install",
			Since:       globals.MinimumRootAuthVersion,
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
	NdbFlavor:           NdbCapabilities,
	PxcFlavor:           PxcCapabilities,
}

// Returns a set of existing capabilities with custom ones
// added (or replaced) to the list
func addCapabilities(flavorFeatures, features FeatureList) FeatureList {
	var fList = make(FeatureList)
	for fName, feature := range flavorFeatures {
		fList[fName] = feature
	}
	for fName, feature := range features {
		fList[fName] = feature
	}
	return fList
}

// Returns a subset of a flavor capabilities
func copyCapabilities(flavor string, names []string) FeatureList {
	var fList = make(FeatureList)
	_, flavorExists := AllCapabilities[flavor]
	if !flavorExists {
		return fList
	}
	for fName, feature := range AllCapabilities[flavor].Features {
		for _, n := range names {
			if fName == n {
				fList[n] = feature
			}
		}
	}
	return fList
}

// Returns true if a given flavor and version support the wanted feature
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
