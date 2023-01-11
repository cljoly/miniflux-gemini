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
	"io"
	"text/template"
	_ "embed"

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

func NewEntry(minifluxEntry *client.Entry) *Entry {
	if minifluxEntry == nil {
		return nil
	}
	return &Entry{
		Entry:         minifluxEntry,
		GeminiContent: "TODO",
	}
}

// Render renders the entry with the gemini template
func (entry *Entry) Render(w io.Writer) error {
	return entryTmpl.Execute(w, entry)
}

