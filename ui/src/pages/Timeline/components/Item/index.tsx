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

import { FC, useState } from 'react';
import { Button, Row, Col } from 'react-bootstrap';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import { Icon, BaseUserCard, DiffContent, FormatTime } from '@/components';
import { TIMELINE_NORMAL_ACTIVITY_TYPE } from '@/common/constants';
import * as Type from '@/common/interface';
import { getTimelineDetail } from '@/services';

interface Props {
  data: Type.TimelineItem;
  objectInfo: Type.TimelineObject;
  isAdmin: boolean;
  revisionList: Type.TimelineItem[];
}
const Index: FC<Props> = ({ data, isAdmin, objectInfo, revisionList }) => {
  const { t } = useTranslation('translation', { keyPrefix: 'timeline' });
  const [isOpen, setIsOpen] = useState(false);
  const [detailData, setDetailData] = useState({
    new_revision: {},
    old_revision: {},
  });

  const handleItemClick = async (id) => {
    if (!isOpen) {
      const revisionItem = revisionList?.find((v) => v.revision_id === id);
      let oldId;
      if (revisionList?.length > 0 && revisionItem) {
        const idIndex = revisionList.indexOf(revisionItem) || 0;
        if (idIndex === revisionList.length - 1) {
          oldId = 0;
        } else {
          oldId = revisionList[idIndex + 1].revision_id;
        }
      }
      const res = await getTimelineDetail({
        new_revision_id: id,
        old_revision_id: oldId,
      });
      setDetailData(res);
    }
    setIsOpen(!isOpen);
  };

  return (
    <>
      <tr>
        <td>
          <FormatTime time={data.created_at} />
          <br />
          {data.cancelled_at > 0 && <FormatTime time={data.cancelled_at} />}
        </td>
        <td className="text-nowrap">
          {(data.activity_type === 'rollback' ||
            data.activity_type === 'edited' ||
            data.activity_type === 'asked' ||
            data.activity_type === 'created' ||
            (objectInfo.object_type === 'answer' &&
              data.activity_type === 'answered')) && (
            <Button
              onClick={() => handleItemClick(data.revision_id)}
              variant="link"
              className="text-body p-0 btn-no-border">
              <Icon
                name="caret-right-fill"
                className={`me-1 ${isOpen ? 'rotate-90-deg' : 'rotate-0-deg'}`}
              />
              {t(data.activity_type)}
            </Button>
          )}
          {data.activity_type === 'accept' && (
            <Link
              to={`/questions/${objectInfo.question_id}/${data?.object_id}`}>
              {t(data.activity_type)}
            </Link>
          )}

          {objectInfo.object_type === 'question' &&
            data.activity_type === 'answered' && (
              <Link
                to={`/questions/${objectInfo.question_id}/${data.object_id}`}>
                {t(data.activity_type)}
              </Link>
            )}

          {data.activity_type === 'commented' && (
            <Link
              to={
                objectInfo.object_type === 'answer'
                  ? `/questions/${objectInfo.question_id}/${objectInfo.answer_id}?commentId=${data.object_id}`
                  : `/questions/${objectInfo.question_id}?commentId=${data.object_id}`
              }>
              {t(data.activity_type)}
            </Link>
          )}

          {TIMELINE_NORMAL_ACTIVITY_TYPE.includes(data.activity_type) && (
            <div>{t(data.activity_type)}</div>
          )}

          {data.cancelled && (
            <div className="text-danger">{t('cancelled')}</div>
          )}
        </td>
        <td>
          {data.activity_type === 'downvote' && !isAdmin ? (
            <div>{t('n_or_a')}</div>
          ) : (
            <BaseUserCard
              className="fs-normal"
              data={data?.user_info}
              showAvatar={false}
            />
          )}
        </td>
        <td>
          <div dangerouslySetInnerHTML={{ __html: data.comment }} />
        </td>
      </tr>
      <tr className={isOpen ? '' : 'd-none'}>
        <td colSpan={5} className="p-0 py-5">
          <Row className="justify-content-center">
            <Col xxl={8}>
              <DiffContent
                objectType={objectInfo.object_type}
                newData={detailData?.new_revision}
                oldData={detailData?.old_revision}
              />
            </Col>
          </Row>
        </td>
      </tr>
    </>
  );
};

export default Index;
