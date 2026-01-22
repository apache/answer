import request from '@/utils/request';

type McpConfig = {
  enabled: boolean;
  type: string;
  url: string;
  http_header: string;
};

export const getMcpConfig = () => {
  return request.get<McpConfig>(`/answer/admin/api/mcp-config`);
};

export const saveMcpConfig = (params: { enabled: boolean }) => {
  return request.put(`/answer/admin/api/mcp-config`, params);
};
