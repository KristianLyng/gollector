/*
 * skogul, http writer internal test cases
 *
 * Copyright (c) 2020 Telenor Norge AS
 * Author(s):
 *  - Håkon Solbjørg <hakon.solbjorg@telenor.com>
 *
 * This library is free software; you can redistribute it and/or
 * modify it under the terms of the GNU Lesser General Public
 * License as published by the Free Software Foundation; either
 * version 2.1 of the License, or (at your option) any later version.
 *
 * This library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public
 * License along with this library; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA
 * 02110-1301  USA
 */
package sender // required for calling HTTP.init()

import (
	"net/http"
	"testing"
)

func TestHttpSenderAlwaysHasContentTypeHeader(t *testing.T) {
	headers := make(map[string]string)
	ht := HTTP{
		Headers: headers,
	}
	ht.init() // this is called by Auto.Alloc()

	if ht.Headers[http.CanonicalHeaderKey("content-type")] == "" {
		t.Errorf("Missing value in Content-Type header when not setting the header value")
	}
}

func TestHttpSenderKeepContentHeader(t *testing.T) {
	textPlain := "text/plain"

	headers := make(map[string]string)
	headers[http.CanonicalHeaderKey("content-type")] = textPlain
	ht := HTTP{
		Headers: headers,
	}
	ht.init()

	if ht.Headers[http.CanonicalHeaderKey("content-type")] == "" {
		t.Errorf("Missing value in Content-Type header when setting the header value")
	}
	if ht.Headers[http.CanonicalHeaderKey("content-type")] != textPlain {
		t.Errorf("Content-Type value was '%s' when we expected '%s'", ht.Headers[http.CanonicalHeaderKey("content-type")], textPlain)
	}
}

func TestHttpSenderKeepContentTypeHeaderEvenInNonCanonicalizedForm(t *testing.T) {
	textPlain := "text/plain"

	headers := make(map[string]string)
	headers["content-type"] = textPlain
	ht := HTTP{
		Headers: headers,
	}
	ht.init()

	if ht.Headers[http.CanonicalHeaderKey("content-type")] == "" {
		t.Errorf("Missing value in Content-Type header when setting the header value")
	}
	if ht.Headers[http.CanonicalHeaderKey("content-type")] != textPlain {
		t.Errorf("Content-Type value was '%s' when we expected '%s'", ht.Headers[http.CanonicalHeaderKey("content-type")], textPlain)
	}
}
