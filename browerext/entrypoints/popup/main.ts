/// <reference types="chrome"/>
import './style.css';

// ===== 环境变量配置 =====
const CONFIG = {
  auth0: {
    domain: import.meta.env.VITE_AUTH0_DOMAIN || 'your-domain.auth0.com',
    clientId: import.meta.env.VITE_AUTH0_CLIENT_ID || 'your-auth0-client-id',
    audience: import.meta.env.VITE_AUTH0_AUDIENCE,
    redirectUri: import.meta.env.VITE_AUTH0_REDIRECT_URI || `https://${chrome.runtime.id}.chromiumapp.org/`
  },
  api: {
    baseUrl: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080',
    version: import.meta.env.VITE_API_VERSION || 'v1',
    timeout: parseInt(import.meta.env.VITE_API_TIMEOUT || '10000')
  },
  debug: import.meta.env.NODE_ENV === 'development'
}

// 调试日志函数
const debugLog = (...args: any[]) => {
  if (CONFIG.debug) {
    console.log('[ThreadMirror Extension]', ...args)
  }
}

// ===== 类型定义 =====
interface AuthState {
  isAuthenticated: boolean
  user: any
  token: string | null
}

interface ApiResponse<T = any> {
  success: boolean
  data?: T
  error?: string
}

// ===== ThreadMirror Extension 主类 =====
class ThreadMirrorExtension {
  private authState: AuthState = {
    isAuthenticated: false,
    user: null,
    token: null
  }

  constructor() {
    debugLog('Extension initialized with config:', CONFIG)
    this.init()
  }

  private async init() {
    await this.loadAuthState()
    this.renderUI()
    this.setupEventListeners()
  }

  // ===== 认证相关方法 =====
  private async loadAuthState() {
    try {
      const result = await chrome.storage.local.get(['authToken', 'authUser'])
      if (result.authToken && result.authUser) {
        this.authState = {
          isAuthenticated: true,
          token: result.authToken,
          user: result.authUser
        }
        debugLog('Loaded auth state:', this.authState)
      }
    } catch (error: any) {
      debugLog('Failed to load auth state:', error)
    }
  }

  private async saveAuthState(token: string, user: any) {
    try {
      await chrome.storage.local.set({
        authToken: token,
        authUser: user
      })
      this.authState = {
        isAuthenticated: true,
        token,
        user
      }
      debugLog('Saved auth state:', this.authState)
    } catch (error: any) {
      debugLog('Failed to save auth state:', error)
      throw error
    }
  }

  private async clearAuthState() {
    try {
      await chrome.storage.local.remove(['authToken', 'authUser'])
      this.authState = {
        isAuthenticated: false,
        token: null,
        user: null
      }
      debugLog('Cleared auth state')
    } catch (error: any) {
      debugLog('Failed to clear auth state:', error)
    }
  }

  // ===== 配置验证 =====
  private validateConfiguration(): string[] {
    const errors: string[] = []
    
    // 检查 Auth0 域名
    if (!CONFIG.auth0.domain || CONFIG.auth0.domain === 'your-domain.auth0.com') {
      errors.push('❌ Auth0 Domain not configured. Please set VITE_AUTH0_DOMAIN in .env.local')
    } else if (!CONFIG.auth0.domain.includes('.auth0.com') && !CONFIG.auth0.domain.includes('.')) {
      errors.push('❌ Invalid Auth0 Domain format. Should be like: your-domain.auth0.com')
    }
    
    // 检查 Auth0 Client ID
    if (!CONFIG.auth0.clientId || CONFIG.auth0.clientId === 'your-auth0-client-id') {
      errors.push('❌ Auth0 Client ID not configured. Please set VITE_AUTH0_CLIENT_ID in .env.local')
    } else if (CONFIG.auth0.clientId.length < 20) {
      errors.push('❌ Invalid Auth0 Client ID format. Should be a long string from Auth0')
    }
    
    // 检查 Redirect URI
    if (!CONFIG.auth0.redirectUri.includes('chromiumapp.org')) {
      errors.push('❌ Invalid redirect URI. Should use chromiumapp.org format')
    }
    
    if (errors.length > 0) {
      errors.unshift('🔧 Configuration Help:')
      errors.push('')
      errors.push('📋 Steps to fix:')
      errors.push('1. Copy env.example to .env.local')
      errors.push('2. Get Auth0 credentials from https://manage.auth0.com/')
      errors.push('3. Update .env.local with your values')
      errors.push('4. Rebuild the extension: npm run build')
    }
    
    return errors
  }

