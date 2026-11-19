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
import qs from 'qs';

import request from '@/utils/request';
import type * as Type from '@/common/interface';

export const useAnswerSearch = (
  params: Type.AdminContentsReq & { question_id?: string },
) => {
  const apiUrl = `/answer/admin/api/answer/page?${qs.stringify(params)}`;
  const { data, error, mutate } = useSWR<Type.ListResult, Error>(
    [apiUrl],
    request.instance.get,
  );
  return {
    data,
    isLoading: !data && !error,
    error,
    mutate,
  };
};

export const changeAnswerStatus = (
  answer_id: string,
  status: Type.AdminAnswerStatus,
) => {
  return request.put('/answer/admin/api/answer/status', {
    answer_id,
    status,
  });
};

export const deleteAnswers = (answerIds: string[]) => {
  return request.delete<Type.BulkDeleteResult>('/answer/admin/api/answers', {
    answer_ids: answerIds,
  });
};
