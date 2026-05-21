import './index.css'
import { Layout } from './components/Layout'
import { SessionProvider } from './components/SessionProvider'

function App() {
  return (
    <SessionProvider>
      <Layout />
    </SessionProvider>
  )
}

export default App
