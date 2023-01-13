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
	_ "embed"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"text/template"

	"github.com/LukeEmmet/html2gemini"
	miniflux "miniflux.app/client"
)

// Gemini templates

func geminiTemplate(name, text string) *template.Template {
	tmpl, err := template.New(name).Parse(text)
	if err != nil {
		panic(err)
	}
	return tmpl
}

var (
	//go:embed templates/entry.gmi
	entryTxt  string
	entryTmpl = geminiTemplate("entry", entryTxt)
)

type TemplatableEntry struct {
	*miniflux.Entry
	GeminiContent string
	query         *url.Values
}

func NewTemplatableEntry(minifluxEntry *miniflux.Entry, query *url.Values) (*TemplatableEntry, error) {
	if minifluxEntry == nil || query == nil {
		return nil, fmt.Errorf("error trying to render nil entry")
	}
	gemini, err := htmlToGemini(minifluxEntry.Content)
	if err != nil {
		return nil, fmt.Errorf("error converting gemini to HTML for entry %d: %w", minifluxEntry.ID, err)
	}

	return &TemplatableEntry{
		Entry:         minifluxEntry,
		GeminiContent: gemini,
		query:         query,
	}, nil
}

// Render renders the entry with the gemini template
func (entry *TemplatableEntry) Render(w io.Writer) error {
	return entryTmpl.Execute(w, entry)
}

// Utility function to copy the params of the current entry
func (entry *TemplatableEntry) copyQuery() url.Values {
	nextMap := make(map[string][]string)
	for k, v := range map[string][]string(*entry.query) {
		nextMap[k] = v
	}

	return url.Values(nextMap)
}

func maxInt(a, b int) int {
	if a >= b {
		return a
	} else {
		return b
	}
}

// Utility function to get the current offset
func (entry *TemplatableEntry) currentOffset() int {
	offset := 0
	offsetParsed, err := strconv.Atoi(entry.query.Get("offset"))
	if err == nil {
		offset = offsetParsed
	}

	return maxInt(offset, 0)
}

// Next returns the parameters to get the next entry in the reading list
func (entry *TemplatableEntry) Next() string {
	query := entry.copyQuery()
	query.Del("offset") // There may already be an offset stored, we want to replace it
	query.Set("offset", fmt.Sprint(entry.currentOffset()+1))
	return query.Encode()
}

// Prev returns the parameters to get the previous entry in the reading list
func (entry *TemplatableEntry) Prev() *string {
	offset := entry.currentOffset()
	if offset == 0 {
		return nil
	}

	query := entry.copyQuery()
	query.Del("offset") // There may already be an offset stored, we want to replace it
	query.Set("offset", fmt.Sprint(offset-1))

	s := query.Encode()
	return &s
}

func htmlToGemini(html string) (gemini string, err error) {
	opts := html2gemini.NewOptions()
	// TODO Customize options
	ctx := html2gemini.NewTraverseContext(*opts)

	return html2gemini.FromString(html, *ctx)
}

