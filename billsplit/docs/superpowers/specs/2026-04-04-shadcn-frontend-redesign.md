# BillSplit Frontend Redesign — shadcn/ui

**Date:** 2026-04-04
**Status:** Approved

## Overview

Replace the existing plain-HTML React frontend with a new frontend built from scratch using Vite, React, TypeScript, Tailwind CSS, and shadcn/ui. The new frontend introduces a persistent sidebar layout, React Router, dark/light mode support, and a blue accent colour theme.

The backend (Go), API contract, and `api.ts` business logic are unchanged. Only the UI layer is replaced.

---

## Approach

Scaffold a fresh Vite + React + TypeScript project in `billsplit/frontend-new/`. Once complete and verified, `frontend-new/` replaces `frontend/` and the old directory is deleted. The Go embed path and Vite build output directory remain `dist/` so the backend requires no changes.

---

## Stack

| Concern | Choice |
|---|---|
| Bundler | Vite 5 |
| Framework | React 18 + TypeScript |
| Styling | Tailwind CSS v3 |
| Component library | shadcn/ui |
| Routing | React Router v6 |
| Toast notifications | Sonner (shadcn integration) |
| Theme | Blue accent, system dark/light default + manual toggle |

---

## Project Structure

```
frontend-new/
  index.html
  vite.config.ts
  tailwind.config.ts
  tsconfig.json
  src/
    main.tsx
    App.tsx                   # Router setup, auth state, token persistence
    api.ts                    # Copied from api.js, typed
    lib/
      utils.ts                # shadcn cn() helper
    components/
      ui/                     # shadcn auto-generated components
      AppShell.tsx            # Sidebar + main content layout
      ThemeProvider.tsx       # Dark/light mode context + toggle
      CreateGroupDialog.tsx   # shadcn Dialog — create group form
      AddExpenseModal.tsx     # shadcn Dialog
      AddSettlementModal.tsx  # shadcn Dialog
    pages/
      Login.tsx
      Register.tsx
      GroupDetail.tsx
```

`GroupList.jsx` is not migrated — the group list moves into the sidebar inside `AppShell`.

---

## Routing

| Path | Component | Auth required |
|---|---|---|
| `/login` | `Login` | No |
| `/register` | `Register` | No |
| `/` | `AppShell` → redirect to first group | Yes |
| `/groups/:groupId` | `AppShell` → `GroupDetail` | Yes |

A `<ProtectedRoute>` wrapper checks for a token in `localStorage` and redirects to `/login` if absent.

---

## Auth & Token Persistence

- On login, the JWT is stored in `localStorage` under the key `billsplit_token`.
- On app load, `App.tsx` reads the token from `localStorage` to restore the session.
- On sign-out, the token is removed from `localStorage` and the user is redirected to `/login`.

---

## Layout: AppShell

The `AppShell` component renders a two-column layout:

**Sidebar (left, fixed width ~220px):**
- App name "BillSplit" at the top
- List of the user's groups — fetched on mount, active group highlighted in blue; each item shows the group name and the user's net balance (green if owed, red if owing)
- "+ New group" button at the bottom of the group list — opens `CreateGroupDialog`
- If the user is an admin: a "Generate invite" button that opens an inline panel showing the generated invite code with a copy button (regular or admin invite, configurable via checkbox)
- Username and dark/light mode toggle at the very bottom

**Main content area (right, flex-grow):**
- Renders the current page (`GroupDetail` or an empty/welcome state if no group is selected)

The sidebar is not shown on `/login` or `/register`.

---

## Theme: Dark/Light Mode

- `ThemeProvider` wraps the app and reads the OS preference via `prefers-color-scheme` on first load.
- The user can override with a toggle in the sidebar (sun/moon icon).
- The chosen preference is persisted to `localStorage` under `billsplit_theme`.
- Implemented via shadcn's recommended `class` strategy on the `<html>` element (Tailwind `dark:` variants).

---

## API Layer (`api.ts`)

`api.js` is copied across and typed. Key types:

```ts
type Group = {
  id: string
  name: string
  currency: string
  balances: Record<string, number>
  members: string[]
}

type EventType = 'expense' | 'settlement' | 'reversal'

type GroupEvent = {
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

type EventsResponse = {
  events: GroupEvent[]
  total: number
}
```

All `api` functions are typed with these shapes as return values.

`parseToken(token: string): { username: string; isAdmin: boolean }` — decodes the JWT payload (no verification, client-side only) and is also typed and exported from `api.ts`.

The `localStorage` key for the auth token is `billsplit_token` (the existing app uses `token` — on first load after the migration the user will need to log in again; this is acceptable).

---

## Error Handling

| Context | Approach |
|---|---|
| Form validation (Login, Register, modals) | Inline error text below the relevant field/form, using shadcn `FormMessage` |
| Background failures (load group, delete expense, etc.) | Sonner toast (bottom-right, auto-dismiss) |

---

## Pages & Components

### Login / Register
- shadcn `Card` centred on screen (no sidebar)
- shadcn `Input`, `Button`, `Label`
- Inline error below form on failure

### GroupDetail
- Header: group name + currency badge, "+ Expense" (primary blue button) and "+ Settle" (secondary) buttons
- Balances section: shadcn `Card` listing each member with their balance in green (positive) or red (negative)
- History list: shadcn `Card` per event, expandable splits on expense rows, muted styling for reversals
- Pagination: previous/next buttons below history list

### AddExpenseModal / AddSettlementModal
- shadcn `Dialog` (replaces current inline conditional renders)
- Same fields as today; shadcn `Input`, `Select`, `Button`

### CreateGroupDialog
- shadcn `Dialog` triggered by "+ New group" in sidebar
- Fields: group name (text input), currency (text input, default GBP)
- Member picker: fetches all users from `api.getUsers()`, filters out the current user; multi-select list with toggle-chip display of selected members (same UX as today)
- Inline validation: at least one member must be selected
- On success: closes dialog, sidebar group list refreshes

---

## Out of Scope

- Backend changes of any kind
- New features (invite flow UI, settle-up recommendations, etc.)
- React Router lazy loading / code splitting
- End-to-end tests