  // ===== Auth0 OAuth 流程 =====
  async login() {
    try {
      this.updateUIState('loading', 'Connecting to Auth0...')
      
      // 详细的配置验证
      const configErrors = this.validateConfiguration()
      if (configErrors.length > 0) {
        throw new Error(`Configuration errors:\n${configErrors.join('\n')}`)
      }

      // 构建 Auth0 授权 URL
      const params = new URLSearchParams({
        response_type: 'code',
        client_id: CONFIG.auth0.clientId,
        redirect_uri: CONFIG.auth0.redirectUri,
        scope: 'openid profile email',
        state: crypto.randomUUID(),
        ...(CONFIG.auth0.audience && { audience: CONFIG.auth0.audience })
      })

      const authUrl = `https://${CONFIG.auth0.domain}/authorize?${params.toString()}`
      debugLog('Starting OAuth flow with URL:', authUrl)
      debugLog('Auth0 Configuration:', {
        domain: CONFIG.auth0.domain,
        clientId: CONFIG.auth0.clientId.substring(0, 8) + '...',
        redirectUri: CONFIG.auth0.redirectUri,
        audience: CONFIG.auth0.audience
      })

      // 启动 OAuth 流程
      let responseUrl: string | undefined
      try {
        responseUrl = await chrome.identity.launchWebAuthFlow({
          url: authUrl,
          interactive: true
        })
        debugLog('OAuth response URL:', responseUrl)
      } catch (error: any) {
        debugLog('OAuth flow error:', error)
        
        // 提供具体的错误信息和解决方案
        let errorMessage = 'Authorization page could not be loaded. '
        
        if (error.message?.includes('network') || error.message?.includes('connection')) {
          errorMessage += 'Please check your internet connection.'
        } else if (error.message?.includes('Invalid') || error.message?.includes('404')) {
          errorMessage += 'Auth0 configuration may be incorrect:\n'
          errorMessage += `• Check your domain: ${CONFIG.auth0.domain}\n`
          errorMessage += `• Check your client ID: ${CONFIG.auth0.clientId.substring(0, 8)}...\n`
          errorMessage += '• Verify settings at https://manage.auth0.com/'
        } else {
          errorMessage += 'This usually means:\n'
          errorMessage += '1. Auth0 domain is incorrect\n'
          errorMessage += '2. Client ID is invalid\n'
          errorMessage += '3. Application is not properly configured in Auth0\n'
          errorMessage += '4. Redirect URI mismatch\n\n'
          errorMessage += 'Please check your .env.local file and Auth0 dashboard.'
        }
        
        throw new Error(errorMessage)
      }

      if (!responseUrl) {
        throw new Error('Authorization was cancelled by user')
      }

      // 解析授权码
      const urlParams = new URL(responseUrl).searchParams
      const code = urlParams.get('code')
      const error = urlParams.get('error')

      if (error) {
        throw new Error(`Auth0 error: ${error}`)
      }

      if (!code) {
        throw new Error('No authorization code received')
      }

      debugLog('Received authorization code:', code.substring(0, 10) + '...')

      // 使用授权码换取访问令牌
      await this.exchangeCodeForToken(code)
      
      // 重新渲染UI并设置事件监听器
      this.renderUI()
      this.setupEventListeners()
      this.updateUIState('success', 'Login successful!')

    } catch (error: any) {
      debugLog('Login failed:', error)
      // 重新渲染登录界面
      this.renderUI()
      this.setupEventListeners()
      this.updateUIState('error', `Login failed: ${error.message}`)
    }
  }

