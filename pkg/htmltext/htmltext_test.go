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
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClearText(t *testing.T) {
	var (
		expected,
		clearedText string
	)

	// test code clear text
	expected = "hello{code...}"
	clearedText = ClearText("<p>hello<pre>var a = \"good\"</pre></p>")
	assert.Equal(t, expected, clearedText)

	// test link clear text
	expected = "hello [example.com]"
	clearedText = ClearText("<p>hello <a href=\"http://example.com/\">example.com</a></p>")
	assert.Equal(t, expected, clearedText)
	clearedText = ClearText("<p>hello<a href=\"https://example.com/\">example.com</a></p>")
	assert.Equal(t, expected, clearedText)

	expected = "hello world"
	clearedText = ClearText("<div> hello</div>\n<div>world</div>")
	assert.Equal(t, expected, clearedText)
}

func TestFetchExcerpt(t *testing.T) {
	var (
		expected,
		text string
	)

	// test english string
	expected = "hello..."
	text = FetchExcerpt("<p>hello world</p>", "...", 5)
	assert.Equal(t, expected, text)

	// test mixed string
	expected = "hello你好..."
	text = FetchExcerpt("<p>hello你好world</p>", "...", 7)
	assert.Equal(t, expected, text)

	// test mixed string with emoticon
	expected = "hello你好😂..."
	text = FetchExcerpt("<p>hello你好😂world</p>", "...", 8)
	assert.Equal(t, expected, text)

	expected = "hello你好"
	text = FetchExcerpt("<p>hello你好</p>", "...", 8)
	assert.Equal(t, expected, text)
}

func TestUrlTitle(t *testing.T) {
	list := []string{
		"hello你好😂...",
		"这是一个，标题，title",
	}
	for _, title := range list {
		formatTitle := UrlTitle(title)
		fmt.Println(formatTitle)
	}
}

