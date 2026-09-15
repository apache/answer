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

import { Suspense, lazy, useEffect, useState } from 'react';
import { RouteObject } from 'react-router-dom';

import Layout from '@/pages/Layout';
import { mergeRoutePlugins } from '@/utils/pluginKit';

import baseRoutes, { RouteNode } from './routes';
import RouteGuard from './RouteGuard';
import RouteErrorBoundary from './RouteErrorBoundary';

const STALE_CHUNK_KEY = 'answer_stale_chunk_reload';

const routeWrapper = (routeNodes: RouteNode[], root: RouteNode[]) => {
  routeNodes.forEach((rn) => {
    if (rn.page === 'pages/Layout') {
      rn.element = rn.guard ? (
        <RouteGuard onEnter={rn.guard} path={rn.path} page={rn.page}>
          <Layout />
        </RouteGuard>
      ) : (
        <Layout />
      );
      rn.errorElement = <RouteErrorBoundary />;
    } else {
      /**
       * cannot use a fully dynamic import statement
       * ref: https://webpack.js.org/api/module-methods/#import-1
       */

      let Ctrl;

      if (typeof rn.page === 'string') {
        const pagePath = rn.page.replace('pages/', '');
        Ctrl = lazy(() =>
          import(`@/pages/${pagePath}`).then(
            (m) => {
              sessionStorage.removeItem(STALE_CHUNK_KEY);
              return m;
            },
            (err) => {
              // A tab opened before an upgrade still runs the old bundle; the hashed chunk
              // names it asks for no longer exist on the server, so reload once to pick up
              // the new index.html instead of rendering the 50X error page.
              if (!sessionStorage.getItem(STALE_CHUNK_KEY)) {
                sessionStorage.setItem(STALE_CHUNK_KEY, '1');
                window.location.reload();
                return new Promise<never>(() => {});
              }
              throw err;
            },
          ),
        );
      } else {
        Ctrl = rn.page;
      }

      rn.element = (
        <Suspense>
          {rn.guard ? (
            <RouteGuard onEnter={rn.guard} path={rn.path} page={rn.page}>
              <Ctrl />
            </RouteGuard>
          ) : (
            <Ctrl />
          )}
        </Suspense>
      );
      rn.errorElement = <RouteErrorBoundary />;
    }
    root.push(rn);
    const children = Array.isArray(rn.children) ? rn.children : null;
    if (children) {
      rn.children = [];
      routeWrapper(children, rn.children);
    }
  });
};

function useMergeRoutes() {
  const [routesState, setRoutes] = useState<RouteObject[]>([]);

  const init = async () => {
    const routes = [];
    const mergedRoutes = await mergeRoutePlugins(baseRoutes).catch(() => []);
    routeWrapper(mergedRoutes, routes);
    setRoutes(routes);
  };

  useEffect(() => {
    init();
  }, []);

  return routesState;
}

export { useMergeRoutes };
