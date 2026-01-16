import qs from 'qs';

import request from '@/utils/request';
import type * as Type from '@/common/interface';

export const getConversationList = (params: Type.Paging) => {
  return request.get<{ count: number; list: Type.ConversationListItem[] }>(
    `/answer/api/v1/ai/conversation/page?${qs.stringify(params)}`,
  );
};

export const getConversationDetail = (id: string) => {
  return request.get<Type.ConversationDetail>(
    `/answer/api/v1/ai/conversation?conversation_id=${id}`,
  );
};

// /answer/api/v1/ai/conversation/vote
export const voteConversation = (params: Type.VoteConversationParams) => {
  return request.post('/answer/api/v1/ai/conversation/vote', params);
};
