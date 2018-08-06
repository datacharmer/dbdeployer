// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2018 Giuseppe Maxia
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

var VersionDef string = "1.8.3" // 2018-08-05

// Compatible version is the version used to mark compatible archives (templates, configuration).
// It is usually major.minor.0, except when we are at version 0.x, when
// every revision may bring incompatibility
var CompatibleVersion string = "1.8.0" // 2018-07-05
