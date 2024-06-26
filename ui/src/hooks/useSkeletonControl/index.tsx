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

import { useEffect, useRef, useState } from 'react';

import { SKELETON_SHOW_TIME } from '@/common/constants';

/**
 * @param needShowFirst whether the skeleton should show at first
 *
 * Why need 'needShowFirst' param?
 * Sometimes we need skeleton screens to take up space in the dom from the start
 *
 * If you set the 'needShowFirst' param as false, If the interface time is too short,
 * the skeleton screen will not be displayed, which can reduce the time occupation
 */
const useSkeletonControl = (isLoading: boolean) => {
  const [isSkeletonShow, setIsSkeletonShow] = useState(false);
  const timer = useRef<NodeJS.Timeout | null>(null);
  const openSkeleton = () => {
    if (timer.current) {
      clearTimeout(timer.current);
    }
    timer.current = setTimeout(() => {
      setIsSkeletonShow(true);
    }, SKELETON_SHOW_TIME);
  };

  const closeSkeleton = () => {
    clearTimeout(timer.current as NodeJS.Timeout);
    setIsSkeletonShow(false);
  };

  useEffect(() => {
    if (isLoading) {
      openSkeleton();
    } else {
      closeSkeleton();
    }
  }, [isLoading]);

  return { isSkeletonShow };
};

export default useSkeletonControl;
