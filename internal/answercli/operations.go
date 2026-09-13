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

package answercli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func (s *commandState) questionCommand() *cobra.Command {
	command := &cobra.Command{Use: "question"}
	command.AddCommand(s.questionSearchCommand(), s.questionGetCommand(), s.questionCreateCommand())
	return command
}

func (s *commandState) questionSearchCommand() *cobra.Command {
	var query, order string
	var page, size int
	command := &cobra.Command{
		Use: "search", Short: "Search questions",
		RunE: func(cmd *cobra.Command, _ []string) error {
			values := url.Values{"q": {strings.TrimSpace(query + " is:question")}, "order": {order}, "page": {fmt.Sprint(page)}, "size": {fmt.Sprint(size)}}
			return s.run(cmd, http.MethodGet, "/answer/api/v1/search", values, nil)
		},
	}
	command.Flags().StringVar(&query, "query", "", "search query")
	command.Flags().StringVar(&order, "order", "relevance", "result order")
	command.Flags().IntVar(&page, "page", 1, "page number")
	command.Flags().IntVar(&size, "size", 30, "page size")
	_ = command.MarkFlagRequired("query")
	return command
}

func (s *commandState) questionGetCommand() *cobra.Command {
	return &cobra.Command{
		Use: "get <question-id>", Args: cobra.ExactArgs(1), Short: "Get a question",
		RunE: func(cmd *cobra.Command, args []string) error {
			return s.run(cmd, http.MethodGet, "/answer/api/v1/question/info", url.Values{"id": {args[0]}}, nil)
		},
	}
}

func (s *commandState) questionCreateCommand() *cobra.Command {
	var title, bodyFile, inputJSON string
	var tags []string
	command := &cobra.Command{
		Use: "create", Short: "Create a question",
		RunE: func(cmd *cobra.Command, _ []string) error {
			body, err := s.createQuestionBody(title, bodyFile, inputJSON, tags)
			if err != nil {
				return s.writeError(err)
			}
			return s.run(cmd, http.MethodPost, "/answer/api/v1/question", nil, body)
		},
	}
	command.Flags().StringVar(&title, "title", "", "question title")
	command.Flags().StringSliceVar(&tags, "tag", nil, "tag slug (repeatable)")
	command.Flags().StringVar(&bodyFile, "body-file", "", "Markdown body file, or - for stdin")
	command.Flags().StringVar(&inputJSON, "input-json", "", "complete JSON request file, or - for stdin")
	return command
}

func (s *commandState) createQuestionBody(title, bodyFile, inputJSON string, tags []string) (any, error) {
	if inputJSON != "" {
		return readJSONObject(s.options.Stdin, inputJSON)
	}
	if title == "" || bodyFile == "" {
		return nil, fmt.Errorf("--title and --body-file are required unless --input-json is used")
	}
	content, err := readInput(s.options.Stdin, bodyFile)
	if err != nil {
		return nil, err
	}
	tagItems := make([]map[string]string, 0, len(tags))
	for _, tag := range tags {
		tagItems = append(tagItems, map[string]string{"slug_name": tag, "display_name": tag})
	}
	return map[string]any{"title": title, "content": string(content), "tags": tagItems}, nil
}

func (s *commandState) answerCommand() *cobra.Command {
	command := &cobra.Command{Use: "answer"}
	command.AddCommand(s.answerListCommand(), s.answerGetCommand(), s.answerCreateCommand())
	return command
}

func (s *commandState) answerListCommand() *cobra.Command {
	var questionID string
	command := &cobra.Command{
		Use: "list", Short: "List answers to a question",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return s.run(cmd, http.MethodGet, "/answer/api/v1/answer/page", url.Values{
				"question_id": {questionID}, "page": {"1"}, "page_size": {"100"}, "order": {"default"},
			}, nil)
		},
	}
	command.Flags().StringVar(&questionID, "question", "", "question ID")
	_ = command.MarkFlagRequired("question")
	return command
}

