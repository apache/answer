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

package schema

import "slices"

type UpdateReactionReq struct {
	ObjectID string `validate:"required" json:"object_id"`
	Emoji    string `validate:"required,oneof=heart smile frown" json:"emoji"`
	Reaction string `validate:"required,oneof=activate deactivate" json:"reaction"`
	UserID   string `json:"-"`
}

type GetReactionReq struct {
	ObjectID string `validate:"required" form:"object_id"`
	UserID   string `json:"-"`
}

// ReactionsSummaryMeta reactions summary meta
type ReactionsSummaryMeta struct {
	Reactions []*ReactionSummaryMeta `json:"reactions"`
}

// ReactionSummaryMeta reaction summary meta
type ReactionSummaryMeta struct {
	Emoji   string   `json:"emoji"`
	UserIDs []string `json:"user_ids"`
}

// AddReactionSummary add user operation to reaction summary
func (r *ReactionsSummaryMeta) AddReactionSummary(emoji, userID string) {
	for _, reaction := range r.Reactions {
		if reaction.Emoji != emoji {
			continue
		}
		exist := slices.Contains(reaction.UserIDs, userID)
		if !exist {
			reaction.UserIDs = append(reaction.UserIDs, userID)
		}
		return
	}
	r.Reactions = append(r.Reactions, &ReactionSummaryMeta{
		Emoji:   emoji,
		UserIDs: []string{userID},
	})
}

// RemoveReactionSummary remove user operation from reaction summary
func (r *ReactionsSummaryMeta) RemoveReactionSummary(emoji, userID string) {
	updatedReactions := make([]*ReactionSummaryMeta, 0)
	for _, reaction := range r.Reactions {
		if reaction.Emoji != emoji && len(reaction.UserIDs) > 0 {
			updatedReactions = append(updatedReactions, reaction)
			continue
		}
		updatedUserIDs := make([]string, 0, len(r.Reactions))
		for _, id := range reaction.UserIDs {
			if id != userID {
				updatedUserIDs = append(updatedUserIDs, id)
			}
		}
		if len(updatedUserIDs) > 0 {
			reaction.UserIDs = updatedUserIDs
			updatedReactions = append(updatedReactions, reaction)
		}
	}
	r.Reactions = updatedReactions
}

// CheckUserInReactionSummary check user's operation if in reaction summary
func (r *ReactionsSummaryMeta) CheckUserInReactionSummary(emoji, userID string) bool {
	for _, reaction := range r.Reactions {
		if reaction.Emoji != emoji {
			continue
		}
		if slices.Contains(reaction.UserIDs, userID) {
			return true
		}
	}
	return false
}

// GetReactionByObjectIdResp get reaction by object id response
type GetReactionByObjectIdResp struct {
	ReactionSummary []*ReactionRespItem `json:"reaction_summary"`
}

// ReactionRespItem reaction response item
type ReactionRespItem struct {
	// Emoji is the reaction emoji
	Emoji string `json:"emoji"`
	// Count is the number of users who reacted
	Count int `json:"count"`
	// Tooltip is the user's name who reacted
	Tooltip string `json:"tooltip"`
	// IsActive is if current user has reacted
	IsActive bool `json:"is_active"`
}
