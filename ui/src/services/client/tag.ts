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

import useSWR from 'swr';

import request from '@/utils/request';
import type * as Type from '@/common/interface';
import { tryLoggedAndActivated } from '@/utils/guard';

export const deleteTag = (id) => {
  return request.delete('/answer/api/v1/tag', {
    tag_id: id,
  });
};
export const modifyTag = (params) => {
  return request.put('/answer/api/v1/tag', params);
};
export const mergeTag = (params: {
  source_tag_id: string;
  target_tag_id: string;
}) => {
  return request.post('/answer/api/v1/tag/merge', params);
};

export const useQuerySynonymsTags = (tagId, status) => {
  const apiUrl =
    status === 'deleted'
      ? ''
      : tagId
        ? `/answer/api/v1/tag/synonyms?tag_id=${tagId}`
        : '';
  return useSWR<{
    synonyms: Type.SynonymsTag[];
    member_actions?: Type.MemberActionItem[];
  }>(apiUrl, request.instance.get);
};

export const saveSynonymsTags = (params) => {
  return request.put('/answer/api/v1/tag/synonym', params);
};

export const useFollowingTags = () => {
  let apiUrl = '';
  if (tryLoggedAndActivated().ok) {
    apiUrl = '/answer/api/v1/tags/following';
  }
  const { data, error, mutate } = useSWR<any[]>(apiUrl, request.instance.get);
  return {
    data,
    isLoading: !data && !error,
    error,
    mutate,
  };
};

export const useTagInfo = ({ id = '', name = '' }) => {
  let apiUrl;
  if (id) {
    apiUrl = `/answer/api/v1/tag?id=${id}`;
  } else if (name) {
    name = encodeURIComponent(name);
    apiUrl = `/answer/api/v1/tag?name=${name}`;
  }
  const { data, error, mutate } = useSWR<Type.TagInfo>(apiUrl, (url) =>
    request.get(url, { allow404: true }),
  );
  return {
    mutate,
    data,
    isLoading: !data && !error,
    error,
  };
};

export const followTags = (params) => {
  return request.put('/answer/api/v1/follow/tags', params);
};

export const getTagsBySlugName = (slugNames: string) => {
  const apiUrl = `/answer/api/v1/tags?tags=${encodeURIComponent(slugNames)}`;
  return request.get<Type.TagInfo[]>(apiUrl);
};

export const createTag = (params: Type.TagBase) => {
  const apiUrl = '/answer/api/v1/tag';
  return request.post<Type.TagInfo>(apiUrl, params);
};

export const unDeleteTag = (id) => {
  return request.post('/answer/api/v1/tag/recover', {
    tag_id: id,
  });
};
