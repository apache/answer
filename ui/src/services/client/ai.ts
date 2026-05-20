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

import qs from 'qs';

import request from '@/utils/request';
import type * as Type from '@/common/interface';

export const getConversationList = (params: Type.Paging) => {
  return request.get<{ count: number; list: Type.ConversationListItem[] }>(
    `/answer/api/v1/ai/conversation/page?${qs.stringify(params)}`,
  );
};

export const getConversationDetail = (id: string) => {
  return request.get<Type.ConversationDetail>(
    `/answer/api/v1/ai/conversation?conversation_id=${id}`,
  );
};

// /answer/api/v1/ai/conversation/vote
export const voteConversation = (params: Type.VoteConversationParams) => {
  return request.post('/answer/api/v1/ai/conversation/vote', params);
};
