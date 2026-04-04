# BillSplit shadcn/ui Frontend Redesign — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the existing plain-HTML React frontend with a fully-typed TypeScript frontend using shadcn/ui, Tailwind CSS, React Router v6, and a persistent sidebar layout.

**Architecture:** Scaffold a fresh Vite + React + TypeScript project in `billsplit/frontend-new/`. The existing `api.js` is rewritten as typed `api.ts`. Once all pages are implemented and verified, `frontend-new/` replaces `frontend/` so the Go embed path (`frontend/dist`) stays unchanged.

**Tech Stack:** Vite 5, React 18, TypeScript, Tailwind CSS v3, shadcn/ui (blue theme), React Router v6, Sonner (toasts), Vitest (unit tests), Lucide React (icons).

**Spec:** `docs/superpowers/specs/2026-04-04-shadcn-frontend-redesign.md`

---

## File Map

| File | Purpose |
|---|---|
| `frontend-new/vite.config.ts` | Vite config: API proxy, `@/` alias, Vitest config |
| `frontend-new/tailwind.config.ts` | Tailwind v3 with shadcn class-dark-mode config |
| `frontend-new/tsconfig.app.json` | TypeScript config with `@/` path alias |
| `frontend-new/src/index.css` | shadcn CSS variables (blue theme, light + dark) |
| `frontend-new/src/main.tsx` | React root mount |
| `frontend-new/src/App.tsx` | Router setup, auth state, ProtectedRoute, Toaster |
| `frontend-new/src/api.ts` | Typed API client + `parseToken` |
| `frontend-new/src/lib/utils.ts` | shadcn `cn()` helper (auto-generated) |
| `frontend-new/src/components/ThemeProvider.tsx` | Dark/light mode context + system preference |
| `frontend-new/src/components/AppShell.tsx` | Sidebar layout wrapping `<Outlet />` |
| `frontend-new/src/components/CreateGroupDialog.tsx` | Create group dialog (name, currency, member picker) |
| `frontend-new/src/components/AddExpenseModal.tsx` | Add expense dialog |
| `frontend-new/src/components/AddSettlementModal.tsx` | Record settlement dialog |
| `frontend-new/src/pages/Login.tsx` | Login page (no sidebar) |
| `frontend-new/src/pages/Register.tsx` | Register page (no sidebar) |
| `frontend-new/src/pages/GroupDetail.tsx` | Group detail: balances + paginated history |
| `frontend-new/src/api.test.ts` | Unit tests for `parseToken` |
| `frontend-new/src/components/expense.test.ts` | Unit tests for `computeSplits` |

---

## Task 1: Scaffold Vite + React + TypeScript project

**Files:**
- Create: `billsplit/frontend-new/` (entire directory)

- [ ] **Step 1: Scaffold the project**

```bash
cd /home/fergal/code/apps/billsplit
source ~/.nvm/nvm.sh
npm create vite@latest frontend-new -- --template react-ts
cd frontend-new
npm install
```

Expected: `frontend-new/` created with `src/App.tsx`, `src/main.tsx`, `vite.config.ts`, `tsconfig.json`, `tsconfig.app.json`.

- [ ] **Step 2: Remove default boilerplate and stub App.tsx**

```bash
cd /home/fergal/code/apps/billsplit/frontend-new
rm -f src/App.css public/vite.svg
rm -rf src/assets
```

Replace `src/App.tsx` with a stub (it will be fully replaced in Task 10):

```typescript
// frontend-new/src/App.tsx — stub, replaced in Task 10
export default function App() {
  return <div>Setting up...</div>
}
```

- [ ] **Step 3: Clear src/index.css (shadcn will populate it)**

Replace `src/index.css` with an empty placeholder:

```css
/* populated by shadcn init */
```

- [ ] **Step 4: Verify the project builds**

```bash
cd /home/fergal/code/apps/billsplit/frontend-new
npm run build
```

Expected: build succeeds with the stub App. If TypeScript errors appear about the stub, they are fine to ignore — the stub only needs to compile.

- [ ] **Step 5: Commit**

```bash
cd /home/fergal/code/apps/billsplit
git add frontend-new/
git commit -m "chore: scaffold frontend-new vite+react+ts project"
```

---

## Task 2: Configure Tailwind CSS v3 and path alias

**Files:**
- Modify: `frontend-new/vite.config.ts`
- Modify: `frontend-new/tsconfig.app.json`
- Create: `frontend-new/tailwind.config.ts`
- Create: `frontend-new/postcss.config.js`

- [ ] **Step 1: Install Tailwind v3, PostCSS, Autoprefixer, and Vitest**

```bash
cd /home/fergal/code/apps/billsplit/frontend-new
source ~/.nvm/nvm.sh
npm install -D tailwindcss@3 postcss autoprefixer
npm install -D vitest @types/node
```

- [ ] **Step 2: Initialise Tailwind (generates postcss.config.js)**

```bash
cd /home/fergal/code/apps/billsplit/frontend-new
npx tailwindcss init -p --ts
```

Expected: `tailwind.config.ts` and `postcss.config.js` created.

- [ ] **Step 3: Replace tailwind.config.ts with shadcn-compatible config**

```typescript
// frontend-new/tailwind.config.ts
import type { Config } from "tailwindcss"

const config = {
  darkMode: ["class"],
  content: [
    "./index.html",
    "./src/**/*.{ts,tsx}",
  ],
  theme: {
    container: {
      center: true,
      padding: "2rem",
      screens: { "2xl": "1400px" },
    },
    extend: {
      colors: {
        border: "hsl(var(--border))",
        input: "hsl(var(--input))",
        ring: "hsl(var(--ring))",
        background: "hsl(var(--background))",
        foreground: "hsl(var(--foreground))",
        primary: {
          DEFAULT: "hsl(var(--primary))",
          foreground: "hsl(var(--primary-foreground))",
        },
        secondary: {
          DEFAULT: "hsl(var(--secondary))",
          foreground: "hsl(var(--secondary-foreground))",
        },
        destructive: {
          DEFAULT: "hsl(var(--destructive))",
          foreground: "hsl(var(--destructive-foreground))",
        },
        muted: {
          DEFAULT: "hsl(var(--muted))",
          foreground: "hsl(var(--muted-foreground))",
        },
        accent: {
          DEFAULT: "hsl(var(--accent))",
          foreground: "hsl(var(--accent-foreground))",
        },
        card: {
          DEFAULT: "hsl(var(--card))",
          foreground: "hsl(var(--card-foreground))",
        },
        popover: {
          DEFAULT: "hsl(var(--popover))",
          foreground: "hsl(var(--popover-foreground))",
        },
      },
      borderRadius: {
        lg: "var(--radius)",
        md: "calc(var(--radius) - 2px)",
        sm: "calc(var(--radius) - 4px)",
      },
      keyframes: {
        "accordion-down": {
          from: { height: "0" },
          to: { height: "var(--accordion-content-height)" },
        },
        "accordion-up": {
          from: { height: "var(--accordion-content-height)" },
          to: { height: "0" },
        },
      },
      animation: {
        "accordion-down": "accordion-down 0.2s ease-out",
        "accordion-up": "accordion-up 0.2s ease-out",
      },
    },
  },
  plugins: [require("tailwindcss-animate")],
} satisfies Config

export default config
```

