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

package plugin

var (
	DefaultCDNFileType = map[string]bool{
		".ico":   true,
		".json":  true,
		".css":   true,
		".js":    true,
		".webp":  true,
		".woff2": true,
		".woff":  true,
		".jpg":   true,
		".svg":   true,
		".png":   true,
		".map":   true,
		".txt":   true,
	}
)

type CDN interface {
	Base
	GetStaticPrefix() string
}

var (
	// CallCDN is a function that calls all registered parsers
	CallCDN,
	registerCDN = MakePlugin[CDN](false)
)

func coordinatedCDNPlugins(slugName string) (coordinatedSlugNames []string) {
	isCDN := false
	_ = CallBase(func(base Base) error {
		cdn, ok := base.(CDN)
		if !ok {
			return nil
		}
		name := cdn.Info().SlugName
		if slugName == name {
			isCDN = true
		} else {
			coordinatedSlugNames = append(coordinatedSlugNames, name)
		}
		return nil
	})
	if isCDN {
		return coordinatedSlugNames
	}
	return nil
}
