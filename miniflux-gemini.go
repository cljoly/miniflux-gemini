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
	"fmt"
	"log"
	"os"

	"miniflux.app/client"
)

func Run() error {
	db, err := NewDB("miniflux-gemini.db")
	if err != nil {
		return fmt.Errorf("NewDB: %w", err)
	}

	// TODO Actually get the certificate
	instance, token, err := db.GetUser("1")
	if err != nil {
		return err
	}
	// TODO Use a pool of clients, with one for each instance?
	miniflux := client.New(instance, token)

	entries, err := miniflux.Entries(&client.Filter{
		Status:    "unread",
		Order:     "published_at",
		Limit:     1,
		Direction: "desc",
	})
	if err != nil {
		return fmt.Errorf("error getting miniflux entries: %w", err)
	}
	if entries.Total < 1 {
		ErrorPage("No entry returned")
	}

	entry := NewEntry(entries.Entries[0])
	entry.Render(os.Stdout)

	return nil
}

// TODO For now, just print
func ErrorPage(msg string) {
	println(msg)
}

func main() {
	err := Run()
	if err != nil {
		log.Fatalf("Run: %v", err)
	}
}

