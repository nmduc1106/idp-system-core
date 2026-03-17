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
    isExporting: boolean;
    onExport: (code: string) => void;
}

const JobTable: React.FC<JobTableProps> = ({
    jobs, onJobUpdate, loading,
    page, totalPages, total,
    searchCode, onSearch, onPageChange,
    isExporting, onExport,
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
            case 'PROCESSING':
            case 'PENDING':
                const isPending = state === 'PENDING';
                return (
                    <div className="flex flex-col gap-1.5 w-[110px]">
                        <span className={`inline-flex items-center text-[10px] font-bold uppercase tracking-widest ${isPending ? 'text-slate-400' : 'text-primary'}`}>
                            <Loader2 className={`w-3 h-3 mr-1 animate-spin ${isPending ? 'opacity-50' : ''}`} />
                            {isPending ? 'Queued' : 'Processing...'}
                        </span>
                        <div className="h-1.5 w-full bg-slate-100 rounded-full overflow-hidden relative border border-slate-200/50">
                            {isPending ? (
                                <div className="absolute top-0 bottom-0 left-0 w-1/4 bg-slate-200 rounded-full"></div>
                            ) : (
                                <div className="absolute top-0 bottom-0 left-0 w-1/2 bg-primary rounded-full animate-indeterminate"></div>
                            )}
                        </div>
                    </div>
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
                <form onSubmit={handleSearchSubmit} className="relative w-full sm:w-auto flex gap-2">
                    <div className="relative flex-1 sm:w-64">
                        <Search className="absolute left-3 top-2.5 h-4 w-4 text-slate-400" />
                        <input
                            type="text"
                            value={searchInput}
                            onChange={(e) => setSearchInput(e.target.value)}
                            placeholder="Search by File Code..."
                            className="w-full pl-10 pr-4 py-2 text-sm border border-slate-200 rounded-lg bg-white focus:ring-2 focus:ring-primary focus:border-primary outline-none"
                        />
                    </div>
                    <button
                        type="button"
                        onClick={() => onExport(searchInput.trim())}
                        disabled={isExporting}
                        className="inline-flex items-center justify-center px-4 py-2 text-sm font-medium text-white bg-emerald-600 border border-transparent rounded-lg hover:bg-emerald-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-emerald-500 disabled:opacity-50 disabled:cursor-not-allowed whitespace-nowrap shadow-sm transition-colors"
                    >
                        {isExporting ? <Loader2 className="w-4 h-4 mr-2 animate-spin" /> : <span className="mr-2 tracking-tighter">📥</span>}
                        Xuất Excel
                    </button>
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
            {selectedResult && (() => {
                let data: any = {};
                try {
                    const resultObj = typeof selectedResult === 'string' ? JSON.parse(selectedResult) : selectedResult;
                    data = resultObj.extracted_data || resultObj;
                } catch (e) {
                    data = selectedResult;
                }

                return (
                    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-sm">
                        <div className="bg-white rounded-xl shadow-2xl max-w-lg w-full max-h-[90vh] flex flex-col border border-slate-200 relative overflow-hidden">

                            {/* Jagged Receipt Top Edge */}
                            <div className="absolute top-0 left-0 w-full h-3 flex text-white fill-current">
                                {[...Array(24)].map((_, i) => (
                                    <svg key={i} className="h-3 w-auto flex-1" viewBox="0 0 10 10" preserveAspectRatio="none">
                                        <polygon points="0,0 10,0 5,10" />
                                    </svg>
                                ))}
                            </div>

                            <div className="px-6 pt-8 pb-4 border-b border-dashed border-slate-300 flex justify-between items-start bg-slate-50">
                                <div>
                                    <h3 className="text-xl font-bold text-slate-900 uppercase tracking-wider">{data.vendor_name || 'UNKNOWN VENDOR'}</h3>
                                    {data.vendor_address && <p className="text-xs text-slate-500 mt-1 max-w-[250px]">{data.vendor_address}</p>}
                                </div>
                                <button onClick={() => setSelectedResult(null)} className="text-slate-400 hover:text-slate-600 transition-colors p-1 bg-white rounded-full shadow-sm border border-slate-200">
                                    <XCircle className="w-5 h-5" />
                                </button>
                            </div>

                            <div className="p-8 overflow-y-auto w-full bg-slate-50 font-mono text-sm custom-scrollbar">
                                <div className="flex justify-between border-b border-dashed border-slate-300 pb-4 mb-5 text-slate-600">
                                    <div className="space-y-1">
                                        <p><span className="text-slate-400">TAX ID:</span> <span className="text-slate-800 font-medium">{data.tax_id || 'N/A'}</span></p>
                                        <p><span className="text-slate-400">DATE:</span> <span className="text-slate-800 font-medium">{data.date || 'N/A'}</span></p>
                                    </div>
                                    <div className="text-right space-y-1">
                                        <p><span className="text-slate-400">INV #:</span> <span className="text-slate-800 font-medium">{data.invoice_number || 'N/A'}</span></p>
                                    </div>
                                </div>

                                <div className="border-y-2 border-solid border-slate-800 py-4 mb-6 mt-4">
                                    <div className="flex justify-between items-center text-xl font-bold text-slate-900">
                                        <span className="uppercase tracking-wider text-sm text-slate-500">Tổng thanh toán hóa đơn</span>
                                        <span>{data.total_amount || '-'}</span>
                                    </div>
                                </div>

                                <div className="mt-8 text-center text-slate-400 text-xs flex flex-col items-center">
                                    <div className="w-full max-w-[200px] h-10 border-y-2 border-slate-800 flex flex-col justify-center mb-3 px-1 gap-0.5">
                                        {/* Pure CSS Barcode Header */}
                                        <div className="flex w-full h-8 justify-between opacity-80 items-end">
                                            {[...Array(24)].map((_, i) => (
                                                <div key={i} className={`bg-slate-800 ${i % 7 === 0 ? 'w-1.5 h-full' : i % 3 === 0 ? 'w-[3px] h-5' : i % 2 === 0 ? 'w-1 h-6' : 'w-[2px] h-full'}`}></div>
                                            ))}
                                        </div>
                                    </div>
                                    <p className="tracking-widest">END OF RECEIPT</p>
                                </div>
                            </div>

                            {/* Jagged Receipt Bottom Edge */}
                            <div className="absolute bottom-0 left-0 w-full h-3 flex text-white fill-current rotate-180">
                                {[...Array(24)].map((_, i) => (
                                    <svg key={i} className="h-3 w-auto flex-1" viewBox="0 0 10 10" preserveAspectRatio="none">
                                        <polygon points="0,0 10,0 5,10" />
                                    </svg>
                                ))}
                            </div>
                        </div>
                    </div>
                );
            })()}
        </div>
    );
};

export default JobTable;
