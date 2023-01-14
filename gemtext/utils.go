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
	"text/template"
	"net/url"
	"strconv"
)

func geminiTemplate(name, text string) *template.Template {
	tmpl, err := template.New(name).Parse(text)
	if err != nil {
		panic(err)
	}
	return tmpl
}

func maxInt(a, b int) int {
	if a >= b {
		return a
	} else {
		return b
	}
}

// Utility function to copy the params of the current entry
func copyQuery(old *url.Values) url.Values {
	nextMap := make(map[string][]string)
	for k, v := range map[string][]string(*old) {
		nextMap[k] = v
	}

	return url.Values(nextMap)
}

// Utility function to get the current offset
func currentOffset(values *url.Values) int {
	offset := 0
	offsetParsed, err := strconv.Atoi(values.Get("offset"))
	if err == nil {
		offset = offsetParsed
	}

	return maxInt(offset, 0)
}

