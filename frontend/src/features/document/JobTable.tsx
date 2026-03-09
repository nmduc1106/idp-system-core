import React, { useEffect, useState } from 'react';
import { Job, streamJobStatus } from './docService';
import { FileText, Loader2, CheckCircle, XCircle, Eye } from 'lucide-react';

interface JobTableProps {
    jobs: Job[];
    onJobUpdate: (updatedJob: Job) => void;
}

const JobTable: React.FC<JobTableProps> = ({ jobs, onJobUpdate }) => {
    const [selectedResult, setSelectedResult] = useState<any | null>(null);

    useEffect(() => {
        // Whenever jobs change, we ensure we are streaming any job that is not completed
        const activeStreams: { [key: string]: EventSource } = {};

        jobs.forEach(job => {
            if (job.state === 'PENDING' || job.state === 'EXTRACTING') {
                const source = streamJobStatus(
                    job.id,
                    (updatedData) => {
                        onJobUpdate(updatedData);
                    },
                    (err) => {
                        console.error(`Streaming error for job ${job.id}`, err);
                    }
                );
                activeStreams[job.id] = source;
            }
        });

        return () => {
            // Cleanup streams when component unmounts or jobs array changes deeply
            Object.values(activeStreams).forEach(source => source.close());
        };
    }, [jobs, onJobUpdate]);

    const getStateBadge = (state: string) => {
        switch (state) {
            case 'COMPLETED':
                return (
                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300 border border-green-200 dark:border-green-800">
                        <CheckCircle className="w-3 h-3 mr-1" /> Completed
                    </span>
                );
            case 'FAILED':
                return (
                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-300 border border-red-200 dark:border-red-800">
                        <XCircle className="w-3 h-3 mr-1" /> Failed
                    </span>
                );
            case 'EXTRACTING':
            case 'PENDING':
                return (
                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-300 border border-blue-200 dark:border-blue-800">
                        <Loader2 className="w-3 h-3 mr-1 animate-spin" /> {state === 'PENDING' ? 'Pending' : 'Extracting...'}
                    </span>
                );
            default:
                return <span className="text-xs text-slate-500">{state}</span>;
        }
    };

    return (
        <div className="bg-white dark:bg-slate-800 rounded-xl shadow-sm border border-slate-200 dark:border-slate-700 overflow-hidden">
            <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-slate-200 dark:divide-slate-700">
                    <thead className="bg-slate-50 dark:bg-slate-900/50">
                        <tr>
                            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                                Document
                            </th>
                            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                                Date Added
                            </th>
                            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                                Status
                            </th>
                            <th scope="col" className="px-6 py-3 text-right text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                                Actions
                            </th>
                        </tr>
                    </thead>
                    <tbody className="bg-white dark:bg-slate-800 divide-y divide-slate-200 dark:divide-slate-700">
                        {jobs.length === 0 ? (
                            <tr>
                                <td colSpan={4} className="px-6 py-8 text-center text-sm text-slate-500 dark:text-slate-400">
                                    No documents uploaded yet.
                                </td>
                            </tr>
                        ) : (
                            jobs.map((job) => (
                                <tr key={job.id} className="hover:bg-slate-50 dark:hover:bg-slate-700/50 transition-colors">
                                    <td className="px-6 py-4 whitespace-nowrap">
                                        <div className="flex items-center">
                                            <div className="flex-shrink-0 h-8 w-8 bg-indigo-100 dark:bg-indigo-900/30 rounded flex items-center justify-center">
                                                <FileText className="h-4 w-4 text-indigo-600 dark:text-indigo-400" />
                                            </div>
                                            <div className="ml-4">
                                                <div className="text-sm font-medium text-slate-900 dark:text-white">
                                                    Doc: {job.document_id.substring(0, 8)}...
                                                </div>
                                                <div className="text-xs text-slate-500 dark:text-slate-400">
                                                    ID: {job.id.substring(0, 8)}...
                                                </div>
                                            </div>
                                        </div>
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap text-sm text-slate-500 dark:text-slate-400">
                                        {new Date(job.created_at).toLocaleString()}
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap">
                                        {getStateBadge(job.state)}
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                                        {job.state === 'COMPLETED' && job.result && (
                                            <button
                                                onClick={() => setSelectedResult(job.result)}
                                                className="inline-flex items-center text-indigo-600 hover:text-indigo-900 dark:text-indigo-400 dark:hover:text-indigo-300"
                                            >
                                                <Eye className="w-4 h-4 mr-1" /> View JSON
                                            </button>
                                        )}
                                    </td>
                                </tr>
                            ))
                        )}
                    </tbody>
                </table>
            </div>

            {/* Basic Result Modal */}
            {selectedResult && (
                <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/50 backdrop-blur-sm">
                    <div className="bg-white dark:bg-slate-800 rounded-xl shadow-2xl max-w-2xl w-full max-h-[80vh] flex flex-col border border-slate-200 dark:border-slate-700">
                        <div className="px-6 py-4 border-b border-slate-200 dark:border-slate-700 flex justify-between items-center bg-slate-50 dark:bg-slate-900/50 rounded-t-xl">
                            <h3 className="text-lg font-bold text-slate-900 dark:text-white">Extraction Result</h3>
                            <button
                                onClick={() => setSelectedResult(null)}
                                className="text-slate-400 hover:text-slate-500 dark:hover:text-slate-300"
                            >
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
