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

package main

import (
	"os"
	"runtime/debug"

	"github.com/apache/answer/internal/answercli"
)

var (
	version   = "dev"
	revision  string
	buildTime string
)

func main() {
	resolvedRevision, resolvedBuildTime, modified := buildMetadata()
	command := answercli.NewRootCommand(answercli.Options{
		Stdin: os.Stdin, Stdout: os.Stdout, Stderr: os.Stderr,
		Version: version, Revision: resolvedRevision, BuildTime: resolvedBuildTime, Modified: modified,
	})
	if err := command.Execute(); err != nil {
		os.Exit(answercli.ExitCode(err))
	}
}

func buildMetadata() (resolvedRevision, resolvedBuildTime string, modified bool) {
	resolvedRevision = revision
	resolvedBuildTime = buildTime
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return resolvedRevision, resolvedBuildTime, false
	}
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			if resolvedRevision == "" {
				resolvedRevision = setting.Value
			}
		case "vcs.time":
			if resolvedBuildTime == "" {
				resolvedBuildTime = setting.Value
			}
		case "vcs.modified":
			modified = setting.Value == "true"
		}
	}
	return resolvedRevision, resolvedBuildTime, modified
}
