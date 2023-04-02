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
	"time"

	"git.sr.ht/~adnano/go-gemini"
	"git.sr.ht/~adnano/go-gemini/certificate"
)

const defaultHost = "devd.io"

var hostFlag = flag.String("host", defaultHost, "hostname to generate a TLS certificate for")

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
	log.Println("Got TLS certificate for:", *hostFlag)

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

	log.Println("Listening on:", server.Addr)
	if err := server.ListenAndServe(context.Background()); err != nil {
		return err
	}

	return nil
}

func main() {
	flag.Parse()

	err := Run()
	if err != nil {
		log.Fatalf("Run: %v", err)
	}
}

