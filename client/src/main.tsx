import './index.css'
import Auth from './auth.tsx'
import Home from './home.tsx'
import Login from './login.tsx'
import NotFoundPage from './not-found-page.tsx'
import Profile from './profile.tsx'
import Register from './register.tsx'
import Root from './root.tsx'
import UnauthorizedPage from './unauthorized-page.tsx'
import Users from './users.tsx'

import { BrowserRouter, Routes, Route } from 'react-router'
import { StrictMode, useEffect } from 'react'
import { createRoot } from 'react-dom/client'
import { hideSplashScreen } from 'vite-plugin-splash-screen/runtime'

function App() {
  useEffect(() => {
    hideSplashScreen();
  });

  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Root />}>
          <Route index element={<Home />} />
          <Route path="/profile" element={<Profile />} />
          <Route path="/users" element={<Users />} />
        </Route>
        <Route path="auth" element={<Auth />}>
          <Route path="login" element={<Login />} />
          <Route path="register" element={<Register />} />
        </Route>
        <Route path="/unauthorized" element={<UnauthorizedPage />} />
        <Route path="*" element={<NotFoundPage />} />
      </Routes>
    </BrowserRouter>
  );
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>
);
