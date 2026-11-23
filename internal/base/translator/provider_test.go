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

package translator

import (
	"os"
	"path/filepath"
	"testing"

	bundles "github.com/apache/answer/i18n"
	"github.com/segmentfault/pacman/i18n"
	"github.com/stretchr/testify/require"
)

func TestNewTranslatorLoadsBundledEnglish(t *testing.T) {
	// Reserved message keys such as "description" beside nested UI labels can
	// make go-i18n reject the entire English bundle and panic on translation.
	dir := t.TempDir()
	for _, name := range []string{"en_US.yaml", "i18n.yaml"} {
		content, err := bundles.I18n.ReadFile(name)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), content, 0o600))
	}

	tr, err := NewTranslator(&I18n{BundleDir: dir})
	require.NoError(t, err)
	require.NotPanics(t, func() {
		require.Equal(t, "Email", tr.Tr(i18n.DefaultLanguage, "email"))
	})
}
