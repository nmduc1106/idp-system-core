import apiClient from '../../utils/apiClient';
import { User } from '../../contexts/AuthContext';

export interface AuthResponse {
    message: string;
    user?: User;
}

export interface RegisterResponse {
    message: string;
}

export const login = async (email: string, password: string): Promise<AuthResponse> => {
    const response = await apiClient.post<AuthResponse>('/auth/login', { email, password });
    return response.data;
};

export const register = async (email: string, password: string, full_name: string): Promise<RegisterResponse> => {
    const response = await apiClient.post<RegisterResponse>('/auth/register', { email, password, full_name });
    return response.data;
};

// Fetch full user profile (including role) from /users/me after login
export const fetchProfile = async (): Promise<User> => {
    const response = await apiClient.get<User>('/users/me');
    return response.data;
};
