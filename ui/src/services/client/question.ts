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

export const useQuestionList = (params: Type.QueryQuestionsReq) => {
  const apiUrl = `/answer/api/v1/question/page?${qs.stringify(params)}`;
  const { data, error } = useSWR<Type.ListResult, Error>(apiUrl, (url) =>
    request.get(url, { allow404: true }),
  );
  return {
    data,
    isLoading: !data && !error,
    error,
  };
};

export const useQuestionRecommendList = (params: Type.QueryQuestionsReq) => {
  const apiUrl = `/answer/api/v1/question/recommend/page?${qs.stringify(
    params,
  )}`;
  const { data, error } = useSWR<Type.ListResult, Error>(
    [apiUrl],
    request.instance.get,
  );
  return {
    data,
    isLoading: !data && !error,
    error,
  };
};

export const useHotQuestions = (
  params: Type.QueryQuestionsReq = {
    page: 1,
    page_size: 6,
    order: 'hot',
    in_days: 7,
  },
) => {
  const apiUrl = `/answer/api/v1/question/page?${qs.stringify(params)}`;
  const { data, error } = useSWR<Type.ListResult, Error>(
    [apiUrl],
    request.instance.get,
  );
  return {
    data,
    isLoading: !data && !error,
    error,
  };
};

export const useSimilarQuestion = (params: {
  question_id: string;
  page_size: number;
}) => {
  const apiUrl = `/answer/api/v1/post/related?${qs.stringify(params)}`;

  const { data, error } = useSWR<Type.ListResult, Error>(
    params.question_id ? apiUrl : null,
    request.instance.get,
  );
  return {
    data,
    isLoading: !data && !error,
    error,
  };
};

export const getInviteUser = (questionId: string) => {
  const apiUrl = '/answer/api/v1/question/invite';
  return request.get<Type.UserInfoBase[]>(apiUrl, {
    params: { id: questionId },
  });
};

export const putInviteUser = (
  questionId: string,
  users: string[],
  imgCode: Type.ImgCodeReq = {},
) => {
  const apiUrl = '/answer/api/v1/question/invite';
  return request.put(apiUrl, {
    id: questionId,
    invite_user: users,
    ...imgCode,
  });
};

export const unDeleteAnswer = (id) => {
  return request.post('/answer/api/v1/post/comment/recover', {
    answer_id: id,
  });
};

export const unDeleteQuestion = (qid) => {
  return request.post('/answer/api/v1/question/recover', {
    question_id: qid,
  });
};
