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
	"text/template"

	"github.com/LukeEmmet/html2gemini"
	"miniflux.app/client"
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

type Entry struct {
	*client.Entry
	GeminiContent string
}

func NewEntry(minifluxEntry *client.Entry) (*Entry, error) {
	if minifluxEntry == nil {
		return nil, fmt.Errorf("error trying to render nil miniflux entry")
	}
	gemini, err := htmlToGemini(minifluxEntry.Content)
	if err != nil {
		return nil, fmt.Errorf("error converting gemini to HTML for entry %d: %w", minifluxEntry.ID, err)
	}

	return &Entry{
		Entry:         minifluxEntry,
		GeminiContent: gemini,
	}, nil
}

// Render renders the entry with the gemini template
func (entry *Entry) Render(w io.Writer) error {
	return entryTmpl.Execute(w, entry)
}

func htmlToGemini(html string) (gemini string, err error) {
	opts := html2gemini.NewOptions()
	// TODO Customize options
	ctx := html2gemini.NewTraverseContext(*opts)

	return html2gemini.FromString(html, *ctx)
}
