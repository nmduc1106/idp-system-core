import React, { useContext } from 'react';
import { Navigate } from 'react-router-dom';
import { AuthContext, AuthContextType } from '../../contexts/AuthContext';

interface AdminRouteProps {
    children: React.ReactNode;
}

const AdminRoute: React.FC<AdminRouteProps> = ({ children }) => {
    const authContext = useContext<AuthContextType | null>(AuthContext);

    if (authContext?.user?.role !== 'ADMIN') {
        return <Navigate to="/" replace />;
    }

    return <>{children}</>;
};

export default AdminRoute;
