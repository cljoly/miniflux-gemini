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
	"context"
	"fmt"
	"log"
	"time"

	"git.sr.ht/~adnano/go-gemini"
	"git.sr.ht/~adnano/go-gemini/certificate"
	"miniflux.app/client"
)

func Run() error {
	db, err := NewDB("miniflux-gemini.db")
	if err != nil {
		return fmt.Errorf("NewDB: %w", err)
	}

	certificates := &certificate.Store{}
	certificates.Register("gm.cj.rs")
	if err := certificates.Load("./certs"); err != nil {
		return err
	}

	mux := &gemini.Mux{}
	mux.HandleFunc("/", todoHandler)
	mux.HandleFunc("/entry", entryHandler)

	minifluxMiddleware, err := NewMinifluxMiddleware(mux)
	if err != nil {
		return err
	}
	userMiddleware, err := NewUserMiddleware(db, minifluxMiddleware)
	if err != nil {
		return err
	}

	server := &gemini.Server{
		Addr:           "gm.cj.rs:1965",
		Handler:        userMiddleware,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		GetCertificate: certificates.Get,
	}

	if err := server.ListenAndServe(context.Background()); err != nil {
		return err
	}

	return nil
}

func entryHandler(ctx context.Context, w gemini.ResponseWriter, r *gemini.Request) {
	miniflux, ok := MinifluxFromContext(ctx)
	if !ok {
		w.WriteHeader(gemini.StatusPermanentFailure, "Unexpected error")
		log.Printf("couldn’t get miniflux")
	}

	entries, err := miniflux.Entries(&client.Filter{
		Status:    "unread",
		Order:     "published_at",
		Limit:     1,
		Direction: "desc",
		// TODO Make this configurable
		CategoryID: 7,
	})
	if err != nil {
		w.WriteHeader(gemini.StatusTemporaryFailure, "Error querying minflux")
		log.Printf("error getting miniflux entries: %w", err)
		return
	}
	if entries.Total < 1 {
		w.WriteHeader(gemini.StatusTemporaryFailure, "No entry returned")
		return
	}

	entry, err := NewEntry(entries.Entries[0])
	if err != nil {
		w.WriteHeader(gemini.StatusPermanentFailure, "Unexpected error")
		log.Printf("error templating entry: %v", err)
		return
	}
	entry.Render(w)
}

func todoHandler(ctx context.Context, w gemini.ResponseWriter, r *gemini.Request) {
	w.WriteHeader(gemini.StatusTemporaryFailure, "Not implemented")
}

func main() {
	err := Run()
	if err != nil {
		log.Fatalf("Run: %v", err)
	}
}

