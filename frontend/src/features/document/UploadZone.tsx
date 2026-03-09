import React, { useState, useCallback } from 'react';
import { UploadCloud, Loader2 } from 'lucide-react';
import { uploadFile, Job } from './docService';

interface UploadZoneProps {
    onUploadSuccess: (job: Job) => void;
}

const UploadZone: React.FC<UploadZoneProps> = ({ onUploadSuccess }) => {
    const [isDragging, setIsDragging] = useState<boolean>(false);
    const [isUploading, setIsUploading] = useState<boolean>(false);
    const [error, setError] = useState<string | null>(null);

    const handleDragOver = useCallback((e: React.DragEvent<HTMLDivElement>) => {
        e.preventDefault();
        setIsDragging(true);
    }, []);

    const handleDragLeave = useCallback((e: React.DragEvent<HTMLDivElement>) => {
        e.preventDefault();
        setIsDragging(false);
    }, []);

    const processFile = async (file: File) => {
        if (!file) return;

        // Simplistic check, usually allow PDF, PNG, JPEG
        setIsUploading(true);
        setError(null);

        try {
            const response = await uploadFile(file);

            // Create a dummy starting pending job to inject into the dashboard state
            const newJob: Job = {
                id: response.job_id,
                document_id: response.document_id,
                user_id: '', // Would be set by backend but we don't strictly need it in UI yet
                state: 'PENDING',
                result: null,
                created_at: new Date().toISOString(),
                updated_at: new Date().toISOString(),
            };

            onUploadSuccess(newJob);
        } catch (err: any) {
            setError(err.response?.data?.message || err.message || 'Failed to upload document.');
        } finally {
            setIsUploading(false);
        }
    };

    const handleDrop = useCallback((e: React.DragEvent<HTMLDivElement>) => {
        e.preventDefault();
        setIsDragging(false);

        if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
            const file = e.dataTransfer.files[0];
            processFile(file);
            e.dataTransfer.clearData();
        }
    }, []);

    const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        if (e.target.files && e.target.files.length > 0) {
            processFile(e.target.files[0]);
        }
    };

    return (
        <div className="w-full">
            <div
                className={`relative border-2 border-dashed rounded-xl p-10 flex flex-col items-center justify-center transition-all duration-200 ease-in-out ${isDragging
                        ? 'border-indigo-500 bg-indigo-50 dark:bg-indigo-900/20'
                        : 'border-slate-300 dark:border-slate-700 hover:border-indigo-400 dark:hover:border-indigo-500 bg-white dark:bg-slate-800'
                    }`}
                onDragOver={handleDragOver}
                onDragLeave={handleDragLeave}
                onDrop={handleDrop}
            >
                <input
                    type="file"
                    className="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
                    onChange={handleFileChange}
                    disabled={isUploading}
                    accept=".pdf,.png,.jpg,.jpeg"
                />

                {isUploading ? (
                    <div className="flex flex-col items-center text-indigo-600 dark:text-indigo-400">
                        <Loader2 className="h-12 w-12 animate-spin mb-4" />
                        <p className="text-lg font-medium">Uploading document...</p>
                    </div>
                ) : (
                    <div className="flex flex-col items-center text-slate-500 dark:text-slate-400">
                        <div className={`p-4 rounded-full mb-4 ${isDragging ? 'bg-indigo-100 dark:bg-indigo-900/50 text-indigo-600 dark:text-indigo-400' : 'bg-slate-100 dark:bg-slate-700 text-slate-400 dark:text-slate-300'}`}>
                            <UploadCloud className="h-10 w-10" />
                        </div>
                        <p className="text-lg font-medium text-slate-700 dark:text-slate-200 mb-1">
                            Drag & drop your document here
                        </p>
                        <p className="text-sm">or click to browse from your computer</p>
                        <p className="text-xs mt-4 text-slate-400">Supports PDF, PNG, and JPEG files.</p>
                    </div>
                )}
            </div>

            {error && (
                <div className="mt-4 p-4 rounded-md bg-red-50 dark:bg-red-900/30 border-l-4 border-red-500 text-red-700 dark:text-red-300 text-sm">
                    {error}
                </div>
            )}
        </div>
    );
};

export default UploadZone;
