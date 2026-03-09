# IDP Frontend

A modern, real-time Intelligent Document Processing frontend built with React, Vite, TypeScript, and Tailwind CSS. It communicates securely via JWT tokens and streams processing statuses using Server-Sent Events (SSE).

## 1. Project Structure (Directory Map)

The `src/` directory is organized to support scalability and feature isolation:

```text
src/
├── assets/         # Static assets like images and global icons
├── components/     # Reusable, feature-agnostic UI elements
│   ├── auth/       # Authentication-related wrappers (e.g., ProtectedRoute)
│   └── layout/     # Page layout structures (e.g., MainLayout, Navbars)
├── contexts/       # Global React Context providers (e.g., AuthContext)
├── features/       # Feature-driven modules containing domain-specific logic and UI
│   ├── auth/       # Login, Registration, and Auth API services
│   └── document/   # Dashboard, Upload, Job tables, and Streaming services
├── utils/          # Generic tools and configs (e.g., Axios API Client)
├── App.tsx         # Main application router and assembly
├── main.tsx        # React DOM entry point and global provider wrapping
└── index.css       # Global Tailwind directives and base styles
```

## 2. Logic vs. UI Separation

The architecture cleanly separates the application's core logic ("the brain") from the visual presentation ("the body") to enhance testability and reusability.

### Logic Management ("The Brain")
These files handle data fetching, global state, authentication workflows, and real-time connections.
- `utils/apiClient.ts` - Singleton Axios instance configured with auto-injecting Bearer tokens.
- `contexts/AuthContext.tsx` - Provides global user and token state, syncing with `localStorage`.
- `features/auth/authService.ts` - Encapsulates `/auth/login` and `/auth/register` API calls.
- `features/document/docService.ts` - Manages `FormData` uploads and native `EventSource` streams for SSE.

### UI Management ("The Body")
These components handle user interactions, state representation, and beautiful Tailwind rendering.
- `features/auth/AuthPage.tsx` - The toggleable Login/Register card UI.
- `features/document/Dashboard.tsx` - The main container managing local `Job[]` state.
- `features/document/UploadZone.tsx` - The drag-and-drop file upload target.
- `features/document/JobTable.tsx` - The dynamic table rendering active jobs and extraction results.
- `components/layout/MainLayout.tsx` - The responsive application shell and navigation bar.

## 3. Core Features & File Mapping

| Feature | Description | Primary Files |
| :--- | :--- | :--- |
| **Authentication** | Secure JWT-based login and registration with automatic token injection. | `AuthContext.tsx`, `AuthPage.tsx`, `authService.ts` |
| **Protected Routing** | Restricts unauthenticated users from accessing the private dashboard. | `ProtectedRoute.tsx`, `App.tsx` |
| **Document Management** | Seamless drag-and-drop upload for PDF, PNG, and JPEG files. | `UploadZone.tsx`, `docService.ts` |
| **Real-time Monitoring** | Zero-refresh Server-Sent Events (SSE) tracking of extraction jobs (PENDING to COMPLETED). | `JobTable.tsx`, `docService.ts` |

## 4. Technical Stack

- **Framework:** [React 18](https://react.dev/)
- **Bundler:** [Vite](https://vitejs.dev/)
- **Language:** [TypeScript](https://www.typescriptlang.org/)
- **Styling:** [Tailwind CSS](https://tailwindcss.com/)
- **Routing:** [React Router (v6)](https://reactrouter.com/)
- **HTTP Client:** [Axios](https://axios-http.com/)
- **Icons:** [Lucide React](https://lucide.dev/)

## 5. How to Run

1. Navigate to the frontend directory:
   ```bash
   cd frontend
   ```
2. Install dependencies:
   ```bash
   npm install
   ```
3. Start the Vite development server:
   ```bash
   npm run dev
   ```
The application will be available at `http://localhost:5173`. Make sure the Go API Gateway is running on `localhost:8080`.
