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

package answercli

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigRoundTripAndCredentialResolution(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, ".config", "answer", "config.yaml")
	config := &Config{
		CurrentProfile: "work",
		Profiles: map[string]Profile{
			"work": {Server: "https://answer.example.com", Token: "answer_pat_secret"},
		},
	}

	require.NoError(t, SaveConfig(path, config))
	loaded, err := LoadConfig(path)
	require.NoError(t, err)
	require.Equal(t, config, loaded)

	profile, err := loaded.Resolve("", Environment{})
	require.NoError(t, err)
	require.Equal(t, "https://answer.example.com", profile.Server)
	require.Equal(t, "answer_pat_secret", profile.Token)

	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		require.NoError(t, err)
		require.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	}
}

func TestEnvironmentOverridesConfiguredProfile(t *testing.T) {
	config := &Config{CurrentProfile: "work", Profiles: map[string]Profile{"work": {Server: "https://old", Token: "old"}}}
	profile, err := config.Resolve("", Environment{Server: "https://new", Token: "new"})
	require.NoError(t, err)
	require.Equal(t, Profile{Server: "https://new", Token: "new"}, profile)
}
