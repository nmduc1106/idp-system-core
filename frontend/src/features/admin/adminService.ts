import apiClient from '../../utils/apiClient';
import { PaginatedResponse, DocumentMeta, JobQuery } from '../document/docService';

export interface SystemStats {
    total_users: number;
    total_jobs: number;
    jobs_by_state: Record<string, number>;
}

export interface AdminJob {
    id: string;
    user_id: string;
    document_id: string;
    state: string;
    result: any;
    retry_count: number;
    error_message?: string;
    trace_id?: string;
    started_at?: string;
    finished_at?: string;
    created_at: string;
    user?: {
        id: string;
        email: string;
        full_name: string;
        role: string;
    };
    document?: DocumentMeta;
}

export interface AdminUser {
    id: string;
    email: string;
    full_name: string;
    role: string;
    created_at: string;
    updated_at: string;
}

export const getSystemStats = async (): Promise<SystemStats> => {
    const response = await apiClient.get<SystemStats>('/admin/stats');
    return response.data;
};

export const getAllJobs = async (q: JobQuery = {}): Promise<PaginatedResponse<AdminJob>> => {
    const params = new URLSearchParams();
    if (q.page) params.append('page', String(q.page));
    if (q.limit) params.append('limit', String(q.limit));
    if (q.status) params.append('status', q.status);
    if (q.file_code) params.append('file_code', q.file_code);

    const response = await apiClient.get<PaginatedResponse<AdminJob>>(`/admin/jobs?${params.toString()}`);
    return response.data;
};

export const getAllUsers = async (): Promise<AdminUser[]> => {
    const response = await apiClient.get<AdminUser[]>('/admin/users');
    return response.data;
};
