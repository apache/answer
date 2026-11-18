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

/*
 * Mirrors how src/utils/localize.ts loads a language at runtime. The import
 * specifier is a template literal resolved through an alias that points
 * outside the frontend root, so whether it resolves is a property of the
 * bundler rather than of this file.
 *
 * Keep this expression identical to the one in the application.
 * check-locale-resolution.js asserts that it still matches.
 */
export const loadLocaleResource = async (langName) => {
  const { default: resConf } = await import(`@i18n/${langName}.yaml`);
  return resConf;
};
