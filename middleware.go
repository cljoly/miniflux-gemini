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
	"log"

	"git.sr.ht/~adnano/go-gemini"
	"miniflux.app/client"
)

func fingerprint(cert *x509.Certificate) string {
	b := sha256.Sum256(cert.Raw)
	s := fmt.Sprintf("%x", b)
	return s
}

// To find various values in context
const userKey = "User"
const minifluxKey = "MinifluxClient"

// UserMiddleware adds the user to context, found by its TLS certificate
type UserMiddleware struct {
	db *SqliteDB
	h  gemini.Handler
}

func NewUserMiddleware(db *SqliteDB, h gemini.Handler) (*UserMiddleware, error) {
	if db == nil {
		return nil, fmt.Errorf(
			"NewUserMiddleware: nil values not allowed",
		)
	}
	return &UserMiddleware{db, h}, nil
}

func (um *UserMiddleware) ServeGemini(ctx context.Context, w gemini.ResponseWriter, r *gemini.Request) {
	tls := r.TLS()
	if len(tls.PeerCertificates) == 0 {
		w.WriteHeader(gemini.StatusCertificateRequired, "Certificate required, ask your admin to add yours")
		return
	}
	fingerprint := fingerprint(tls.PeerCertificates[0])

	user, err := um.db.GetUser(fingerprint)
	if err == ErrUserNotFound {
		w.WriteHeader(gemini.StatusCertificateNotAuthorized,
			fmt.Sprintf(
				"Unknown certificate, ask your admin to add yours: %q",
				fingerprint,
			))
		return
	}
	if err != nil {
		w.WriteHeader(gemini.StatusPermanentFailure, "Internal Error")
		log.Printf("error getting user in db: %v", err)
		return
	}

	ctx2 := context.WithValue(ctx, userKey, &user)
	um.h.ServeGemini(ctx2, w, r)
}

func UserFromContext(ctx context.Context) (*User, bool) {
	user, ok := ctx.Value(userKey).(*User)
	return user, ok
}

// MinifluxMiddleware depends on UserMiddleware being called first
type MinifluxMiddleware struct {
	h gemini.Handler
	// TODO Store a map of clients per instance? (need to be thread safe)
}

func NewMinifluxMiddleware(h gemini.Handler) (*MinifluxMiddleware, error) {
	return &MinifluxMiddleware{h}, nil
}

func (mm *MinifluxMiddleware) ServeGemini(ctx context.Context, w gemini.ResponseWriter, r *gemini.Request) {
	user, ok := UserFromContext(ctx)
	if !ok {
		w.WriteHeader(gemini.StatusPermanentFailure, "Internal Error")
		log.Printf("Miniflux middleware error: couldn’t find user. Was UserMiddleware called first?", user)
		return
	}

	miniflux := client.New(user.instance, user.token)

	ctx = context.WithValue(ctx, minifluxKey, miniflux)
	mm.h.ServeGemini(ctx, w, r)
}

func MinifluxFromContext(ctx context.Context) (*client.Client, bool) {
	c, ok := ctx.Value(minifluxKey).(*client.Client)
	return c, ok
}
