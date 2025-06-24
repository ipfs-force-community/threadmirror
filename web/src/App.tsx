import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import User from '@pages/User';
import ThreadDetail from '@pages/ThreadDetail';
import UserLgoinOut from '@components/UserLoginOut';
import './App.css';


function App() {
  return (
    <Router>
      <div className="App">
        <header className="App-header">
          <UserLgoinOut />
        </header>
        <Routes>
          <Route path="/" element={<User />} />
          <Route path="/user/:name" element={<User />} />
          <Route path="/thread/:id" element={<ThreadDetail />} />
        </Routes>
      </div>
    </Router>
  );
}

export default App;