import React, { useState, useCallback, useEffect } from 'react';
import MainLayout from '../../components/layout/MainLayout';
import UploadZone from './UploadZone';
import JobTable from './JobTable';
import { Job, getUserJobs, PaginatedResponse } from './docService';

const Dashboard: React.FC = () => {
    const [pageData, setPageData] = useState<PaginatedResponse<Job> | null>(null);
    const [page, setPage] = useState(1);
    const [searchCode, setSearchCode] = useState('');
    const [loading, setLoading] = useState(true);

    const fetchJobs = useCallback(async () => {
        console.log(`[REFETCH] 🔄 Fetching jobs... Page: ${page} | SearchCode: ${searchCode}`);
        setLoading(true);
        try {
            const data = await getUserJobs({ page, limit: 10, file_code: searchCode || undefined });
            console.log(`[REFETCH] ✅ Successfully fetched jobs. Count: ${data.data?.length || 0}`);
            setPageData(data);
        } catch (err) {
            console.error('[REFETCH] ❌ Failed to fetch jobs:', err);
        } finally {
            setLoading(false);
        }
    }, [page, searchCode]);

    useEffect(() => {
        fetchJobs();
    }, [fetchJobs]);

    // After upload success: reset to page 1 and refetch from DB
    // This ensures the table gets the fully joined data (document metadata, timestamps)
    const handleUploadSuccess = useCallback(() => {
        console.log('[UPLOAD] 🏁 Upload success callback triggered in Dashboard. Resetting to page 1...');
        setPage(1);
        // Small delay to let DB commit propagate
        setTimeout(() => fetchJobs(), 300);
    }, [fetchJobs]);

    const handleJobUpdate = useCallback((updatedJob: Job) => {
        console.log(`[DASHBOARD] 🔄 handleJobUpdate triggered for Job: ${updatedJob.id} | New Status: ${updatedJob.state}`);
        setPageData((prev) => {
            if (!prev) return prev;
            return {
                ...prev,
                data: prev.data.map((j) => {
                    if (j.id === updatedJob.id) {
                        console.log(`[DASHBOARD] 🔄 Merging updated data into existing Job ${j.id}...`);
                        return { ...j, state: updatedJob.state, result: updatedJob.result, error: updatedJob.error };
                    }
                    return j;
                }),
            };
        });
    }, []);

    const handleSearch = (code: string) => {
        setSearchCode(code);
        setPage(1);
    };

    const jobs = pageData?.data || [];

    return (
        <MainLayout>
            <div className="space-y-8">
                <header>
                    <h1 className="text-2xl font-bold text-slate-900">Document Extraction</h1>
                    <p className="mt-1 text-sm text-slate-500">
                        Upload your documents here to extract structured data automatically.
                    </p>
                </header>

                <section className="max-w-3xl">
                    <UploadZone onUploadSuccess={handleUploadSuccess} />
                </section>

                <section>
                    <JobTable
                        jobs={jobs}
                        onJobUpdate={handleJobUpdate}
                        loading={loading}
                        page={page}
                        totalPages={pageData?.total_pages || 1}
                        total={pageData?.total || 0}
                        searchCode={searchCode}
                        onSearch={handleSearch}
                        onPageChange={setPage}
                    />
                </section>
            </div>
        </MainLayout>
    );
};

export default Dashboard;
