import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import './index.css'
import App from './App.tsx'
import { ScreenProvider } from './config/ScreenContext.tsx'

document.documentElement.setAttribute('data-theme', 'light')

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <BrowserRouter>
      <ScreenProvider>
        <App />
      </ScreenProvider>
    </BrowserRouter>
  </StrictMode>,
)
