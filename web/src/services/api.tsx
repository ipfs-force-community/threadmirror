import { mockPosts } from '@data/mockData';
import {
  PostsApi,
  Configuration,
  PostsGetRequest,
  PostsIdGetRequest,
  PostsGet200Response as PostsGetResponse,
  PostsIdGet200Response as PostsIdGetResponse,
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

const postsApi = new PostsApi(config);

const fetchUserPosts = async (request: PostsGetRequest = {}) => {
  return await postsApi.postsGet(request);
};


const fetchPostDetail = async (request: PostsIdGetRequest) => {
  return await postsApi.postsIdGet(request);
};


const fetchUserPostsMock = async (request: PostsGetRequest = {}) => {
  console.log('use mock data for fetchUserTwitter');
  const startIndex = (request?.offset || 0) * (request?.limit || 0);
  const endIndex = startIndex + (request?.limit || 0);
  const posts = mockPosts.slice(startIndex, endIndex);
  posts.forEach(post => {
    console.log('post--------->', post.id);
  });
  return {
    meta: {
      total: mockPosts?.length || 0,
      limit: request.limit,
      offset: request.offset,
    },
    data: posts,
  } as PostsGetResponse;
};

const fetchPostDetailMock = async (request: PostsIdGetRequest) => {
  console.log('use mock data for fetchTwitterDetail');
  return {
    data: mockPosts.find(post => post.id === request.id),
  } as PostsIdGetResponse;
};

// 导出兼容 mock 模式的函数
export const useApiService = () => {
  return {
    fetchGetPosts: useMock ? fetchUserPostsMock : fetchUserPosts,
    fetchGetPostsId: useMock ? fetchPostDetailMock : fetchPostDetail
  };
};