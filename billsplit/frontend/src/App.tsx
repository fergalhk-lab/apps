// frontend/src/App.tsx
import { BrowserRouter, Routes, Route, Navigate, Outlet } from 'react-router-dom'
import { Toaster } from '@/components/ui/sonner'
import { ThemeProvider } from '@/components/ThemeProvider'
import { TOKEN_KEY } from '@/api'
import Login from '@/pages/Login'
import Register from '@/pages/Register'
import AppShell from '@/components/AppShell'
import GroupDetail from '@/pages/GroupDetail'

function ProtectedRoute() {
  const token = localStorage.getItem(TOKEN_KEY)
  return token ? <Outlet /> : <Navigate to="/login" replace />
}

export default function App() {
  return (
    <ThemeProvider>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />
          <Route element={<ProtectedRoute />}>
            <Route element={<AppShell />}>
              <Route index element={<Navigate to="/groups" replace />} />
              <Route path="/groups" element={<div className="p-8 text-muted-foreground">Select a group from the sidebar.</div>} />
              <Route path="/groups/:groupId" element={<GroupDetail />} />
            </Route>
          </Route>
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </BrowserRouter>
      <Toaster position="bottom-right" />
    </ThemeProvider>
  )
}
