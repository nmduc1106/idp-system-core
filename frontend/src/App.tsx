import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import AuthPage from './features/auth/AuthPage';
import Dashboard from './features/document/Dashboard';
import AdminDashboard from './features/admin/AdminDashboard';
import AdminJobs from './features/admin/AdminJobs';
import AdminUsers from './features/admin/AdminUsers';
import ProtectedRoute from './components/auth/ProtectedRoute';
import AdminRoute from './components/auth/AdminRoute';

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

            {/* Admin Routes */}
            <Route
                path="/admin/dashboard"
                element={
                    <ProtectedRoute>
                        <AdminRoute>
                            <AdminDashboard />
                        </AdminRoute>
                    </ProtectedRoute>
                }
            />
            <Route
                path="/admin/jobs"
                element={
                    <ProtectedRoute>
                        <AdminRoute>
                            <AdminJobs />
                        </AdminRoute>
                    </ProtectedRoute>
                }
            />
            <Route
                path="/admin/users"
                element={
                    <ProtectedRoute>
                        <AdminRoute>
                            <AdminUsers />
                        </AdminRoute>
                    </ProtectedRoute>
                }
            />

            {/* Catch-all */}
            <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
    );
};

export default App;
