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

package gravatar

import (
	"github.com/apache/answer/internal/base/constant"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAvatarURL(t *testing.T) {
	type args struct {
		email string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "answer@answer.com",
			args: args{email: "answer@answer.com"},
			want: "https://www.gravatar.com/avatar/7296942c1f63d97f6c124705142009867638f7b3dbcdadd0cb1bcb40e427eb8e",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, GetAvatarURL(constant.DefaultGravatarBaseURL, tt.args.email))
		})
	}
}

func TestResize(t *testing.T) {
	type args struct {
		originalAvatarURL string
		sizePixel         int
	}
	tests := []struct {
		name                 string
		args                 args
		wantResizedAvatarURL string
	}{
		{
			name: "original url",
			args: args{
				originalAvatarURL: "https://www.gravatar.com/avatar/b2be4e4438f08a5e885be8de5f41fdd7",
				sizePixel:         128,
			},
			wantResizedAvatarURL: "https://www.gravatar.com/avatar/b2be4e4438f08a5e885be8de5f41fdd7?s=128",
		},
		{
			name: "already resized url",
			args: args{
				originalAvatarURL: "https://www.gravatar.com/avatar/b2be4e4438f08a5e885be8de5f41fdd7?s=128",
				sizePixel:         64,
			},
			wantResizedAvatarURL: "https://www.gravatar.com/avatar/b2be4e4438f08a5e885be8de5f41fdd7?s=64",
		},
		{
			name: "empty url",
			args: args{
				originalAvatarURL: "",
				sizePixel:         64,
			},
			wantResizedAvatarURL: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.wantResizedAvatarURL, Resize(tt.args.originalAvatarURL, tt.args.sizePixel), "Resize(%v, %v)", tt.args.originalAvatarURL, tt.args.sizePixel)
		})
	}
}
