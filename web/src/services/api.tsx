import { mockMentions } from '@data/mockData';
import {
  MentionsApi,
  Configuration,
  MentionsGetRequest,
  MentionsIdGetRequest,
  MentionsGet200Response as MentionsGetResponse,
  MentionsIdGet200Response as MentionsIdGetResponse,
} from '@client/index';
import { getAuthToken } from '@utils/cookie';

const useMock = process.env.REACT_APP_USE_MOCK === 'true' || false;
const config = new Configuration({
  basePath: process.env.REACT_APP_API_BASE_URL || 'http://localhost:8080/api/v1',
  middleware: [
    {
      pre: async (context) => {
        const token = await getAuthToken();
        if (token) {
          context.init.headers = {
            ...context.init.headers,
            'Authorization': `Bearer ${token}`
          };
        }
      }
    }
  ]
});

const mentionsApi = new MentionsApi(config);

const fetchUserMentions = async (request: MentionsGetRequest = {}) => {
  return await mentionsApi.mentionsGet(request);
};


const fetchMentionDetail = async (request: MentionsIdGetRequest) => {
  return await mentionsApi.mentionsIdGet(request);
};


const fetchUserMentionsMock = async (request: MentionsGetRequest = {}) => {
  console.log('use mock data for fetchUserTwitter');
  const startIndex = (request?.offset || 0) * (request?.limit || 0);
  const endIndex = Math.min(startIndex + (request?.limit || 0), mockMentions.length);
  const mentions = mockMentions.slice(startIndex, endIndex);
  console.log('[mock]mentions--------->', '[', startIndex, ',', endIndex, ']');

  return {
    meta: {
      total: mockMentions?.length || 0,
      limit: request.limit,
      offset: request.offset,
    },
    data: mentions,
  } as MentionsGetResponse;
};

const fetchMentionDetailMock = async (request: MentionsIdGetRequest) => {
  console.log('use mock data for fetchTwitterDetail');
  return {
    data: mockMentions.find(mention => mention.id === request.id),
  } as MentionsIdGetResponse;
};

// 导出兼容 mock 模式的函数
export const useApiService = () => {
  return {
    fetchGetMentions: useMock ? fetchUserMentionsMock : fetchUserMentions,
    fetchGetMentionsId: useMock ? fetchMentionDetailMock : fetchMentionDetail
  };
};