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
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
)

const (
	maxImagesPerMessage = 4
	maxImageBytes       = 4 << 20 // decoded size per image: 4MB
)

// allowedImagePrefixes accepts base64 data URLs of web-safe raster formats and
// plain HTTPS image URLs.
var allowedImagePrefixes = []string{
	"data:image/png;base64,",
	"data:image/jpeg;base64,",
	"data:image/webp;base64,",
	"https://",
}

// ValidateAndPrepareImages validates attachments and wraps them together with
// the text into a single MultiContent message. The returned message must be
// used as-is (Content stays empty so it never conflicts with MultiContent).
func ValidateAndPrepareImages(text string, images []string) (*openai.ChatCompletionMessage, error) {
	if len(images) == 0 {
		return &openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: text,
		}, nil
	}
	if len(images) > maxImagesPerMessage {
		return nil, fmt.Errorf("at most %d images per message", maxImagesPerMessage)
	}
	for _, img := range images {
		okPrefix := false
		for _, p := range allowedImagePrefixes {
			if strings.HasPrefix(img, p) {
				okPrefix = true
				break
			}
		}
		if !okPrefix {
			return nil, fmt.Errorf("unsupported image source (use PNG/JPEG/WebP base64 or HTTPS URL)")
		}
		if strings.HasPrefix(img, "data:image/") {
			idx := strings.Index(img, "base64,")
			raw, err := base64.StdEncoding.DecodeString(img[idx+len("base64,"):])
			if err != nil {
				return nil, fmt.Errorf("invalid image data")
			}
			if len(raw) > maxImageBytes {
				return nil, fmt.Errorf("image too large (max 4MB)")
			}
		}
	}

	msg := &openai.ChatCompletionMessage{
		Role:         openai.ChatMessageRoleUser,
		MultiContent: make([]openai.ChatMessagePart, 0, len(images)+1),
	}
	msg.MultiContent = append(msg.MultiContent, openai.ChatMessagePart{
		Type: openai.ChatMessagePartTypeText,
		Text: text,
	})
	for _, img := range images {
		msg.MultiContent = append(msg.MultiContent, openai.ChatMessagePart{
			Type:     openai.ChatMessagePartTypeImageURL,
			ImageURL: &openai.ChatMessageImageURL{URL: img},
		})
	}
	return msg, nil
}
