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

import { pathFactory } from '@/router/pathFactory';

export default (data: any) => {
  if (!data?.object_id) {
    return {
      itemLink: '',
      itemId: '',
      itemTimePrefix: '',
    };
  }

  const {
    object_type = '',
    object_id = '',
    question_id = '',
    answer_id = '',
    comment_id = '',
    url_title = '',
  } = data;
  let itemLink = '';
  let itemId = '';
  let itemTimePrefix = '';

  if (object_type === 'question') {
    itemLink = pathFactory.questionLanding(String(object_id), url_title);
    itemId = String(question_id);
    itemTimePrefix = 'asked';
  } else if (object_type === 'answer') {
    itemLink = pathFactory.answerLanding({
      // @ts-ignore
      questionId: question_id,
      slugTitle: url_title,
      answerId: String(object_id),
    });
    itemId = String(object_id);
    itemTimePrefix = 'answered';
  } else if (object_type === 'comment') {
    if (question_id && answer_id) {
      itemLink = `${pathFactory.answerLanding({
        questionId: question_id,
        slugTitle: url_title,
        answerId: answer_id,
      })}?commentId=${comment_id}`;
    } else {
      itemLink = `${pathFactory.questionLanding(
        String(question_id),
        url_title,
      )}?commentId=${comment_id}`;
    }
    itemId = String(comment_id);
    itemTimePrefix = 'commented';
  }

  return {
    itemLink,
    itemId,
    itemTimePrefix,
  };
};
