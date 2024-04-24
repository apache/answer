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

	"github.com/Chain-Zhang/pinyin"
	"github.com/Machiel/slugify"
	strip "github.com/grokify/html-strip-tags-go"

	"github.com/apache/incubator-answer/pkg/checker"
	"github.com/apache/incubator-answer/pkg/converter"
)

// min() and max() can be removed starting from Go1.21
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ClearText clear HTML, get the clear text
func ClearText(html string) (text string) {
	if len(html) == 0 {
		text = html
		return
	}

	var (
		re        *regexp.Regexp
		codeReg   = `(?ism)<(pre)>.*<\/pre>`
		codeRepl  = "{code...}"
		linkReg   = `(?ism)<a.*?[^<]>(.*)?<\/a>`
		linkRepl  = " [$1] "
		spaceReg  = ` +`
		spaceRepl = " "
	)
	re = regexp.MustCompile(codeReg)
	html = re.ReplaceAllString(html, codeRepl)

	re = regexp.MustCompile(linkReg)
	html = re.ReplaceAllString(html, linkRepl)

	text = strings.NewReplacer(
		"\n", " ",
		"\r", " ",
		"\t", " ",
	).Replace(strip.StripTags(html))

	// replace multiple spaces to one space
	re = regexp.MustCompile(spaceReg)
	text = strings.TrimSpace(re.ReplaceAllString(text, spaceRepl))
	return
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
	ret := ""
	rs := []rune(s)
	for i := 0; i < len(rs); i++ {
		if len(string(rs[i])) != 4 {
			ret += string(rs[i])
		}
	}
	return ret
}

func convertChinese(content string) string {
	has := checker.IsChinese(content)
	if !has {
		return content
	}
	str, err := pinyin.New(content).Split("-").Mode(pinyin.WithoutTone).Convert()
	if err != nil {
		return content
	}
	return str
}

func cutLongTitle(title string) string {
	if len(title) > 150 {
		return title[0:150]
	}
	return title
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
		text = text + trimMarker
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

func GetPicByUrl(Url string) string {
	res, err := http.Get(Url)
	if err != nil {
		return ""
	}
	defer res.Body.Close()
	pix, err := io.ReadAll(res.Body)
	if err != nil {
		return ""
	}
	return string(pix)
}
