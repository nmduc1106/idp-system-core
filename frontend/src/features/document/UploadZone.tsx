import React, { useState, useCallback } from 'react';
import { UploadCloud, Loader2, FileText, AlertCircle } from 'lucide-react';
import { uploadFile } from './docService';

interface UploadZoneProps {
    onUploadSuccess: () => void;
}

const FILE_CODE_REGEX = /^[a-zA-Z0-9\-_]{1,50}$/;
const ALLOWED_TYPES = ['application/pdf', 'image/png', 'image/jpeg'];
const ALLOWED_EXTENSIONS = '.pdf,.png,.jpg,.jpeg';

const UploadZone: React.FC<UploadZoneProps> = ({ onUploadSuccess }) => {
    const [isDragging, setIsDragging] = useState(false);
    const [isUploading, setIsUploading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    // Form fields
    const [selectedFile, setSelectedFile] = useState<File | null>(null);
    const [fileCode, setFileCode] = useState('');
    const [notes, setNotes] = useState('');
    const [fileCodeError, setFileCodeError] = useState('');

    const validateFileCode = (value: string): boolean => {
        if (!value.trim()) {
            setFileCodeError('File code is required');
            return false;
        }
        if (!FILE_CODE_REGEX.test(value.trim())) {
            setFileCodeError('Only letters, numbers, dashes, and underscores. Max 50 chars.');
            return false;
        }
        setFileCodeError('');
        return true;
    };

    const validateFile = (file: File): boolean => {
        if (!ALLOWED_TYPES.includes(file.type)) {
            setError('Invalid file type. Only PDF, PNG, and JPEG are allowed.');
            return false;
        }
        return true;
    };

    // Strip HTML tags from notes as a sanitization measure
    const sanitizeText = (text: string): string => {
        return text.replace(/<[^>]*>/g, '').trim();
    };

    const handleSubmit = async () => {
        if (!selectedFile) {
            setError('Please select a file.');
            return;
        }
        if (!validateFile(selectedFile)) return;
        if (!validateFileCode(fileCode)) return;

        setIsUploading(true);
        setError(null);

        try {
            console.log('[UPLOAD] 📤 Initiating upload for file...', { 
                fileName: selectedFile.name, 
                fileCode: fileCode.trim(), 
                notesLength: notes.length 
            });
            const sanitizedNotes = sanitizeText(notes);
            await uploadFile(selectedFile, fileCode.trim(), sanitizedNotes);

            console.log('[UPLOAD] ✅ Upload API call successful! Triggering Dashboard refetch...');
            // Signal Dashboard to refetch from API (gets fully joined data)
            onUploadSuccess();

            // Reset form
            setSelectedFile(null);
            setFileCode('');
            setNotes('');
        } catch (err: any) {
            console.error('[UPLOAD] ❌ Upload API call failed:', err);
            setError(err.response?.data?.error || err.message || 'Failed to upload document.');
        } finally {
            setIsUploading(false);
        }
    };

    const handleDragOver = useCallback((e: React.DragEvent) => {
        e.preventDefault();
        setIsDragging(true);
    }, []);

    const handleDragLeave = useCallback((e: React.DragEvent) => {
        e.preventDefault();
        setIsDragging(false);
    }, []);

    const handleDrop = useCallback((e: React.DragEvent) => {
        e.preventDefault();
        setIsDragging(false);
        if (e.dataTransfer.files?.[0]) {
            const file = e.dataTransfer.files[0];
            if (ALLOWED_TYPES.includes(file.type)) {
                setSelectedFile(file);
                setError(null);
            } else {
                setError('Invalid file type. Only PDF, PNG, and JPEG are allowed.');
            }
        }
    }, []);

    const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        if (e.target.files?.[0]) {
            setSelectedFile(e.target.files[0]);
            setError(null);
        }
    };

    return (
        <div className="w-full space-y-5">
            {/* File Drop Zone */}
            <div
                className={`relative border-2 border-dashed rounded-xl p-8 flex flex-col items-center justify-center transition-all duration-200 ${
                    isDragging
                        ? 'border-primary bg-primary/5'
                        : selectedFile
                        ? 'border-emerald-400 bg-emerald-50'
                        : 'border-slate-300 hover:border-primary bg-white'
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
                    accept={ALLOWED_EXTENSIONS}
                />

                {selectedFile ? (
                    <div className="flex items-center gap-3 text-emerald-700">
                        <FileText className="h-8 w-8" />
                        <div>
                            <p className="font-medium">{selectedFile.name}</p>
                            <p className="text-xs text-slate-500">{(selectedFile.size / 1024).toFixed(1)} KB · Click or drop to replace</p>
                        </div>
                    </div>
                ) : (
                    <div className="flex flex-col items-center text-slate-500">
                        <div className={`p-3 rounded-full mb-3 ${isDragging ? 'bg-primary/10 text-primary' : 'bg-slate-100 text-slate-400'}`}>
                            <UploadCloud className="h-8 w-8" />
                        </div>
                        <p className="font-medium text-slate-700 mb-0.5">Drag & drop your document here</p>
                        <p className="text-sm">or click to browse · PDF, PNG, JPEG</p>
                    </div>
                )}
            </div>

            {/* Metadata Form */}
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-1">
                        File Code <span className="text-red-500">*</span>
                    </label>
                    <input
                        type="text"
                        value={fileCode}
                        onChange={(e) => {
                            setFileCode(e.target.value);
                            if (fileCodeError) validateFileCode(e.target.value);
                        }}
                        placeholder="e.g. INV-2026-001"
                        maxLength={50}
                        className={`w-full px-3 py-2 text-sm border rounded-lg focus:ring-2 focus:ring-primary focus:border-primary outline-none transition ${
                            fileCodeError ? 'border-red-300 bg-red-50' : 'border-slate-200 bg-white'
                        }`}
                        disabled={isUploading}
                    />
                    {fileCodeError && <p className="mt-1 text-xs text-red-500 flex items-center gap-1"><AlertCircle className="h-3 w-3" />{fileCodeError}</p>}
                </div>
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-1">Notes</label>
                    <textarea
                        value={notes}
                        onChange={(e) => setNotes(e.target.value)}
                        placeholder="Optional notes about this document..."
                        rows={1}
                        className="w-full px-3 py-2 text-sm border border-slate-200 rounded-lg bg-white focus:ring-2 focus:ring-primary focus:border-primary outline-none transition resize-none"
                        disabled={isUploading}
                    />
                </div>
            </div>

            {/* Submit Button */}
            <button
                onClick={handleSubmit}
                disabled={isUploading || !selectedFile}
                className="w-full sm:w-auto px-6 py-2.5 bg-primary text-white text-sm font-semibold rounded-lg shadow-sm hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed transition flex items-center justify-center gap-2"
            >
                {isUploading ? <Loader2 className="h-4 w-4 animate-spin" /> : <UploadCloud className="h-4 w-4" />}
                {isUploading ? 'Uploading...' : 'Upload Document'}
            </button>

            {/* Error */}
            {error && (
                <div className="p-3 rounded-lg bg-red-50 border border-red-200 text-red-700 text-sm flex items-center gap-2">
                    <AlertCircle className="h-4 w-4 flex-shrink-0" />
                    {error}
                </div>
            )}
        </div>
    );
};

export default UploadZone;