- [ ] **Step 4: Install tailwindcss-animate (required by shadcn)**

```bash
cd /home/fergal/code/apps/billsplit/frontend-new
source ~/.nvm/nvm.sh
npm install -D tailwindcss-animate
```

- [ ] **Step 5: Update vite.config.ts with path alias, API proxy, and Vitest config**

```typescript
// frontend-new/vite.config.ts
import path from "path"
import react from "@vitejs/plugin-react"
import { defineConfig } from "vite"

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  server: {
    proxy: {
      "/api": "http://localhost:8080",
    },
  },
  build: {
    outDir: "dist",
  },
  test: {
    environment: "node",
    globals: true,
  },
})
```

- [ ] **Step 6: Update tsconfig.app.json with path alias**

Replace the `compilerOptions` section to add `baseUrl` and `paths`:

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForClassFields": true,
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "react-jsx",
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true,
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"]
    }
  },
  "include": ["src"]
}
```

- [ ] **Step 7: Commit**

```bash
cd /home/fergal/code/apps/billsplit
git add frontend-new/
git commit -m "chore: configure tailwind v3 and path alias for frontend-new"
```

---

## Task 3: Initialize shadcn/ui and install components

**Files:**
- Create: `frontend-new/components.json`
- Create: `frontend-new/src/lib/utils.ts`
- Modify: `frontend-new/src/index.css` (shadcn CSS variables)
- Create: `frontend-new/src/components/ui/` (shadcn component files)

- [ ] **Step 1: Run shadcn init**

```bash
cd /home/fergal/code/apps/billsplit/frontend-new
source ~/.nvm/nvm.sh
npx shadcn@latest init
```

Answer the prompts:
- Which style would you like to use? → **Default**
- Which color would you like to use as base color? → **Blue**
- Would you like to use CSS variables for theming? → **yes**

Expected: `components.json` created, `src/lib/utils.ts` created, `src/index.css` populated with CSS variables.

- [ ] **Step 2: Add all required shadcn components**

```bash
cd /home/fergal/code/apps/billsplit/frontend-new
source ~/.nvm/nvm.sh
npx shadcn@latest add button input label card dialog select badge separator sonner
```

Expected: files created under `src/components/ui/`: `button.tsx`, `input.tsx`, `label.tsx`, `card.tsx`, `dialog.tsx`, `select.tsx`, `badge.tsx`, `separator.tsx`, `sonner.tsx`. The `sonner` package is also installed as a dependency.

- [ ] **Step 3: Verify src/index.css has CSS variables for both light and dark**

Open `src/index.css` and confirm it contains both `:root { ... }` and `.dark { ... }` blocks with `--primary`, `--background`, `--foreground`, etc. If it doesn't, the `shadcn init` didn't run correctly — re-run it.

- [ ] **Step 4: Commit**

```bash
cd /home/fergal/code/apps/billsplit
git add frontend-new/
git commit -m "chore: init shadcn/ui with blue theme and add ui components"
```

---

## Task 4: Install React Router v6 and Lucide icons

**Files:** No new files — just package installation.

- [ ] **Step 1: Install dependencies**

```bash
cd /home/fergal/code/apps/billsplit/frontend-new
source ~/.nvm/nvm.sh
npm install react-router-dom
npm install lucide-react
```

- [ ] **Step 2: Verify installed**

```bash
cd /home/fergal/code/apps/billsplit/frontend-new
node -e "require('./node_modules/react-router-dom/package.json'); console.log('ok')"
```

Expected: `ok`

- [ ] **Step 3: Commit**

```bash
cd /home/fergal/code/apps/billsplit
git add frontend-new/package.json frontend-new/package-lock.json
git commit -m "chore: install react-router-dom and lucide-react"
```

---

## Task 5: Write api.ts with TypeScript types

**Files:**
- Create: `frontend-new/src/api.ts`
- Create: `frontend-new/src/api.test.ts`

- [ ] **Step 1: Write the failing test for parseToken**

```typescript
// frontend-new/src/api.test.ts
import { describe, it, expect } from 'vitest'
import { parseToken } from './api'

describe('parseToken', () => {
  it('decodes a valid JWT payload', () => {
    // Build a fake JWT: header.payload.signature
    // payload = { username: 'alice', isAdmin: true }
    const payload = btoa(JSON.stringify({ username: 'alice', isAdmin: true }))
    const token = `header.${payload}.sig`
    const result = parseToken(token)
    expect(result.username).toBe('alice')
    expect(result.isAdmin).toBe(true)
  })

  it('returns safe defaults for a malformed token', () => {
    const result = parseToken('not-a-jwt')
    expect(result.username).toBe('')
    expect(result.isAdmin).toBe(false)
  })
})
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd /home/fergal/code/apps/billsplit/frontend-new
source ~/.nvm/nvm.sh
npx vitest run src/api.test.ts
```

Expected: FAIL — `parseToken` is not defined.

- [ ] **Step 3: Write api.ts**

```typescript
// frontend-new/src/api.ts

const BASE = '/api'
export const TOKEN_KEY = 'billsplit_token'

function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token)
}

export function clearToken(): void {
  localStorage.removeItem(TOKEN_KEY)
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  const token = getToken()
  if (token) headers['Authorization'] = `Bearer ${token}`

  const res = await fetch(BASE + path, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  })

  if (res.status === 204) return null as T
  const data = await res.json()
  if (!res.ok) throw Object.assign(new Error(data.error || 'Request failed'), { status: res.status })
  return data as T
}

export type Group = {
  id: string
  name: string
  currency: string
  balances: Record<string, number>
  members: string[]
  netBalance: number
}

export type EventType = 'expense' | 'settlement' | 'reversal'

export type GroupEvent = {
  id: string
  type: EventType
  description?: string
  amount: number
  paidBy?: string
  from?: string
  to?: string
  splits?: Record<string, number>
  createdAt: string
  reversedEventId?: string
}

