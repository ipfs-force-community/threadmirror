import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';
import App from './App';
import reportWebVitals from './reportWebVitals';
import { Auth0Provider } from '@auth0/auth0-react';
import history from "./utils/history";


const root = ReactDOM.createRoot(
  document.getElementById('root') as HTMLElement
);
root.render(
  <Auth0Provider
    domain={process.env.REACT_APP_AUTH0_DOMAIN || ''}
    clientId={process.env.REACT_APP_AUTH0_CLIENT_ID || ''}
    onRedirectCallback={(appState: any) => {
      history.push(
        appState && appState.returnTo ? appState.returnTo : window.location.pathname
      );
    }}
    authorizationParams={{
      redirect_uri: window.location.origin,
      ...(process.env.REACT_APP_AUTH0_AUDIENCE ? { audience: process.env.REACT_APP_AUTH0_AUDIENCE } : null),
    }}
  >
    <App />
  </Auth0Provider>,
  // <React.StrictMode>
  //   <App />
  // </React.StrictMode>
);

reportWebVitals();
