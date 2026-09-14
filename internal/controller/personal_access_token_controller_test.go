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
	"time"

	"github.com/stretchr/testify/require"
)

func TestPATCreationRequiresAuthenticationWithinConfiguredWindow(t *testing.T) {
	now := time.Date(2026, time.September, 14, 12, 0, 0, 0, time.UTC)
	require.True(t, isPATAuthenticationRecent(now.Add(-60*time.Minute).Unix(), now, 60))
	require.False(t, isPATAuthenticationRecent(now.Add(-61*time.Minute).Unix(), now, 60))
	require.False(t, isPATAuthenticationRecent(0, now, 60))
	require.False(t, isPATAuthenticationRecent(now.Add(time.Minute).Unix(), now, 60))
}