export type EventsResponse = {
  events: GroupEvent[]
  total: number
}

export type UserSummary = { id: string }

export type TokenPayload = { username: string; isAdmin: boolean }

export const api = {
  register: (username: string, password: string, inviteCode: string) =>
    request<void>('POST', '/auth/register', { username, password, inviteCode }),

  login: (username: string, password: string) =>
    request<{ token: string }>('POST', '/auth/login', { username, password }),

  getGroups: () => request<Group[]>('GET', '/groups'),

  getUsers: () => request<{ users: UserSummary[] }>('GET', '/users'),

  createGroup: (name: string, currency: string, members: string[]) =>
    request<Group>('POST', '/groups', { name, currency, members }),

  getGroup: (id: string) => request<Group>('GET', `/groups/${id}`),

  getExpenses: (groupId: string, limit = 20, offset = 0) =>
    request<EventsResponse>('GET', `/groups/${groupId}/expenses?limit=${limit}&offset=${offset}`),

  addExpense: (groupId: string, payload: {
    description: string
    amount: number
    paidBy: string
    splits: Record<string, number>
  }) => request<void>('POST', `/groups/${groupId}/expenses`, payload),

  deleteExpense: (groupId: string, eventId: string) =>
    request<void>('DELETE', `/groups/${groupId}/expenses/${eventId}`),

  addSettlement: (groupId: string, payload: { from: string; to: string; amount: number }) =>
    request<void>('POST', `/groups/${groupId}/settlements`, payload),

  leaveGroup: (groupId: string, username: string) =>
    request<void>('DELETE', `/groups/${groupId}/members/${username}`),

  generateInvite: (isAdmin: boolean) =>
    request<{ code: string }>('POST', '/admin/invites', { isAdmin }),
}

// NOTE: Only decodes the payload — does NOT verify the signature.
// The server enforces admin authorisation; this is UI display only.
export function parseToken(token: string): TokenPayload {
  try {
    return JSON.parse(atob(token.split('.')[1])) as TokenPayload
  } catch {
    return { username: '', isAdmin: false }
  }
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd /home/fergal/code/apps/billsplit/frontend-new
source ~/.nvm/nvm.sh
npx vitest run src/api.test.ts
```

Expected: PASS — 2 tests passing.

- [ ] **Step 5: Commit**

```bash
cd /home/fergal/code/apps/billsplit
git add frontend-new/src/api.ts frontend-new/src/api.test.ts
git commit -m "feat: add typed api.ts with parseToken tests"
```

---

## Task 6: Write ThemeProvider.tsx

**Files:**
- Create: `frontend-new/src/components/ThemeProvider.tsx`

- [ ] **Step 1: Write ThemeProvider.tsx**

```typescript
// frontend-new/src/components/ThemeProvider.tsx
import { createContext, useContext, useEffect, useState } from 'react'

type Theme = 'light' | 'dark' | 'system'

interface ThemeContextValue {
  theme: Theme
  resolvedTheme: 'light' | 'dark'
  setTheme: (theme: Theme) => void
}

const ThemeContext = createContext<ThemeContextValue | null>(null)
const STORAGE_KEY = 'billsplit_theme'

function getSystemDark(): boolean {
  return window.matchMedia('(prefers-color-scheme: dark)').matches
}

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [theme, setThemeState] = useState<Theme>(
    () => (localStorage.getItem(STORAGE_KEY) as Theme) || 'system'
  )
  const [systemDark, setSystemDark] = useState(getSystemDark)

  // Keep systemDark in sync with OS preference changes
  useEffect(() => {
    const media = window.matchMedia('(prefers-color-scheme: dark)')
    const handler = (e: MediaQueryListEvent) => setSystemDark(e.matches)
    media.addEventListener('change', handler)
    return () => media.removeEventListener('change', handler)
  }, [])

  const resolvedTheme: 'light' | 'dark' =
    theme === 'system' ? (systemDark ? 'dark' : 'light') : theme

  // Apply class to <html> element for Tailwind dark: variants
  useEffect(() => {
    const root = document.documentElement
    root.classList.remove('light', 'dark')
    root.classList.add(resolvedTheme)
  }, [resolvedTheme])

  function setTheme(t: Theme) {
    localStorage.setItem(STORAGE_KEY, t)
    setThemeState(t)
  }

  return (
    <ThemeContext.Provider value={{ theme, resolvedTheme, setTheme }}>
      {children}
    </ThemeContext.Provider>
  )
}

export function useTheme() {
  const ctx = useContext(ThemeContext)
  if (!ctx) throw new Error('useTheme must be used within ThemeProvider')
  return ctx
}
```

- [ ] **Step 2: Commit**

```bash
cd /home/fergal/code/apps/billsplit
git add frontend-new/src/components/ThemeProvider.tsx
git commit -m "feat: add ThemeProvider with system preference + manual toggle"
```

---

## Task 7: Write AppShell.tsx

**Files:**
- Create: `frontend-new/src/components/AppShell.tsx`

AppShell renders the full-height two-column layout. The sidebar fetches and lists groups; the main area renders `<Outlet />` (the current page from React Router).

- [ ] **Step 1: Write AppShell.tsx**

```typescript
// frontend-new/src/components/AppShell.tsx
import { useEffect, useState } from 'react'
import { Link, Outlet, useMatch, useNavigate } from 'react-router-dom'
import { Sun, Moon, Monitor, Plus } from 'lucide-react'
import { toast } from 'sonner'
import { api, clearToken, parseToken, TOKEN_KEY, type Group } from '@/api'
import { useTheme } from '@/components/ThemeProvider'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { Badge } from '@/components/ui/badge'
import CreateGroupDialog from '@/components/CreateGroupDialog'

