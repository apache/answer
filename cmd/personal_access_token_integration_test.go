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

package answercmd

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	bundles "github.com/apache/answer/i18n"
	"github.com/apache/answer/internal/base/conf"
	"github.com/apache/answer/internal/base/data"
	serverconfig "github.com/apache/answer/internal/base/server"
	"github.com/apache/answer/internal/base/translator"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/migrations"
	"github.com/apache/answer/internal/router"
	"github.com/apache/answer/internal/schema"
	pat "github.com/apache/answer/internal/service/personal_access_token"
	"github.com/apache/answer/internal/service/service_config"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"
)

// TestPATHTTPServerProcess runs the real application in an isolated test process.
func TestPATHTTPServerProcess(t *testing.T) {
	if os.Getenv("ANSWER_PAT_TEST_PROCESS") != "1" {
		return
	}
	dir := os.Getenv("ANSWER_PAT_TEST_DATA")
	logPath = filepath.Join(dir, "logs")
	os.Args = []string{"answer", "run", "-C", dir}
	Main()
}

func TestPersonalAccessTokenHTTPWorkflow(t *testing.T) {
	api := newPATHTTPFixture(t)
	admin := api.login(t, "admin@example.test")
	member := api.login(t, "member@example.test")
	var security schema.SiteSecurityReq
	api.request(t, http.MethodGet, "/answer/admin/api/siteinfo/security", admin, nil, http.StatusOK, &security)
	require.False(t, security.PersonalAccessTokensEnabled)

	create := func(session string, scopes []string, status int) schema.PersonalAccessTokenCreateResp {
		t.Helper()
		var result schema.PersonalAccessTokenCreateResp
		api.request(t, http.MethodPost, "/answer/api/v1/personal-access-tokens", session, schema.PersonalAccessTokenCreateReq{
			Name: "http-fixture", Scopes: scopes, ExpiresAt: time.Now().Add(time.Hour).Unix(),
		}, status, &result)
		return result
	}
	scopes := []string{pat.ScopeQuestionRead, pat.ScopeQuestionCreate, pat.ScopeAnswerRead, pat.ScopeAnswerCreate, pat.ScopeVoteWrite}
	create(member, scopes, http.StatusForbidden)
	security.PersonalAccessTokensEnabled = true
	api.request(t, http.MethodPut, "/answer/admin/api/siteinfo/security", admin, security, http.StatusOK, nil)
	adminPAT := create(admin, scopes, http.StatusOK)
	memberPAT := create(member, scopes, http.StatusOK)
	readPAT := create(admin, []string{pat.ScopeQuestionRead}, http.StatusOK)
	adminBearer, memberBearer := "Bearer "+adminPAT.Token, "Bearer "+memberPAT.Token
	for bearer, owner := range map[string]string{adminBearer: "1", memberBearer: "2"} {
		var current schema.PersonalAccessTokenCurrentResp
		api.request(t, http.MethodGet, "/answer/api/v1/personal-access-tokens/current", bearer, nil, http.StatusOK, &current)
		require.Equal(t, owner, current.User.ID)
	}

	var question struct {
		ID string `json:"id"`
	}
	questionBody := map[string]any{
		"title":   "How does delegated PAT access preserve account permissions?",
		"content": "An eligible fixture user creates this question using its own scoped personal access token.",
		"tags":    []map[string]string{{"slug_name": "pat-fixture", "display_name": "PAT fixture"}},
	}
	api.request(t, http.MethodPost, "/answer/api/v1/question", "Bearer "+readPAT.Token, questionBody, http.StatusForbidden, nil)
	api.request(t, http.MethodPost, "/answer/api/v1/question", memberBearer, questionBody, http.StatusOK, &question)
	require.NotEmpty(t, question.ID)
	var answer schema.GetAnswerInfoResp
	api.request(t, http.MethodPost, "/answer/api/v1/answer", adminBearer, map[string]string{
		"question_id": question.ID, "content": "The token limits routes while the existing account permissions continue to govern the action.",
	}, http.StatusOK, &answer)
	require.NotNil(t, answer.Info)

	// Each owner can vote on the other user's content, then retract the vote.
	for direction, count := range map[string]int{"up": 1, "down": -1} {
		for bearer, objectID := range map[string]string{adminBearer: question.ID, memberBearer: answer.Info.ID} {
			var vote schema.VoteResp
			api.request(t, http.MethodPost, "/answer/api/v1/vote/"+direction, bearer,
				map[string]any{"object_id": objectID, "is_cancel": false}, http.StatusOK, &vote)
			require.EqualValues(t, count, vote.Votes)
			api.request(t, http.MethodPost, "/answer/api/v1/vote/"+direction, bearer,
				map[string]any{"object_id": objectID, "is_cancel": true}, http.StatusOK, &vote)
			require.Zero(t, vote.Votes)
		}
	}
	api.request(t, http.MethodPost, "/answer/api/v1/vote/up", memberBearer,
		map[string]string{"object_id": question.ID}, http.StatusBadRequest, nil)
	api.request(t, http.MethodGet, "/answer/api/v1/personal-access-tokens", adminBearer, nil, http.StatusForbidden, nil)
	api.request(t, http.MethodGet, "/answer/api/v1/question/info?id="+question.ID, memberPAT.Token, nil, http.StatusUnauthorized, nil)
	api.request(t, http.MethodGet, "/answer/api/v1/question/info?id="+question.ID+"&Authorization="+url.QueryEscape(memberPAT.Token), "", nil, http.StatusUnauthorized, nil)

	// A session belonging to another user cannot revoke this owner's PAT.
	revokePath := "/answer/api/v1/personal-access-tokens?id=" + strconv.FormatInt(memberPAT.ID, 10)
	api.request(t, http.MethodDelete, revokePath, admin, nil, http.StatusOK, nil)
	api.request(t, http.MethodGet, "/answer/api/v1/personal-access-tokens/current", memberBearer, nil, http.StatusOK, nil)
	security.PersonalAccessTokensEnabled = false
	api.request(t, http.MethodPut, "/answer/admin/api/siteinfo/security", admin, security, http.StatusOK, nil)
	for _, bearer := range []string{adminBearer, memberBearer} {
		api.request(t, http.MethodGet, "/answer/api/v1/personal-access-tokens/current", bearer, nil, http.StatusForbidden, nil)
	}
	security.PersonalAccessTokensEnabled = true
	api.request(t, http.MethodPut, "/answer/admin/api/siteinfo/security", admin, security, http.StatusOK, nil)
	api.request(t, http.MethodGet, "/answer/api/v1/personal-access-tokens/current", memberBearer, nil, http.StatusOK, nil)
	api.request(t, http.MethodDelete, revokePath, member, nil, http.StatusOK, nil)
	api.request(t, http.MethodGet, "/answer/api/v1/personal-access-tokens/current", memberBearer, nil, http.StatusUnauthorized, nil)
	security.PersonalAccessTokensEnabled = false
	api.request(t, http.MethodPut, "/answer/admin/api/siteinfo/security", admin, security, http.StatusOK, nil)
	security.PersonalAccessTokensEnabled = true
	api.request(t, http.MethodPut, "/answer/admin/api/siteinfo/security", admin, security, http.StatusOK, nil)
	api.request(t, http.MethodGet, "/answer/api/v1/personal-access-tokens/current", memberBearer, nil, http.StatusUnauthorized, nil)
	api.request(t, http.MethodGet, "/answer/api/v1/personal-access-tokens/current", adminBearer, nil, http.StatusOK, nil)
}

