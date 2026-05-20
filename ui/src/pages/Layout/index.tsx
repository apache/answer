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

import { FC, memo, useEffect } from 'react';
import { Outlet, useLocation, ScrollRestoration } from 'react-router-dom';
import { HelmetProvider } from 'react-helmet-async';

import { SWRConfig } from 'swr';
import classnames from 'classnames';

import {
  toastStore,
  loginToContinueStore,
  errorCodeStore,
  siteSecurityStore,
  themeSettingStore,
} from '@/stores';
import {
  Header,
  Toast,
  Customize,
  CustomizeTheme,
  PageTags,
  HttpErrorContent,
} from '@/components';
import { LoginToContinueModal, BadgeModal } from '@/components/Modal';
import { changeTheme, Storage, scrollToElementTop } from '@/utils';
import { useQueryNotificationStatus } from '@/services';
import { useExternalToast } from '@/hooks';
import { EXTERNAL_CONTENT_DISPLAY_MODE } from '@/common/constants';

const Layout: FC = () => {
  const location = useLocation();
  const { msg: toastMsg, variant, clear: toastClear } = toastStore();
  const externalToast = useExternalToast();
  const externalContentDisplay = siteSecurityStore(
    (state) => state.external_content_display,
  );
  const closeToast = () => {
    toastClear();
  };
  const { code: httpStatusCode, reset: httpStatusReset } = errorCodeStore();
  const { show: showLoginToContinueModal } = loginToContinueStore();
  const { data: notificationData } = useQueryNotificationStatus();
  const layout = themeSettingStore((state) => state.layout);
  useEffect(() => {
    // handle footnote links
    const fixFootnoteLinks = () => {
      const footnoteLinks = document.querySelectorAll(
        'a[href^="#"]:not([data-footnote-fixed])',
      );

      footnoteLinks.forEach((link) => {
        link.setAttribute('data-footnote-fixed', 'true');
        const href = link.getAttribute('href');
        link.addEventListener('click', (e) => {
          e.preventDefault();
          const targetId = href?.substring(1) || '';
          const targetElement = document.getElementById(targetId);

          if (targetElement) {
            window.history.pushState(null, '', `${location.pathname}${href}`);

            scrollToElementTop(targetElement);
          }
        });
      });

      if (window.location.hash) {
        const { hash } = window.location;
        const targetElement = document.getElementById(hash.substring(1));

        if (targetElement) {
          setTimeout(() => {
            scrollToElementTop(targetElement);
          }, 100);
        }
      }
    };
    fixFootnoteLinks();

    const observer = new MutationObserver(() => {
      fixFootnoteLinks();
    });

    observer.observe(document.body, {
      childList: true,
      subtree: true,
      attributes: true,
      attributeFilter: ['id', 'href'],
    });

    const handleHashChange = () => {
      if (window.location.hash) {
        const { hash } = window.location;
        const targetElement = document.getElementById(hash.substring(1));

        if (targetElement) {
          setTimeout(() => {
            scrollToElementTop(targetElement);
          }, 100);
        }
      }
    };

    window.addEventListener('hashchange', handleHashChange);

    return () => {
      observer.disconnect();
      window.removeEventListener('hashchange', handleHashChange);
    };
  }, [location.pathname]);

  useEffect(() => {
    httpStatusReset();
  }, [location]);

  useEffect(() => {
    const systemThemeQuery = window.matchMedia('(prefers-color-scheme: dark)');
    function handleSystemThemeChange(event) {
      if (event.matches) {
        changeTheme('dark');
      } else {
        changeTheme('light');
      }
    }

    systemThemeQuery.addListener(handleSystemThemeChange);

    return () => {
      systemThemeQuery.removeListener(handleSystemThemeChange);
    };
  }, []);

  const replaceImgSrc = () => {
    const storageUserExternalMode = Storage.get(EXTERNAL_CONTENT_DISPLAY_MODE);
    const images = document.querySelectorAll(
      'img:not([data-processed])',
    ) as NodeListOf<HTMLImageElement>;

    images.forEach((img) => {
      // Mark as processed to avoid duplication
      img.setAttribute('data-processed', 'true');

      if (
        img.src &&
        storageUserExternalMode !== 'always' &&
        !img.src.startsWith('/') &&
        !img.src.startsWith('data:') &&
        !img.src.startsWith('blob:') &&
        !img.src.startsWith(window.location.origin)
      ) {
        externalToast.onShow();
        img.dataset.src = img.src;
        img.removeAttribute('src');
      }
    });
  };

  useEffect(() => {
    // Controlling the loading of external image resources
    const observer = new MutationObserver((mutationsList) => {
      let hasNewImages = false;
      mutationsList.forEach((mutation) => {
        if (mutation.type === 'childList') {
          mutation.addedNodes.forEach((node) => {
            if (
              node.nodeName === 'IMG' ||
              (node.nodeType === 1 &&
                (node as Element).querySelectorAll('img:not([data-processed])')
                  .length > 0)
            ) {
              hasNewImages = true;
            }
          });
        }
      });
      if (hasNewImages) {
        replaceImgSrc();
      }
    });

    if (externalContentDisplay !== 'always_display') {
      observer.observe(document.body, { childList: true, subtree: true });
    }

    return () => observer.disconnect();
  }, [externalContentDisplay]);
  return (
    <HelmetProvider>
      <PageTags />
      <CustomizeTheme />
      <SWRConfig
        value={{
          revalidateOnFocus: false,
        }}>
        <Header />
        <div
          className={classnames(
            'position-relative page-wrap d-flex flex-column flex-fill',
            layout === 'Fixed-width' ? 'container-xxl' : '',
          )}>
          {httpStatusCode ? (
            <HttpErrorContent httpCode={httpStatusCode} />
          ) : (
            <Outlet />
          )}
        </div>
        <Toast msg={toastMsg} variant={variant} onClose={closeToast} />
        <Customize />
        <LoginToContinueModal visible={showLoginToContinueModal} />
        <BadgeModal
          badge={notificationData?.badge_award}
          visible={Boolean(notificationData?.badge_award)}
        />
        <ScrollRestoration />
      </SWRConfig>
    </HelmetProvider>
  );
};

export default memo(Layout);