func TestUrlTitleTable(t *testing.T) {
	// Long pure-Arabic title: 50 copies of the same Arabic word, joined by spaces.
	// Unidecode of "كيف" is "kyf", so the slug becomes "kyf-" repeated and
	// exceeds cutLongTitle's 150-byte cap.
	longArabic := strings.Repeat("كيف ", 50)
	wantLongArabic := strings.Repeat("kyf-", 37) + "ky" // 37*4 + 2 = 150 bytes

	cases := []struct {
		name  string
		title string
		want  string
	}{
		{
			name:  "empty",
			title: "",
			want:  "topic",
		},
		{
			name:  "pure latin unchanged",
			title: "hello world",
			want:  "hello-world",
		},
		{
			// Pinyin conversion drops Latin runes by design — matches pre-fix behavior.
			name:  "pure chinese unchanged",
			title: "这是一个，标题，title",
			want:  "zhe-shi-yi-ge-biao-ti",
		},
		{
			// The fix: previously collapsed to "topic" for all of these scripts.
			// Outputs are an ASCII approximation, not linguistically correct
			// romanization — see PR description.
			name:  "arabic transliterated",
			title: "كيف حالك",
			want:  "kyf-hlk",
		},
		{
			name:  "mixed latin and arabic",
			title: "مرحبا hello",
			want:  "mrhb-hello",
		},
		{
			name:  "thai transliterated",
			title: "ไทย ไทย",
			want:  "aithy-aithy",
		},
		{
			name:  "japanese hiragana transliterated",
			title: "こんにちは",
			want:  "konnichiha",
		},
		{
			// Japanese with Han-block kanji is caught by the pre-existing pinyin
			// pre-step (Chinese reading, not Japanese), so this path is unchanged
			// by this PR. Pinning to document the existing behavior.
			name:  "japanese kanji goes through pinyin path unchanged",
			title: "日本",
			want:  "ri-ben",
		},
		{
			name:  "korean transliterated",
			title: "안녕하세요",
			want:  "annyeonghaseyo",
		},
		{
			name:  "hebrew transliterated",
			title: "שלום עולם",
			want:  "shlvm-vlm",
		},
		{
			name:  "cyrillic transliterated",
			title: "Привет мир",
			want:  "privet-mir",
		},
		{
			name:  "emoji only falls back to topic",
			title: "😂😂😂",
			want:  "topic",
		},
		{
			name:  "long arabic truncates at cutLongTitle boundary",
			title: longArabic,
			want:  wantLongArabic,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := UrlTitle(tc.title)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestUrlTitleTransliterationToggle(t *testing.T) {
	defer SetTransliterateNonLatin(true)

	SetTransliterateNonLatin(false)
	// With transliteration off, pure-Arabic titles collapse to the existing
	// "topic" fallback (the pre-fix behavior).
	assert.Equal(t, "topic", UrlTitle("كيف حالك"))

	SetTransliterateNonLatin(true)
	assert.Equal(t, "kyf-hlk", UrlTitle("كيف حالك"))
}

func TestFindFirstMatchedWord(t *testing.T) {
	var (
		expectedWord,
		actualWord string
		expectedIndex,
		actualIndex int
	)

	text := "Hello, I have 中文 and 😂 and I am supposed to work fine."

	// test find nothing
	expectedWord, expectedIndex = "", 0
	actualWord, actualIndex = findFirstMatchedWord(text, []string{"youcantfindme"})
	assert.Equal(t, expectedWord, actualWord)
	assert.Equal(t, expectedIndex, actualIndex)

	// test find one word
	expectedWord, expectedIndex = "文", 17
	actualWord, actualIndex = findFirstMatchedWord(text, []string{"文"})
	assert.Equal(t, expectedWord, actualWord)
	assert.Equal(t, expectedIndex, actualIndex)

	// test find multiple matched words
	expectedWord, expectedIndex = "Hello", 0
	actualWord, actualIndex = findFirstMatchedWord(text, []string{"Hello", "文"})
	assert.Equal(t, expectedWord, actualWord)
	assert.Equal(t, expectedIndex, actualIndex)
}

func TestGetRuneRange(t *testing.T) {
	var (
		expectedBegin,
		expectedEnd,
		actualBegin,
		actualEnd int
	)

	runeText := []rune("Hello, I have 中文 and 😂.")
	runeLen := len(runeText)

	// test get range of negative offset and negative limit
	expectedBegin, expectedEnd = 0, 0
	actualBegin, actualEnd = getRuneRange(runeText, -1, -1)
	assert.Equal(t, expectedBegin, actualBegin)
	assert.Equal(t, expectedEnd, actualEnd)

	// test get range of exceeding offset and exceeding limit
	expectedBegin, expectedEnd = runeLen, runeLen
	actualBegin, actualEnd = getRuneRange(runeText, runeLen+1, runeLen+1)
	assert.Equal(t, expectedBegin, actualBegin)
	assert.Equal(t, expectedEnd, actualEnd)

	// test get range of normal offset and exceeding limit
	expectedBegin, expectedEnd = 3, runeLen
	actualBegin, actualEnd = getRuneRange(runeText, 3, runeLen)
	assert.Equal(t, expectedBegin, actualBegin)
	assert.Equal(t, expectedEnd, actualEnd)

	// test get range of normal offset and normal limit
	expectedBegin, expectedEnd = 3, 10
	actualBegin, actualEnd = getRuneRange(runeText, 3, 7)
	assert.Equal(t, expectedBegin, actualBegin)
	assert.Equal(t, expectedEnd, actualEnd)
}

func TestFetchRangedExcerpt(t *testing.T) {
	var (
		expected,
		actual string
	)

	// test english string
	expected = "hello..."
	actual = FetchRangedExcerpt("<p>hello world</p>", "...", 0, 5)
	assert.Equal(t, expected, actual)

	// test string with offset
	expected = "...llo你好..."
	actual = FetchRangedExcerpt("<p>hello你好world</p>", "...", 2, 5)
	assert.Equal(t, expected, actual)

	// test mixed string with emoticon with offset
	expected = "...你好😂..."
	actual = FetchRangedExcerpt("<p>hello你好😂world</p>", "...", 5, 3)
	assert.Equal(t, expected, actual)

	// test mixed string with offset and exceeding limit
	expected = "...你好😂world"
	actual = FetchRangedExcerpt("<p>hello你好😂world</p>", "...", 5, 100)
	assert.Equal(t, expected, actual)
}

func TestCutLongTitle(t *testing.T) {
	// Short title, no cutting needed
	short := "hello"
	assert.Equal(t, short, cutLongTitle(short))

	// Exactly max bytes, no cutting needed
	exact150 := strings.Repeat("a", 150)
	assert.Len(t, cutLongTitle(exact150), 150)

	// Just over max bytes, should be cut
	exact151 := strings.Repeat("a", 151)
	assert.Len(t, cutLongTitle(exact151), 150)

	// Multi-byte rune at boundary gets removed properly
	asciiPart := strings.Repeat("a", 149) // 149 bytes
	multiByteChar := "中"                  // 3 bytes - will span bytes 149-151
	title := asciiPart + multiByteChar    // 152 bytes total

	assert.Equal(t, asciiPart, cutLongTitle(title))
}

func TestFetchMatchedExcerpt(t *testing.T) {
	var (
		expected,
		actual string
	)

	html := "<p>Hello, I have 中文 and 😂 and I am supposed to work fine</p>"

	// test find nothing
	// it should return from the begin with double trimLength text
	expected = "Hello, I h..."
	actual = FetchMatchedExcerpt(html, []string{"youcantfindme"}, "...", 5)
	assert.Equal(t, expected, actual)

	// test find the word at the end
	// it should return the word beginning with double trimLenth plus len(word)
	expected = "... work fine"
	actual = FetchMatchedExcerpt(html, []string{"youcant", "fine"}, "...", 3)
	assert.Equal(t, expected, actual)

	// test find multiple words
	// it should return the first matched word with trimmedText
	expected = "... have 中文 and 😂..."
	actual = FetchMatchedExcerpt(html, []string{"中文", "😂"}, "...", 6)
	assert.Equal(t, expected, actual)
}
