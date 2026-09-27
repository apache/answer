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

package bulk_delete

import "github.com/apache/answer/internal/schema"

// Execute applies a delete operation once per unique ID and records partial failures.
func Execute(ids []string, deleteFunc func(string) error) *schema.BulkDeleteResp {
	resp := &schema.BulkDeleteResp{
		SucceededIDs: make([]string, 0),
		FailedIDs:    make([]string, 0),
	}
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}

		if err := deleteFunc(id); err != nil {
			resp.FailedIDs = append(resp.FailedIDs, id)
			continue
		}
		resp.SucceededIDs = append(resp.SucceededIDs, id)
	}
	return resp
}
