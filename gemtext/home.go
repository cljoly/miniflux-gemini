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

package gemtext

import (
	_ "embed"
	"fmt"
	"io"
	"net/url"

	miniflux "miniflux.app/client"
)

var (
	//go:embed templates/home.gmi
	homeTxt  string
	homeTmpl = geminiTemplate("home", homeTxt)
)

type Home struct {
	Categories *miniflux.Categories
	query      *url.Values
}

func NewHome(categories *miniflux.Categories, query *url.Values) (*Home, error) {
	if categories == nil || query == nil {
		return nil, fmt.Errorf("error trying to render with nil categories")
	}

	return &Home{
		Categories: categories,
		query:      query,
	}, nil
}

func (home *Home) Params() string {
	return home.query.Encode()
}

func (home *Home) Render(w io.Writer) error {
	return homeTmpl.Execute(w, home)
}

