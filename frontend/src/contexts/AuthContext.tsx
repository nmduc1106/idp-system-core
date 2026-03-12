import React, { createContext, useState, useEffect } from 'react';
import apiClient from '../utils/apiClient';

export interface User {
    id?: string;
    email: string;
    full_name?: string;
    role?: string;
}

export interface AuthContextType {
    user: User | null;
    login: (userData: User) => void;
    logout: () => void;
    isLoading: boolean;
}

export const AuthContext = createContext<AuthContextType | null>(null);

export const AuthProvider = ({ children }: { children: React.ReactNode }) => {
    const [user, setUser] = useState<User | null>(null);
    const [isLoading, setIsLoading] = useState<boolean>(true);

    useEffect(() => {
        // Initialize state from localStorage (only UI aspects, no tokens)
        const storedUser = localStorage.getItem('user');

        if (storedUser) {
            try {
                setUser(JSON.parse(storedUser));
            } catch (e) {
                console.error('Failed to parse user from localStorage', e);
            }
        }
        setIsLoading(false);
    }, []);

    const login = (userData: User) => {
        setUser(userData);
        localStorage.setItem('user', JSON.stringify(userData));
    };

    const logout = async () => {
        try {
            // Call backend to invalidate refresh token in Redis & clear HttpOnly cookies
            await apiClient.post('/auth/logout');
        } catch (err) {
            // Even if the API call fails (e.g. token expired), we must still clear local state
            console.error('Logout API call failed:', err);
        } finally {
            setUser(null);
            localStorage.removeItem('user');
        }
    };

    if (isLoading) {
        return null; // Or a loading spinner
    }

    return (
        <AuthContext.Provider value={{ user, login, logout, isLoading }}>
            {children}
        </AuthContext.Provider>
    );
};
