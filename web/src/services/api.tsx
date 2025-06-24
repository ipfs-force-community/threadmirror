import { mockTweetData, userProfile } from '@data/mockData';
import { UserListResponse, UserTwitterResponse, TwitterDetailResponse } from '@src/types';

const API_BASE_URL = process.env.REACT_APP_API_BASE_URL || 'http://localhost:63303/api';
const useMock = process.env.REACT_APP_USE_MOCK === 'true' || false;

const fetchUserListApi = async (limit: number, offset: number) => {
    const response = await fetch(`${API_BASE_URL}/user-list?limit=${limit}&offset=${offset}`);
    if (!response.ok) throw new Error('Failed to fetch user list');
    return response.json() as Promise<UserListResponse[]>;
};

const fetchUserTwitterApi = async (id: string, limit: number, offset: number) => {
    const response = await fetch(`${API_BASE_URL}/user/${id}?limit=${limit}&offset=${offset}`);
    if (!response.ok) throw new Error('Failed to fetch user Twitter data');
    return response.json() as Promise<UserTwitterResponse>;
};

const fetchTwitterDetailApi = async (id: string) => {
    const response = await fetch(`${API_BASE_URL}/twitter/${id}`);
    if (!response.ok) throw new Error('Failed to fetch Twitter detail');
    return response.json() as Promise<TwitterDetailResponse>;
};


// mock

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


// export 
export const fetchUserList = useMock ? fetchUserListMock : fetchUserListApi;
export const fetchUserTwitter = useMock ? fetchUserTwitterMock : fetchUserTwitterApi;
export const fetchTwitterDetail = useMock ? fetchTwitterDetailMock : fetchTwitterDetailApi;