export default function AppShell() {
  const [groups, setGroups] = useState<Group[]>([])
  const [showCreateGroup, setShowCreateGroup] = useState(false)
  const [showInvitePanel, setShowInvitePanel] = useState(false)
  const [inviteIsAdmin, setInviteIsAdmin] = useState(false)
  const [inviteCode, setInviteCode] = useState('')
  const { theme, setTheme } = useTheme()
  const navigate = useNavigate()
  const match = useMatch('/groups/:groupId')
  const activeGroupId = match?.params.groupId ?? null

  const token = localStorage.getItem(TOKEN_KEY) ?? ''
  const { username, isAdmin } = parseToken(token)

  useEffect(() => {
    loadGroups()
  }, [])

  async function loadGroups() {
    try {
      const list = await api.getGroups()
      setGroups(list ?? [])
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to load groups')
    }
  }

  function handleLogout() {
    clearToken()
    navigate('/login')
  }

  async function handleGenerateInvite() {
    try {
      const res = await api.generateInvite(inviteIsAdmin)
      setInviteCode(res.code)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to generate invite')
    }
  }

  function cycleTheme() {
    if (theme === 'system') setTheme('light')
    else if (theme === 'light') setTheme('dark')
    else setTheme('system')
  }

  const ThemeIcon = theme === 'light' ? Sun : theme === 'dark' ? Moon : Monitor

  return (
    <div className="flex h-screen bg-background text-foreground">
      {/* Sidebar */}
      <aside className="w-56 flex-shrink-0 border-r border-border flex flex-col">
        {/* Logo */}
        <div className="px-4 py-4">
          <span className="font-bold text-lg">BillSplit</span>
        </div>

        <Separator />

        {/* Groups list */}
        <div className="flex-1 overflow-y-auto py-2">
          <p className="px-4 py-1 text-xs font-medium text-muted-foreground uppercase tracking-wider">
            Groups
          </p>
          {groups.map(g => (
            <Link
              key={g.id}
              to={`/groups/${g.id}`}
              className={`flex items-center justify-between px-4 py-2 text-sm rounded-md mx-2 my-0.5 transition-colors ${
                g.id === activeGroupId
                  ? 'bg-primary/10 text-primary font-medium'
                  : 'text-foreground hover:bg-muted'
              }`}
            >
              <span className="truncate">{g.name}</span>
              <span
                className={`text-xs font-medium ml-2 flex-shrink-0 ${
                  g.netBalance >= 0 ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'
                }`}
              >
                {g.netBalance >= 0 ? '+' : ''}{g.netBalance.toFixed(2)}
              </span>
            </Link>
          ))}

          {/* New group button */}
          <button
            onClick={() => setShowCreateGroup(true)}
            className="flex items-center gap-1 px-4 py-2 mx-2 mt-1 text-sm text-muted-foreground hover:text-foreground hover:bg-muted rounded-md w-full transition-colors"
          >
            <Plus className="h-3 w-3" />
            New group
          </button>
        </div>

        <Separator />

        {/* Admin invite panel */}
        {isAdmin && (
          <div className="px-4 py-2">
            <button
              onClick={() => { setShowInvitePanel(v => !v); setInviteCode('') }}
              className="text-xs text-muted-foreground hover:text-foreground transition-colors"
            >
              Generate invite
            </button>
            {showInvitePanel && (
              <div className="mt-2 space-y-2">
                <label className="flex items-center gap-2 text-xs">
                  <input
                    type="checkbox"
                    checked={inviteIsAdmin}
                    onChange={e => setInviteIsAdmin(e.target.checked)}
                  />
                  Admin invite
                </label>
                <Button size="sm" variant="outline" className="w-full text-xs" onClick={handleGenerateInvite}>
                  Generate
                </Button>
                {inviteCode && (
                  <div className="flex gap-1">
                    <input
                      readOnly
                      value={inviteCode}
                      className="flex-1 text-xs border rounded px-2 py-1 bg-muted font-mono"
                    />
                    <Button
                      size="sm"
                      variant="outline"
                      className="text-xs px-2"
                      onClick={() =>
                        navigator.clipboard.writeText(inviteCode)
                          .then(() => toast.success('Copied!'))
                          .catch(() => toast.error('Failed to copy'))
                      }
                    >
                      Copy
                    </Button>
                  </div>
                )}
              </div>
            )}
          </div>
        )}

        {/* Bottom: username + theme toggle */}
        <div className="px-4 py-3 flex items-center justify-between">
          <button
            onClick={handleLogout}
            className="text-sm text-muted-foreground hover:text-foreground transition-colors truncate"
          >
            {username}
          </button>
          <button
            onClick={cycleTheme}
            className="text-muted-foreground hover:text-foreground transition-colors ml-2 flex-shrink-0"
            title={`Theme: ${theme}`}
          >
            <ThemeIcon className="h-4 w-4" />
          </button>
        </div>
      </aside>

      {/* Main content */}
      <main className="flex-1 overflow-y-auto">
        <Outlet context={{ reloadGroups: loadGroups }} />
      </main>

      {showCreateGroup && (
        <CreateGroupDialog
          currentUsername={username}
          onClose={() => setShowCreateGroup(false)}
          onCreated={(groupId) => {
            setShowCreateGroup(false)
            loadGroups()
            navigate(`/groups/${groupId}`)
          }}
        />
      )}
    </div>
  )
}
```

- [ ] **Step 2: Commit**

```bash
cd /home/fergal/code/apps/billsplit
git add frontend-new/src/components/AppShell.tsx
git commit -m "feat: add AppShell sidebar layout component"
```

---

## Task 8: Write Login.tsx

**Files:**
- Create: `frontend-new/src/pages/Login.tsx`

- [ ] **Step 1: Write Login.tsx**

```typescript
// frontend-new/src/pages/Login.tsx
import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { api, setToken } from '@/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

