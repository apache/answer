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

package badge

import (
	"testing"

	"github.com/apache/answer/internal/entity"
	"github.com/stretchr/testify/require"
)

// a ladder of single award badges: thresholds 1, 6, 15, 80 and 200 accepted answers
func tierBadges() []*entity.Badge {
	return []*entity.Badge{
		{ID: "10000000000000001", Name: "first", Single: entity.BadgeSingleAward, Param: `{"amount":"1"}`},
		{ID: "10000000000000002", Name: "second", Single: entity.BadgeSingleAward, Param: `{"amount":"6"}`},
		{ID: "10000000000000003", Name: "third", Single: entity.BadgeSingleAward, Param: `{"amount":"15"}`},
		{ID: "10000000000000004", Name: "fourth", Single: entity.BadgeSingleAward, Param: `{"amount":"80"}`},
		{ID: "10000000000000005", Name: "fifth", Single: entity.BadgeSingleAward, Param: `{"amount":"200"}`},
	}
}

func TestHighestTierReached_OnlyTheTopTierIsHandedOut(t *testing.T) {
	tiers := tierBadges()

	require.Nil(t, highestTierReached(tiers, 0), "no accepted answers, no badge")
	require.Equal(t, "first", highestTierReached(tiers, 1).Name)
	require.Equal(t, "second", highestTierReached(tiers, 6).Name)
	require.Equal(t, "second", highestTierReached(tiers, 14).Name)
	require.Equal(t, "third", highestTierReached(tiers, 15).Name, "crossing several thresholds at once gives one badge")
	require.Equal(t, "fourth", highestTierReached(tiers, 199).Name)
	require.Equal(t, "fifth", highestTierReached(tiers, 200).Name)
}

func TestHighestHeldTierOf_KnowsTheTierTheUserAlreadyHas(t *testing.T) {
	tiers := tierBadges()

	require.Equal(t, int64(0), highestHeldTierOf(tiers, map[string]bool{}))
	require.Equal(t, int64(80), highestHeldTierOf(tiers, map[string]bool{"10000000000000004": true}),
		"granted by an administrator")
	require.Equal(t, int64(80), highestHeldTierOf(tiers, map[string]bool{
		"10000000000000004": true, "10000000000000001": true,
	}), "the highest one held counts, not the last one granted")
}

// The bug itself: a user holding the fourth tier was awarded the first one as soon as an answer of
// theirs was accepted.
func TestTierNotAwardedBelowTheTierAlreadyHeld(t *testing.T) {
	tiers := tierBadges()
	held := highestHeldTierOf(tiers, map[string]bool{"10000000000000004": true})

	best := highestTierReached(tiers, 1) // their first answer is accepted
	require.Equal(t, "first", best.Name)
	require.False(t, best.GetIntParam("amount") > held, "the lower badge must not be awarded")

	best = highestTierReached(tiers, 200) // they reach the top threshold
	require.Equal(t, "fifth", best.Name)
	require.True(t, best.GetIntParam("amount") > held, "a higher tier is still awarded")

	// and the top tier is not awarded twice
	heldTop := highestHeldTierOf(tiers, map[string]bool{"10000000000000005": true})
	require.False(t, highestTierReached(tiers, 200).GetIntParam("amount") > heldTop)
}

func TestSplitSingleAwardBadges_MultiAwardKeepsOldBehaviour(t *testing.T) {
	badges := append(tierBadges(), &entity.Badge{
		ID: "10000000000000006", Name: "per-post", Single: entity.BadgeMultiAward, Param: `{"amount":"3"}`,
	})

	tiers, others := splitSingleAwardBadges(badges)
	require.Len(t, tiers, 5)
	require.Len(t, others, 1)
	require.Equal(t, "per-post", others[0].Name)
}
