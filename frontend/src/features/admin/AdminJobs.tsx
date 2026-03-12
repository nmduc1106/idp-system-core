import React, { useState, useEffect, useCallback } from 'react';
import MainLayout from '../../components/layout/MainLayout';
import { getAllJobs, AdminJob } from './adminService';
import { PaginatedResponse } from '../document/docService';
import { Loader2, Search, FileText, ChevronLeft, ChevronRight } from 'lucide-react';

const stateBadge = (state: string) => {
    const styles: Record<string, string> = {
        COMPLETED: 'bg-emerald-100 text-emerald-700 border-emerald-200',
        FAILED: 'bg-rose-100 text-rose-700 border-rose-200',
        PENDING: 'bg-amber-100 text-amber-700 border-amber-200',
        EXTRACTING: 'bg-sky-100 text-sky-700 border-sky-200',
    };
    return (
        <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold border ${styles[state] || 'bg-slate-100 text-slate-700 border-slate-200'}`}>
            {state}
        </span>
    );
};

const AdminJobs: React.FC = () => {
    const [pageData, setPageData] = useState<PaginatedResponse<AdminJob> | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [page, setPage] = useState(1);
    const [searchCode, setSearchCode] = useState('');
    const [searchInput, setSearchInput] = useState('');

    const fetchJobs = useCallback(async () => {
        setLoading(true);
        try {
            const data = await getAllJobs({ page, limit: 10, file_code: searchCode || undefined });
            setPageData(data);
        } catch (err: any) {
            setError(err.response?.data?.error || 'Failed to load jobs');
        } finally {
            setLoading(false);
        }
    }, [page, searchCode]);

    useEffect(() => {
        fetchJobs();
    }, [fetchJobs]);

    const handleSearchSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        setSearchCode(searchInput.trim());
        setPage(1);
    };

    const rawData = pageData?.data;
    const jobs = Array.isArray(rawData) ? rawData : [];
    const totalPages = pageData?.total_pages || 1;
    const total = pageData?.total || 0;

    return (
        <MainLayout>
            <div className="space-y-6">
                <header className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
                    <div>
                        <h1 className="text-2xl font-bold text-slate-900">All Jobs</h1>
                        <p className="mt-1 text-sm text-slate-500">{total} total processing jobs across all users.</p>
                    </div>
                    <form onSubmit={handleSearchSubmit} className="relative w-full sm:w-72">
                        <Search className="absolute left-3 top-2.5 h-4 w-4 text-slate-400" />
                        <input
                            type="text"
                            placeholder="Search by File Code..."
                            value={searchInput}
                            onChange={(e) => setSearchInput(e.target.value)}
                            className="w-full pl-10 pr-4 py-2 text-sm border border-slate-200 rounded-lg bg-white focus:ring-2 focus:ring-primary focus:border-primary outline-none"
                        />
                    </form>
                </header>

                {error && <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg">{error}</div>}

                <div className="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
                    <div className="overflow-x-auto">
                        <table className="min-w-full divide-y divide-slate-200">
                            <thead className="bg-slate-50">
                                <tr>
                                    <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500 uppercase tracking-wider">Document</th>
                                    <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500 uppercase tracking-wider">File Code</th>
                                    <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500 uppercase tracking-wider">Uploader</th>
                                    <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500 uppercase tracking-wider">Status</th>
                                    <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500 uppercase tracking-wider">Created</th>
                                    <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500 uppercase tracking-wider">Retries</th>
                                </tr>
                            </thead>
                            <tbody className="bg-white divide-y divide-slate-100">
                                {loading ? (
                                    <tr>
                                        <td colSpan={6} className="px-6 py-12 text-center">
                                            <Loader2 className="h-6 w-6 animate-spin text-primary mx-auto" />
                                        </td>
                                    </tr>
                                ) : jobs.length === 0 ? (
                                    <tr>
                                        <td colSpan={6} className="px-6 py-12 text-center text-slate-400">
                                            <FileText className="h-8 w-8 mx-auto mb-2 opacity-40" />
                                            No jobs found.
                                        </td>
                                    </tr>
                                ) : (
                                    jobs.map((job) => (
                                        <tr key={job.id} className="hover:bg-slate-50 transition-colors">
                                            <td className="px-6 py-4 whitespace-nowrap">
                                                <div className="text-sm font-medium text-slate-900">
                                                    {job.document?.file_name || job.id.substring(0, 8) + '...'}
                                                </div>
                                                <div className="text-xs text-slate-500">{job.id.substring(0, 8)}...</div>
                                            </td>
                                            <td className="px-6 py-4 whitespace-nowrap">
                                                {job.document?.file_code ? (
                                                    <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-mono font-medium bg-slate-100 text-slate-700 border border-slate-200">
                                                        {job.document.file_code}
                                                    </span>
                                                ) : (
                                                    <span className="text-xs text-slate-400">—</span>
                                                )}
                                            </td>
                                            <td className="px-6 py-4 whitespace-nowrap">
                                                <div className="text-sm font-medium text-slate-900">{job.user?.full_name || '—'}</div>
                                                <div className="text-xs text-slate-500">{job.user?.email || '—'}</div>
                                            </td>
                                            <td className="px-6 py-4 whitespace-nowrap">{stateBadge(job.state)}</td>
                                            <td className="px-6 py-4 whitespace-nowrap text-sm text-slate-500">
                                                {new Date(job.created_at).toLocaleString()}
                                            </td>
                                            <td className="px-6 py-4 whitespace-nowrap text-sm text-slate-500">{job.retry_count}</td>
                                        </tr>
                                    ))
                                )}
                            </tbody>
                        </table>
                    </div>

                    {/* Pagination Controls */}
                    {totalPages > 1 && (
                        <div className="flex items-center justify-between px-6 py-3 border-t border-slate-200 bg-slate-50">
                            <p className="text-sm text-slate-500">
                                Page {page} of {totalPages} · {total} total
                            </p>
                            <div className="flex items-center gap-2">
                                <button
                                    onClick={() => setPage((p) => Math.max(1, p - 1))}
                                    disabled={page <= 1}
                                    className="inline-flex items-center px-3 py-1.5 text-sm border border-slate-200 rounded-lg bg-white hover:bg-slate-50 disabled:opacity-40 disabled:cursor-not-allowed transition"
                                >
                                    <ChevronLeft className="h-4 w-4 mr-1" /> Prev
                                </button>
                                <button
                                    onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                                    disabled={page >= totalPages}
                                    className="inline-flex items-center px-3 py-1.5 text-sm border border-slate-200 rounded-lg bg-white hover:bg-slate-50 disabled:opacity-40 disabled:cursor-not-allowed transition"
                                >
                                    Next <ChevronRight className="h-4 w-4 ml-1" />
                                </button>
                            </div>
                        </div>
                    )}
                </div>
            </div>
        </MainLayout>
    );
};

export default AdminJobs;
