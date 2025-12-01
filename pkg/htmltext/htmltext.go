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

package htmltext

import (
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/Machiel/slugify"
	"github.com/apache/answer/pkg/checker"
	"github.com/apache/answer/pkg/converter"
	strip "github.com/grokify/html-strip-tags-go"
	"github.com/mozillazg/go-pinyin"
)

var (
	reCode         = regexp.MustCompile(`(?ism)<(pre)>.*<\/pre>`)
	reCodeReplace  = "{code...}"
	reLink         = regexp.MustCompile(`(?ism)<a.*?[^<]>(.*)?<\/a>`)
	reLinkReplace  = " [$1] "
	reSpace        = regexp.MustCompile(` +`)
	reSpaceReplace = " "

	spaceReplacer = strings.NewReplacer(
		"\n", " ",
		"\r", " ",
		"\t", " ",
	)
)

// ClearText clear HTML, get the clear text
func ClearText(html string) string {
	if html == "" {
		return html
	}

	html = reCode.ReplaceAllString(html, reCodeReplace)
	html = reLink.ReplaceAllString(html, reLinkReplace)

	text := spaceReplacer.Replace(strip.StripTags(html))

	// replace multiple spaces to one space
	return strings.TrimSpace(reSpace.ReplaceAllString(text, reSpaceReplace))
}

func UrlTitle(title string) (text string) {
	title = convertChinese(title)
	title = clearEmoji(title)
	title = slugify.Slugify(title)
	title = url.QueryEscape(title)
	title = cutLongTitle(title)
	if len(title) == 0 {
		title = "topic"
	}
	return title
}

func clearEmoji(s string) string {
	var ret strings.Builder
	rs := []rune(s)
	for i := range rs {
		if len(string(rs[i])) != 4 {
			ret.WriteString(string(rs[i]))
		}
	}
	return ret.String()
}

func convertChinese(content string) string {
	has := checker.IsChinese(content)
	if !has {
		return content
	}
	return strings.Join(pinyin.LazyConvert(content, nil), "-")
}

func cutLongTitle(title string) string {
	maxBytes := 150
	if len(title) <= maxBytes {
		return title
	}

	truncated := title[:maxBytes]
	for len(truncated) > 0 && !utf8.ValidString(truncated) {
		truncated = truncated[:len(truncated)-1]
	}
	return truncated
}

// FetchExcerpt return the excerpt from the HTML string
func FetchExcerpt(html, trimMarker string, limit int) (text string) {
	return FetchRangedExcerpt(html, trimMarker, 0, limit)
}

// findFirstMatchedWord returns the first matched word and its index
func findFirstMatchedWord(text string, words []string) (string, int) {
	if len(text) == 0 || len(words) == 0 {
		return "", 0
	}

	words = converter.UniqueArray(words)
	firstWord := ""
	firstIndex := len(text)

	for _, word := range words {
		if idx := strings.Index(text, word); idx != -1 && idx < firstIndex {
			firstIndex = idx
			firstWord = word
		}
	}

	if firstIndex != len(text) {
		return firstWord, firstIndex
	}

	return "", 0
}

// getRuneRange returns the valid begin and end indexes of the runeText
func getRuneRange(runeText []rune, offset, limit int) (begin, end int) {
	runeLen := len(runeText)

	limit = min(runeLen, max(0, limit))
	begin = min(runeLen, max(0, offset))
	end = min(runeLen, begin+limit)

	return
}

// FetchRangedExcerpt returns a ranged excerpt from the HTML string.
// Note: offset is a rune index, not a byte index
func FetchRangedExcerpt(html, trimMarker string, offset int, limit int) (text string) {
	if len(html) == 0 {
		text = html
		return
	}

	runeText := []rune(ClearText(html))
	begin, end := getRuneRange(runeText, offset, limit)
	text = string(runeText[begin:end])

	if begin > 0 {
		text = trimMarker + text
	}
	if end < len(runeText) {
		text += trimMarker
	}

	return
}

// FetchMatchedExcerpt returns the matched excerpt according to the words
func FetchMatchedExcerpt(html string, words []string, trimMarker string, trimLength int) string {
	text := ClearText(html)
	matchedWord, matchedIndex := findFirstMatchedWord(text, words)
	runeIndex := utf8.RuneCountInString(text[0:matchedIndex])

	trimLength = max(0, trimLength)
	runeOffset := runeIndex - trimLength
	runeLimit := trimLength + trimLength + utf8.RuneCountInString(matchedWord)

	textRuneCount := utf8.RuneCountInString(text)
	if runeOffset+runeLimit > textRuneCount {
		// Reserved extra chars before the matched word
		runeOffset = textRuneCount - runeLimit
	}

	return FetchRangedExcerpt(html, trimMarker, runeOffset, runeLimit)
}

func GetPicByUrl(url string) string {
	res, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer func() {
		_ = res.Body.Close()
	}()
	pix, err := io.ReadAll(res.Body)
	if err != nil {
		return ""
	}
	return string(pix)
}
