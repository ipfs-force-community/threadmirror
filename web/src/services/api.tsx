import { mockThreads } from '@data/mockData';
import {
  MentionsApi,
  DefaultApi,
  Configuration,
  MentionsGetRequest,
  ThreadIdGetRequest,
  MentionsGet200Response as MentionsGetResponse,
  ThreadIdGet200Response as ThreadIdGetResponse,
} from '@client/index';
import { getAuthToken } from '@utils/cookie';

const useMock = process.env.REACT_APP_USE_MOCK === 'true' || false;
const API_BASE_PATH = process.env.REACT_APP_API_BASE_URL || 'http://localhost:8080/api/v1';

const config = new Configuration({
  basePath: API_BASE_PATH,
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
const defaultApi = new DefaultApi(config);

const fetchUserMentions = async (request: MentionsGetRequest = {}) => {
  return await mentionsApi.mentionsGet(request);
};

const fetchThreadDetail = async (request: ThreadIdGetRequest) => {
  return await defaultApi.threadIdGet(request);
};

const fetchUserMentionsMock = async (request: MentionsGetRequest = {}) => {
  console.log('use mock data for fetchUserTwitter');
  const startIndex = (request?.offset || 0) * (request?.limit || 0);
  const endIndex = Math.min(startIndex + (request?.limit || 0), mockThreads.length);
  const _mentions = mockThreads.slice(startIndex, endIndex);
  console.log('[mock]mentions--------->', '[', startIndex, ',', endIndex, ']');

  return {
    meta: {
      total: mockThreads?.length || 0,
      limit: request.limit,
      offset: request.offset,
    },
    data: {}, // TODO: fix mock
  } as MentionsGetResponse;
};

const fetchThreadDetailMock = async (request: ThreadIdGetRequest) => {
  console.log('use mock data for fetchThreadDetail');
  return {
    data: mockThreads.find(thread => thread.id === request.id),
  } as ThreadIdGetResponse;
};

const fetchShareImage = async (threadId: string) => {
  if (useMock) {
    // Return placeholder blob in mock mode
    return new Blob();
  }

  const token = await getAuthToken();
  const headers: Record<string, string> = {};
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const response = await fetch(`${API_BASE_PATH}/share?thread_id=${threadId}`, {
    method: 'GET',
    headers,
  });

  if (!response.ok) {
    throw new Error(`Failed to download share image: ${response.statusText}`);
  }

  return await response.blob();
};

// 导出兼容 mock 模式的函数
export const useApiService = () => {
  return {
    fetchGetMentions: useMock ? fetchUserMentionsMock : fetchUserMentions,
    fetchGetThreadId: useMock ? fetchThreadDetailMock : fetchThreadDetail,
    fetchGetShare: fetchShareImage,
  };
};