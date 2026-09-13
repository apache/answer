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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

type Options struct {
	ConfigPath string
	Stdin      io.Reader
	Stdout     io.Writer
	Stderr     io.Writer
	HTTPClient *http.Client
	Version    string
}

type commandState struct {
	options    Options
	configPath string
	profile    string
}

func NewRootCommand(options Options) *cobra.Command {
	if options.Stdin == nil {
		options.Stdin = os.Stdin
	}
	if options.Stdout == nil {
		options.Stdout = os.Stdout
	}
	if options.Stderr == nil {
		options.Stderr = os.Stderr
	}
	state := &commandState{options: options, configPath: options.ConfigPath}
	root := &cobra.Command{
		Use:           "answer-cli",
		Version:       options.Version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetIn(options.Stdin)
	root.SetOut(options.Stdout)
	root.SetErr(options.Stderr)
	root.PersistentFlags().StringVar(&state.configPath, "config", state.configPath, "configuration file")
	root.PersistentFlags().StringVar(&state.profile, "profile", "", "profile name")
	root.AddCommand(state.authCommand(), state.questionCommand(), state.answerCommand(), state.voteCommand(), state.tagCommand())
	return root
}

func (s *commandState) authCommand() *cobra.Command {
	authCommand := &cobra.Command{Use: "auth"}
	authCommand.AddCommand(s.authLoginCommand(), s.authLogoutCommand(), &cobra.Command{
		Use:   "status",
		Short: "Show the configured PAT and user",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := s.client()
			if err != nil {
				return s.writeError(err)
			}
			data, err := client.Do(cmd.Context(), http.MethodGet, "/answer/api/v1/personal-access-tokens/current", nil, nil)
			if err != nil {
				return s.writeError(err)
			}
			return writeSuccess(s.options.Stdout, data)
		},
	})
	return authCommand
}

func (s *commandState) authLoginCommand() *cobra.Command {
	var server, profileName string
	var withToken, allowInsecureHTTP bool
	command := &cobra.Command{
		Use: "login", Short: "Save an existing personal access token",
		RunE: func(_ *cobra.Command, _ []string) error {
			if !withToken {
				return s.writeError(fmt.Errorf("--with-token is required"))
			}
			tokenBytes, err := io.ReadAll(s.options.Stdin)
			if err != nil {
				return s.writeError(err)
			}
			token := string(bytes.TrimSpace(tokenBytes))
			if server == "" || token == "" {
				return s.writeError(fmt.Errorf("--server and a token on stdin are required"))
			}
			if err := ValidateServerURL(server, allowInsecureHTTP); err != nil {
				return s.writeError(err)
			}
			if profileName == "" {
				profileName = "default"
			}
			path, err := s.resolvedConfigPath()
			if err != nil {
				return s.writeError(err)
			}
			config, err := LoadConfig(path)
			if err != nil {
				return s.writeError(err)
			}
			config.Profiles[profileName] = Profile{Server: server, Token: token, AllowInsecureHTTP: allowInsecureHTTP}
			config.CurrentProfile = profileName
			if err := SaveConfig(path, config); err != nil {
				return s.writeError(err)
			}
			return writeSuccess(s.options.Stdout, mustJSON(map[string]string{"profile": profileName, "server": server}))
		},
	}
	command.Flags().StringVar(&server, "server", "", "Answer server URL")
	command.Flags().StringVar(&profileName, "name", "default", "profile name")
	command.Flags().BoolVar(&withToken, "with-token", false, "read the PAT from stdin")
	command.Flags().BoolVar(&allowInsecureHTTP, "allow-insecure-http", false, "allow non-loopback HTTP")
	return command
}

