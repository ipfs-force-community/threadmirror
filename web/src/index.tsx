import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';
import App from './App';
import reportWebVitals from './reportWebVitals';
import { Auth0Provider } from '@auth0/auth0-react';
import { CookiesProvider } from 'react-cookie';


const root = ReactDOM.createRoot(
  document.getElementById('root') as HTMLElement
);
root.render(
  <Auth0Provider
    domain={process.env.REACT_APP_AUTH0_DOMAIN || ''}
    clientId={process.env.REACT_APP_AUTH0_CLIENT_ID || ''}
    onRedirectCallback={(appState: any) => {
      // 让Auth0处理完整的回调流程，避免手动干预导致状态验证失败
      const targetUrl = appState?.returnTo || window.location.pathname;
      
      // 清理URL中的Auth0回调参数，避免刷新时出现"invalid state"错误
      const url = new URL(window.location.href);
      const auth0Params = ['code', 'state', 'session_state', 'error', 'error_description'];
      
      let hasAuth0Params = false;
      auth0Params.forEach(param => {
        if (url.searchParams.has(param)) {
          url.searchParams.delete(param);
          hasAuth0Params = true;
        }
      });
      
      // 如果URL中有Auth0参数或需要重定向到不同路径，则更新URL
      if (hasAuth0Params || targetUrl !== window.location.pathname) {
        const cleanUrl = targetUrl + (url.search ? url.search : '');
        window.history.replaceState({}, '', cleanUrl);
      }
    }}
    authorizationParams={{
      redirect_uri: window.location.origin,
      ...(process.env.REACT_APP_AUTH0_AUDIENCE ? { audience: process.env.REACT_APP_AUTH0_AUDIENCE } : null),
    }}
    cacheLocation="localstorage"
    useRefreshTokens={true}
  >
    <CookiesProvider>
      <App />
    </CookiesProvider>
  </Auth0Provider>,
  // <React.StrictMode>
  //   <App />
  // </React.StrictMode>
);

reportWebVitals();
