/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package controller

import (
	"strings"
	"testing"

	"github.com/apache/answer/ui"
	"github.com/stretchr/testify/require"
)

// GetStyle scrapes the script and stylesheet paths out of the built
// index.html and every server-rendered page reuses them. The scrape is
// coupled to the exact attribute order and attribute set that the frontend
// build tool writes into those tags, and nothing in the system reports a
// mismatch: the frontend build still succeeds, the dev server still works,
// the binary still compiles, and the server-rendered pages simply come back
// with no script tags and no stylesheet.
//
// Assert the coupling directly so a change to the emitted tag shape fails
// here instead of shipping.
func TestGetStyleResolvesBuiltAssets(t *testing.T) {
	const builtIndexPath = "build/index.html"

	raw, err := ui.Build.ReadFile(builtIndexPath)
	if err != nil {
		t.Skipf("no frontend build embedded at %s; build the frontend and re-run: %v", builtIndexPath, err)
	}

	scripts, css := GetStyle()

	require.NotEmpty(t, scripts,
		"no script sources parsed out of %s; server-rendered pages would load without any JavaScript", builtIndexPath)
	for i, src := range scripts {
		require.NotEmpty(t, src, "script source %d parsed out of %s is empty", i, builtIndexPath)
	}

	require.NotEmpty(t, css,
		"no stylesheet href parsed out of %s; server-rendered pages would load unstyled", builtIndexPath)
	for i, href := range css {
		require.NotEmpty(t, href,
			"stylesheet href %d parsed out of %s is empty; server-rendered pages would load unstyled", i, builtIndexPath)
	}

	// Finding every stylesheet matters as much as finding one. The build emits
	// more than a single entry stylesheet, and a parser that stopped at the
	// first one would still satisfy every assertion above while half the page's
	// CSS silently stopped loading. That regression has happened once already.
	//
	// Count them again by a deliberately different and cruder method than the
	// parser uses, so the two have to agree. It is a lower bound: a build that
	// quotes attributes differently drives this to zero and the comparison
	// simply stops constraining, which is why it supplements the assertions
	// above rather than replacing them.
	declared := strings.Count(string(raw), `rel="stylesheet"`)
	require.GreaterOrEqual(t, len(css), declared,
		"%s declares at least %d stylesheets but only %d were parsed out of it; "+
			"server-rendered pages would load missing part of their CSS",
		builtIndexPath, declared, len(css))
}
