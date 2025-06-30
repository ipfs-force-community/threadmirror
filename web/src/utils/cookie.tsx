import { Cookies } from 'react-cookie';

// 默认cookie配置，过期时间7天
export const getDefaultCookieOptions = () => ({
    path: '/',
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax' as 'lax',
    maxAge: 24 * 60 * 60, // 1天
    domain: window.location.hostname
});

// 创建cookies实例
const cookies = new Cookies();

// 设置cookie
export const setCookieValue = (name: string, value: any, customOptions = {}) => {
    const options = { ...getDefaultCookieOptions(), ...customOptions };

    if (typeof value === 'object') {
        cookies.set(name, JSON.stringify(value), options);
    } else {
        cookies.set(name, value, options);
    }
};

// 获取cookie
export const getCookieValue = <T = any,>(name: string): T | null => {
    const value = cookies.get(name);
    if (!value) return null;

    // 如果预期返回类型不是字符串，尝试解析JSON
    if (typeof value === 'string') {
        try {
            // 只有看起来像JSON的字符串才尝试解析
            if (value.startsWith('{') || value.startsWith('[')) {
                return JSON.parse(value) as T;
            }
        } catch (e) {
            console.warn(`无法解析cookie ${name} 的JSON值:`, e);
        }
    }

    return value as unknown as T;
};

// 删除cookie
export const removeCookieValue = (name: string, customOptions = {}) => {
    const options = { path: '/', ...customOptions };
    cookies.remove(name, options);
};

// 认证相关cookie处理
export const AUTH_TOKEN_KEY = 'auth_token';
export const USER_INFO_KEY = 'user_info';

// 保存认证信息
export const saveAuthCookies = (token: string, userInfo: any, customOptions = {}) => {
    setCookieValue(AUTH_TOKEN_KEY, token, customOptions);
    setCookieValue(USER_INFO_KEY, userInfo, customOptions);
};

// 获取认证token
export const getAuthToken = (): string | null => {
    return getCookieValue<string>(AUTH_TOKEN_KEY);
};

// 获取用户信息
export const getUserInfo = <T = any,>(): T | null => {
    return getCookieValue<T>(USER_INFO_KEY);
};

// 清除认证信息
export const clearAuthCookies = (customOptions = {}) => {
    removeCookieValue(AUTH_TOKEN_KEY, customOptions);
    removeCookieValue(USER_INFO_KEY, customOptions);
};

// 检查是否已认证
export const isAuthenticated = (): boolean => {
    return !!getAuthToken();
};

// 检查用户是否已经登录（更完整的检查）
export const isUserLoggedIn = (): boolean => {
    const token = getAuthToken();
    const userInfo = getUserInfo();
    return !!(token && userInfo);
};

