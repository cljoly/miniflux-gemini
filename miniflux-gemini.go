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
	"flag"
	"fmt"
	"log"
	"strconv"
	"time"

	"cj.rs/miniflux-gemini/gemtext"
	"git.sr.ht/~adnano/go-gemini"
	"git.sr.ht/~adnano/go-gemini/certificate"
	minifluxClient "miniflux.app/client"
)

var hostFlag = flag.String("host", "devd.io", "hostname to generate a TLS certificate for")

func Run() error {
	db, err := NewDB("miniflux-gemini.db")
	if err != nil {
		return fmt.Errorf("NewDB: %w", err)
	}

	certificates := &certificate.Store{}
	certificates.Register(*hostFlag)
	if err := certificates.Load("./certs"); err != nil {
		return err
	}

	mux := &gemini.Mux{}
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/entry", entryHandler)
	mux.HandleFunc("/mark_as", markAsHandler)

	userMiddleware, err := NewUserMiddleware(db, mux)
	if err != nil {
		return err
	}

	server := &gemini.Server{
		Addr:           "0.0.0.0:1965",
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

func getMiniflux(ctx context.Context, w gemini.ResponseWriter) *minifluxClient.Client {
	user, ok := UserFromContext(ctx)
	if !ok {
		w.WriteHeader(gemini.StatusPermanentFailure, "Unexpected error")
		log.Printf("couldn’t get user")
		return nil
	}
	miniflux := minifluxClient.New(user.instance, user.token)
	return miniflux
}

// markAsHandler changes the status of the entry as given
func markAsHandler(ctx context.Context, w gemini.ResponseWriter, r *gemini.Request) {
	query := r.URL.Query()

	status := query.Get("_status")
	switch status {
	case minifluxClient.EntryStatusRead,
		minifluxClient.EntryStatusUnread:
		// valid, continue
	default:
		w.WriteHeader(gemini.StatusBadRequest, "missing or invalid status")
		return
	}

	idString := query.Get("_id")
	if idString == "" {
		w.WriteHeader(gemini.StatusBadRequest, "missing id")
		return
	}
	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		w.WriteHeader(gemini.StatusBadRequest, "invalid id")
		return
	}

	// Remove params for this action, we will pass back the other params
	query.Del("_status")
	query.Del("_id")

	miniflux := getMiniflux(ctx, w)
	if miniflux == nil {
		return
	}

	err = miniflux.UpdateEntries([]int64{id}, status)
	if err != nil {
		w.WriteHeader(gemini.StatusCGIError, "miniflux error")
		log.Printf("error updating entry %v: %v", id, err)
		return
	}

	// Save the params and attempt to keep mostly the same position in the
	// article list (this may get back to the same article if the reading
	// list has read articles)
	w.WriteHeader(gemini.StatusRedirect, fmt.Sprintf("/entry?%s", query.Encode()))
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
	err = gemtextHome.Render(w)
	if err != nil {
		log.Printf("error rendering home template: %v", err)
		return
	}
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
	err = gemtextEntry.Render(w)
	if err != nil {
		log.Printf("error rendering entry: %v", err)
		return
	}
}

func todoHandler(ctx context.Context, w gemini.ResponseWriter, r *gemini.Request) {
	w.WriteHeader(gemini.StatusTemporaryFailure, "Not implemented")
}

func main() {
	flag.Parse()

	err := Run()
	if err != nil {
		log.Fatalf("Run: %v", err)
	}
}

