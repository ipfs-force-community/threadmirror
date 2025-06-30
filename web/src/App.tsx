import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { Link } from 'react-router-dom';
import { Toaster } from 'sonner';
import UserPosts from '@pages/UserPosts';
import PostDetail from '@pages/PostDetail';
import UserLgoinOut from '@components/UserLoginOut';
import './App.css';


function App() {
  return (
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
            <h1 className="project-title">Thread Monitor</h1>
          </Link>
          <UserLgoinOut />
        </header>
        <Routes>
          <Route path="/" element={<UserPosts />} />
          <Route path="/posts/:id" element={<PostDetail />} />
        </Routes>
      </div>
    </Router>
  );
}

export default App;