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
	"strconv"

	miniflux "miniflux.app/client"
)

// Treat the unread articles as a list navigated by offsets. If new articles
// are added, the offset won’t send back to the same articles, but that should
// be infrequent enough

type ArticleList struct {
	miniflux.Filter
}

// Default parameters used in filters, that defines the basis for the article
// list
func defaultFilter() miniflux.Filter {
	return miniflux.Filter{
		Status:    "unread",
		Order:     "published_at",
		Direction: "desc",
	}
}

func NewArticleList() ArticleList {
	return ArticleList{defaultFilter()}
}

// ExtendFilter takes net/url.Url.Values (in the generic form of a map) and
// completes the default filter
func (al *ArticleList) Extend(values map[string][]string) {

	if k, exists := values["status"]; exists {
		al.Status = k[0]
	}
	if k, exists := values["offset"]; exists {
		offset, err := strconv.Atoi(k[0])
		if err == nil {
			al.Offset = offset
		}
	}
	if k, exists := values["limit"]; exists {
		limit, err := strconv.Atoi(k[0])
		if err == nil {
			al.Limit = limit
		}
	}
	if k, exists := values["order"]; exists {
		al.Order = k[0]
	}
	if k, exists := values["direction"]; exists {
		al.Direction = k[0]
	}
	if k, exists := values["starred"]; exists {
		al.Starred = k[0]
	}
	if k, exists := values["before"]; exists {
		before, err := strconv.ParseInt(k[0], 10, 64)
		if err == nil {
			al.Before = before
		}
	}
	if k, exists := values["after"]; exists {
		after, err := strconv.ParseInt(k[0], 10, 64)
		if err == nil {
			al.After = after
		}
	}
	if k, exists := values["afterEntryID"]; exists {
		afterEntryID, err := strconv.ParseInt(k[0], 10, 64)
		if err == nil {
			al.AfterEntryID = afterEntryID
		}
	}
	if k, exists := values["beforeEntryID"]; exists {
		beforeEntryID, err := strconv.ParseInt(k[0], 10, 64)
		if err == nil {
			al.BeforeEntryID = beforeEntryID
		}
	}
	if k, exists := values["search"]; exists {
		al.Search = k[0]
	}
	if k, exists := values["categoryID"]; exists {
		categoryID, err := strconv.ParseInt(k[0], 10, 64)
		if err == nil {
			al.CategoryID = categoryID
		}
	}
	if k, exists := values["feedID"]; exists {
		feedID, err := strconv.ParseInt(k[0], 10, 64)
		if err == nil {
			al.FeedID = feedID
		}
	}
	if k, exists := values["statuses"]; exists {
		al.Statuses = k
	}
}

// First returns the first entry
func (al *ArticleList) First(client *miniflux.Client) (*miniflux.Entry, error) {
	prevLimit := al.Filter.Limit
	al.Filter.Limit = 1
	entrySet, err := client.Entries(&al.Filter)
	if err != nil {
		return nil, err
	}
	al.Filter.Limit = prevLimit

	entry := new(miniflux.Entry)
	if entrySet.Total > 0 {
		entry = entrySet.Entries[0]

	}
	return entry, err
}

