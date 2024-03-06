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
import { Col, Card } from 'react-bootstrap';

interface Props {
  count?: number;
}

const Index: FC<Props> = ({ count = 8 }) => {
  const list = new Array(count).fill(0).map((v, i) => v + i);
  return (
    <>
      {list.map((v) => (
        <Col
          key={v}
          xs={12}
          lg={3}
          md={4}
          sm={6}
          className="mb-4 placeholder-glow">
          <Card className="h-100">
            <Card.Body className="d-flex flex-column align-items-start">
              <div
                className="placeholder align-top w-25 mb-3"
                style={{ height: '24px' }}
              />

              <p
                className="placeholder small text-truncate-3 w-100"
                style={{ height: '42px' }}
              />
              <div className="d-flex align-items-center">
                <div
                  className="placeholder me-2"
                  style={{ width: '80px', height: '31px' }}
                />
                <span
                  className="placeholder text-secondary small text-nowrap"
                  style={{ width: '100px', height: '21px' }}
                />
              </div>
            </Card.Body>
          </Card>
        </Col>
      ))}
    </>
  );
};

export default memo(Index);
