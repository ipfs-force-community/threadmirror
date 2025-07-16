import { defineConfig } from 'wxt';

// See https://wxt.dev/api/config.html
export default defineConfig({
  manifestVersion: 3,
  manifest: {
    name: 'ThreadMirror Extension',
    description: 'Browser extension for ThreadMirror backend API integration with Auth0 authentication',
    version: '1.2.0',
    permissions: [
      'activeTab', 
      'storage',
      'notifications',
      'identity'  // Required for OAuth flows
    ],
    host_permissions: [
      'http://localhost:8080/*',
      'https://*.auth0.com/*'  // Auth0 domain permissions
    ],
    action: {
      default_popup: 'popup.html',
      default_title: 'ThreadMirror Extension'
    }
  },
  
  // Vite configuration for environment variables
  vite: () => ({
    define: {
      // 确保环境变量在构建时被注入
      __DEV__: JSON.stringify(process.env.NODE_ENV === 'development'),
    },
    envPrefix: ['VITE_'], // 只有以 VITE_ 开头的环境变量会被暴露到客户端
  }),

  // Build configuration
  outDir: '.output'
});
