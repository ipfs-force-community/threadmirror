import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { Link } from 'react-router-dom';
import { Toaster } from 'sonner';
import UserMentions from '@pages/UserMentions';
import MentionDetail from '@pages/MentionDetail';
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
            <h1 className="project-title">Thread Mirror</h1>
          </Link>
          <UserLgoinOut />
        </header>
        <Routes>
          <Route path="/" element={<UserMentions />} />
          <Route path="/mentions/:id" element={<MentionDetail />} />
        </Routes>
      </div>
    </Router>
  );
}

export default App;