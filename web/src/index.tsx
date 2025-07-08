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
      if (targetUrl !== window.location.pathname) {
        window.history.replaceState({}, '', targetUrl);
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
