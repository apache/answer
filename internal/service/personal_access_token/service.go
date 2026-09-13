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

package personal_access_token

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/apache/answer/internal/entity"
)

const (
	TokenPrefix = "answer_pat_"

	ScopeQuestionRead   = "question.read"
	ScopeQuestionCreate = "question.create"
	ScopeAnswerRead     = "answer.read"
	ScopeAnswerCreate   = "answer.create"
	ScopeVoteWrite      = "vote.write"
)

var (
	ErrInvalidToken  = errors.New("invalid personal access token")
	ErrInvalidScope  = errors.New("invalid personal access token scope")
	ErrInvalidExpiry = errors.New("invalid personal access token expiry")
)

var allowedScopes = map[string]struct{}{
	ScopeQuestionRead:   {},
	ScopeQuestionCreate: {},
	ScopeAnswerRead:     {},
	ScopeAnswerCreate:   {},
	ScopeVoteWrite:      {},
}

// Repository persists Personal Access Tokens.
type Repository interface {
	Create(ctx context.Context, token *entity.PersonalAccessToken) error
	FindByHash(ctx context.Context, hash string) (*entity.PersonalAccessToken, bool, error)
	ListByUserID(ctx context.Context, userID string) ([]*entity.PersonalAccessToken, error)
	Revoke(ctx context.Context, userID string, id int64, revokedAt time.Time) (bool, error)
	RevokeAllByUserID(ctx context.Context, userID string, revokedAt time.Time) error
}

type Option func(*Service)

// Service manages Personal Access Token lifecycle and authentication.
type Service struct {
	repo           Repository
	now            func() time.Time
	generateSecret func() (string, error)
}

func NewService(repo Repository, options ...Option) *Service {
	s := &Service{
		repo:           repo,
		now:            time.Now,
		generateSecret: randomSecret,
	}
	for _, option := range options {
		option(s)
	}
	return s
}

func WithClock(now func() time.Time) Option {
	return func(service *Service) { service.now = now }
}

func WithSecretGenerator(generate func() (string, error)) Option {
	return func(service *Service) { service.generateSecret = generate }
}

type CreateInput struct {
	UserID    string
	Name      string
	Scopes    []string
	ExpiresAt time.Time
}

type Status string

const (
	StatusActive  Status = "active"
	StatusExpired Status = "expired"
	StatusRevoked Status = "revoked"
)

type TokenInfo struct {
	ID          int64
	UserID      string
	Name        string
	TokenSuffix string
	Scopes      []string
	CreatedAt   time.Time
	ExpiresAt   time.Time
	RevokedAt   time.Time
	Status      Status
}

type CreatedToken struct {
	Token string
	Info  TokenInfo
}

type AuthenticatedToken struct {
	ID        int64
	UserID    string
	Name      string
	Scopes    []string
	ExpiresAt time.Time
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*CreatedToken, error) {
	now := s.now().UTC()
	if input.ExpiresAt.After(now.Add(365*24*time.Hour)) || !input.ExpiresAt.After(now) {
		return nil, ErrInvalidExpiry
	}
	scopes, err := CanonicalScopes(input.Scopes)
	if err != nil {
		return nil, err
	}
	secret, err := s.generateSecret()
	if err != nil {
		return nil, err
	}
	rawToken := TokenPrefix + secret
	scopeJSON, err := json.Marshal(scopes)
	if err != nil {
		return nil, err
	}
	token := &entity.PersonalAccessToken{
		UserID:      input.UserID,
		Name:        strings.TrimSpace(input.Name),
		TokenHash:   Hash(rawToken),
		TokenSuffix: suffix(rawToken),
		Scopes:      string(scopeJSON),
		CreatedAt:   now,
		ExpiresAt:   input.ExpiresAt.UTC(),
	}
	if err := s.repo.Create(ctx, token); err != nil {
		return nil, err
	}
	return &CreatedToken{Token: rawToken, Info: toInfo(token, scopes)}, nil
}

func (s *Service) List(ctx context.Context, userID string) ([]TokenInfo, error) {
	tokens, err := s.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]TokenInfo, 0, len(tokens))
	for _, token := range tokens {
		var scopes []string
		if err := json.Unmarshal([]byte(token.Scopes), &scopes); err != nil {
			return nil, err
		}
		info := toInfo(token, scopes)
		info.Status = statusAt(token, s.now().UTC())
		result = append(result, info)
	}
	return result, nil
}

func (s *Service) Revoke(ctx context.Context, userID string, id int64) error {
	_, err := s.repo.Revoke(ctx, userID, id, s.now().UTC())
	return err
}

func (s *Service) RevokeAll(ctx context.Context, userID string) error {
	return s.repo.RevokeAllByUserID(ctx, userID, s.now().UTC())
}

func (s *Service) Authenticate(ctx context.Context, rawToken string) (*AuthenticatedToken, error) {
	if !strings.HasPrefix(rawToken, TokenPrefix) || len(rawToken) <= len(TokenPrefix)+4 {
		return nil, ErrInvalidToken
	}
	token, found, err := s.repo.FindByHash(ctx, Hash(rawToken))
	if err != nil {
		return nil, err
	}
	if !found || !token.RevokedAt.IsZero() || !token.ExpiresAt.After(s.now().UTC()) {
		return nil, ErrInvalidToken
	}
	var scopes []string
	if err := json.Unmarshal([]byte(token.Scopes), &scopes); err != nil {
		return nil, ErrInvalidToken
	}
	scopes, err = CanonicalScopes(scopes)
	if err != nil {
		return nil, ErrInvalidToken
	}
	return &AuthenticatedToken{
		ID: token.ID, UserID: token.UserID, Name: token.Name,
		Scopes: scopes, ExpiresAt: token.ExpiresAt,
	}, nil
}

func CanonicalScopes(scopes []string) ([]string, error) {
	if len(scopes) == 0 {
		return nil, ErrInvalidScope
	}
	unique := make(map[string]struct{}, len(scopes))
	for _, scope := range scopes {
		if _, ok := allowedScopes[scope]; !ok {
			return nil, ErrInvalidScope
		}
		unique[scope] = struct{}{}
	}
	result := make([]string, 0, len(unique))
	for scope := range unique {
		result = append(result, scope)
	}
	sort.Strings(result)
	return result, nil
}

func Hash(rawToken string) string {
	digest := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(digest[:])
}

func randomSecret() (string, error) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(secret), nil
}

func suffix(token string) string {
	if len(token) <= 4 {
		return token
	}
	return token[len(token)-4:]
}

func toInfo(token *entity.PersonalAccessToken, scopes []string) TokenInfo {
	return TokenInfo{
		ID: token.ID, UserID: token.UserID, Name: token.Name,
		TokenSuffix: token.TokenSuffix, Scopes: scopes, CreatedAt: token.CreatedAt,
		ExpiresAt: token.ExpiresAt, RevokedAt: token.RevokedAt,
	}
}

func statusAt(token *entity.PersonalAccessToken, now time.Time) Status {
	if !token.RevokedAt.IsZero() {
		return StatusRevoked
	}
	if !token.ExpiresAt.After(now) {
		return StatusExpired
	}
	return StatusActive
}
