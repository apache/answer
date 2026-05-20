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
import useSWR from 'swr';

import request from '@/utils/request';
import type * as Type from '@/common/interface';

export const uploadImage = (params: { file: File; type: Type.UploadType }) => {
  const form = new FormData();
  form.append('source', String(params.type));
  form.append('file', params.file);
  return request.post('/answer/api/v1/file', form);
};

export const queryQuestionByTitle = (title: string) => {
  return request.get(`/answer/api/v1/question/similar?title=${title}`);
};

export const useQueryTags = (params) => {
  const apiUrl = `/answer/api/v1/tags/page?${qs.stringify(params, {
    skipNulls: true,
  })}`;
  const { data, error, mutate } = useSWR<Type.ListResult>(apiUrl, (url) =>
    request.get(url, { allow404: true }),
  );
  return {
    data,
    isLoading: !data && !error,
    error,
    mutate,
  };
};

export const useQueryRevisions = (object_id: string | undefined) => {
  return useSWR<Record<string, any>>(
    object_id ? `/answer/api/v1/revisions?object_id=${object_id}` : '',
    request.instance.get,
  );
};

export const useQueryComments = (params) => {
  if (params.page === 0) {
    params.query_cond = 'vote';
    params.page = 1;
  } else {
    // only first page need commentId
    params.query_cond = '';
    delete params.comment_id;
  }
  return useSWR<Type.ListResult>(
    `/answer/api/v1/comment/page?${qs.stringify(params, {
      skipNulls: true,
    })}`,
    request.instance.get,
  );
};

export const updateComment = (params) => {
  return request.put('/answer/api/v1/comment', params);
};

export const deleteComment = (id, imgCode: Type.ImgCodeReq = {}) => {
  return request.delete('/answer/api/v1/comment', {
    comment_id: id,
    ...imgCode,
  });
};

export const addComment = (params) => {
  return request.post('/answer/api/v1/comment', params);
};

export const updateReaction = (params) => {
  return request.put('/answer/api/v1/meta/reaction', params);
};

export const queryReactions = (object_id: string) => {
  return request.get<Type.ReactionItems>(
    `/answer/api/v1/meta/reaction?object_id=${object_id}`,
  );
};

export const queryTags = (tag: string) => {
  return request.get(
    `/answer/api/v1/question/tags?tag=${encodeURIComponent(tag)}`,
  );
};

export const useQueryAnswerInfo = (id: string) => {
  return useSWR<{
    info;
    question;
  }>(`/answer/api/v1/answer/info?id=${id}`, request.instance.get);
};

export const modifyQuestion = (
  params: Type.QuestionParams & { id: string; edit_summary: string },
) => {
  return request.put(`/answer/api/v1/question`, params);
};

export const modifyAnswer = (params: Type.AnswerParams) => {
  return request.put(`/answer/api/v1/answer`, params);
};

export const login = (params: Type.LoginReqParams) => {
  return request.post<Type.UserInfoRes>(
    '/answer/api/v1/user/login/email',
    params,
  );
};

export const register = (params: Type.RegisterReqParams) => {
  return request.post<any>('/answer/api/v1/user/register/email', params);
};

export const logout = () => {
  return request.get('/answer/api/v1/user/logout');
};

export const resendEmail = (params?: Type.ImgCodeReq) => {
  params = qs.parse(
    qs.stringify(params, {
      skipNulls: true,
    }),
  );
  return request.post('/answer/api/v1/user/email/verification/send', {
    ...params,
  });
};

/**
 * @description get login userinfo
 * @returns {UserInfo}
 */
export const getLoggedUserInfo = (config = { passingError: false }) => {
  return request.get<Type.UserInfoRes>('/answer/api/v1/user/info', config);
};

export const modifyUserInfo = (params: Type.ModifyUserReq) => {
  return request.put('/answer/api/v1/user/info', params);
};

export const modifyPassword = (params: Type.ModifyPasswordReq) => {
  return request.put('/answer/api/v1/user/password', params);
};

export const resetPassword = (params: Type.PasswordResetReq) => {
  return request.post('/answer/api/v1/user/password/reset', params);
};

export const replacementPassword = (params: Type.PasswordReplaceReq) => {
  return request.post('/answer/api/v1/user/password/replacement', params);
};

export const activateAccount = (code: string) => {
  return request.post(`/answer/api/v1/user/email/verification`, { code });
};

export const checkImgCode = (k: Type.CaptchaKey) => {
  const apiUrl = `/answer/api/v1/user/action/record`;
  return request.get<Type.ImgCodeRes>(apiUrl, {
    params: {
      action: k,
    },
  });
};

