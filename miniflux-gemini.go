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

	"cj.rs/miniflux-gemini/gemtext"
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
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/entry", entryHandler)
	mux.HandleFunc("/mark_as_read", markAsReadHandler)

	userMiddleware, err := NewUserMiddleware(db, mux)
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
	user, ok := UserFromContext(ctx)
	if !ok {
		w.WriteHeader(gemini.StatusPermanentFailure, "Unexpected error")
		log.Printf("couldn’t get user")
		return nil
	}
	miniflux := client.New(user.instance, user.token)
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

func homeHandler(ctx context.Context, w gemini.ResponseWriter, r *gemini.Request) {
	miniflux := getMiniflux(ctx, w)
	if miniflux == nil {
		w.WriteHeader(gemini.StatusTemporaryFailure, "miniflux error")
		log.Println("couldn't create miniflux client")
		return
	}
	query := r.URL.Query()

	categories, err := miniflux.Categories()
	if err != nil {
		w.WriteHeader(gemini.StatusTemporaryFailure, "Error querying minflux")
		log.Printf("error getting miniflux categories: %v", err)
		return
	}

	gemtextHome, err := gemtext.NewHome(&categories, &query)
	gemtextHome.Render(w)
}

func entryHandler(ctx context.Context, w gemini.ResponseWriter, r *gemini.Request) {
	articleList := NewArticleList()
	miniflux := getMiniflux(ctx, w)
	if miniflux == nil {
		w.WriteHeader(gemini.StatusTemporaryFailure, "miniflux error")
		log.Println("couldn't create miniflux client")
		return
	}

	query := r.URL.Query()
	articleList.Extend(query)

	entry, err := articleList.First(miniflux)
	if err != nil {
		w.WriteHeader(gemini.StatusTemporaryFailure, "Error querying minflux")
		log.Printf("error getting miniflux entries: %v", err)
		return
	}
	if entry == nil {
		w.WriteHeader(gemini.StatusTemporaryFailure, "No entry returned")
		return
	}

	gemtextEntry, err := gemtext.NewTemplatableEntry(entry, &query)
	if err != nil {
		w.WriteHeader(gemini.StatusPermanentFailure, "Unexpected error")
		log.Printf("error templating entry: %v", err)
		return
	}
	gemtextEntry.Render(w)
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

