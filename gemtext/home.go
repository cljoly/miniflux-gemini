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
	Categories []*RichCategory
	query      *url.Values
}

// Categories enriched with their Feeds
type RichCategory struct {
	*miniflux.Category
	Feeds []*miniflux.Feed
}

func NewHome(categories *miniflux.Categories, feeds *miniflux.Feeds, query *url.Values) (*Home, error) {
	if categories == nil || query == nil {
		return nil, fmt.Errorf("error trying to render with nil arguments")
	}
	nbrCategories := len(*categories)

	feedsByCategory := feedsByCategoryID(feeds, nbrCategories)

	richCategories := make([]*RichCategory, nbrCategories)
	for i, category := range *categories {
		richCategories[i] = &RichCategory{
			Category: category,
			Feeds:    feedsByCategory[category.ID],
		}
	}

	return &Home{
		Categories: richCategories,
		query:      query,
	}, nil
}

func feedsByCategoryID(feeds *miniflux.Feeds, categoryHint int) map[int64][]*miniflux.Feed {
	feedsByCategory := make(map[int64][]*miniflux.Feed, categoryHint)
	for _, feed := range *feeds {
		if feeds, ok := feedsByCategory[feed.Category.ID]; ok {
			feedsByCategory[feed.Category.ID] = append(feeds, feed)
		} else {
			feedsByCategory[feed.Category.ID] = make([]*miniflux.Feed, 0)
		}
	}

	return feedsByCategory
}

func (home *Home) Params(key_values ...string) (string, error) {
	return params(home.query, key_values...)
}

func (home *Home) Render(w io.Writer) error {
	return homeTmpl.Execute(w, home)
}

