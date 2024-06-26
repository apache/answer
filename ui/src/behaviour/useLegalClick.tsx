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

import { MouseEvent, useCallback } from 'react';

import { useLegalPrivacy, useLegalTos } from '@/services/client/legal';

export const useLegalClick = () => {
  const { data: tos } = useLegalTos();
  const { data: privacy } = useLegalPrivacy();

  const legalClick = useCallback(
    (evt: MouseEvent, type: 'tos' | 'privacy') => {
      evt.stopPropagation();
      const contentText =
        type === 'tos'
          ? tos?.terms_of_service_original_text
          : privacy?.privacy_policy_original_text;
      let matchUrl: URL | undefined;
      try {
        if (contentText) {
          matchUrl = new URL(contentText);
        }
        // eslint-disable-next-line no-empty
      } catch (ex) {}
      if (matchUrl) {
        evt.preventDefault();
        window.open(matchUrl.toString());
      }
    },
    [tos, privacy],
  );

  return legalClick;
};