type patHTTPFixture struct {
	url    string
	client *http.Client
}

const patFixturePassword = "fixture-password-only"

func (f *patHTTPFixture) login(t *testing.T, email string) string {
	t.Helper()
	var result schema.UserLoginResp
	f.request(t, http.MethodPost, "/answer/api/v1/user/login/email", "", map[string]string{
		"e_mail": email, "pass": patFixturePassword,
	}, http.StatusOK, &result)
	f.request(t, http.MethodPost, "/answer/api/v1/user/reauthenticate", result.AccessToken,
		map[string]string{"password": patFixturePassword}, http.StatusOK, nil)
	return result.AccessToken
}

func (f *patHTTPFixture) request(t *testing.T, method, path, token string, body any, status int, out any) {
	t.Helper()
	var payload bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&payload).Encode(body))
	}
	request, err := http.NewRequestWithContext(t.Context(), method, f.url+path, &payload)
	require.NoError(t, err)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	request.Header.Set("Authorization", token)
	response, err := f.client.Do(request)
	require.NoError(t, err)
	defer func() { require.NoError(t, response.Body.Close()) }()
	var envelope struct {
		Reason string          `json:"reason"`
		Data   json.RawMessage `json:"data"`
	}
	require.NoError(t, json.NewDecoder(response.Body).Decode(&envelope))
	require.Equal(t, status, response.StatusCode, "%s %s: %s", method, path, envelope.Reason)
	if out != nil && status == http.StatusOK {
		require.NoError(t, json.Unmarshal(envelope.Data, out))
	}
}

