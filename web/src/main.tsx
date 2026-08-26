import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { NomenApp } from './shell/NomenApp'
import './styles.css'

const root = document.getElementById('root')

if (!root) {
  throw new Error('Nomen UI root is missing')
}

createRoot(root).render(
  <StrictMode>
    <NomenApp />
  </StrictMode>,
)
