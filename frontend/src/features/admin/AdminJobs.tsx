import React, { useState, useEffect } from 'react';
import MainLayout from '../../components/layout/MainLayout';
import { getAllJobs, AdminJob } from './adminService';
import { Loader2, Search, FileText } from 'lucide-react';

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
    const [jobs, setJobs] = useState<AdminJob[]>([]);
    const [filtered, setFiltered] = useState<AdminJob[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [search, setSearch] = useState('');

    useEffect(() => {
        const fetchJobs = async () => {
            try {
                const data = await getAllJobs();
                setJobs(data);
                setFiltered(data);
            } catch (err: any) {
                setError(err.response?.data?.error || 'Failed to load jobs');
            } finally {
                setLoading(false);
            }
        };
        fetchJobs();
    }, []);

    useEffect(() => {
        const q = search.toLowerCase();
        setFiltered(
            jobs.filter(
                (j) =>
                    j.id.toLowerCase().includes(q) ||
                    j.state.toLowerCase().includes(q) ||
                    j.user?.email?.toLowerCase().includes(q) ||
                    j.user?.full_name?.toLowerCase().includes(q)
            )
        );
    }, [search, jobs]);

    if (loading) {
        return (
            <MainLayout>
                <div className="flex items-center justify-center h-64">
                    <Loader2 className="h-8 w-8 animate-spin text-primary" />
                </div>
            </MainLayout>
        );
    }

    return (
        <MainLayout>
            <div className="space-y-6">
                <header className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
                    <div>
                        <h1 className="text-2xl font-bold text-slate-900">All Jobs</h1>
                        <p className="mt-1 text-sm text-slate-500">{jobs.length} total processing jobs across all users.</p>
                    </div>
                    <div className="relative w-full sm:w-72">
                        <Search className="absolute left-3 top-2.5 h-4 w-4 text-slate-400" />
                        <input
                            type="text"
                            placeholder="Search by ID, status, email..."
                            value={search}
                            onChange={(e) => setSearch(e.target.value)}
                            className="w-full pl-10 pr-4 py-2 text-sm border border-slate-200 rounded-lg bg-white focus:ring-2 focus:ring-primary focus:border-primary outline-none"
                        />
                    </div>
                </header>

                {error && <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg">{error}</div>}

                <div className="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
                    <div className="overflow-x-auto">
                        <table className="min-w-full divide-y divide-slate-200">
                            <thead className="bg-slate-50">
                                <tr>
                                    <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500 uppercase tracking-wider">Job ID</th>
                                    <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500 uppercase tracking-wider">Uploader</th>
                                    <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500 uppercase tracking-wider">Status</th>
                                    <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500 uppercase tracking-wider">Created</th>
                                    <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500 uppercase tracking-wider">Retries</th>
                                </tr>
                            </thead>
                            <tbody className="bg-white divide-y divide-slate-100">
                                {filtered.length === 0 ? (
                                    <tr>
                                        <td colSpan={5} className="px-6 py-12 text-center text-slate-400">
                                            <FileText className="h-8 w-8 mx-auto mb-2 opacity-40" />
                                            No jobs found.
                                        </td>
                                    </tr>
                                ) : (
                                    filtered.map((job) => (
                                        <tr key={job.id} className="hover:bg-slate-50 transition-colors">
                                            <td className="px-6 py-4 whitespace-nowrap text-sm font-mono text-slate-700">
                                                {job.id.substring(0, 8)}...
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
                </div>
            </div>
        </MainLayout>
    );
};

export default AdminJobs;
