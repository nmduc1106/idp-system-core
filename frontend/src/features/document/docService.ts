import apiClient from '../../utils/apiClient';

// --- Shared Pagination Types ---
export interface PaginatedResponse<T> {
    data: T[];
    total: number;
    page: number;
    limit: number;
    total_pages: number;
}

export interface DocumentMeta {
    id: string;
    file_name: string;
    file_code: string;
    notes?: string;
    mime_type: string;
    file_size: number;
    original_filename: string;
    created_at: string;
}

export interface Job {
    id: string;
    document_id: string;
    user_id: string;
    state: 'PENDING' | 'EXTRACTING' | 'COMPLETED' | 'FAILED';
    result: any;
    error?: string;
    created_at: string;
    updated_at: string;
    document?: DocumentMeta;
}

export interface UploadResponse {
    message: string;
    job_id: string;
    doc_id: string;
}

export interface JobQuery {
    page?: number;
    limit?: number;
    status?: string;
    file_code?: string;
}

// --- Upload (with metadata) ---
export const uploadFile = async (file: File, fileCode: string, notes: string): Promise<UploadResponse> => {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('file_code', fileCode);
    formData.append('notes', notes);

    const response = await apiClient.post<UploadResponse>('/upload', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
    });
    return response.data;
};

// --- Paginated User Jobs ---
export const getUserJobs = async (q: JobQuery = {}): Promise<PaginatedResponse<Job>> => {
    const params = new URLSearchParams();
    if (q.page) params.append('page', String(q.page));
    if (q.limit) params.append('limit', String(q.limit));
    if (q.status) params.append('status', q.status);
    if (q.file_code) params.append('file_code', q.file_code);

    const response = await apiClient.get<PaginatedResponse<Job>>(`/jobs?${params.toString()}`);
    return response.data;
};

// --- Export to Excel ---
export const exportJobsExcel = async (fileCode?: string): Promise<void> => {
    const params = new URLSearchParams();
    if (fileCode) params.append('file_code', fileCode);

    const response = await apiClient.get(`/jobs/export?${params.toString()}`, {
        responseType: 'blob',
    });

    const url = window.URL.createObjectURL(new Blob([response.data]));
    const link = document.createElement('a');
    link.href = url;

    // Format: IDP_Report_2026-03-15.xlsx
    const dateStr = new Date().toISOString().split('T')[0];
    link.setAttribute('download', `IDP_Report_${dateStr}.xlsx`);

    document.body.appendChild(link);
    link.click();

    // Cleanup
    document.body.removeChild(link);
    window.URL.revokeObjectURL(url);
};

// --- SSE Stream ---
export const streamJobStatus = (
    jobId: string,
    onMessage: (data: Job) => void,
    onError: (err: Event) => void
): EventSource => {
    console.log(`[SSE] 🟢 Opening stream for Job ID: ${jobId}`);

    // Tự động nhận diện môi trường (VPS hay Local)
    const baseUrl = (import.meta as any).env.VITE_API_BASE_URL
        ? `${(import.meta as any).env.VITE_API_BASE_URL}/api/v1`
        : 'http://localhost:8080/api/v1';

    const url = `${baseUrl}/jobs/${jobId}/stream`;
    const source = new EventSource(url, { withCredentials: true });

    source.onopen = () => console.log(`[SSE] 🌐 Connection successfully opened for Job: ${jobId}`);

    source.onmessage = (e: MessageEvent) => {
        try {
            console.log(`[SSE] 📥 Raw message received for Job ${jobId}:`, e.data);
            const raw = JSON.parse(e.data);
            const data: Job = {
                ...raw,
                id: raw.id || raw.job_id,
                state: raw.state || raw.status,
            };
            console.log(`[SSE] 🧩 Parsed message for Job ${jobId}:`, data);

            onMessage(data);

            if (data.state === 'COMPLETED' || data.state === 'FAILED') {
                console.log(`[SSE] 🛑 Terminal state reached (${data.state}) for Job ${jobId}. Closing internal stream.`);
                source.close();
            }
        } catch (err) {
            console.error(`[SSE] ❌ Failed to parse SSE message for Job ${jobId}:`, err);
        }
    };

    source.onerror = (err: Event) => {
        console.error(`[SSE] ❌ Connection error/closed for Job ${jobId}:`, err);
        onError(err);
        source.close();
    };

    return source;
};
