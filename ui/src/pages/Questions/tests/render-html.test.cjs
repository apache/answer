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

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const test = require('node:test');
const ts = require('typescript');
const { JSDOM } = require('jsdom');

const dom = new JSDOM('<!doctype html><html><body></body></html>', {
  url: 'http://localhost/',
});
Object.assign(globalThis, {
  window: dom.window,
  document: dom.window.document,
  HTMLElement: dom.window.HTMLElement,
  MutationObserver: dom.window.MutationObserver,
  IS_REACT_ACT_ENVIRONMENT: true,
});
const React = require('react');
const { createRoot } = require('react-dom/client');
const { act } = React;
const noop = () => {};
const translate = (key) => key;
const passthrough = ({ children }) => children;
const Empty = () => null;
let queryData;
let activeTargets;
let initializedTargets;
let mutationCount;

// Model the public plugin contract: effects subscribe to the supplied element.
// UI/services are unrelated to target lifecycle and are stubbed below.
function useRenderHtmlPlugin(element) {
  React.useEffect(() => {
    if (!element) return undefined;
    const target = element instanceof HTMLElement ? element : element.current;
    if (!target) return undefined;
    initializedTargets.push(target);
    activeTargets.add(target);
    const observer = new MutationObserver(() => {
      mutationCount += 1;
    });
    observer.observe(target, { childList: true, subtree: true });
    return () => {
      observer.disconnect();
      activeTargets.delete(target);
    };
  }, [element]);
}

function loadComponent(relativePath) {
  const filename = path.resolve(__dirname, '..', relativePath);
  const source = fs.readFileSync(filename, 'utf8');
  const { outputText } = ts.transpileModule(source, {
    fileName: filename,
    compilerOptions: {
      target: ts.ScriptTarget.ES2020,
      module: ts.ModuleKind.CommonJS,
      jsx: ts.JsxEmit.ReactJSX,
      esModuleInterop: true,
    },
  });
  const Form = Object.assign(Empty, {
    Group: Empty,
    Label: Empty,
    Select: Empty,
    Control: Object.assign(Empty, { Feedback: Empty }),
  });
  const bootstrap = new Proxy(
    {
      Form,
      Row: passthrough,
      Col: passthrough,
      Card: Object.assign(Empty, { Header: Empty, Body: Empty }),
    },
    {
      get: (object, key) => object[key] || Empty,
    },
  );
  const mocks = {
    '@/components': {
      Tag: Empty,
      Actions: Empty,
      Operate: Empty,
      UserCard: Empty,
      BaseUserCard: Empty,
      Icon: Empty,
      Comment: passthrough,
      FormatTime: Empty,
      ImgViewer: passthrough,
      Editor: Empty,
      htmlRender: noop,
    },
    '@/utils/pluginKit': { useRenderHtmlPlugin, useCaptchaPlugin: () => null },
    '@/utils': {
      formatCount: String,
      guard: {},
      scrollToElementTop: noop,
      bgFadeOut: noop,
      handleFormError: noop,
    },
    '@/hooks': { usePageTags: noop, usePromptWithUnload: noop },
    '@/services': {
      acceptanceAnswer: noop,
      following: noop,
      useQueryAnswerInfo: () => ({ data: queryData }),
      useQueryRevisions: () => ({ data: [] }),
      modifyAnswer: noop,
    },
    '@/router/pathFactory': {
      pathFactory: { questionLanding: () => '/questions/1' },
    },
    'react-bootstrap': bootstrap,
    'react-i18next': { useTranslation: () => ({ t: translate }) },
    'react-router-dom': {
      Link: passthrough,
      useSearchParams: () => [new URLSearchParams()],
      useParams: () => ({ aid: '1', qid: '2' }),
      useNavigate: () => noop,
    },
    './index.scss': {},
  };
  const module = { exports: {} };
  const localRequire = (id) =>
    Object.hasOwn(mocks, id) ? mocks[id] : require(id);
  vm.runInThisContext(
    `(function(require, module, exports) {${outputText}\n})`,
    {
      filename,
    },
  )(localRequire, module, module.exports);
  return module.exports.default;
}

const data = {
  id: 'answer-1',
  question_id: 'question-1',
  is_followed: false,
  html: '<p>\\(x+1\\)</p>',
  question: { title: 'Question', html: '<p>\\(x+1\\)</p>' },
};

for (const [name, file, selector] of [
  ['answer', 'Detail/components/Answer/index.tsx', 'article.fmt'],
  ['question', 'Detail/components/Question/index.tsx', 'article.fmt'],
  [
    'answer editor question',
    'EditAnswer/index.tsx',
    '.question-content-wrap .content',
  ],
]) {
  for (const delayed of [false, true]) {
    test(`${name}: plugins attach to content on ${delayed ? 'asynchronous' : 'first'} mount and clean up`, async () => {
      activeTargets = new Set();
      initializedTargets = [];
      mutationCount = 0;
      queryData = delayed ? undefined : data;
      const Component = loadComponent(file);
      const container = document.createElement('div');
      document.body.append(container);
      const root = createRoot(container);
      const props = (value) => ({
        data: value || {},
        aid: 'different-answer',
        callback: noop,
        initPage: noop,
      });
      try {
        await act(async () =>
          root.render(React.createElement(Component, props(queryData))),
        );
        if (delayed) {
          queryData = data;
          await act(async () =>
            root.render(React.createElement(Component, props(data))),
          );
        }
        const target = container.querySelector(selector);
        assert.ok(target, 'content DOM is mounted');
        assert.equal(activeTargets.size, 1, 'exactly one active plugin target');
        assert.ok(
          activeTargets.has(target),
          'plugin targets content, not the wrapper/comments',
        );
        assert.ok(
          initializedTargets.includes(target),
          'plugin initialized without unrelated user updates',
        );

        const updated = {
          ...data,
          html: '<p>\\(y+2\\)</p>',
          question: { ...data.question, html: '<p>\\(y+2\\)</p>' },
        };
        queryData = updated;
        await act(async () =>
          root.render(React.createElement(Component, props(updated))),
        );
        assert.ok(target.textContent.includes('y+2'));
        assert.ok(
          mutationCount > 0,
          'the plugin still observes asynchronous content changes',
        );
        assert.equal(activeTargets.size, 1);
        if (name !== 'answer editor question') {
          await act(async () =>
            root.render(React.createElement(Component, props({}))),
          );
          assert.equal(
            activeTargets.size,
            0,
            'removing content releases the old target',
          );
          await act(async () =>
            root.render(React.createElement(Component, props(updated))),
          );
          const replacement = container.querySelector(selector);
          assert.notEqual(replacement, target);
          assert.ok(
            activeTargets.has(replacement),
            'plugins attach when content returns',
          );
        }
        await act(async () => root.unmount());
        assert.equal(
          activeTargets.size,
          0,
          'unmount releases the plugin subscription',
        );
      } finally {
        if (container.firstChild) await act(async () => root.unmount());
        container.remove();
      }
    });
  }
}
