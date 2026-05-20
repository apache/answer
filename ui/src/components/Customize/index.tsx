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
import { useLocation } from 'react-router-dom';

import { customizeStore } from '@/stores';

const CUSTOM_MARK_HEAD = 'customize_head';
const CUSTOM_MARK_HEADER = 'customize_header';
const CUSTOM_MARK_FOOTER = 'customize_footer';

const makeMarker = (mark) => {
  return `<!--${mark}-->`;
};

const ActivateScriptNodes = (el, part) => {
  let startMarkNode;
  const scriptList: HTMLScriptElement[] = [];
  const { childNodes } = el;
  for (let i = 0; i < childNodes.length; i += 1) {
    const node = childNodes[i];
    if (node.nodeType === 8 && node.nodeValue === part) {
      if (!startMarkNode) {
        startMarkNode = node;
      } else {
        // this is the endMarkNode
        break;
      }
    }
    if (
      startMarkNode &&
      node.nodeType === 1 &&
      node.nodeName.toLowerCase() === 'script'
    ) {
      scriptList.push(node);
    }
  }
  scriptList?.forEach((so) => {
    const script = document.createElement('script');
    script.text = `(() => {${so.text}})();`;
    for (let i = 0; i < so.attributes.length; i += 1) {
      const attr = so.attributes[i];
      script.setAttribute(attr.name, attr.value);
    }
    el.replaceChild(script, so);
  });
};

type pos = 'afterbegin' | 'beforeend';
const renderCustomArea = (el, part, pos: pos, content: string = '') => {
  let startMarkNode;
  let endMarkNode;
  const { childNodes } = el;
  for (let i = 0; i < childNodes.length; i += 1) {
    const node = childNodes[i];
    if (node.nodeType === 8 && node.nodeValue === part) {
      if (!startMarkNode) {
        startMarkNode = node;
      } else {
        endMarkNode = node;
        break;
      }
    }
  }

  if (startMarkNode && endMarkNode) {
    while (
      startMarkNode.nextSibling &&
      startMarkNode.nextSibling !== endMarkNode
    ) {
      el.removeChild(startMarkNode.nextSibling);
    }
  }
  if (startMarkNode) {
    el.removeChild(startMarkNode);
  }
  if (endMarkNode) {
    el.removeChild(endMarkNode);
  }
  el.insertAdjacentHTML(pos, makeMarker(part));
  el.insertAdjacentHTML(pos, content);
  el.insertAdjacentHTML(pos, makeMarker(part));
  ActivateScriptNodes(el, part);
};
const handleCustomHead = (content) => {
  const el = document.head;
  renderCustomArea(el, CUSTOM_MARK_HEAD, 'beforeend', content);
};

const handleCustomHeader = (content) => {
  const el = document.body;
  renderCustomArea(el, CUSTOM_MARK_HEADER, 'afterbegin', content);
};

const handleCustomFooter = (content) => {
  const el = document.body;
  renderCustomArea(el, CUSTOM_MARK_FOOTER, 'beforeend', content);
};

const Index: FC = () => {
  const { custom_head, custom_header, custom_footer } = customizeStore(
    (state) => state,
  );
  const { pathname } = useLocation();

  useEffect(() => {
    const isSeo = document.querySelector('meta[name="go-template"]');
    if (!isSeo) {
      setTimeout(() => {
        handleCustomHead(custom_head);
      }, 1000);
      handleCustomHeader(custom_header);
      handleCustomFooter(custom_footer);
    } else {
      isSeo.remove();
    }
  }, [custom_head, custom_header, custom_footer]);

  useEffect(() => {
    /**
     * description:  Activate scripts with data-client attribute when route changes
     */
    const allScript = document.body.querySelectorAll('script[data-client]');
    allScript.forEach((scriptNode) => {
      const script = document.createElement('script');
      script.setAttribute('data-client', 'true');
      // If the script is already wrapped in an IIFE, use it directly; otherwise, wrap it in an IIFE
      if (
        /^\s*\(\s*function\s*\(\s*\)\s*{/.test(
          (scriptNode as HTMLScriptElement).text,
        ) ||
        /^\s*\(\s*\(\s*\)\s*=>\s*{/.test((scriptNode as HTMLScriptElement).text)
      ) {
        script.text = (scriptNode as HTMLScriptElement).text;
      } else {
        script.text = `(() => {${(scriptNode as HTMLScriptElement).text}})();`;
      }
      for (let i = 0; i < scriptNode.attributes.length; i += 1) {
        const attr = scriptNode.attributes[i];
        if (attr.name !== 'data-client') {
          script.setAttribute(attr.name, attr.value);
        }
      }
      scriptNode.parentElement?.replaceChild(script, scriptNode);
    });
  }, [pathname]);

  return null;
};

export default memo(Index);
