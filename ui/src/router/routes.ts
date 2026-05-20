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

import type { IndexRouteObject, NonIndexRouteObject } from 'react-router-dom';

import { guard } from '@/utils';
import type { TGuardFunc } from '@/utils/guard';
import { editCheck } from '@/services';
import { isEditable } from '@/utils/guard';

type IndexRouteNode = Omit<IndexRouteObject, 'children'>;
type NonIndexRouteNode = Omit<NonIndexRouteObject, 'children'>;
type UnionRouteNode = IndexRouteNode | NonIndexRouteNode;

export type RouteNode = UnionRouteNode & {
  page: string;
  children?: RouteNode[];
  /**
   * a method to auto guard route before route enter
   * if the `ok` field in guard returned `TGuardResult` is true,
   * it means the guard passed then enter the route.
   * if guard returned the `TGuardResult` has `redirect` field,
   * then auto redirect route to the `redirect` target.
   */
  guard?: TGuardFunc;
};

const routes: RouteNode[] = [
  {
    path: '/',
    page: 'pages/Layout',
    loader: async () => {
      await guard.setupApp();
      return null;
    },
    guard: () => {
      const gr = guard.shouldLoginRequired();
      if (!gr.ok) {
        return gr;
      }
      return {
        ok: true,
      };
    },
    children: [
      // question and answer
      {
        // side nav layout
        page: 'pages/SideNavLayout',
        children: [
          {
            index: true,
            page: 'pages/Questions',
          },
          {
            path: 'questions',
            page: 'pages/Questions',
          },
          {
            path: 'questions/ask',
            page: 'pages/Questions/Ask',
            guard: () => {
              return guard.askRedirect();
            },
          },
          {
            path: 'questions/add',
            page: 'pages/Questions/Ask',
            guard: () => {
              return guard.activated();
            },
          },
          {
            path: 'posts/:qid/edit',
            page: 'pages/Questions/Ask',
            guard: () => {
              return guard.activated();
            },
          },
          {
            path: 'posts/:qid/:aid/edit',
            page: 'pages/Questions/EditAnswer',
            loader: async ({ params }) => {
              const ret = await editCheck(params.aid as string, true);
              return ret;
            },
            guard: (args) => {
              return isEditable(args);
            },
          },
          {
            path: 'questions/:qid',
            page: 'pages/Questions/Detail',
          },
          {
            path: 'questions/:qid/:slugPermalink',
            page: 'pages/Questions/Detail',
          },
          {
            path: 'questions/:qid/:slugPermalink/:aid',
            page: 'pages/Questions/Detail',
          },
          {
            path: '/questions/linked/:qid',
            page: 'pages/Questions/Linked',
            guard: () => {
              return guard.linkedRedirect();
            },
          },
          {
            path: '/linked/:qid',
            page: 'pages/Questions/Linked',
          },
          {
            path: '/search',
            page: 'pages/Search',
          },
          // tags
          {
            path: 'tags',
            page: 'pages/Tags',
          },
          {
            path: 'tags/create',
            page: 'pages/Tags/Create',
            guard: () => {
              return guard.isAdminOrModerator();
            },
          },
          {
            path: 'tags/:tagName',
            page: 'pages/Tags/Detail',
          },
          {
            path: 'tags/:tagName/questions',
            page: 'pages/Tags/Detail',
          },
          {
            path: 'tags/:tagName/info',
            page: 'pages/Tags/Info',
          },
          {
            path: 'tags/:tagId/edit',
            page: 'pages/Tags/Edit',
            guard: () => {
              return guard.activated();
            },
          },
          // for users
          {
            path: 'users',
            page: 'pages/Users',
          },
          {
            path: 'users/:username',
            page: 'pages/Users/Personal',
          },
          {
            path: 'users/:username/:tabName',
            page: 'pages/Users/Personal',
          },
          {
            path: 'users/settings',
            page: 'pages/Users/Settings',
            guard: () => {
              return guard.logged();
            },
            children: [
              {
                index: true,
                page: 'pages/Users/Settings/Profile',
              },
              {
                path: 'profile',
                page: 'pages/Users/Settings/Profile',
              },
              {
                path: 'notify',
                page: 'pages/Users/Settings/Notification',
              },
              {
                path: 'account',
                page: 'pages/Users/Settings/Account',
              },
              {
                path: 'interface',
                page: 'pages/Users/Settings/Interface',
              },
              {
                path: ':slug_name',
                page: 'pages/Users/Settings/Plugins',
              },
            ],
          },
          {
            path: 'users/notifications/:type/:subType?',
            page: 'pages/Users/Notifications',
          },
          {
            path: '/posts/:qid/timeline',
            page: 'pages/Timeline',
            guard: () => {
              return guard.logged();
            },
          },
          {
            path: '/posts/:qid/:aid/timeline',
            page: 'pages/Timeline',
            guard: () => {
              return guard.logged();
            },
          },
          {
            path: '/tags/:tid/timeline',
            page: 'pages/Timeline',
            guard: () => {
              return guard.logged();
            },
          },
          // for review
          {
            path: 'review',
            page: 'pages/Review',
          },
          {
            path: '/badges',
            page: 'pages/Badges/index',
          },
          {
            path: '/badges/:badge_id',
            page: 'pages/Badges/Detail/index',
          },
        ],
      },
      {
        path: 'users/login',
        page: 'pages/Users/Login',
        guard: () => {
          const notLogged = guard.notLogged();
          if (notLogged.ok) {
            return notLogged;
          }

          return guard.notActivated();
        },
      },
      {
        path: 'users/register',
        page: 'pages/Users/Register',
        guard: () => {
          const allowNew = guard.allowNewRegistration();
          if (!allowNew.ok) {
            return allowNew;
          }
          const notLogged = guard.notLogged();
          if (notLogged.ok) {
            const sa = guard.singUpAgent();
            if (!sa.ok) {
              return sa;
            }
          }
          return notLogged;
        },
      },
      {
        path: 'users/logout',
        page: 'pages/Users/Logout',
        guard: () => {
          return guard.loggedRedirectHome();
        },
      },
      {
        path: 'users/account-recovery',
        page: 'pages/Users/AccountForgot',
        guard: () => {
          return guard.notLogged();
        },
      },
      {
        path: 'users/change-email',
        page: 'pages/Users/ChangeEmail',
      },
      {
        path: 'users/password-reset',
        page: 'pages/Users/PasswordReset',
      },
      {
        path: 'users/account-activation',
        page: 'pages/Users/ActiveEmail',
      },
      {
        path: 'users/account-activation/success',
        page: 'pages/Users/ActivationResult',
        guard: () => {
          return guard.activated();
        },
      },
      {
        path: '/users/account-activation/failed',
        page: 'pages/Users/ActivationResult',
        guard: () => {
          return guard.notActivated();
        },
      },
      {
        path: '/users/confirm-new-email',
        page: 'pages/Users/ConfirmNewEmail',
      },
      {
        path: '/users/account-suspended',
        page: 'pages/Users/Suspended',
        guard: () => {
          return guard.notLogged();
        },
      },
      {
        path: '/users/confirm-email',
        page: 'pages/Users/OauthBindEmail',
      },
      {
        path: '/users/auth-landing',
        page: 'pages/Users/AuthCallback',
      },
      // for admin
      {
        path: 'admin',
        page: 'pages/Admin',
        loader: async () => {
          await guard.pullLoggedUser();
          return null;
        },
        guard: () => {
          return guard.admin();
        },
        children: [
          {
            index: true,
            page: 'pages/Admin/Dashboard',
          },
          {
            path: 'dashboard',
            page: 'pages/Admin/Dashboard',
          },
          {
            path: 'qa/questions',
            page: 'pages/Admin/Questions',
          },
          {
            path: 'qa/answers',
            page: 'pages/Admin/Answers',
          },
          {
            path: 'qa/settings',
            page: 'pages/Admin/QaSettings',
          },
          {
            path: 'tags/settings',
            page: 'pages/Admin/TagsSettings',
          },
          {
            path: 'security',
            page: 'pages/Admin/Security',
          },
          {
            path: 'themes',
            page: 'pages/Admin/Themes',
          },
          {
            path: 'customize',
            page: 'pages/Admin/CssAndHtml',
          },
          {
            path: 'general',
            page: 'pages/Admin/General',
          },
          {
            path: 'interface',
            page: 'pages/Admin/Interface',
          },
          {
            path: 'users',
            page: 'pages/Admin/Users',
          },
          {
            path: 'users/settings',
            page: 'pages/Admin/UsersSettings',
          },
          {
            path: 'users/:user_id',
            page: 'pages/Admin/UserOverview',
          },
          {
            path: 'smtp',
            page: 'pages/Admin/Smtp',
          },
          {
            path: 'branding',
            page: 'pages/Admin/Branding',
          },
          {
            path: 'rules/policies',
            page: 'pages/Admin/Policies',
          },
          {
            path: 'files',
            page: 'pages/Admin/Files',
          },
          {
            path: 'seo',
            page: 'pages/Admin/Seo',
          },
          {
            path: 'login',
            page: 'pages/Admin/Login',
          },
          {
            path: 'rules/privileges',
            page: 'pages/Admin/Privileges',
          },
          {
            path: 'installed-plugins',
            page: 'pages/Admin/Plugins/Installed',
          },
          {
            path: ':slug_name',
            page: 'pages/Admin/Plugins/Config',
          },
          {
            path: 'badges',
            page: 'pages/Admin/Badges',
          },
          {
            path: 'ai-assistant',
            page: 'pages/Admin/AiAssistant',
          },
          {
            path: 'ai-settings',
            page: 'pages/Admin/AiSettings',
          },
          {
            path: 'apikeys',
            page: 'pages/Admin/Apikeys',
          },
          {
            path: 'mcp',
            page: 'pages/Admin/Mcp',
          },
        ],
      },
      {
        path: '/user-center/auth',
        page: 'pages/UserCenter/Auth',
        guard: () => {
          const notLogged = guard.notLogged();
          return notLogged;
        },
      },
      {
        path: '/user-center/auth-failed',
        page: 'pages/UserCenter/AuthFailed',
      },
      {
        path: '*',
        page: 'pages/404',
      },
      {
        path: '50x',
        page: 'pages/50X',
      },
      // ai
      {
        page: 'pages/SideNavLayoutWithoutFooter',
        children: [
          {
            path: '/ai-assistant',
            page: 'pages/AiAssistant',
            guard: () => {
              return guard.logged();
            },
          },
          {
            path: '/ai-assistant/:id',
            page: 'pages/AiAssistant',
            guard: () => {
              return guard.logged();
            },
          },
        ],
      },
    ],
  },
  {
    path: '/',
    page: 'pages/Layout',
    loader: async () => {
      await guard.setupApp();
      return null;
    },
    children: [
      {
        page: 'pages/SideNavLayout',
        children: [
          {
            page: 'pages/Legal',
            children: [
              {
                path: 'tos',
                page: 'pages/Legal/Tos',
              },
              {
                path: 'privacy',
                page: 'pages/Legal/Privacy',
              },
            ],
          },
        ],
      },
      {
        path: '/users/unsubscribe',
        page: 'pages/Users/Unsubscribe',
      },
      {
        path: '403',
        page: 'pages/403',
      },
    ],
  },
  {
    path: '/install',
    page: 'pages/Install',
  },
  {
    path: '/maintenance',
    page: 'pages/Maintenance',
  },
];
export default routes;
