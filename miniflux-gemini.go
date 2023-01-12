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
	"strconv"
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
	mux.HandleFunc("/mark_as_read", markAsReadHandler)

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

func getMiniflux(ctx context.Context, w gemini.ResponseWriter) *client.Client {
	miniflux, ok := MinifluxFromContext(ctx)
	if !ok {
		w.WriteHeader(gemini.StatusPermanentFailure, "Unexpected error")
		log.Printf("couldn’t get miniflux")
		return nil
	}
	return miniflux
}

func markAsReadHandler(ctx context.Context, w gemini.ResponseWriter, r *gemini.Request) {
	query := r.URL.Query()
	idString := query.Get("id")
	if idString == "" {
		w.WriteHeader(gemini.StatusBadRequest, "missing id")
		return
	}
	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		w.WriteHeader(gemini.StatusBadRequest, "invalid id")
		return
	}

	miniflux := getMiniflux(ctx, w)
	if miniflux == nil {
		return
	}

	err = miniflux.UpdateEntries([]int64{id}, "read")
	if err != nil {
		w.WriteHeader(gemini.StatusCGIError, "miniflux error")
		log.Printf("miniflux error: %v", err)
		return
	}
	// Redirect to next entry
	w.WriteHeader(gemini.StatusRedirect, fmt.Sprintf("/entry?nextOf=%d", id))
}

func entryHandler(ctx context.Context, w gemini.ResponseWriter, r *gemini.Request) {
	filter := client.Filter{
		Status:    "unread",
		Order:     "published_at",
		Limit:     1,
		Direction: "desc",
		// TODO Make this configurable
		CategoryID: 7,
	}

	query := r.URL.Query()
	idString := query.Get("nextOf")
	if idString != "" {
		id, err := strconv.ParseInt(idString, 10, 64)
		if err != nil {
			w.WriteHeader(gemini.StatusBadRequest, "invalid id")
			return
		}
		filter.AfterEntryID = id
	}

	handleEntry(ctx, w, &filter)
}

// Generic function to render various entries, once a filter to find the entry has been figured out.
// TODO Handle multiple entries?
func handleEntry(ctx context.Context, w gemini.ResponseWriter, filter *client.Filter) {
	miniflux := getMiniflux(ctx, w)
	if miniflux == nil {
		return
	}

	entries, err := miniflux.Entries(filter)
	if err != nil {
		w.WriteHeader(gemini.StatusTemporaryFailure, "Error querying minflux")
		log.Printf("error getting miniflux entries: %v", err)
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

