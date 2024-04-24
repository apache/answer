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

import debounce from 'lodash/debounce';

import {
  DRAFT_QUESTION_STORAGE_KEY,
  DRAFT_ANSWER_STORAGE_KEY,
} from '@/common/constants';
import { storageExpires as storage } from '@/utils';

export type QuestionDraft = {
  params: {
    title: string;
    content: string;
    tags: any[];
    answer_content: string;
  };
  callback?: () => void;
};

export type AnswerDraft = {
  questionId: string;
  content: string;
  callback?: () => void;
};

type DraftType = {
  type: 'question' | 'answer';
};

export type DraftParams = QuestionDraft | AnswerDraft;

class SaveDraft {
  type: DraftType['type'];

  status: 'save' | 'remove';

  constructor({ type = 'question' }: DraftType) {
    this.type = type;
    this.status = 'save';
  }

  save = debounce((data: DraftParams) => {
    // TODO
    if (this.status === 'remove') {
      return;
    }
    if (this.type === 'question') {
      const { params, callback } = data as QuestionDraft;

      this.storeDraft(params, callback);
    }

    if (this.type === 'answer') {
      const { content, questionId, callback } = data as AnswerDraft;
      if (!questionId || !content) {
        return;
      }

      this.storeDraft({ content, questionId }, callback);
    }
  }, 3000);

  remove() {
    this.status = 'remove';
    const that = this;
    if (this.type === 'question') {
      storage.remove(DRAFT_QUESTION_STORAGE_KEY, () => {
        that.status = 'save';
      });
    }
    if (this.type === 'answer') {
      storage.remove(DRAFT_ANSWER_STORAGE_KEY, () => {
        that.status = 'save';
      });
    }
  }

  private storeDraft = (params: any, callback) => {
    const key =
      this.type === 'question'
        ? DRAFT_QUESTION_STORAGE_KEY
        : DRAFT_ANSWER_STORAGE_KEY;
    storage.set(key, params);
    callback?.();
  };
}

export default SaveDraft;
