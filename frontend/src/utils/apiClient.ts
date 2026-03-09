import axios, { AxiosError } from 'axios';

const apiClient = axios.create({
    baseURL: 'http://localhost:8080/api/v1',
    withCredentials: true,
});

apiClient.interceptors.response.use(
    (response) => response,
    (error: AxiosError) => {
        return Promise.reject(error);
    }
);

export default apiClient;
