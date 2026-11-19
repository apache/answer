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

import { memo, useEffect, useRef } from 'react';

import { customizeStore } from '@/stores';

/**
 * Re-create every script node in the container, because the browser runs a
 * script only when it is inserted as a node. dangerouslySetInnerHTML parses
 * script tags without running them.
 */
const activateScriptNodes = (el: HTMLElement) => {
  el.querySelectorAll('script').forEach((so) => {
    const script = document.createElement('script');
    // Wrap a classic script, to keep its declarations out of the page scope.
    // Copy every other script as it is: a module has its own scope, and a data
    // block such as application/ld+json must keep its content. The test matches
    // the types the browser runs.
    const type = so.getAttribute('type')?.trim().toLowerCase();
    const isClassic =
      !type || /^(text|application)\/(x-)?(java|ecma)script$/.test(type);
    // The line breaks matter: a script that ends in a line comment would
    // otherwise swallow the closing brace.
    script.text = isClassic ? `(() => {\n${so.text}\n})();` : so.text;
    for (let i = 0; i < so.attributes.length; i += 1) {
      const attr = so.attributes[i];
      script.setAttribute(attr.name, attr.value);
    }
    so.parentNode?.replaceChild(script, so);
  });
};

const Index = () => {
  const customBanner = customizeStore((state) => state.custom_banner);
  const slotRef = useRef<HTMLDivElement>(null);
  // The content whose scripts already ran. It stops the scripts from running
  // again when React re-runs this effect for the same content.
  const activatedBanner = useRef<string | null>(null);

  useEffect(() => {
    if (!customBanner) {
      activatedBanner.current = null;
      return;
    }
    if (!slotRef.current || customBanner === activatedBanner.current) {
      return;
    }
    activatedBanner.current = customBanner;
    activateScriptNodes(slotRef.current);
  }, [customBanner]);

  if (!customBanner) {
    return null;
  }

  return (
    <div
      id="custom-banner-slot"
      ref={slotRef}
      dangerouslySetInnerHTML={{ __html: customBanner }}
    />
  );
};

export default memo(Index);
