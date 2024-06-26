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

import { seoSettingStore } from '@/stores';

const tagLanding = (slugName: string) => {
  const r = slugName ? `/tags/${encodeURIComponent(slugName)}` : '/tags';
  return r;
};

const tagInfo = (slugName: string) => {
  const r = slugName ? `/tags/${encodeURIComponent(slugName)}/info` : '/tags';
  return r;
};

const tagEdit = (tagId: string) => {
  const r = `/tags/${tagId}/edit`;
  return r;
};

const questionLanding = (questionId: string, slugTitle: string = '') => {
  const { seo } = seoSettingStore.getState();
  if (!questionId) {
    return slugTitle ? `/questions/null/${slugTitle}` : '/questions/null';
  }
  // @ts-ignore
  if (/[13]/.test(seo.permalink) && slugTitle) {
    return `/questions/${questionId}/${encodeURIComponent(slugTitle)}`;
  }

  return `/questions/${questionId}`;
};

const answerLanding = (params: {
  questionId: string;
  slugTitle?: string;
  answerId: string;
}) => {
  const questionLandingUrl = questionLanding(
    params.questionId,
    params.slugTitle,
  );
  return `${questionLandingUrl}/${params.answerId}`;
};

export const pathFactory = {
  tagLanding,
  tagInfo,
  tagEdit,
  questionLanding,
  answerLanding,
};