export default function Login() {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const navigate = useNavigate()

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    try {
      const { token } = await api.login(username, password)
      setToken(token)
      navigate('/')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed')
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-background px-4">
      <Card className="w-full max-w-sm">
        <CardHeader>
          <CardTitle className="text-2xl">BillSplit</CardTitle>
          <CardDescription>Sign in to your account</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="username">Username</Label>
              <Input
                id="username"
                placeholder="Username"
                value={username}
                onChange={e => setUsername(e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                type="password"
                placeholder="Password"
                value={password}
                onChange={e => setPassword(e.target.value)}
                required
              />
            </div>
            {error && <p className="text-sm text-destructive">{error}</p>}
            <Button type="submit" className="w-full">Sign in</Button>
          </form>
          <p className="mt-4 text-sm text-center text-muted-foreground">
            No account?{' '}
            <Link to="/register" className="text-primary hover:underline">
              Register
            </Link>
          </p>
        </CardContent>
      </Card>
    </div>
  )
}
```

- [ ] **Step 2: Commit**

```bash
cd /home/fergal/code/apps/billsplit
git add frontend-new/src/pages/Login.tsx
git commit -m "feat: add Login page with shadcn Card/Input/Button"
```

---

## Task 9: Write Register.tsx

**Files:**
- Create: `frontend-new/src/pages/Register.tsx`

- [ ] **Step 1: Write Register.tsx**

```typescript
// frontend-new/src/pages/Register.tsx
import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { api } from '@/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

export default function Register() {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [inviteCode, setInviteCode] = useState('')
  const [error, setError] = useState('')
  const navigate = useNavigate()

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    try {
      await api.register(username, password, inviteCode)
      navigate('/login')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Registration failed')
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-background px-4">
      <Card className="w-full max-w-sm">
        <CardHeader>
          <CardTitle className="text-2xl">BillSplit</CardTitle>
          <CardDescription>Create your account</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="username">Username</Label>
              <Input
                id="username"
                placeholder="Username"
                value={username}
                onChange={e => setUsername(e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                type="password"
                placeholder="Password"
                value={password}
                onChange={e => setPassword(e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="invite">Invite code</Label>
              <Input
                id="invite"
                placeholder="Invite code"
                value={inviteCode}
                onChange={e => setInviteCode(e.target.value)}
                required
              />
            </div>
            {error && <p className="text-sm text-destructive">{error}</p>}
            <Button type="submit" className="w-full">Create account</Button>
          </form>
          <p className="mt-4 text-sm text-center text-muted-foreground">
            Already have an account?{' '}
            <Link to="/login" className="text-primary hover:underline">
              Sign in
            </Link>
          </p>
        </CardContent>
      </Card>
    </div>
  )
}
```

- [ ] **Step 2: Commit**

```bash
cd /home/fergal/code/apps/billsplit
git add frontend-new/src/pages/Register.tsx
git commit -m "feat: add Register page"
```

---

## Task 10: Write App.tsx (routing + auth + Toaster)

**Files:**
- Create: `frontend-new/src/App.tsx`
- Modify: `frontend-new/src/main.tsx`

- [ ] **Step 1: Write App.tsx**

```typescript
// frontend-new/src/App.tsx
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
```

- [ ] **Step 2: Write main.tsx**

```typescript
// frontend-new/src/main.tsx
import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import './index.css'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
)
```

- [ ] **Step 3: Commit**

```bash
cd /home/fergal/code/apps/billsplit
git add frontend-new/src/App.tsx frontend-new/src/main.tsx
git commit -m "feat: wire up routing, ProtectedRoute, ThemeProvider, and Toaster"
```

---

## Task 11: Write GroupDetail.tsx

**Files:**
- Create: `frontend-new/src/pages/GroupDetail.tsx`

- [ ] **Step 1: Write GroupDetail.tsx**

```typescript
// frontend-new/src/pages/GroupDetail.tsx
import { useState, useEffect } from 'react'
import { useParams, useOutletContext } from 'react-router-dom'
import { toast } from 'sonner'
import { api, type Group, type GroupEvent } from '@/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import AddExpenseModal from '@/components/AddExpenseModal'
import AddSettlementModal from '@/components/AddSettlementModal'

type OutletContext = { reloadGroups: () => void }

const LIMIT = 20

export default function GroupDetail() {
  const { groupId } = useParams<{ groupId: string }>()
  const { reloadGroups } = useOutletContext<OutletContext>()
  const [group, setGroup] = useState<Group | null>(null)
  const [events, setEvents] = useState<GroupEvent[]>([])
  const [total, setTotal] = useState(0)
  const [offset, setOffset] = useState(0)
  const [expandedId, setExpandedId] = useState<string | null>(null)
  const [showExpense, setShowExpense] = useState(false)
  const [showSettlement, setShowSettlement] = useState(false)

  useEffect(() => {
    if (groupId) {
      loadGroup()
      loadEvents(0)
    }
  }, [groupId])

  async function loadGroup() {
    try {
      const g = await api.getGroup(groupId!)
      setGroup(g)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to load group')
    }
  }

  async function loadEvents(newOffset: number) {
    try {
      const res = await api.getExpenses(groupId!, LIMIT, newOffset)
      setEvents(res.events ?? [])
      setTotal(res.total ?? 0)
      setOffset(newOffset)
      setExpandedId(null)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to load expenses')
    }
  }

  async function handleDeleteExpense(eventId: string) {
    try {
      await api.deleteExpense(groupId!, eventId)
      loadGroup()
      loadEvents(offset)
      reloadGroups()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to cancel expense')
    }
  }

  function refresh() {
    loadGroup()
    loadEvents(0)
    reloadGroups()
  }

  if (!group) return <div className="p-8 text-muted-foreground">Loading...</div>

  return (
    <div className="p-6 max-w-2xl mx-auto space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-2xl font-bold">{group.name}</h1>
          <Badge variant="secondary" className="mt-1">{group.currency}</Badge>
        </div>
        <div className="flex gap-2">
          <Button onClick={() => setShowExpense(true)}>+ Expense</Button>
          <Button variant="outline" onClick={() => setShowSettlement(true)}>+ Settle</Button>
        </div>
      </div>

      {/* Balances */}
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-sm font-medium text-muted-foreground uppercase tracking-wider">
            Balances
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-2">
          {Object.entries(group.balances ?? {}).map(([user, bal]) => (
            <div key={user} className="flex justify-between items-center">
              <span className="text-sm">{user}</span>
              <span className={`text-sm font-semibold ${bal >= 0 ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'}`}>
                {bal >= 0 ? '+' : ''}{bal.toFixed(2)}
              </span>
            </div>
          ))}
        </CardContent>
      </Card>

      {/* History */}
      <div className="space-y-2">
        <h2 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">History</h2>

        {events.length === 0 && (
          <p className="text-sm text-muted-foreground py-4 text-center">No transactions yet.</p>
        )}

        {events.map(e => (
          <Card key={e.id} className={e.type === 'expense' ? 'cursor-pointer hover:bg-muted/50 transition-colors' : ''}>
            <CardContent className="py-3 px-4">
              {e.type === 'expense' && (
                <div onClick={() => setExpandedId(e.id === expandedId ? null : e.id)}>
                  <div className="flex justify-between items-start">
                    <div>
                      <p className="font-medium text-sm">{e.description}</p>
                      <p className="text-xs text-muted-foreground">
                        Paid by {e.paidBy} · {new Date(e.createdAt).toLocaleDateString()}
                      </p>
                    </div>
                    <div className="text-right ml-4">
                      <p className="font-semibold text-sm">{group.currency} {e.amount.toFixed(2)}</p>
                      <p className="text-xs text-muted-foreground">{expandedId === e.id ? '▲' : '▼'} splits</p>
                    </div>
                  </div>
                  {expandedId === e.id && (
                    <div className="mt-3 pt-3 border-t border-border space-y-1">
                      {Object.entries(e.splits ?? {})
                        .sort(([a], [b]) => a.localeCompare(b))
                        .map(([member, amount]) => (
                          <div key={member} className="flex justify-between text-xs text-muted-foreground max-w-xs">
                            <span>{member}</span>
                            <span>{group.currency} {amount.toFixed(2)}</span>
                          </div>
                        ))}
                      <div className="pt-2">
                        <Button
                          size="sm"
                          variant="destructive"
                          className="text-xs h-7"
                          onClick={ev => { ev.stopPropagation(); handleDeleteExpense(e.id) }}
                        >
                          Cancel expense
                        </Button>
                      </div>
                    </div>
                  )}
                </div>
              )}

              {e.type === 'settlement' && (
                <div className="flex justify-between items-center">
                  <div>
                    <p className="font-medium text-sm">Settlement</p>
                    <p className="text-xs text-muted-foreground">
                      {e.from} → {e.to} · {new Date(e.createdAt).toLocaleDateString()}
                    </p>
                  </div>
                  <p className="font-semibold text-sm">{group.currency} {e.amount.toFixed(2)}</p>
                </div>
              )}

              {e.type === 'reversal' && (
                <p className="text-xs text-muted-foreground italic">
                  Expense cancelled (#{e.reversedEventId?.slice(0, 8)})
                </p>
              )}
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Pagination */}
      {(offset > 0 || offset + LIMIT < total) && (
        <div className="flex gap-2">
          {offset > 0 && (
            <Button variant="outline" size="sm" onClick={() => loadEvents(offset - LIMIT)}>← Prev</Button>
          )}
          {offset + LIMIT < total && (
            <Button variant="outline" size="sm" onClick={() => loadEvents(offset + LIMIT)}>Next →</Button>
          )}
        </div>
      )}

      {showExpense && (
        <AddExpenseModal
          group={group}
          onClose={() => setShowExpense(false)}
          onSaved={() => { setShowExpense(false); refresh() }}
        />
      )}
      {showSettlement && (
        <AddSettlementModal
          group={group}
          onClose={() => setShowSettlement(false)}
          onSaved={() => { setShowSettlement(false); refresh() }}
        />
      )}
    </div>
  )
}
```

- [ ] **Step 2: Commit**

```bash
cd /home/fergal/code/apps/billsplit
git add frontend-new/src/pages/GroupDetail.tsx
git commit -m "feat: add GroupDetail page with balances, history, and pagination"
```

---

## Task 12: Write AddExpenseModal.tsx + test computeSplits

**Files:**
- Create: `frontend-new/src/components/AddExpenseModal.tsx`
- Create: `frontend-new/src/components/expense.test.ts`

The split calculation logic is pure and worth testing. Extract it to a testable function first.

- [ ] **Step 1: Write the failing test for computeSplits**

```typescript
// frontend-new/src/components/expense.test.ts
import { describe, it, expect } from 'vitest'
import { computeSplits } from './AddExpenseModal'

describe('computeSplits', () => {
  const members = ['Alice', 'Bob', 'Carol']

  it('splits equally', () => {
    const result = computeSplits('equal', 90, members, {}, {})
    expect(result).toEqual({ Alice: 30, Bob: 30, Carol: 30 })
  })

  it('splits by ratio', () => {
    const ratios = { Alice: '2', Bob: '1', Carol: '1' }
    const result = computeSplits('ratio', 100, members, ratios, {})
    expect(result?.Alice).toBe(50)
    expect(result?.Bob).toBe(25)
    expect(result?.Carol).toBe(25)
  })

  it('splits by fixed amounts', () => {
    const fixed = { Alice: '50', Bob: '30', Carol: '20' }
    const result = computeSplits('fixed', 100, members, {}, fixed)
    expect(result).toEqual({ Alice: 50, Bob: 30, Carol: 20 })
  })

  it('returns null for zero ratio sum', () => {
    const ratios = { Alice: '0', Bob: '0', Carol: '0' }
    expect(computeSplits('ratio', 100, members, ratios, {})).toBeNull()
  })
})
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd /home/fergal/code/apps/billsplit/frontend-new
source ~/.nvm/nvm.sh
npx vitest run src/components/expense.test.ts
```

Expected: FAIL — `computeSplits` not found.

- [ ] **Step 3: Write AddExpenseModal.tsx (with exported computeSplits)**

```typescript
// frontend-new/src/components/AddExpenseModal.tsx
import { useState } from 'react'
import { api, type Group } from '@/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

type SplitMode = 'equal' | 'ratio' | 'fixed'

// Exported for testing
export function computeSplits(
  mode: SplitMode,
  total: number,
  members: string[],
  ratios: Record<string, string>,
  fixed: Record<string, string>,
): Record<string, number> | null {
  if (isNaN(total) || total <= 0) return null

  if (mode === 'equal') {
    const share = parseFloat((total / members.length).toFixed(2))
    return Object.fromEntries(members.map(m => [m, share]))
  }

  if (mode === 'ratio') {
    const nums = members.map(m => parseFloat(ratios[m]) || 0)
    const sum = nums.reduce((a, b) => a + b, 0)
    if (sum === 0) return null
    return Object.fromEntries(
      members.map((m, i) => [m, parseFloat(((nums[i] / sum) * total).toFixed(2))])
    )
  }

  return Object.fromEntries(members.map(m => [m, parseFloat(fixed[m]) || 0]))
}

interface Props {
  group: Group
  onClose: () => void
  onSaved: () => void
}

export default function AddExpenseModal({ group, onClose, onSaved }: Props) {
  const [description, setDescription] = useState('')
  const [amount, setAmount] = useState('')
  const [paidBy, setPaidBy] = useState(group.members[0] ?? '')
  const [splitMode, setSplitMode] = useState<SplitMode>('equal')
  const [ratios, setRatios] = useState<Record<string, string>>(
    () => Object.fromEntries(group.members.map(m => [m, '1']))
  )
  const [fixed, setFixed] = useState<Record<string, string>>(
    () => Object.fromEntries(group.members.map(m => [m, '0']))
  )
  const [error, setError] = useState('')

  const total = parseFloat(amount)
  const splits = computeSplits(splitMode, total, group.members, ratios, fixed)
  const splitsTotal = splits ? Object.values(splits).reduce((a, b) => a + b, 0) : 0
  const splitsMismatch = amount && splits && Math.abs(splitsTotal - total) > 0.01

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    if (!splits) { setError('Invalid split configuration'); return }
    if (splitsMismatch) {
      setError(`Splits sum to ${splitsTotal.toFixed(2)} but total is ${total.toFixed(2)}`)
      return
    }
    try {
      await api.addExpense(group.id, { description, amount: total, paidBy, splits })
      onSaved()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to add expense')
    }
  }

  return (
    <Dialog open onOpenChange={open => !open && onClose()}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Add expense</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label>Description</Label>
            <Input
              placeholder="Description"
              value={description}
              onChange={e => setDescription(e.target.value)}
              required
            />
          </div>
          <div className="space-y-2">
            <Label>Amount ({group.currency})</Label>
            <Input
              type="number"
              step="0.01"
              min="0.01"
              placeholder="0.00"
              value={amount}
              onChange={e => setAmount(e.target.value)}
              required
            />
          </div>
          <div className="space-y-2">
            <Label>Paid by</Label>
            <Select value={paidBy} onValueChange={setPaidBy}>
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                {group.members.map(m => <SelectItem key={m} value={m}>{m}</SelectItem>)}
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label>Split</Label>
            <Select value={splitMode} onValueChange={v => setSplitMode(v as SplitMode)}>
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="equal">Equal split</SelectItem>
                <SelectItem value="ratio">By ratio</SelectItem>
                <SelectItem value="fixed">Fixed amounts</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {splitMode === 'ratio' && (
            <div className="space-y-2">
              {group.members.map(m => (
                <div key={m} className="flex items-center gap-3">
                  <Label className="w-24 truncate">{m}</Label>
                  <Input
                    type="number"
                    step="0.1"
                    min="0"
                    className="w-24"
                    value={ratios[m]}
                    onChange={e => setRatios({ ...ratios, [m]: e.target.value })}
                  />
                </div>
              ))}
            </div>
          )}

          {splitMode === 'fixed' && (
            <div className="space-y-2">
              {group.members.map(m => (
                <div key={m} className="flex items-center gap-3">
                  <Label className="w-24 truncate">{m}</Label>
                  <Input
                    type="number"
                    step="0.01"
                    min="0"
                    className="w-24"
                    value={fixed[m]}
                    onChange={e => setFixed({ ...fixed, [m]: e.target.value })}
                  />
                </div>
              ))}
              {amount && (
                <p className={`text-xs ${splitsMismatch ? 'text-destructive' : 'text-green-600 dark:text-green-400'}`}>
                  Splits total: {splitsTotal.toFixed(2)} / {total.toFixed(2)}
                </p>
              )}
            </div>
          )}

          {error && <p className="text-sm text-destructive">{error}</p>}
          <div className="flex gap-2 justify-end">
            <Button type="button" variant="outline" onClick={onClose}>Cancel</Button>
            <Button type="submit">Add</Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd /home/fergal/code/apps/billsplit/frontend-new
source ~/.nvm/nvm.sh
npx vitest run src/components/expense.test.ts
```

Expected: PASS — 4 tests passing.

- [ ] **Step 5: Commit**

```bash
cd /home/fergal/code/apps/billsplit
git add frontend-new/src/components/AddExpenseModal.tsx frontend-new/src/components/expense.test.ts
git commit -m "feat: add AddExpenseModal with tested computeSplits logic"
```

---

## Task 13: Write AddSettlementModal.tsx

**Files:**
- Create: `frontend-new/src/components/AddSettlementModal.tsx`

- [ ] **Step 1: Write AddSettlementModal.tsx**

```typescript
// frontend-new/src/components/AddSettlementModal.tsx
import { useState } from 'react'
import { api, type Group } from '@/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

interface Props {
  group: Group
  onClose: () => void
  onSaved: () => void
}

export default function AddSettlementModal({ group, onClose, onSaved }: Props) {
  const [from, setFrom] = useState(group.members[0] ?? '')
  const [to, setTo] = useState(group.members[1] ?? group.members[0] ?? '')
  const [amount, setAmount] = useState('')
  const [error, setError] = useState('')

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    try {
      await api.addSettlement(group.id, { from, to, amount: parseFloat(amount) })
      onSaved()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to record settlement')
    }
  }

  return (
    <Dialog open onOpenChange={open => !open && onClose()}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Record settlement</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label>From</Label>
            <Select value={from} onValueChange={setFrom}>
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                {group.members.map(m => <SelectItem key={m} value={m}>{m}</SelectItem>)}
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label>To</Label>
            <Select value={to} onValueChange={setTo}>
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                {group.members.map(m => <SelectItem key={m} value={m}>{m}</SelectItem>)}
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label>Amount ({group.currency})</Label>
            <Input
              type="number"
              step="0.01"
              min="0.01"
              placeholder="0.00"
              value={amount}
              onChange={e => setAmount(e.target.value)}
              required
            />
          </div>
          {error && <p className="text-sm text-destructive">{error}</p>}
          <div className="flex gap-2 justify-end">
            <Button type="button" variant="outline" onClick={onClose}>Cancel</Button>
            <Button type="submit">Record</Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}
```

- [ ] **Step 2: Commit**

```bash
cd /home/fergal/code/apps/billsplit
git add frontend-new/src/components/AddSettlementModal.tsx
git commit -m "feat: add AddSettlementModal"
```

---

## Task 14: Write CreateGroupDialog.tsx

**Files:**
- Create: `frontend-new/src/components/CreateGroupDialog.tsx`

- [ ] **Step 1: Write CreateGroupDialog.tsx**

```typescript
// frontend-new/src/components/CreateGroupDialog.tsx
import { useState, useEffect } from 'react'
import { api, type UserSummary } from '@/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { X } from 'lucide-react'

interface Props {
  currentUsername: string
  onClose: () => void
  onCreated: (groupId: string) => void
}

export default function CreateGroupDialog({ currentUsername, onClose, onCreated }: Props) {
  const [name, setName] = useState('')
  const [currency, setCurrency] = useState('GBP')
  const [availableUsers, setAvailableUsers] = useState<UserSummary[]>([])
  const [selectedMembers, setSelectedMembers] = useState<UserSummary[]>([])
  const [error, setError] = useState('')

  useEffect(() => {
    api.getUsers()
      .then(res => setAvailableUsers((res.users ?? []).filter(u => u.id !== currentUsername)))
      .catch(err => setError(err instanceof Error ? err.message : 'Failed to load users'))
  }, [currentUsername])

  function toggleMember(user: UserSummary) {
    setSelectedMembers(prev =>
      prev.some(m => m.id === user.id)
        ? prev.filter(m => m.id !== user.id)
        : [...prev, user]
    )
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    if (selectedMembers.length === 0) {
      setError('Add at least one other member')
      return
    }
    try {
      const group = await api.createGroup(name, currency, selectedMembers.map(u => u.id))
      onCreated(group.id)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create group')
    }
  }

  return (
    <Dialog open onOpenChange={open => !open && onClose()}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Create group</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label>Group name</Label>
            <Input
              placeholder="Weekend trip"
              value={name}
              onChange={e => setName(e.target.value)}
              required
            />
          </div>
          <div className="space-y-2">
            <Label>Currency</Label>
            <Input
              placeholder="GBP"
              value={currency}
              onChange={e => setCurrency(e.target.value)}
              required
            />
          </div>
          <div className="space-y-2">
            <Label>Members</Label>
            {/* Selected members chips */}
            <div className="min-h-10 border rounded-md p-2 flex flex-wrap gap-1">
              {selectedMembers.length === 0 && (
                <span className="text-xs text-muted-foreground italic self-center">
                  Click below to add members…
                </span>
              )}
              {selectedMembers.map(u => (
                <Badge key={u.id} variant="secondary" className="flex items-center gap-1">
                  {u.id}
                  <button
                    type="button"
                    onClick={() => toggleMember(u)}
                    className="ml-1 hover:text-destructive"
                  >
                    <X className="h-3 w-3" />
                  </button>
                </Badge>
              ))}
            </div>
            {/* Available users list */}
            <div className="max-h-36 overflow-y-auto border rounded-md divide-y">
              {availableUsers.map(u => {
                const selected = selectedMembers.some(m => m.id === u.id)
                return (
                  <button
                    key={u.id}
                    type="button"
                    onClick={() => toggleMember(u)}
                    className={`w-full flex items-center gap-2 px-3 py-2 text-sm text-left transition-colors ${
                      selected ? 'bg-primary/10 text-primary' : 'hover:bg-muted'
                    }`}
                  >
                    <span className="w-4 text-center">{selected ? '✓' : ''}</span>
                    {u.id}
                  </button>
                )
              })}
              {availableUsers.length === 0 && !error && (
                <p className="text-xs text-muted-foreground p-3">No other users found.</p>
              )}
            </div>
          </div>
          {error && <p className="text-sm text-destructive">{error}</p>}
          <div className="flex gap-2 justify-end">
            <Button type="button" variant="outline" onClick={onClose}>Cancel</Button>
            <Button type="submit">Create</Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}
```

- [ ] **Step 2: Commit**

```bash
cd /home/fergal/code/apps/billsplit
git add frontend-new/src/components/CreateGroupDialog.tsx
git commit -m "feat: add CreateGroupDialog with member picker"
```

---

## Task 15: Run all tests and verify dev server

**Files:** No new files.

- [ ] **Step 1: Run all unit tests**

```bash
cd /home/fergal/code/apps/billsplit/frontend-new
source ~/.nvm/nvm.sh
npx vitest run
```

Expected: all tests pass (api.test.ts + expense.test.ts = 6 tests total).

- [ ] **Step 2: Build the project to check for TypeScript errors**

```bash
cd /home/fergal/code/apps/billsplit/frontend-new
source ~/.nvm/nvm.sh
npm run build
```

Expected: build succeeds with no TypeScript errors. If there are errors, fix them before continuing. Common issues:
- Unused imports (TypeScript strict mode)
- Missing `React` import (not needed with React 17+ JSX transform, but check `tsconfig.app.json` has `"jsx": "react-jsx"`)
- Type mismatches in component props

- [ ] **Step 3: Start the Go backend (in a separate terminal, needed for the proxy)**

This step requires the Go backend to be running. In a separate terminal:
```bash
cd /home/fergal/code/apps/billsplit
# Ensure docker-compose services (minio) are running if needed:
# docker-compose up -d
go run . 
```

The backend listens on `:8080`.

- [ ] **Step 4: Start the Vite dev server**

```bash
cd /home/fergal/code/apps/billsplit/frontend-new
source ~/.nvm/nvm.sh
npm run dev
```

Expected: dev server starts at `http://localhost:5173` (or similar port).

- [ ] **Step 5: Manual smoke test**

Open `http://localhost:5173` in a browser and verify:
1. Redirects to `/login` when no token exists
2. Login form renders with shadcn Card/Input/Button styling
3. Login succeeds and redirects to `/groups`
4. Sidebar appears with group list
5. Clicking a group navigates to `/groups/:id` and loads GroupDetail
6. Balances show in green/red
7. "+ Expense" button opens AddExpenseModal dialog
8. "+ Settle" button opens AddSettlementModal dialog
9. "+ New group" opens CreateGroupDialog
10. Theme toggle cycles through light/dark/system
11. Sign out (click username) clears token and redirects to login

Fix any issues found before proceeding to the swap.

---

## Task 16: Swap frontend-new → frontend

**Files:**
- Delete: `billsplit/frontend/` (archived first)
- Rename: `billsplit/frontend-new/` → `billsplit/frontend/`

Do this step only after Task 15 smoke test passes.

- [ ] **Step 1: Build the new frontend**

```bash
cd /home/fergal/code/apps/billsplit/frontend-new
source ~/.nvm/nvm.sh
npm run build
```

Expected: `frontend-new/dist/` populated with `index.html` and assets.

- [ ] **Step 2: Archive the old frontend**

```bash
cd /home/fergal/code/apps/billsplit
mv frontend frontend-old
```

- [ ] **Step 3: Rename frontend-new to frontend**

```bash
cd /home/fergal/code/apps/billsplit
mv frontend-new frontend
```

- [ ] **Step 4: Verify the Go embed path still works**

```bash
cd /home/fergal/code/apps/billsplit
go build ./...
```

Expected: compiles without error. The `//go:embed frontend/dist` directive in `main.go` now picks up the new frontend's build output.

- [ ] **Step 5: Delete the old frontend**

```bash
cd /home/fergal/code/apps/billsplit
rm -rf frontend-old
```

- [ ] **Step 6: Commit**

```bash
cd /home/fergal/code/apps/billsplit
git add -A
git commit -m "feat: replace frontend with shadcn/ui TypeScript redesign"
```

- [ ] **Step 7: Final smoke test with embedded build**

```bash
cd /home/fergal/code/apps/billsplit
go run .
```

Open `http://localhost:8080` in a browser and verify the app loads correctly with the new UI. This tests the Go embed path end-to-end.

---

## Appendix: Running tests

```bash
cd /home/fergal/code/apps/billsplit/frontend
source ~/.nvm/nvm.sh
npx vitest run           # run all tests once
npx vitest               # watch mode
```

## Appendix: Starting dev environment

Terminal 1 (backend):
```bash
cd /home/fergal/code/apps/billsplit
docker-compose up -d     # if minio needed
go run .
```

Terminal 2 (frontend dev server):
```bash
cd /home/fergal/code/apps/billsplit/frontend
source ~/.nvm/nvm.sh
npm run dev
```