func (s *commandState) answerGetCommand() *cobra.Command {
	return &cobra.Command{
		Use: "get <answer-id>", Args: cobra.ExactArgs(1), Short: "Get an answer",
		RunE: func(cmd *cobra.Command, args []string) error {
			return s.run(cmd, http.MethodGet, "/answer/api/v1/answer/info", url.Values{"id": {args[0]}}, nil)
		},
	}
}

func (s *commandState) answerCreateCommand() *cobra.Command {
	var questionID, bodyFile, inputJSON string
	command := &cobra.Command{
		Use: "create", Short: "Create an answer",
		RunE: func(cmd *cobra.Command, _ []string) error {
			var body any
			var err error
			if inputJSON != "" {
				body, err = readJSONObject(s.options.Stdin, inputJSON)
			} else {
				if questionID == "" || bodyFile == "" {
					return s.writeError(fmt.Errorf("--question and --body-file are required unless --input-json is used"))
				}
				var content []byte
				content, err = readInput(s.options.Stdin, bodyFile)
				body = map[string]any{"question_id": questionID, "content": string(content)}
			}
			if err != nil {
				return s.writeError(err)
			}
			return s.run(cmd, http.MethodPost, "/answer/api/v1/answer", nil, body)
		},
	}
	command.Flags().StringVar(&questionID, "question", "", "question ID")
	command.Flags().StringVar(&bodyFile, "body-file", "", "Markdown body file, or - for stdin")
	command.Flags().StringVar(&inputJSON, "input-json", "", "complete JSON request file, or - for stdin")
	return command
}

func (s *commandState) voteCommand() *cobra.Command {
	command := &cobra.Command{Use: "vote"}
	command.AddCommand(s.voteActionCommand("up", false), s.voteActionCommand("down", false), s.voteRetractCommand())
	return command
}

func (s *commandState) voteActionCommand(direction string, cancel bool) *cobra.Command {
	return &cobra.Command{
		Use: direction + " <object-id>", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return s.run(cmd, http.MethodPost, "/answer/api/v1/vote/"+direction, nil,
				map[string]any{"object_id": args[0], "is_cancel": cancel})
		},
	}
}

func (s *commandState) voteRetractCommand() *cobra.Command {
	var direction string
	command := &cobra.Command{
		Use: "retract <object-id>", Args: cobra.ExactArgs(1), Short: "Retract an upvote or downvote",
		RunE: func(cmd *cobra.Command, args []string) error {
			if direction != "up" && direction != "down" {
				return s.writeError(fmt.Errorf("--direction must be up or down"))
			}
			return s.run(cmd, http.MethodPost, "/answer/api/v1/vote/"+direction, nil,
				map[string]any{"object_id": args[0], "is_cancel": true})
		},
	}
	command.Flags().StringVar(&direction, "direction", "", "vote direction to retract")
	_ = command.MarkFlagRequired("direction")
	return command
}

func (s *commandState) tagCommand() *cobra.Command {
	command := &cobra.Command{Use: "tag"}
	var query string
	search := &cobra.Command{
		Use: "search", Short: "Search tags",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return s.run(cmd, http.MethodGet, "/answer/api/v1/question/tags", url.Values{"tag": {query}}, nil)
		},
	}
	search.Flags().StringVar(&query, "query", "", "tag search query")
	command.AddCommand(search)
	return command
}

func (s *commandState) run(cmd *cobra.Command, method, path string, query url.Values, body any) error {
	client, err := s.client()
	if err != nil {
		return s.writeError(err)
	}
	data, err := client.Do(cmd.Context(), method, path, query, body)
	if err != nil {
		return s.writeError(err)
	}
	return writeSuccess(s.options.Stdout, data)
}

func readJSONObject(stdin io.Reader, path string) (map[string]any, error) {
	content, err := readInput(stdin, path)
	if err != nil {
		return nil, err
	}
	result := make(map[string]any)
	if err := json.Unmarshal(content, &result); err != nil {
		return nil, fmt.Errorf("decode JSON input: %w", err)
	}
	return result, nil
}

func readInput(stdin io.Reader, path string) ([]byte, error) {
	if path == "-" {
		return io.ReadAll(stdin)
	}
	return os.ReadFile(path)
}
