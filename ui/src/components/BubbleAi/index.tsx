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

import { FC, useEffect, useState, useRef } from 'react';
import { Button } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import { marked } from 'marked';
import copy from 'copy-to-clipboard';

import { voteConversation } from '@/services';
import { Icon, htmlRender } from '@/components';

interface IProps {
  canType?: boolean;
  chatId: string;
  isLast: boolean;
  isCompleted: boolean;
  content: string;
  minHeight?: number;
  actionData: {
    helpful: number;
    unhelpful: number;
  };
}

const BubbleAi: FC<IProps> = ({
  canType = false,
  isLast,
  isCompleted,
  content,
  chatId = '',
  actionData,
  minHeight = 0,
}) => {
  const { t } = useTranslation('translation', { keyPrefix: 'ai_assistant' });
  const [displayContent, setDisplayContent] = useState('');
  const [copyText, setCopyText] = useState<string>(t('copy'));
  const [isHelpful, setIsHelpful] = useState(false);
  const [isUnhelpful, setIsUnhelpful] = useState(false);
  const [canShowAction, setCanShowAction] = useState(false);
  const typewriterRef = useRef<{
    timer: NodeJS.Timeout | null;
    index: number;
    isTyping: boolean;
  }>({
    timer: null,
    index: 0,
    isTyping: false,
  });
  const fmtContainer = useRef<HTMLDivElement>(null);
  // add ref for ScrollIntoView
  const containerRef = useRef<HTMLDivElement>(null);

  const handleCopy = () => {
    const res = copy(displayContent);
    if (res) {
      setCopyText(t('copied', { keyPrefix: 'messages' }));
      setTimeout(() => {
        setCopyText(t('copy'));
      }, 1200);
    }
  };

  const handleVote = (voteType: 'helpful' | 'unhelpful') => {
    const isCancel =
      (voteType === 'helpful' && isHelpful) ||
      (voteType === 'unhelpful' && isUnhelpful);
    voteConversation({
      chat_completion_id: chatId,
      cancel: isCancel,
      vote_type: voteType,
    }).then(() => {
      setIsHelpful(voteType === 'helpful' && !isCancel);
      setIsUnhelpful(voteType === 'unhelpful' && !isCancel);
    });
  };

  useEffect(() => {
    if ((!canType || !isLast) && content) {
      // 如果不是最后一个消息，直接返回，不进行打字效果
      if (typewriterRef.current.timer) {
        clearInterval(typewriterRef.current.timer);
        typewriterRef.current.timer = null;
      }
      setDisplayContent(content);
      setCanShowAction(true);
      typewriterRef.current.timer = null;
      typewriterRef.current.isTyping = false;
      return;
    }
    // 当内容变化时，清理之前的计时器
    if (typewriterRef.current.timer) {
      clearInterval(typewriterRef.current.timer);
      typewriterRef.current.timer = null;
    }

    // 如果内容为空，则直接返回
    if (!content) {
      setDisplayContent('');
      return;
    }

    // 如果内容比当前显示的短，则重置
    if (content.length < displayContent.length) {
      setDisplayContent('');
      typewriterRef.current.index = 0;
    }

    // 如果内容与显示内容相同，不需要做任何事
    if (content === displayContent) {
      return;
    }

    typewriterRef.current.isTyping = true;

    // start typing animation
    typewriterRef.current.timer = setInterval(() => {
      const currentIndex = typewriterRef.current.index;
      if (currentIndex < content.length) {
        const remainingLength = content.length - currentIndex;
        const baseRandomNum = Math.floor(Math.random() * 3) + 2;
        let randomNum = Math.min(baseRandomNum, remainingLength);

        // 简单的单词边界检查（可选）
        const nextChar = content[currentIndex + randomNum];
        const prevChar = content[currentIndex + randomNum - 1];

        // 如果下一个字符是字母，当前字符也是字母，尝试调整到空格处
        if (
          nextChar &&
          /[a-zA-Z]/.test(nextChar) &&
          /[a-zA-Z]/.test(prevChar)
        ) {
          // 向前找1-2个字符，看看有没有空格
          for (
            let i = 1;
            i <= 2 && currentIndex + randomNum - i > currentIndex;
            i += 1
          ) {
            if (content[currentIndex + randomNum - i] === ' ') {
              randomNum = randomNum - i + 1;
              break;
            }
          }
          // 向后找1-2个字符，看看有没有空格
          for (
            let i = 1;
            i <= 2 && currentIndex + randomNum + i < content.length;
            i += 1
          ) {
            if (content[currentIndex + randomNum + i] === ' ') {
              randomNum = randomNum + i + 1;
              break;
            }
          }
        }

        const nextIndex = currentIndex + randomNum;
        const newContent = content.substring(0, nextIndex);
        setDisplayContent(newContent);
        typewriterRef.current.index = nextIndex;
        setCanShowAction(false);
      } else {
        clearInterval(typewriterRef.current.timer as NodeJS.Timeout);
        typewriterRef.current.timer = null;
        typewriterRef.current.isTyping = false;
        setCanShowAction(false);
      }
    }, 30);

    // eslint-disable-next-line consistent-return
    return () => {
      if (typewriterRef.current.timer) {
        clearInterval(typewriterRef.current.timer);
        typewriterRef.current.timer = null;
      }
    };
  }, [content, isCompleted]);

  useEffect(() => {
    setIsHelpful(actionData.helpful > 0);
    setIsUnhelpful(actionData.unhelpful > 0);
  }, [actionData]);

  useEffect(() => {
    if (fmtContainer.current && isCompleted) {
      htmlRender(fmtContainer.current, {
        copySuccessText: t('copied', { keyPrefix: 'messages' }),
        copyText: t('copy', { keyPrefix: 'messages' }),
      });
      const links = fmtContainer.current.querySelectorAll('a');
      links.forEach((link) => {
        link.setAttribute('target', '_blank');
      });
      setCanShowAction(true);
    }
  }, [isCompleted, fmtContainer.current]);

  return (
    <div
      className="rounded bubble-ai"
      ref={containerRef}
      style={{ minHeight: `${minHeight}px`, overflowAnchor: 'none' }}>
      <div id={chatId}>
        <div
          className="fmt text-break text-wrap"
          ref={fmtContainer}
          style={{ transition: 'all 0.2s ease' }}
          dangerouslySetInnerHTML={{ __html: marked.parse(displayContent) }}
        />

        {canShowAction && (
          <div className="action">
            <Button
              variant="link"
              className="p-0 link-secondary small me-3"
              onClick={handleCopy}>
              <Icon name="copy" />
              <span className="ms-1">{copyText}</span>
            </Button>
            <Button
              variant="link"
              className={`p-0 small me-3 ${isHelpful ? 'link-primary active' : 'link-secondary'}`}
              onClick={() => handleVote('helpful')}>
              <Icon name="hand-thumbs-up-fill" />
              <span className="ms-1">Helpful</span>
            </Button>
            <Button
              variant="link"
              className={`p-0 small me-3 ${isUnhelpful ? 'link-primary active' : 'link-secondary'}`}
              onClick={() => handleVote('unhelpful')}>
              <Icon name="hand-thumbs-down-fill" />
              <span className="ms-1">Unhelpful</span>
            </Button>
          </div>
        )}
      </div>
    </div>
  );
};

export default BubbleAi;
