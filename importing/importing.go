// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2020 Giuseppe Maxia
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

package importing

import (
	"database/sql"
	"fmt"

	"github.com/go-sql-driver/mysql"
)

type DB struct {
	*sql.DB
}

func ParamsToConfig(host, user, password string, port int) *mysql.Config {
	var config = mysql.Config{
		User:                    user,
		Net:                     "tcp",
		Addr:                    fmt.Sprintf("%s:%d", host, port),
		Passwd:                  password,
		AllowCleartextPasswords: true,
		AllowNativePasswords:    true,
	}
	return &config
}

func Connect(config *mysql.Config) (*DB, error) {
	dsn := config.FormatDSN()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("error connecting to server %s - %s", config.Addr, err)
	}
	return &DB{db}, nil
}

func (db *DB) GetSingleResult(config *mysql.Config, query string, result interface{}) error {
	err := db.QueryRow(query).Scan(result)
	if err != nil {
		return fmt.Errorf("error getting version from server %s: - %s", config.Addr, err)
	}

	return nil
}
