import './index.css'
import Home from './home.tsx'
import Login from './login.tsx'
import NotFoundPage from './not-found-page.tsx'

import { BrowserRouter, Routes, Route } from 'react-router'
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="auth">
          <Route path="login" element={<Login />} />
        </Route>
        <Route path="*" element={<NotFoundPage />} />
      </Routes>
    </BrowserRouter>
  </StrictMode>,
)
