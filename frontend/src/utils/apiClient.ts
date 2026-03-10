import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios';

const apiClient = axios.create({
    baseURL: 'http://localhost:8080/api/v1',
    withCredentials: true,
});

// Track if a refresh is already in progress to avoid duplicate calls
let isRefreshing = false;
let failedQueue: { resolve: (value?: unknown) => void; reject: (reason?: unknown) => void; config: InternalAxiosRequestConfig }[] = [];

const processQueue = (error: AxiosError | null) => {
    failedQueue.forEach((prom) => {
        if (error) {
            prom.reject(error);
        } else {
            prom.resolve(apiClient(prom.config));
        }
    });
    failedQueue = [];
};

apiClient.interceptors.response.use(
    (response) => response,
    async (error: AxiosError) => {
        const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };

        // Only intercept 401s that are NOT from login or refresh endpoints
        if (
            error.response?.status === 401 &&
            !originalRequest._retry &&
            !originalRequest.url?.includes('/auth/login') &&
            !originalRequest.url?.includes('/auth/refresh')
        ) {
            if (isRefreshing) {
                // Queue the request until the refresh completes
                return new Promise((resolve, reject) => {
                    failedQueue.push({ resolve, reject, config: originalRequest });
                });
            }

            originalRequest._retry = true;
            isRefreshing = true;

            try {
                await apiClient.post('/auth/refresh');
                processQueue(null);
                // Retry the original request with the new cookie
                return apiClient(originalRequest);
            } catch (refreshError) {
                processQueue(refreshError as AxiosError);
                // Refresh failed — redirect to login
                localStorage.removeItem('user');
                window.location.href = '/login';
                return Promise.reject(refreshError);
            } finally {
                isRefreshing = false;
            }
        }

        return Promise.reject(error);
    }
);

export default apiClient;
