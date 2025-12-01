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

import { FC, memo } from 'react';
import { Outlet } from 'react-router-dom';

import { SideNav, Footer } from '@/components';

import '@/common/sideNavLayout.scss';

const Index: FC = () => {
  return (
    <div className="d-flex">
      <div
        className="position-sticky px-3 border-end pt-4 d-none d-lg-block"
        id="pcSideNav">
        <SideNav />
      </div>
      <div className="flex-fill w-100 overflow-x-hidden">
        <div className="d-flex justify-content-center px-0 px-md-4">
          <div className="answer-container main-mx-with">
            <Outlet />
          </div>
        </div>
        <div className="d-flex justify-content-center">
          <div className="main-mx-with">
            <Footer />
          </div>
        </div>
      </div>
    </div>
  );
};

export default memo(Index);
