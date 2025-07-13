import { BrowserRouter as Router, Routes, Route, useLocation } from 'react-router-dom';
import { Link } from 'react-router-dom';
import { Toaster } from 'sonner';
import { useAuth0 } from '@auth0/auth0-react';
import UserMentions from '@pages/UserMentions';
import MentionDetail from '@pages/MentionDetail';
import TwitterScraperPage from '@pages/TwitterScraper';
import UserLgoinOut from '@components/UserLoginOut';
import './App.css';
import { AuthProvider, useAuthContext } from './AuthContext';

// Floating Action Button Component
const FloatingActionButton = () => {
  const { isLoggedIn } = useAuthContext();
  const location = useLocation();
  
  // 不在 /scrape 页面显示
  if (location.pathname === '/scrape') {
    return null;
  }
  // 不在 /thread/ 详情页面显示
  if (location.pathname.startsWith('/thread/')) {
    return null;
  }
  // 未登录不显示按钮
  if (!isLoggedIn) {
    return null;
  }

  return (
    <Link to="/scrape" className="fab">
      <span className="fab-icon">🧵</span>
      <span className="fab-text">Scrape Thread</span>
    </Link>
  );
};

function App() {
  return (
    <AuthProvider>
      <Router
        future={{
          v7_startTransition: true,
          v7_relativeSplatPath: true,
        }}
      >
        <div className="App">
          <Toaster
            position="top-center"
            richColors={false}
            closeButton
            theme="light"
            toastOptions={{
              style: {
                background: '#fffceb',
                color: '#7c6c3b',
                border: '1px solid #f9e9a2',
              }
            }}
          />
          <header className="App-header">
            <Link to="/" className="project-title-link">
              <h1 className="project-title">Thread Mirror</h1>
            </Link>
            <UserLgoinOut />
          </header>
          <Routes>
            <Route path="/" element={<UserMentions />} />
            <Route path="/scrape" element={<TwitterScraperPage />} />
            <Route path="/mentions/:id" element={<MentionDetail />} />
            <Route path="/thread/:id" element={<MentionDetail />} />
          </Routes>
          <FloatingActionButton />
        </div>
      </Router>
    </AuthProvider>
  );
}

export default App;