  private async exchangeCodeForToken(code: string) {
    try {
      this.updateUIState('loading', 'Exchanging authorization code...')

      const response = await fetch(`${CONFIG.api.baseUrl}/api/${CONFIG.api.version}/auth/oauth/callback`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          code,
          redirect_uri: CONFIG.auth0.redirectUri
        }),
        signal: AbortSignal.timeout(CONFIG.api.timeout)
      })

      if (!response.ok) {
        const errorText = await response.text()
        throw new Error(`HTTP ${response.status}: ${errorText}`)
      }

      const data: ApiResponse<{ token: string; user: any }> = await response.json()
      
      if (!data.success || !data.data) {
        throw new Error(data.error || 'Invalid response from server')
      }

      await this.saveAuthState(data.data.token, data.data.user)
      debugLog('Token exchange successful')

    } catch (error: any) {
      debugLog('Token exchange failed:', error)
      throw error
    }
  }

  async logout() {
    try {
      this.updateUIState('loading', 'Logging out...')

      // 可选：调用后端登出接口
      if (this.authState.token) {
        try {
          await fetch(`${CONFIG.api.baseUrl}/api/${CONFIG.api.version}/auth/logout`, {
            method: 'POST',
            headers: {
              'Authorization': `Bearer ${this.authState.token}`
            },
            signal: AbortSignal.timeout(CONFIG.api.timeout)
          })
        } catch (error: any) {
          debugLog('Backend logout failed, but continuing:', error)
        }
      }

      await this.clearAuthState()
      // 重新渲染UI并设置事件监听器
      this.renderUI()
      this.setupEventListeners()
      this.updateUIState('success', 'Logged out successfully')

    } catch (error: any) {
      debugLog('Logout failed:', error)
      // 重新渲染当前UI
      this.renderUI()
      this.setupEventListeners()
      this.updateUIState('error', `Logout failed: ${error.message}`)
    }
  }

  // ===== API 调用示例 =====
  private async testApiCall() {
    try {
      this.updateUIState('loading', 'Testing API connection...')

      const response = await fetch(`${CONFIG.api.baseUrl}/api/${CONFIG.api.version}/health`, {
        headers: {
          ...(this.authState.token && { 'Authorization': `Bearer ${this.authState.token}` })
        },
        signal: AbortSignal.timeout(CONFIG.api.timeout)
      })

      if (!response.ok) {
        throw new Error(`API test failed: HTTP ${response.status}`)
      }

      const data = await response.json()
      // 恢复UI并显示成功消息
      this.renderUI()
      this.setupEventListeners()
      this.updateUIState('success', `API connection successful!`)
      debugLog('API test successful:', data)

    } catch (error: any) {
      debugLog('API test failed:', error)
      // 恢复UI并显示错误消息
      this.renderUI()
      this.setupEventListeners()
      this.updateUIState('error', `API test failed: ${error.message}`)
    }
  }

  // ===== UI 渲染和事件处理 =====
  private renderUI() {
    const app = document.getElementById('app')!
    
    if (this.authState.isAuthenticated) {
      app.innerHTML = `
        <div class="container">
          <div class="header">
            <h1>ThreadMirror</h1>
            <p>Successfully connected</p>
          </div>
          
          <div id="content">
            <div class="profile-section">
              <div class="user-info">
                ${this.authState.user?.picture 
                  ? `<img src="${this.authState.user.picture}" alt="User Avatar" class="avatar">`
                  : `<div class="avatar-placeholder">${(this.authState.user?.name || 'U')[0].toUpperCase()}</div>`
                }
                <div class="user-details">
                  <h3>${this.authState.user?.name || 'User'}</h3>
                  <p>${this.authState.user?.email || 'No email available'}</p>
                </div>
              </div>
              
              <div class="actions">
                <button id="test-api" class="btn btn-secondary">Test API Connection</button>
                <button id="logout" class="btn btn-secondary">Logout</button>
              </div>
            </div>
          </div>
        </div>
      `
    } else {
      app.innerHTML = `
        <div class="container">
          <div class="header">
            <h1>ThreadMirror</h1>
            <p>Browser Extension</p>
          </div>
          
          <div id="content">
            <div class="login-section">
              <div class="icon">🔐</div>
              <h2>Welcome to ThreadMirror</h2>
              <p>Connect with your Auth0 account to start using ThreadMirror features.</p>
              <button id="login" class="btn btn-primary">Login with Auth0</button>
            </div>
          </div>
        </div>
      `
    }
  }

  private setupEventListeners() {
    // 登录按钮
    const loginBtn = document.getElementById('login')
    if (loginBtn) {
      loginBtn.addEventListener('click', () => this.login())
    }

    // 登出按钮  
    const logoutBtn = document.getElementById('logout')
    if (logoutBtn) {
      logoutBtn.addEventListener('click', () => this.logout())
    }

    // API 测试按钮
    const testApiBtn = document.getElementById('test-api')
    if (testApiBtn) {
      testApiBtn.addEventListener('click', () => this.testApiCall())
    }
  }

  private updateUIState(type: 'loading' | 'success' | 'error', message: string) {
    // 移除之前的消息
    const existingMessage = document.querySelector('.message')
    if (existingMessage) {
      existingMessage.remove()
    }

    // 如果是加载状态，显示在content区域
    if (type === 'loading') {
      const content = document.getElementById('content')
      if (content) {
        content.innerHTML = `
          <div class="loading">
            <div class="spinner"></div>
            <p>${message}</p>
          </div>
        `
      }
      return
    }

    // 创建新的消息元素
    const messageEl = document.createElement('div')
    messageEl.className = `message ${type}`
    messageEl.textContent = message
    
    // 添加到容器
    const container = document.querySelector('.container')
    if (container) {
      container.appendChild(messageEl)
      
      // 5秒后自动移除（成功和错误消息）
      setTimeout(() => {
        if (messageEl.parentNode) {
          messageEl.remove()
        }
      }, 5000)
    }
  }
}

// ===== 初始化扩展 =====
document.addEventListener('DOMContentLoaded', () => {
  debugLog('DOM loaded, initializing extension...')
  new ThreadMirrorExtension()
}) 