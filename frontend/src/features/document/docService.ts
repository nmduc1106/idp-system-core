import apiClient from '../../utils/apiClient';

export interface Job {
    id: string;
    document_id: string;
    user_id: string;
    state: 'PENDING' | 'EXTRACTING' | 'COMPLETED' | 'FAILED';
    result: any;
    created_at: string;
    updated_at: string;
}

export interface UploadResponse {
    message: string;
    job_id: string;
    document_id: string;
}

export const uploadFile = async (file: File): Promise<UploadResponse> => {
    const formData = new FormData();
    formData.append('file', file);

    const response = await apiClient.post<UploadResponse>('/upload', formData, {
        headers: {
            'Content-Type': 'multipart/form-data',
        },
    });

    return response.data;
};

export const streamJobStatus = (
    jobId: string,
    onMessage: (data: Job) => void,
    onError: (err: Event) => void
): EventSource => {
    const url = `http://localhost:8080/api/v1/jobs/${jobId}/stream`;

    const source = new EventSource(url, { withCredentials: true });

    source.onmessage = (e: MessageEvent) => {
        try {
            const data: Job = JSON.parse(e.data);
            onMessage(data);

            if (data.state === 'COMPLETED' || data.state === 'FAILED') {
                source.close();
            }
        } catch (err) {
            console.error('Failed to parse SSE message', err);
        }
    };

    source.onerror = (err: Event) => {
        console.error('SSE Error:', err);
        onError(err);
        source.close(); // Close on error to prevent infinite reconnection loops if unauthorized
    };

    return source;
};
