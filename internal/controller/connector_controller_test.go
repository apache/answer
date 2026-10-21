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

import "testing"

type stateRequiringConnector struct {
	required bool
}

func (c stateRequiringConnector) ConnectorRequireState() bool {
	return c.required
}

func TestConnectorRequiresState(t *testing.T) {
	tests := []struct {
		name      string
		connector any
		want      bool
	}{
		{name: "legacy connector", connector: struct{}{}, want: false},
		{name: "connector opts out", connector: stateRequiringConnector{required: false}, want: false},
		{name: "connector opts in", connector: stateRequiringConnector{required: true}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := connectorRequiresState(tt.connector); got != tt.want {
				t.Fatalf("connectorRequiresState() = %t, want %t", got, tt.want)
			}
		})
	}
}
