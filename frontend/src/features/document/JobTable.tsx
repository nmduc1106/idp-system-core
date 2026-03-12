import React, { useEffect, useState, useRef } from 'react';
import { Job, streamJobStatus } from './docService';
import { FileText, Loader2, CheckCircle, XCircle, Eye, Search, ChevronLeft, ChevronRight } from 'lucide-react';

interface JobTableProps {
    jobs: Job[];
    onJobUpdate: (updatedJob: Job) => void;
    loading: boolean;
    page: number;
    totalPages: number;
    total: number;
    searchCode: string;
    onSearch: (code: string) => void;
    onPageChange: (page: number) => void;
}

const JobTable: React.FC<JobTableProps> = ({
    jobs, onJobUpdate, loading,
    page, totalPages, total,
    searchCode, onSearch, onPageChange,
}) => {
    const [selectedResult, setSelectedResult] = useState<any | null>(null);
    const [searchInput, setSearchInput] = useState(searchCode);
    const streamsRef = useRef<{ [key: string]: EventSource }>({});

    // Persist latest onJobUpdate in a ref to avoid stale closures inside the SSE callback
    // without triggering the useEffect dependency array
    const onJobUpdateRef = useRef(onJobUpdate);
    useEffect(() => {
        onJobUpdateRef.current = onJobUpdate;
    }, [onJobUpdate]);

    useEffect(() => {
        const activeStreams = streamsRef.current;

        jobs.forEach((job) => {
            // Safely check state and status to handle any DTO mapping mismatches
            const currentState = (job.state || (job as any).status) as string;
            const isPending = currentState === 'PENDING' || currentState === 'EXTRACTING' || currentState === 'PROCESSING';

            // Only open a stream if the job is active AND we don't already have one
            if (isPending && !activeStreams[job.id]) {
                console.log(`[SSE] 🟢 Initializing stream in JobTable for Job ID: ${job.id}`);
                const source = streamJobStatus(
                    job.id,
                    (updatedData) => {
                        console.log(`[UI] ⬆️ Calling onJobUpdate for Job: ${job.id} with status: ${updatedData.state}`);
                        // Use the ref to always get the latest callback
                        onJobUpdateRef.current(updatedData);
                        // Close stream if terminal
                        if (updatedData.state === 'COMPLETED' || updatedData.state === 'FAILED') {
                            source.close();
                            delete activeStreams[job.id];
                        }
                    },
                    (err) => {
                        console.error(`Streaming error for job ${job.id}`, err);
                        source.close();
                        delete activeStreams[job.id];
                    }
                );
                activeStreams[job.id] = source;
            }
        });

        // Cleanup: We do NOT aggressively loop and close streams here.
        // If we did, every time `jobs` updates, we would tear down all active connections.
        // We rely on the `onmessage` terminal state (COMPLETED/FAILED) to close them,
        // OR a true unmount (which we can't perfectly distinguish here without a separate effect, 
        // but since they self-close on completion, it's safer to leave them running than to flap them).
    }, [jobs]); // <--- CRITICAL: jobs MUST be in the dependency array to react to new jobs!

    // Dedicated unmount cleanup to prevent memory leaks if user navigates away
    useEffect(() => {
        return () => {
            console.log('[SSE] 🧹 JobTable unmounting, closing all active streams...');
            Object.values(streamsRef.current).forEach((s) => s.close());
            streamsRef.current = {};
        };
    }, []);

    const handleSearchSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        onSearch(searchInput.trim());
    };

    const getStateBadge = (state: string) => {
        switch (state) {
            case 'COMPLETED':
                return (
                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-emerald-100 text-emerald-700 border border-emerald-200">
                        <CheckCircle className="w-3 h-3 mr-1" /> Completed
                    </span>
                );
            case 'FAILED':
                return (
                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-rose-100 text-rose-700 border border-rose-200">
                        <XCircle className="w-3 h-3 mr-1" /> Failed
                    </span>
                );
            case 'EXTRACTING':
            case 'PENDING':
                return (
                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-700 border border-blue-200">
                        <Loader2 className="w-3 h-3 mr-1 animate-spin" /> {state === 'PENDING' ? 'Pending' : 'Extracting...'}
                    </span>
                );
            default:
                return <span className="text-xs text-slate-500">{state}</span>;
        }
    };

    return (
        <div className="space-y-4">
            {/* Header Bar: Title + Search */}
            <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
                <div>
                    <h2 className="text-lg font-semibold text-slate-900">Processing History</h2>
                    <p className="text-sm text-slate-500">{total} document{total !== 1 ? 's' : ''} total</p>
                </div>
                <form onSubmit={handleSearchSubmit} className="relative w-full sm:w-64">
                    <Search className="absolute left-3 top-2.5 h-4 w-4 text-slate-400" />
                    <input
                        type="text"
                        value={searchInput}
                        onChange={(e) => setSearchInput(e.target.value)}
                        placeholder="Search by File Code..."
                        className="w-full pl-10 pr-4 py-2 text-sm border border-slate-200 rounded-lg bg-white focus:ring-2 focus:ring-primary focus:border-primary outline-none"
                    />
                </form>
            </div>

            {/* Table */}
            <div className="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
                <div className="overflow-x-auto">
                    <table className="min-w-full divide-y divide-slate-200">
                        <thead className="bg-slate-50">
                            <tr>
                                <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500 uppercase tracking-wider">Document</th>
                                <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500 uppercase tracking-wider">File Code</th>
                                <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500 uppercase tracking-wider">Created</th>
                                <th className="px-6 py-3 text-left text-xs font-semibold text-slate-500 uppercase tracking-wider">Status</th>
                                <th className="px-6 py-3 text-right text-xs font-semibold text-slate-500 uppercase tracking-wider">Actions</th>
                            </tr>
                        </thead>
                        <tbody className="bg-white divide-y divide-slate-100">
                            {loading && jobs.length === 0 ? (
                                <tr>
                                    <td colSpan={5} className="px-6 py-12 text-center">
                                        <Loader2 className="h-6 w-6 animate-spin text-primary mx-auto" />
                                    </td>
                                </tr>
                            ) : jobs.length === 0 ? (
                                <tr>
                                    <td colSpan={5} className="px-6 py-12 text-center text-slate-400">
                                        <FileText className="h-8 w-8 mx-auto mb-2 opacity-40" />
                                        No documents found.
                                    </td>
                                </tr>
                            ) : (
                                jobs.map((job) => (
                                    <tr key={job.id} className="hover:bg-slate-50 transition-colors">
                                        <td className="px-6 py-4 whitespace-nowrap">
                                            <div className="flex items-center gap-3">
                                                <div className="flex-shrink-0 h-8 w-8 bg-primary/10 rounded flex items-center justify-center">
                                                    <FileText className="h-4 w-4 text-primary" />
                                                </div>
                                                <div>
                                                    <div className="text-sm font-medium text-slate-900">
                                                        {job.document?.file_name || job.document?.original_filename || `Doc: ${job.document_id?.substring(0, 8) || '—'}...`}
                                                    </div>
                                                    <div className="text-xs text-slate-500">ID: {job.id?.substring(0, 8) || '—'}...</div>
                                                </div>
                                            </div>
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
                                        <td className="px-6 py-4 whitespace-nowrap text-sm text-slate-500">
                                            {job.created_at ? new Date(job.created_at).toLocaleString() : '—'}
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap">{getStateBadge(job.state)}</td>
                                        <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                                            {job.state === 'COMPLETED' && job.result && (
                                                <button
                                                    onClick={() => setSelectedResult(job.result)}
                                                    className="inline-flex items-center text-primary hover:text-primary/80"
                                                >
                                                    <Eye className="w-4 h-4 mr-1" /> View
                                                </button>
                                            )}
                                        </td>
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
                            Page {page} of {totalPages}
                        </p>
                        <div className="flex items-center gap-2">
                            <button
                                onClick={() => onPageChange(page - 1)}
                                disabled={page <= 1}
                                className="inline-flex items-center px-3 py-1.5 text-sm border border-slate-200 rounded-lg bg-white hover:bg-slate-50 disabled:opacity-40 disabled:cursor-not-allowed transition"
                            >
                                <ChevronLeft className="h-4 w-4 mr-1" /> Prev
                            </button>
                            <button
                                onClick={() => onPageChange(page + 1)}
                                disabled={page >= totalPages}
                                className="inline-flex items-center px-3 py-1.5 text-sm border border-slate-200 rounded-lg bg-white hover:bg-slate-50 disabled:opacity-40 disabled:cursor-not-allowed transition"
                            >
                                Next <ChevronRight className="h-4 w-4 ml-1" />
                            </button>
                        </div>
                    </div>
                )}
            </div>

            {/* Result Modal */}
            {selectedResult && (
                <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/50 backdrop-blur-sm">
                    <div className="bg-white rounded-xl shadow-2xl max-w-2xl w-full max-h-[80vh] flex flex-col border border-slate-200">
                        <div className="px-6 py-4 border-b border-slate-200 flex justify-between items-center bg-slate-50 rounded-t-xl">
                            <h3 className="text-lg font-bold text-slate-900">Extraction Result</h3>
                            <button onClick={() => setSelectedResult(null)} className="text-slate-400 hover:text-slate-500">
                                <XCircle className="w-6 h-6" />
                            </button>
                        </div>
                        <div className="p-6 overflow-y-auto w-full">
                            <pre className="bg-slate-900 text-slate-300 p-4 rounded-lg overflow-x-auto text-sm">
                                {JSON.stringify(selectedResult, null, 2)}
                            </pre>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default JobTable;