func (s *commandState) authLogoutCommand() *cobra.Command {
	return &cobra.Command{
		Use: "logout", Short: "Remove the local credential without revoking it on the server",
		RunE: func(_ *cobra.Command, _ []string) error {
			path, err := s.resolvedConfigPath()
			if err != nil {
				return s.writeError(err)
			}
			config, err := LoadConfig(path)
			if err != nil {
				return s.writeError(err)
			}
			profileName := s.profile
			if profileName == "" {
				profileName = config.CurrentProfile
			}
			profile, exists := config.Profiles[profileName]
			if exists {
				profile.Token = ""
				config.Profiles[profileName] = profile
			}
			if err := SaveConfig(path, config); err != nil {
				return s.writeError(err)
			}
			return writeSuccess(s.options.Stdout, mustJSON(map[string]any{"profile": profileName, "revoked": false}))
		},
	}
}

func (s *commandState) resolvedConfigPath() (string, error) {
	if s.configPath != "" {
		return s.configPath, nil
	}
	if path := os.Getenv("ANSWER_CONFIG"); path != "" {
		return path, nil
	}
	return DefaultConfigPath()
}

func (s *commandState) client() (*Client, error) {
	path, err := s.resolvedConfigPath()
	if err != nil {
		return nil, err
	}
	config, err := LoadConfig(path)
	if err != nil {
		return nil, err
	}
	profile, err := config.Resolve(s.profile, Environment{
		Profile: os.Getenv("ANSWER_PROFILE"),
		Server:  os.Getenv("ANSWER_SERVER"),
		Token:   os.Getenv("ANSWER_TOKEN"),
	})
	if err != nil {
		return nil, err
	}
	return NewClient(profile, s.options.HTTPClient)
}

func mustJSON(value any) json.RawMessage {
	content, _ := json.Marshal(value)
	return content
}

func writeSuccess(output io.Writer, data json.RawMessage) error {
	result := struct {
		OK   bool            `json:"ok"`
		Data json.RawMessage `json:"data"`
	}{OK: true, Data: data}
	return json.NewEncoder(output).Encode(result)
}

func (s *commandState) writeError(err error) error {
	result := struct {
		OK    bool     `json:"ok"`
		Error cliError `json:"error"`
	}{OK: false, Error: normalizeError(err)}
	_ = json.NewEncoder(s.options.Stdout).Encode(result)
	return err
}

type cliError struct {
	Type         string          `json:"type"`
	HTTPStatus   int             `json:"http_status,omitempty"`
	ServerReason string          `json:"server_reason,omitempty"`
	Message      string          `json:"message"`
	Data         json.RawMessage `json:"data,omitempty"`
}

func ExitCode(err error) int {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return 2
	}
	if apiErr.OutcomeUnknown || apiErr.HTTPStatus == 0 || apiErr.HTTPStatus >= 500 {
		return 6
	}
	if apiErr.HTTPStatus == http.StatusUnauthorized {
		return 3
	}
	if apiErr.HTTPStatus == http.StatusForbidden {
		return 4
	}
	return 5
}

func normalizeError(err error) cliError {
	result := cliError{Type: "client_error", Message: err.Error()}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return result
	}
	result.HTTPStatus = apiErr.HTTPStatus
	result.ServerReason = apiErr.Reason
	result.Message = apiErr.Message
	result.Data = apiErr.Data
	switch {
	case apiErr.OutcomeUnknown:
		result.Type = "outcome_unknown"
	case apiErr.Reason == "error.object.captcha_verification_failed":
		result.Type = "captcha_required"
	case apiErr.HTTPStatus == http.StatusUnauthorized:
		result.Type = "invalid_token"
	case apiErr.Reason == "error.personal_access_token.insufficient_scope":
		result.Type = "insufficient_token_scope"
	case apiErr.HTTPStatus == http.StatusForbidden:
		result.Type = "permission_denied"
	default:
		result.Type = "server_error"
	}
	if result.Message == "" {
		result.Message = fmt.Sprintf("Answer request failed with HTTP %d", result.HTTPStatus)
	}
	return result
}

func Execute(ctx context.Context, options Options, args []string) error {
	command := NewRootCommand(options)
	command.SetArgs(args)
	return command.ExecuteContext(ctx)
}
