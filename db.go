// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright Clément Joly and contributors.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.


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

type User struct {
	certFingerprint, instance, token string
}

var ErrUserNotFound = fmt.Errorf("User not found in DB")

// TODO Return a full User struct and don’t panic
func (s *SqliteDB) GetUser(certFingerprint string) (user User, err error) {
	user.certFingerprint = certFingerprint
	row := s.db.QueryRow("SELECT instance, token FROM Users WHERE certFingerprint=?1", certFingerprint)
	err = row.Scan(&user.instance, &user.token)
	if err == sql.ErrNoRows {
		return user, ErrUserNotFound
	}
	if err != nil {
		return user, fmt.Errorf("error reading user with cert %q: %w", certFingerprint, err)
	}
	return user, nil
}

