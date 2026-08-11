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

package tag_common

import "testing"

// The bug: the search term was formatted into LOWER(%s), which put the function
// name into the value rather than applying it to the column, so the query became
// slug_name LIKE '%LOWER(coco)%' and matched nothing. Only display_name did any
// work, and that is case-sensitive on Postgres -- so typing a tag the way tags
// are actually written returned "no such tag".
func TestSearchTermIsLoweredNotWrapped(t *testing.T) {
	for _, in := range []string{"Coco", "COCO", "coco"} {
		got := searchTermForTag(in)
		if got != "coco" {
			t.Errorf("searchTermForTag(%q) = %q, want %q", in, got, "coco")
		}
	}
}
