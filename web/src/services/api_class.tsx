import { useAuth0 } from '@auth0/auth0-react';

interface ApiRequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE';
  body?: Record<string, any>;
  headers?: Record<string, string>;
  query?: Record<string, string | number>;
}

// 错误处理类
class ApiError extends Error {
  constructor(public status: number, public statusText: string, message: string) {
    super(message);
    this.name = 'ApiError';
  }
}

// API 客户端类
class ApiClient {
  private baseUrl: string;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
  }

  // 构造完整 URL，附加查询参数
  private buildUrl(endpoint: string, query?: Record<string, string | number>): string {
    const url = new URL(endpoint.startsWith('http') ? endpoint : `${this.baseUrl}${endpoint}`);
    if (query) {
      Object.entries(query).forEach(([key, value]) => url.searchParams.append(key, String(value)));
    }
    return url.toString();
  }

  // 核心 API 调用方法
  async request<T>(
    endpoint: string,
    options: ApiRequestOptions = {},
    token?: string
  ): Promise<T> {
    const { method = 'GET', body, headers = {}, query } = options;
    const url = this.buildUrl(endpoint, query);

    // 构造请求头
    const requestHeaders: HeadersInit = {
      'Content-Type': 'application/json',
      ...headers,
    };

    // 如果提供了 token，添加到 Authorization 头
    if (token) {
      requestHeaders['Authorization'] = `Bearer ${token}`;
    }
    
    try {
      console.log('发送请求到:', url);
      const response = await fetch(url, {
        method,
        headers: requestHeaders,
        body: body ? JSON.stringify(body) : undefined,
      });
      console.log('收到响应:', response);

      // 确保response不是undefined
      if (!response) {
        throw new ApiError(0, 'No Response', 'Server did not return a response');
      }

      // 处理非 2xx 响应
      if (!response.ok) {
        throw new ApiError(response.status, response.statusText, `Request failed: ${response.statusText}`);
      }

      // 解析JSON响应
      const data = await response.json();
      return data;
    } catch (error) {
      console.error('API请求错误:', error);
      if (error instanceof ApiError) {
        throw error;
      }
      throw new ApiError(0, 'Network Error', error instanceof Error ? error.message : 'Unknown error');
    }
  }
}

// 创建 API 客户端实例
const API_BASE_URL = process.env.REACT_APP_API_BASE_URL || 'http://localhost:63303/api/v1';
const apiClient = new ApiClient(API_BASE_URL);

// React Hook 用于获取带令牌的 API 调用
export const useApi = () => {
  const { getAccessTokenSilently, isAuthenticated } = useAuth0();

  const request = async <T extends unknown>(
    endpoint: string,
    options: ApiRequestOptions = {}
  ): Promise<T> => {
    if (!isAuthenticated) {
      throw new Error('User is not authenticated');
    }

    try {
      // 获取访问令牌
      const token = await getAccessTokenSilently();
      console.log('获取到token:', token ? '成功' : '失败');
      // 使用 API 客户端发起请求
      return await apiClient.request<T>(endpoint, options, token);
    } catch (error) {
      console.error('请求处理错误:', error);
      if (error instanceof Error && error.message.includes('login_required')) {
        throw new Error('Please log in to access this resource');
      }
      throw error;
    }
  };

  return { request };
};
