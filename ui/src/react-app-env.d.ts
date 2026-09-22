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

/// <reference types="vite/client" />

// Without this, ImportMetaEnv falls back to an index signature and a typo
// in import.meta.env.REACT_APP_* compiles clean. strictImportMetaEnv turns
// that off, so only the keys declared below (plus vite/client's own
// BASE_URL, MODE, DEV, PROD, SSR) are valid to read.
interface ViteTypeOptions {
  strictImportMetaEnv: unknown;
}

interface ImportMetaEnv {
  readonly REACT_APP_API_URL: string;
  readonly REACT_APP_BASE_URL: string;
}

declare module '*.yaml';

declare module '*.ico';