export const setNotice = (params: Type.SetNoticeReq) => {
  return request.post('/answer/api/v1/user/notice/set', params);
};

export const saveQuestion = (params: Type.QuestionParams) => {
  return request.post('/answer/api/v1/question', params);
};

export const questionDetail = (id: string) => {
  return request.get<Type.QuestionDetailRes>(
    `/answer/api/v1/question/info?id=${id}`,
    { allow404: true },
  );
};

export const useQuestionLink = (params: {
  question_id: string;
  page: number;
  page_size: number;
  order?: string;
}) => {
  const apiUrl = `/answer/api/v1/question/link?${qs.stringify(params)}`;
  const { data, error } = useSWR<Type.ListResult, Error>(
    [apiUrl, params],
    request.instance.get,
  );
  return {
    data,
    isLoading: !data && !error,
    error,
  };
};

export const getAnswers = (params: Type.AnswersReq) => {
  const apiUrl = `/answer/api/v1/answer/page?${qs.stringify(params)}`;
  return request.get<Type.ListResult<Type.AnswerItem>>(apiUrl);
};

export const postAnswer = (params: Type.PostAnswerReq) => {
  return request.post('/answer/api/v1/answer', params);
};

export const bookmark = (params: {
  group_id: string;
  object_id: string;
  bookmark: boolean;
}) => {
  return request.post('/answer/api/v1/collection/switch', params);
};

export const postVote = (
  params: { object_id: string; is_cancel: boolean } & Type.ImgCodeReq,
  type: 'down' | 'up',
) => {
  return request.post(`/answer/api/v1/vote/${type}`, params);
};

export const following = (params: {
  object_id: string;
  is_cancel: boolean;
}) => {
  return request.post<{ follows: number; is_followed: boolean }>(
    '/answer/api/v1/follow',
    params,
  );
};

export const acceptanceAnswer = (params: {
  answer_id?: string;
  question_id: string;
}) => {
  return request.post('/answer/api/v1/answer/acceptance', params);
};

export const reportList = ({
  type,
  action,
  isBackend = false,
}: Type.ReportParams & { isBackend }) => {
  let api = '/answer/api/v1/reasons';
  if (isBackend) {
    api = '/answer/admin/api/reasons';
  }
  return request.get(`${api}?object_type=${type}&action=${action}`);
};

export const postReport = (
  params: {
    source: Type.ReportType;
    content: string;
    object_id: string;
    report_type: number;
  } & Type.ImgCodeReq,
) => {
  return request.post('/answer/api/v1/report', params);
};

export const deleteQuestion = (params: {
  id: string;
  captcha_code?: string;
  captcha_id?: string;
}) => {
  return request.delete('/answer/api/v1/question', params);
};

export const deleteAnswer = (params: {
  id: string;
  captcha_code?: string;
  captcha_id?: string;
}) => {
  return request.delete('/answer/api/v1/answer', params);
};

export const closeQuestion = (params: {
  id: string;
  close_msg?: string;
  close_type: number;
}) => {
  return request.put('/answer/api/v1/question/status', params);
};

export const changeEmail = (params: { e_mail: string; pass?: string }) => {
  return request.post('/answer/api/v1/user/email/change/code', params);
};

export const changeEmailVerify = (params: { code: string }) => {
  return request.put('/answer/api/v1/user/email', params);
};

export const getAppSettings = () => {
  return request.get<Type.SiteSettings>('/answer/api/v1/siteinfo');
};

export const reopenQuestion = (params: { question_id: string }) => {
  return request.put('/answer/api/v1/question/reopen', params);
};

export const unsubscribe = (code: string) => {
  const apiUrl = '/answer/api/v1/user/notification/unsubscribe';
  return request.put(apiUrl, { code });
};

export const markdownToHtml = (content: string) => {
  const apiUrl = '/answer/api/v1/post/render';
  return request.post(apiUrl, { content });
};

export const saveQuestionWithAnswer = (params: Type.QuestionWithAnswer) => {
  return request.post('/answer/api/v1/question/answer', params);
};

export const questionOperation = (params: Type.QuestionOperationReq) => {
  return request.put('/answer/api/v1/question/operation', params);
};

export const getPluginsStatus = () => {
  return request.get<Type.ActivatedPlugin[]>('/answer/api/v1/plugin/status');
};

export const deletePermanently = (type: string) => {
  return request.delete('/answer/admin/api/delete/permanently', { type });
};