func newPATHTTPFixture(t *testing.T) *patHTTPFixture {
	t.Helper()
	dir := t.TempDir()
	for _, name := range []string{"conf", "cache", "i18n", "uploads", "logs"} {
		require.NoError(t, os.MkdirAll(filepath.Join(dir, name), 0o700))
	}
	entries, err := bundles.I18n.ReadDir(".")
	require.NoError(t, err)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		content, readErr := bundles.I18n.ReadFile(entry.Name())
		require.NoError(t, readErr)
		require.NoError(t, os.WriteFile(filepath.Join(dir, "i18n", entry.Name()), content, 0o600))
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	address := listener.Addr().String()
	require.NoError(t, listener.Close())
	dbConfig := &data.Database{Driver: "sqlite3", Connection: filepath.Join(dir, "answer.db")}
	db, err := data.NewDB(false, dbConfig)
	require.NoError(t, err)
	require.NoError(t, migrations.NewMentor(t.Context(), db, &migrations.InitNeedUserInputData{
		Language: "en_US", SiteName: "PAT fixture", SiteURL: "http://" + address,
		ContactEmail: "admin@example.test", AdminName: "pat-admin",
		AdminEmail: "admin@example.test", AdminPassword: patFixturePassword,
		ExternalContentDisplay: "always_display",
	}).InitDB())
	password, err := bcrypt.GenerateFromPassword([]byte(patFixturePassword), bcrypt.MinCost)
	require.NoError(t, err)
	_, err = db.Insert(&entity.User{ID: "2", Username: "pat-member", DisplayName: "PAT member",
		EMail: "member@example.test", Pass: string(password), MailStatus: entity.EmailStatusAvailable,
		Status: entity.UserStatusAvailable, Rank: 10000})
	require.NoError(t, err)
	_, err = db.Insert(&entity.UserRoleRel{UserID: "2", RoleID: 1})
	require.NoError(t, err)
	_, err = db.Insert(&entity.Tag{ID: "10030000000009999", SlugName: "pat-fixture", DisplayName: "PAT fixture",
		UserID: "1", Status: entity.TagStatusAvailable, Recommend: true})
	require.NoError(t, err)
	require.NoError(t, db.Close())
	config, err := yaml.Marshal(&conf.AllConfig{
		Server:    &conf.Server{HTTP: &serverconfig.HTTP{Addr: address}},
		Data:      &conf.Data{Database: dbConfig, Cache: &data.CacheConf{FilePath: filepath.Join(dir, "cache", "cache.db")}},
		I18n:      &translator.I18n{BundleDir: filepath.Join(dir, "i18n")},
		Swaggerui: &router.SwaggerConfig{}, UI: &serverconfig.UI{},
		ServiceConfig: &service_config.ServiceConfig{UploadPath: filepath.Join(dir, "uploads"),
			CleanOrphanUploadsPeriodHours: 48, PurgeDeletedFilesPeriodDays: 30},
	})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "conf", "config.yaml"), config, 0o600))
	ctx, cancel := context.WithCancel(t.Context())
	process := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestPATHTTPServerProcess$")
	process.Env = append(os.Environ(), "ANSWER_PAT_TEST_PROCESS=1", "ANSWER_PAT_TEST_DATA="+dir,
		"SITE_ADDR=", "SWAGGER_HOST=", "SWAGGER_ADDRESS_PORT=", "ANSWER_TEMPLATE_PATH=")
	logFile, err := os.Create(filepath.Join(dir, "server.log"))
	require.NoError(t, err)
	process.Stdout, process.Stderr = logFile, logFile
	require.NoError(t, process.Start())
	t.Cleanup(func() {
		cancel()
		_ = process.Wait()
		require.NoError(t, logFile.Close())
		if t.Failed() {
			logs, readErr := os.ReadFile(logFile.Name())
			require.NoError(t, readErr)
			t.Log(string(logs))
		}
	})
	fixture := &patHTTPFixture{url: "http://" + address, client: &http.Client{Timeout: 5 * time.Second}}
	require.Eventually(t, func() bool {
		response, requestErr := fixture.client.Get(fixture.url + "/healthz")
		if requestErr != nil {
			return false
		}
		_ = response.Body.Close()
		return response.StatusCode == http.StatusOK
	}, 20*time.Second, 100*time.Millisecond)
	return fixture
}
