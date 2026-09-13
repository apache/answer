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
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var ErrProfileNotConfigured = errors.New("answer CLI profile is not configured")

type Profile struct {
	Server            string `yaml:"server" json:"server"`
	Token             string `yaml:"token" json:"-"`
	AllowInsecureHTTP bool   `yaml:"allow_insecure_http,omitempty" json:"allow_insecure_http,omitempty"`
}

type Config struct {
	CurrentProfile string             `yaml:"current_profile" json:"current_profile"`
	Profiles       map[string]Profile `yaml:"profiles" json:"profiles"`
}

type Environment struct {
	Profile string
	Server  string
	Token   string
}

func DefaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "answer", "config.yaml"), nil
}

func LoadConfig(path string) (*Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Profiles: make(map[string]Profile)}, nil
		}
		return nil, err
	}
	config := &Config{}
	if err := yaml.Unmarshal(content, config); err != nil {
		return nil, err
	}
	if config.Profiles == nil {
		config.Profiles = make(map[string]Profile)
	}
	return config, nil
}

func SaveConfig(path string, config *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	if err := os.Chmod(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	content, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, content, 0o600); err != nil {
		return err
	}
	return os.Chmod(path, 0o600)
}

func (c *Config) Resolve(requestedProfile string, environment Environment) (Profile, error) {
	profileName := requestedProfile
	if environment.Profile != "" {
		profileName = environment.Profile
	}
	if profileName == "" {
		profileName = c.CurrentProfile
	}
	profile, exists := c.Profiles[profileName]
	if !exists && environment.Server == "" && environment.Token == "" {
		return Profile{}, ErrProfileNotConfigured
	}
	if environment.Server != "" {
		profile.Server = environment.Server
	}
	if environment.Token != "" {
		profile.Token = environment.Token
	}
	if profile.Server == "" || profile.Token == "" {
		return Profile{}, ErrProfileNotConfigured
	}
	return profile, nil
}
