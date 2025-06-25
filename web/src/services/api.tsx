import { mockTweetData, userProfile } from '@data/mockData';
import { UserListResponse, UserTwitterResponse, TwitterDetailResponse } from '@src/types';
import { useApi } from './api_class';

const useMock = process.env.REACT_APP_USE_MOCK === 'true' || false;
console.log('useMock', useMock);
export const useUserApi = () => {
  const { request } = useApi();

  // 获取用户列表
  const fetchUserList = async (limit: number, offset: number) => {
    return await request<UserListResponse[]>('/user-list', {
      query: { limit, offset },
    });
  };

  // 获取用户 Twitter 数据
  const fetchUserTwitter = async (id: string, limit: number, offset: number) => {
    return await request<UserTwitterResponse>(`/user/${id}`, {
      query: { limit, offset },
    });
  };

  // 获取 Twitter 详情
  const fetchTwitterDetail = async (id: string) => {
    return await request<TwitterDetailResponse>(`/posts/${id}`);
  };

  return {
    fetchUserList,
    fetchUserTwitter,
    fetchTwitterDetail
  };
};

// mock 实现
const fetchUserListMock = async (limit: number, offset: number) => {
  throw new Error('Mock not implemented for fetchUserList');
};

const fetchUserTwitterMock = async (id: string, limit: number, offset: number) => {
  console.log('use mock data for fetchUserTwitter');
  return {
    auth: userProfile,
    msg: [mockTweetData, mockTweetData, mockTweetData],
    // total: mockTweetData.length,
  } as UserTwitterResponse;
};

const fetchTwitterDetailMock = async (id: string) => {
  console.log('use mock data for fetchTwitterDetail');
  return mockTweetData as TwitterDetailResponse;
};

// 导出兼容 mock 模式的函数
export const useApiService = () => {
  const api = useUserApi();
  
  return {
    fetchUserList: useMock ? fetchUserListMock : api.fetchUserList,
    fetchUserTwitter: useMock ? fetchUserTwitterMock : api.fetchUserTwitter,
    fetchTwitterDetail: useMock ? fetchTwitterDetailMock : api.fetchTwitterDetail
  };
};