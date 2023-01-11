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
	"crypto/sha256"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"git.sr.ht/~adnano/go-gemini"
	"git.sr.ht/~adnano/go-gemini/certificate"
	"miniflux.app/client"
)

type Handler struct{}

func (h Handler) Handle(r gemini.Request) *gemini.Response {
	body := io.NopCloser(strings.NewReader("Testing server"))
	return &gemini.Response{
		Status: 20,
		Meta:   "text/gemini",
		Body:   body,
	}
}

func fingerprint(cert *x509.Certificate) [32]byte {
	b := sha256.Sum256(cert.Raw)
	return b
}

func tmpHandler(ctx context.Context, w gemini.ResponseWriter, r *gemini.Request) {
	tls := r.TLS()
	if len(tls.PeerCertificates) == 0 {
		w.WriteHeader(gemini.StatusCertificateRequired, "Certificate required, ask your admin to add yours")
		return
	}
	fingerprint := fingerprint(tls.PeerCertificates[0])
	fmt.Printf("%+v", fingerprint)
}

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

	certificates := &certificate.Store{}
	certificates.Register("gm.cj.rs")
	if err := certificates.Load("./certs"); err != nil {
		return err
	}

	mux := &gemini.Mux{}
	mux.HandleFunc("/", tmpHandler)

	server := &gemini.Server{
		Addr:           "gm.cj.rs:1965",
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		GetCertificate: certificates.Get,
	}

	if err := server.ListenAndServe(context.Background()); err != nil {
		return err
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
		return fmt.Errorf("error getting miniflux entries: %w", err)
	}
	if entries.Total < 1 {
		ErrorPage("No entry returned")
	}

	entry, err := NewEntry(entries.Entries[0])
	if err != nil {
		return fmt.Errorf("error templating entry: %w", err)
	}
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

