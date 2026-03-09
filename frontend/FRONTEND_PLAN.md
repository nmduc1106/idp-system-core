# IDP Frontend Master Plan (React + Vite + TypeScript + Tailwind)

## 🗺️ PHASE 1: Core Setup
**Objective:** Configure Tailwind, set up the API client with automatic token injection, and establish global authentication state.
- [x] 1.1 **Tailwind Config:** Update `tailwind.config.js` and `src/index.css`.
- [x] 1.2 **API Client:** Create `src/utils/apiClient.ts` (Axios instance fetching token from `localStorage`).
- [x] 1.3 **Auth Context:** Create `src/contexts/AuthContext.tsx` (Manage User/Token state, login/logout actions).

## 🗺️ PHASE 2: Auth Feature (Security Gate)
**Objective:** Build a seamless Login/Register UI to acquire JWT tokens.
- [x] 2.1 **Auth Service:** Create `src/features/auth/authService.ts` (API calls for `/login` and `/register`).
- [x] 2.2 **Auth UI:** Create `src/features/auth/AuthPage.tsx` (Toggleable Login/Register form using `lucide-react`).

## 🗺️ PHASE 3: Dashboard & Real-time (Control Station)
**Objective:** The core workspace for drag-and-drop uploads and real-time SSE status tracking.
- [x] 3.1 **Layout:** Create `src/components/layout/MainLayout.tsx` (Navbar, user info, logout).
- [x] 3.2 **Doc Service:** Create `src/features/document/docService.ts` (Upload API and SSE EventSource logic).
- [x] 3.3 **Upload Zone:** Create `src/features/document/UploadZone.tsx` (Drag & drop area).
- [x] 3.4 **Job Table:** Create `src/features/document/JobTable.tsx` (Real-time SSE connected data table).
- [x] 3.5 **Dashboard:** Create `src/features/document/Dashboard.tsx` (Combine UploadZone and JobTable).

## 🗺️ PHASE 4: App Routing & Assembly
**Objective:** Connect all pieces using React Router.
- [x] 4.1 **Router Setup:** Update `src/App.tsx` (Define `/login` and protected `/` routes).
- [x] 4.2 **Entry Point:** Update `src/main.tsx` (Wrap App in AuthContext and BrowserRouter).

## 🗺️ PHASE 5: Security Refactor (HttpOnly Cookies)
**Objective:** Transition to secure, HttpOnly cookie-based authentication, remove token handling from LocalStorage, and establish standard security guidelines to prevent XSS.
- [x] 5.1 **Security Guidelines:** Create `SECURITY_GUIDELINES.md` outlining XSS prevention and SSE specifics.
- [x] 5.2 **API Client Refactor:** Update `src/utils/apiClient.ts` with `withCredentials: true`.
- [x] 5.3 **Context & Routing Refactor:** Remove token from `AuthContext.tsx`, `AuthPage.tsx`, and `ProtectedRoute.tsx`.
- [x] 5.4 **Streaming Security:** Remove URL-based tokens from `docService.ts`, ensuring EventSource passes explicit credentials.
