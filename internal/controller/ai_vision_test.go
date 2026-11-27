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
	"bytes"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/sashabaranov/go-openai"
)

// minimalPNGBytes builds a tiny blob that starts with the canonical PNG file
// signature at runtime. The validation inspects the declared MIME type, base64
// integrity, size and image magic bytes, so a fully decodable image is not
// required; we construct it on the fly instead of embedding a base64 blob in
// the source.
func minimalPNGBytes() []byte {
	sig := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A} // \x89PNG\r\n\x1a\n
	return append(sig, bytes.Repeat([]byte{0x00}, 40)...)
}

func jpegBytes() []byte {
	return append([]byte{0xFF, 0xD8, 0xFF, 0xE0}, bytes.Repeat([]byte{0x00}, 40)...)
}

func webpBytes() []byte {
	b := append([]byte("RIFF"), []byte{0x28, 0x00, 0x00, 0x00}...)
	return append(append(b, []byte("WEBP")...), bytes.Repeat([]byte{0x00}, 32)...)
}

func dataURL() string {
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(minimalPNGBytes())
}

func TestValidateAndPrepareImagesOK(t *testing.T) {
	msg, err := ValidateAndPrepareImages("看这张图", []string{dataURL()})
	if err != nil {
		t.Fatal(err)
	}
	if len(msg.MultiContent) != 2 {
		t.Fatalf("expect 2 parts got %d", len(msg.MultiContent))
	}
	if msg.MultiContent[0].Type != openai.ChatMessagePartTypeText ||
		msg.MultiContent[0].Text != "看这张图" {
		t.Fatalf("bad text part: %+v", msg.MultiContent[0])
	}
	if msg.MultiContent[1].ImageURL.URL != dataURL() {
		t.Fatalf("bad image url part")
	}
}

func TestValidateAndPrepareImagesNoImages(t *testing.T) {
	msg, err := ValidateAndPrepareImages("纯文本", nil)
	if err != nil {
		t.Fatal(err)
	}
	if msg.Content != "纯文本" || msg.MultiContent != nil {
		t.Fatalf("expect plain content message, got %+v", msg)
	}
}

func TestValidateAndPrepareImagesMagicBytes(t *testing.T) {
	// Matching signatures of the accepted formats pass.
	for _, tc := range []struct {
		mime string
		raw  []byte
	}{
		{"png", minimalPNGBytes()},
		{"jpeg", jpegBytes()},
		{"webp", webpBytes()},
	} {
		url := "data:image/" + tc.mime + ";base64," + base64.StdEncoding.EncodeToString(tc.raw)
		if _, err := ValidateAndPrepareImages("hi", []string{url}); err != nil {
			t.Fatalf("expect %s payload accepted, got %v", tc.mime, err)
		}
	}
	// A declared image MIME type with non-image content is rejected.
	text := "data:image/png;base64," + base64.StdEncoding.EncodeToString([]byte("<script>alert(1)</script>"))
	if _, err := ValidateAndPrepareImages("hi", []string{text}); err == nil ||
		!strings.Contains(err.Error(), "not a valid") {
		t.Fatalf("expect non-image content error, got %v", err)
	}
}

func TestValidateAndPrepareImagesLimits(t *testing.T) {
	five := []string{dataURL(), dataURL(), dataURL(), dataURL(), dataURL()}
	if _, err := ValidateAndPrepareImages("hi", five); err == nil ||
		!strings.Contains(err.Error(), "at most") {
		t.Fatalf("expect limit error, got %v", err)
	}
	if _, err := ValidateAndPrepareImages("", []string{"http://a.com/x.png"}); err == nil {
		t.Fatal("expect reject non-https URL")
	}
	if _, err := ValidateAndPrepareImages("", []string{"data:image/gif;base64," + base64.StdEncoding.EncodeToString(minimalPNGBytes())}); err == nil {
		t.Fatal("expect reject unsupported mime")
	}
	junk := base64.StdEncoding.EncodeToString(make([]byte, 5<<20))
	if _, err := ValidateAndPrepareImages("", []string{"data:image/png;base64," + junk}); err == nil ||
		!strings.Contains(err.Error(), "too large") {
		t.Fatalf("expect size error, got %v", err)
	}
	if _, err := ValidateAndPrepareImages("", []string{"data:text/html;base64," + junk}); err == nil {
		t.Fatal("expect reject non-image mime")
	}
}
