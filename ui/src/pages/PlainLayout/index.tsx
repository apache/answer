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

/* A layout that is nothing but a main landmark.
 *
 * Most pages reach their `main` through `SideNavLayout` or `Admin`. The rest —
 * signing in, registering, recovering an account, the error pages — hang
 * straight off `pages/Layout`, which wraps the header and would therefore put
 * the navigation inside `main` if the landmark were added there.
 *
 * A pathless layout route solves that without touching a single page: the
 * children keep their paths, their order and their guards, and gain a landmark
 * they can be skipped to.
 */

import { FC, memo } from 'react';
import { Outlet } from 'react-router-dom';

const Index: FC = () => {
  return (
    <main>
      <Outlet />
    </main>
  );
};

export default memo(Index);
