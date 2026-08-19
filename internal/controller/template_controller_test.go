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

	if _, err := ui.Build.ReadFile(builtIndexPath); err != nil {
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
}
