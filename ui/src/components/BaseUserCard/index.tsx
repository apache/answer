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

import { memo, FC } from 'react';
import { Link } from 'react-router-dom';

import { Avatar } from '@/components';
import { formatCount } from '@/utils';

interface Props {
  data: any;
  showAvatar?: boolean;
  avatarSize?: string;
  showReputation?: boolean;
  avatarSearchStr?: string;
  className?: string;
  avatarClass?: string;
  nameMaxWidth?: string;
}

const Index: FC<Props> = ({
  data,
  showAvatar = true,
  avatarClass = '',
  avatarSize = '24px',
  className = 'small',
  avatarSearchStr = 's=48',
  showReputation = true,
  nameMaxWidth = '300px',
}) => {
  return (
    <div className={`d-flex align-items-center  text-secondary ${className}`}>
      {data?.status !== 'deleted' ? (
        <Link
          to={`/users/${data?.username}`}
          onClick={(e) => {
            e.stopPropagation();
          }}
          className="d-flex align-items-center">
          {showAvatar && (
            <Avatar
              avatar={data?.avatar}
              size={avatarSize}
              className={`me-1 ${avatarClass}`}
              searchStr={avatarSearchStr}
              alt={data?.display_name}
            />
          )}
          <span
            className="me-1 name-ellipsis"
            style={{ maxWidth: nameMaxWidth }}>
            {data?.display_name}
          </span>
        </Link>
      ) : (
        <>
          {showAvatar && (
            <Avatar
              avatar={data?.avatar}
              size={avatarSize}
              className={`me-1 ${avatarClass}`}
              searchStr={avatarSearchStr}
              alt={data?.display_name}
            />
          )}
          <span className="me-1 name-ellipsis">{data?.display_name}</span>
        </>
      )}

      {showReputation && (
        <span className="fw-bold" title="Reputation">
          {formatCount(data?.rank)}
        </span>
      )}
    </div>
  );
};

export default memo(Index);
