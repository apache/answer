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

import React, { memo, FC } from 'react';
import { Link } from 'react-router-dom';

import classNames from 'classnames';

import { Tag } from '@/common/interface';
import { pathFactory } from '@/router/pathFactory';

interface IProps {
  data: Tag;
  href?: string;
  className?: string;
  textClassName?: string;
}

const Index: FC<IProps> = ({
  data,
  href,
  className = '',
  textClassName = '',
}) => {
  href ||= pathFactory.tagLanding(data.slug_name);

  return (
    <Link
      to={href}
      onClick={(e) => {
        e.stopPropagation();
      }}
      className={classNames(
        'badge-tag rounded-1',
        data.reserved && 'badge-tag-reserved',
        data.recommend && 'badge-tag-required',
        className,
      )}>
      <span className={textClassName}>
        {data.display_name || data.slug_name}
      </span>
    </Link>
  );
};

export default memo(Index);
