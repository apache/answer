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

import { create } from 'zustand';

import { AdminSettingsWrite } from '@/common/interface';

interface IProps {
  write: AdminSettingsWrite;
  update: (params: AdminSettingsWrite) => void;
}

const Index = create<IProps>((set) => ({
  write: {
    restrict_answer: true,
    min_tags: 1,
    min_content: 6,
    recommend_tags: [],
    required_tag: false,
    reserved_tags: [],
    max_image_size: 4,
    max_attachment_size: 8,
    max_image_megapixel: 40,
    authorized_image_extensions: [],
    authorized_attachment_extensions: [],
  },
  update: (params) =>
    set((state) => {
      const o = { ...state.write, ...params };
      return {
        write: o,
      };
    }),
}));

export default Index;
