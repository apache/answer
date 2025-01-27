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

import { memo, FC, useState, useEffect } from 'react';
import { Dropdown, OverlayTrigger, Tooltip } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import { FacebookShareButton, TwitterShareButton } from 'next-share';
import copy from 'copy-to-clipboard';

import { BASE_ORIGIN } from '@/router/alias';
import { loggedUserInfoStore } from '@/stores';

interface IProps {
  type: 'answer' | 'question';
  qid: any;
  aid?: any;
  title: string;
  // slugTitle: string;
}

const Index: FC<IProps> = ({ type, qid, aid, title }) => {
  const user = loggedUserInfoStore((state) => state.user);
  const [show, setShow] = useState(false);
  const [showTip, setShowTip] = useState(false);
  const [canSystemShare, setSystemShareState] = useState(false);
  const { t } = useTranslation();
  let baseUrl =
    type === 'question'
      ? `${BASE_ORIGIN}/questions/${qid}`
      : `${BASE_ORIGIN}/questions/${qid}/${aid}`;
  if (user.id) {
    baseUrl = `${baseUrl}?share=${user.username}`;
  }

  const closeShare = () => {
    setShowTip(false);
    setShow(false);
  };

  const handleCopy = (evt) => {
    evt.preventDefault();
    evt.stopPropagation();
    let copyText = baseUrl;
    if (title) {
      copyText = `${title} ${baseUrl}`;
    }
    copy(copyText);
    setShowTip(true);
    setTimeout(closeShare, 1000);
  };

  const systemShare = () => {
    navigator.share({
      title,
      text: `${title} - Answer：`,
      url: baseUrl,
    });
  };
  useEffect(() => {
    if (window.navigator?.canShare?.({ text: 'can_share' })) {
      setSystemShareState(true);
    }
  }, []);
  return (
    <Dropdown show={show} onToggle={closeShare}>
      <Dropdown.Toggle
        id="dropdown-share"
        as="a"
        className="no-toggle small link-secondary pointer d-flex"
        onClick={() => setShow(true)}
        style={{ lineHeight: '23px' }}>
        {t('share.name')}
      </Dropdown.Toggle>
      <Dropdown.Menu style={{ minWidth: '195px' }}>
        <OverlayTrigger
          trigger="click"
          placement="left"
          show={showTip}
          overlay={<Tooltip>{t('share.copied')}</Tooltip>}>
          <Dropdown.Item onClick={handleCopy} eventKey="copy">
            {t('share.copy')}
          </Dropdown.Item>
        </OverlayTrigger>
        <Dropdown.Item eventKey="facebook">
          <FacebookShareButton
            title={title}
            url={baseUrl}
            className="w-100 py-1 px-3 text-start">
            {t('share.facebook')}
          </FacebookShareButton>
        </Dropdown.Item>
        <Dropdown.Item>
          <TwitterShareButton
            title={title}
            url={baseUrl}
            className="w-100 py-1 px-3 text-start">
            {t('share.twitter')}
          </TwitterShareButton>
        </Dropdown.Item>
        {canSystemShare && (
          <Dropdown.Item onClick={systemShare}>{t('share.via')}</Dropdown.Item>
        )}
      </Dropdown.Menu>
    </Dropdown>
  );
};

export default memo(Index);
