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

package dir

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDirSize(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.txt"), []byte("12345"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "b.txt"), []byte("1234567890"), 0o600))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sub"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sub", "c.txt"), []byte("1234567"), 0o600))

	size, err := DirSize(dir)
	require.NoError(t, err)
	assert.Equal(t, int64(22), size)
}

// A path that does not exist must be reported as an error. Before the fix
// filepath.Walk hands the callback a nil FileInfo for the root, and
// `info.IsDir()` panicked with a nil pointer dereference instead.
func TestDirSizeMissingPath(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "no-such-dir")

	size, err := DirSize(missing)
	require.Error(t, err)
	assert.Equal(t, int64(0), size)
}

// Same for a path whose parent is missing, and for a relative path that
// cannot be resolved from the working directory.
func TestDirSizeMissingPathsDoNotPanic(t *testing.T) {
	for _, missing := range []string{"no/such/relative/dir", "/no/such/absolute/dir"} {
		t.Run(missing, func(t *testing.T) {
			size, err := DirSize(missing)
			require.Error(t, err)
			assert.Equal(t, int64(0), size)
		})
	}
}
