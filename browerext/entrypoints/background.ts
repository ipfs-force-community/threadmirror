import { defineBackground } from 'wxt/utils/define-background';

// Auth0 Background Script for ThreadMirror Extension

interface AuthTokenInfo {
  token: string;
  user: any;
  expiresAt: number;
}

interface AuthMessage {
  type: 'CHECK_AUTH' | 'LOGOUT' | 'API_CALL';
}

interface AuthResponse {
  success: boolean;
  isLoggedIn?: boolean;
  user?: any;
  message?: string;
  needLogin?: boolean;
}

// 图标状态管理
const ICONS = {
  default: {
    '16': 'icon/icon-16.png',
    '32': 'icon/icon-32.png',
    '48': 'icon/icon-48.png',
    '96': 'icon/icon-96.png',
    '128': 'icon/icon-128.png'
  },
  loading: {
    '16': 'icon/loading-16.png',
    '32': 'icon/loading-32.png', 
    '48': 'icon/loading-48.png',
    '96': 'icon/loading-96.png',
    '128': 'icon/loading-128.png'
  },
  success: {
    '16': 'icon/success-16.png',
    '32': 'icon/success-32.png',
    '48': 'icon/success-48.png', 
    '96': 'icon/success-96.png',
    '128': 'icon/success-128.png'
  }
};

export default defineBackground(() => {
  console.log('ThreadMirror Extension - Auth0 version started', { id: browser.runtime.id });

  class AuthManager {
    private backendUrl: string = 'http://localhost:8080';

    constructor() {
      this.init();
    }

    private init() {
      // 监听来自popup的消息
      browser.runtime.onMessage.addListener((message: AuthMessage, sender, sendResponse) => {
        this.handleMessage(message, sender, sendResponse);
        return true; // 异步响应
      });

      // 监听扩展安装/启动事件
      browser.runtime.onInstalled.addListener(() => {
        console.log('ThreadMirror Extension installed');
        this.setDefaultIcon();
      });

      browser.runtime.onStartup.addListener(() => {
        console.log('ThreadMirror Extension started');
        this.setDefaultIcon();
      });

      // 初始化时设置默认图标
      this.setDefaultIcon();
    }

    private async handleMessage(message: AuthMessage, sender: any, sendResponse: (response: AuthResponse) => void) {
      try {
        switch (message.type) {
          case 'CHECK_AUTH':
            await this.handleCheckAuth(sendResponse);
            break;
          case 'LOGOUT':
            await this.handleLogout(sendResponse);
            break;
          case 'API_CALL':
            await this.handleApiCall(sendResponse);
            break;
          default:
            sendResponse({ success: false, message: '未知消息类型' });
        }
      } catch (error) {
        console.error('处理消息失败:', error);
        sendResponse({ success: false, message: '处理请求时发生错误' });
      }
    }

    private async handleCheckAuth(sendResponse: (response: AuthResponse) => void) {
      try {
        const result = await browser.storage.local.get(['auth_token', 'user_info', 'token_expires']);
        const { auth_token, user_info, token_expires } = result;

        if (auth_token && user_info && token_expires) {
          const now = Date.now();
          if (now < parseInt(token_expires)) {
            sendResponse({
              success: true,
              isLoggedIn: true,
              user: user_info
            });
            return;
          } else {
            // Token过期，清除存储
            await browser.storage.local.clear();
          }
        }

        sendResponse({
          success: true,
          isLoggedIn: false
        });
      } catch (error) {
        console.error('检查认证状态失败:', error);
        sendResponse({
          success: false,
          message: '检查认证状态失败'
        });
      }
    }

    private async handleLogout(sendResponse: (response: AuthResponse) => void) {
      try {
        // 获取当前token，用于通知后端
        const result = await browser.storage.local.get(['auth_token']);
        const token = result.auth_token;

        // 清除本地存储
        await browser.storage.local.clear();

        // 通知后端登出（可选）
        if (token) {
          try {
            await fetch(`${this.backendUrl}/api/v1/auth/logout`, {
              method: 'POST',
              headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
              }
            });
          } catch (error) {
            console.log('后端登出请求失败，但本地已清除认证信息');
          }
        }

        this.setDefaultIcon();
        sendResponse({
          success: true,
          message: '已成功登出'
        });
      } catch (error) {
        console.error('登出失败:', error);
        sendResponse({
          success: false,
          message: '登出失败'
        });
      }
    }

    private async handleApiCall(sendResponse: (response: AuthResponse) => void) {
      try {
        this.setLoadingIcon();

        // 检查认证状态
        const result = await browser.storage.local.get(['auth_token', 'token_expires']);
        const { auth_token, token_expires } = result;

        if (!auth_token || !token_expires) {
          this.setDefaultIcon();
          sendResponse({
            success: false,
            message: '请先登录',
            needLogin: true
          });
          return;
        }

        // 检查token是否过期
        const now = Date.now();
        if (now >= parseInt(token_expires)) {
          await browser.storage.local.clear();
          this.setDefaultIcon();
          sendResponse({
            success: false,
            message: '登录已过期，请重新登录',
            needLogin: true
          });
          return;
        }

        // 调用后端API
        const response = await fetch(`${this.backendUrl}/api/v1/health`, {
          method: 'GET',
          headers: {
            'Authorization': `Bearer ${auth_token}`,
            'Content-Type': 'application/json'
          }
        });

        if (response.status === 401) {
          // Token无效，清除认证状态
          await browser.storage.local.clear();
          this.setDefaultIcon();
          sendResponse({
            success: false,
            message: '认证失效，请重新登录',
            needLogin: true
          });
          return;
        }

        if (!response.ok) {
          throw new Error(`API调用失败: ${response.status} ${response.statusText}`);
        }

        const data = await response.json();
        
        // 显示成功状态
        this.setSuccessIcon();
        
        // 显示成功通知
        browser.notifications.create({
          type: 'basic',
          iconUrl: 'icon/success-48.png',
          title: 'ThreadMirror',
          message: 'API调用成功！'
        });

        sendResponse({
          success: true,
          message: `API调用成功: ${JSON.stringify(data)}`
        });

        // 2秒后恢复默认图标
        setTimeout(() => {
          this.setDefaultIcon();
        }, 2000);

      } catch (error) {
        console.error('API调用失败:', error);
        this.setDefaultIcon();
        
        sendResponse({
          success: false,
          message: error instanceof Error ? error.message : 'API调用失败'
        });
      }
    }

    private setIcon(iconPaths: Record<string, string>) {
      browser.action.setIcon({ path: iconPaths }).catch(console.error);
    }

    private setDefaultIcon() {
      this.setIcon(ICONS.default);
    }

    private setLoadingIcon() {
      this.setIcon(ICONS.loading);
    }

    private setSuccessIcon() {
      this.setIcon(ICONS.success);
    }
  }

  // 初始化认证管理器
  new AuthManager();
});
