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

import (
	"errors"
	"reflect"
	"testing"
)

func TestExecute(t *testing.T) {
	called := make([]string, 0)
	result := Execute([]string{"1", "2", "1", "3"}, func(id string) error {
		called = append(called, id)
		if id == "2" {
			return errors.New("delete failed")
		}
		return nil
	})

	if !reflect.DeepEqual(called, []string{"1", "2", "3"}) {
		t.Fatalf("unexpected calls: %v", called)
	}
	if !reflect.DeepEqual(result.SucceededIDs, []string{"1", "3"}) {
		t.Fatalf("unexpected successful IDs: %v", result.SucceededIDs)
	}
	if !reflect.DeepEqual(result.FailedIDs, []string{"2"}) {
		t.Fatalf("unexpected failed IDs: %v", result.FailedIDs)
	}
}

func TestExecuteReturnsEmptyArrays(t *testing.T) {
	result := Execute(nil, func(string) error { return nil })
	if result.SucceededIDs == nil || result.FailedIDs == nil {
		t.Fatal("expected initialized result arrays")
	}
}
