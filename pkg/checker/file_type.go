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

package checker

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/segmentfault/pacman/log"
	"golang.org/x/image/webp"
)

// IsUnAuthorizedExtension check whether the file extension is not in the allowedExtensions
// WANING Only checks the file extension is not reliable, but `http.DetectContentType` and `mimetype` are not reliable for all file types.
func IsUnAuthorizedExtension(fileName string, allowedExtensions []string) bool {
	ext := strings.ToLower(strings.Trim(filepath.Ext(fileName), "."))
	return !slices.Contains(allowedExtensions, ext)
}

// DecodeAndCheckImageFile currently answers support image type is
// `image/jpeg, image/jpg, image/png, image/gif, image/webp`
func DecodeAndCheckImageFile(localFilePath string, maxImageMegapixel int) bool {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(localFilePath), "."))
	switch ext {
	case "jpg", "jpeg", "png", "gif":
		if !decodeAndCheckImageFile(localFilePath, maxImageMegapixel, ext, formatSpecificConfigCheck) {
			return false
		}
		if !decodeAndCheckImageFile(localFilePath, maxImageMegapixel, ext, formatSpecificImageCheck) {
			return false
		}
	case "webp":
		if !decodeAndCheckImageFile(localFilePath, maxImageMegapixel, ext, webpImageConfigCheck) {
			return false
		}
		if !decodeAndCheckImageFile(localFilePath, maxImageMegapixel, ext, webpImageCheck) {
			return false
		}
	}
	return true
}

func decodeAndCheckImageFile(localFilePath string, maxImageMegapixel int, ext string,
	checker func(file io.Reader, ext string, maxImageMegapixel int) error) bool {
	file, err := os.Open(localFilePath)
	if err != nil {
		log.Errorf("open file error: %v", err)
		return false
	}
	defer func() {
		_ = file.Close()
	}()

	if err = checker(file, ext, maxImageMegapixel); err != nil {
		log.Errorf("check image format error: %v", err)
		return false
	}
	return true
}

// formatSpecificConfigCheck decodes image config using a format-specific decoder
// based on the file extension. This avoids calling image.DecodeConfig() which
// dispatches by magic bytes and can invoke unintended decoders (e.g., TIFF)
// registered by transitive dependencies.
func formatSpecificConfigCheck(file io.Reader, ext string, maxImageMegapixel int) error {
	var config image.Config
	var err error
	switch ext {
	case "jpg", "jpeg":
		config, err = jpeg.DecodeConfig(file)
	case "png":
		config, err = png.DecodeConfig(file)
	case "gif":
		config, err = gif.DecodeConfig(file)
	default:
		return fmt.Errorf("unsupported image format: %s", ext)
	}
	if err != nil {
		return fmt.Errorf("decode image config error: %v", err)
	}
	if imageSizeTooLarge(config, maxImageMegapixel) {
		return fmt.Errorf("image size too large")
	}
	return nil
}

// formatSpecificImageCheck fully decodes the image using a format-specific decoder.
func formatSpecificImageCheck(file io.Reader, ext string, _ int) error {
	var err error
	switch ext {
	case "jpg", "jpeg":
		_, err = jpeg.Decode(file)
	case "png":
		_, err = png.Decode(file)
	case "gif":
		_, err = gif.Decode(file)
	default:
		return fmt.Errorf("unsupported image format: %s", ext)
	}
	if err != nil {
		return fmt.Errorf("decode image error: %v", err)
	}
	return nil
}

func webpImageConfigCheck(file io.Reader, _ string, maxImageMegapixel int) error {
	config, err := webp.DecodeConfig(file)
	if err != nil {
		return fmt.Errorf("decode webp image config error: %v", err)
	}
	if imageSizeTooLarge(config, maxImageMegapixel) {
		return fmt.Errorf("image size too large")
	}
	return nil
}

func webpImageCheck(file io.Reader, _ string, _ int) error {
	_, err := webp.Decode(file)
	if err != nil {
		return fmt.Errorf("decode webp image error: %v", err)
	}
	return nil
}

func imageSizeTooLarge(config image.Config, maxImageMegapixel int) bool {
	return config.Width*config.Height > maxImageMegapixel
}
