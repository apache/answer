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
import { ListGroupItem } from 'react-bootstrap';

interface Props {
  count?: number;
}

const Index: FC<Props> = ({ count = 10 }) => {
  const list = new Array(count).fill(0).map((v, i) => v + i);
  return (
    <>
      {list.map((v) => (
        <ListGroupItem
          className="bg-transparent py-3 px-2 border-start-0 border-end-0 placeholder-glow"
          key={v}>
          <div
            className="placeholder h5 align-top d-block"
            style={{ height: '21px', width: '35%' }}
          />

          <div
            className="placeholder w-75 h5 align-top"
            style={{ height: '24px' }}
          />

          <div
            className="placeholder w-100 d-block align-top mb-2"
            style={{ height: '21px' }}
          />
          <div
            className="placeholder w-100 d-block align-top mb-2"
            style={{ height: '21px' }}
          />

          <div
            className="placeholder w-50 align-top mb-12"
            style={{ height: '24px' }}
          />

          <div
            className="placeholder w-25 align-top d-block"
            style={{ height: '21px' }}
          />
        </ListGroupItem>
      ))}
    </>
  );
};

export default memo(Index);
