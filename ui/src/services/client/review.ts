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

// import useSWR from 'swr';

import request from '@/utils/request';
import * as Type from '@/common/interface';

export const getSuggestReviewList = (page: number) => {
  const apiUrl = `/answer/api/v1/revisions/unreviewed?page=${page}`;
  return request.get<Type.SuggestReviewResp>(apiUrl);
};

export const getReviewType = () => {
  return request.get<Type.ReviewTypeItem[]>('/answer/api/v1/reviewing/type');
};

export const getFlagReviewPostList = (page: number) => {
  const apiUrl = `/answer/api/v1/report/unreviewed/post?page=${page}`;
  return request.get<Type.FlagReviewResp>(apiUrl);
};

export const putFlagReviewAction = (params: Type.PutFlagReviewParams) => {
  return request.put('/answer/api/v1/report/review', params);
};

export const getPendingReviewPostList = (page: number, objectId?: string) => {
  const search = objectId ? `&object_id=${objectId}` : '';
  const apiUrl = `/answer/api/v1/review/pending/post/page?page=${page}${search}`;
  return request.get<Type.QueuedReviewResp>(apiUrl);
};

export const putPendingReviewAction = (params: {
  review_id: number;
  status: 'approve' | 'reject';
}) => {
  return request.put('/answer/api/v1/review/pending/post', params);
};
