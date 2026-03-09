# Security Guidelines

Security is a primary concern for the IDP system. All future UI developments MUST adhere to the following rules:

## 1. No Sensitive Data in LocalStorage
Never store JWT tokens, passwords, or PII (Personally Identifiable Information) in `localStorage` or `sessionStorage`. These storage mechanisms are vulnerable to XSS attacks. Authentication tokens must be handled via HttpOnly, Secure, SameSite cookies.

## 2. XSS Prevention
Always rely on React's built-in data binding to escape dynamic content before rendering. 
- **NEVER** use `dangerouslySetInnerHTML` unless explicitly required and the input has undergone rigorous, server-side-grade sanitization (e.g., via DOMPurify).

## 3. CSRF Protection
Cross-Site Request Forgery must be prevented.
- Our `apiClient` Axios instance MUST be configured with `withCredentials: true` to ensure cookies are sent securely.
- The Go backend MUST set cookies with `SameSite=Lax` (or `Strict`) and `Secure=true`.

## 4. Input Sanitization
Treat all user input as untrusted. 
- All forms must validate lengths, types, and formats.
- Inputs must be trimmed and sanitized before being submitted to the backend to prevent injection vectors.

## 5. Secure SSE
Server-Sent Events (`EventSource`) natively do not support custom authorization headers.
- Rely on the HttpOnly Cookie implementation. The browser will automatically attach the valid session cookie to the SSE stream request if `withCredentials` behavior is respected. Token strings must NOT be passed into URL query parameters.
