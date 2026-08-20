import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { TesseraApp } from './shell/TesseraApp'
import './styles.css'

const root = document.getElementById('root')

if (!root) {
  throw new Error('Tessera UI root is missing')
}

createRoot(root).render(
  <StrictMode>
    <TesseraApp />
  </StrictMode>,
)
