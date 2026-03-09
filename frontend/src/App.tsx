import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import AuthPage from './features/auth/AuthPage';
import Dashboard from './features/document/Dashboard';
import ProtectedRoute from './components/auth/ProtectedRoute';

const App: React.FC = () => {
    return (
        <Routes>
            <Route path="/login" element={<AuthPage />} />

            <Route
                path="/"
                element={
                    <ProtectedRoute>
                        <Dashboard />
                    </ProtectedRoute>
                }
            />

            {/* Catch-all route to redirect back to protected dashboard if authenticated, else login */}
            <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
    );
};

export default App;
