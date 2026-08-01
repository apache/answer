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

import { Suspense, lazy, useEffect, useState, type ComponentType } from 'react';
import { RouteObject } from 'react-router-dom';

import Layout from '@/pages/Layout';
import { mergeRoutePlugins } from '@/utils/pluginKit';

import baseRoutes, { RouteNode } from './routes';
import RouteGuard from './RouteGuard';
import RouteErrorBoundary from './RouteErrorBoundary';

type PageLoader = () => Promise<{ default: ComponentType }>;

// Page components live at `pages/<Name>/index.tsx`, or, for nested pages,
// `pages/<Name>/<Sub>/index.tsx` and `pages/<Name>/<Sub>/<Leaf>/index.tsx`.
// `route.page` values reference either the directory (`Foo/Bar`) or the
// directory plus explicit `/index`; both are normalized to the same lookup
// key below. The glob is bounded to those three depths, and `components`
// subtrees are excluded, so a component's own `index.tsx` never becomes a
// route chunk. `Layout` is excluded too: it is imported statically below
// and handled outside this lookup.
const pageModules = import.meta.glob<{ default: ComponentType }>([
  '../pages/*/index.tsx',
  '../pages/*/*/index.tsx',
  '../pages/*/*/*/index.tsx',
  '!../pages/**/components/**',
  '!../pages/Layout/**',
]);

const pagesByKey: Record<string, PageLoader> = {};
Object.keys(pageModules).forEach((filePath) => {
  const key = filePath
    .replace(/^\.\.\/pages\//, '')
    .replace(/\/index\.tsx$/, '');
  pagesByKey[key] = pageModules[filePath];
});

const missingPageLoader =
  (page: string, lookupKey: string): PageLoader =>
  () =>
    Promise.reject(
      new Error(
        `No page module found for "${page}" (looked up as "${lookupKey}"). ` +
          `Known page keys: ${Object.keys(pagesByKey).sort().join(', ')}`,
      ),
    );

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
      // The import target must be statically analyzable so each page
      // resolves to its own lazy-loaded chunk rather than one shared bundle.
      let Ctrl;

      if (typeof rn.page === 'string') {
        const pagePath = rn.page.replace('pages/', '').replace(/\/index$/, '');
        Ctrl = lazy(
          pagesByKey[pagePath] ?? missingPageLoader(rn.page, pagePath),
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
