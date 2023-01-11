/*
The AGPLv3 License (AGPLv3)

Copyright (c) 2023 Clément Joly

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"database/sql"
	_ "embed"
	"fmt"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

type SqliteDB struct {
	db *sql.DB
}

// Open and create the tables of the SQLite database
func NewDB(f string) (*SqliteDB, error) {
	db, err := sql.Open("sqlite", f)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	if _, err = db.Exec(schema); err != nil {
		return nil, err
	}

	return &SqliteDB{db: db}, nil
}

// TODO Return a full User struct and don’t panic
func (s *SqliteDB) GetUser(cert string) (instance, token string, err error) {
	row := s.db.QueryRow("SELECT instance, token FROM Users WHERE cert=?1", cert)
	err = row.Scan(&instance, &token)
	if err == sql.ErrNoRows {
		return instance, token, fmt.Errorf("Certificate not found")
	}
	if err != nil {
		return instance, token, fmt.Errorf("error reading user with cert %q: %w", cert, err)
	}
	return instance, token, nil
}

