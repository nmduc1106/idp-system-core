import React, { useState, useCallback } from 'react';
import MainLayout from '../../components/layout/MainLayout';
import UploadZone from './UploadZone';
import JobTable from './JobTable';
import { Job } from './docService';

const Dashboard: React.FC = () => {
    const [jobs, setJobs] = useState<Job[]>([]);

    const handleNewJob = useCallback((job: Job) => {
        // Add job to the top of the list
        setJobs((prevJobs) => [job, ...prevJobs]);
    }, []);

    const handleJobUpdate = useCallback((updatedJob: Job) => {
        setJobs((prevJobs) =>
            prevJobs.map(job => (job.id === updatedJob.id ? updatedJob : job))
        );
    }, []);

    return (
        <MainLayout>
            <div className="space-y-8 animate-in fade-in duration-500">
                <header>
                    <h1 className="text-2xl font-bold text-slate-900 dark:text-white shadow-sm:hidden">
                        Document Extraction
                    </h1>
                    <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">
                        Upload your documents here to extract structured data automatically.
                    </p>
                </header>

                <section className="max-w-3xl">
                    <UploadZone onUploadSuccess={handleNewJob} />
                </section>

                <section>
                    <div className="flex items-center justify-between mb-4">
                        <h2 className="text-lg font-semibold text-slate-900 dark:text-white flex items-center">
                            Processing History
                        </h2>
                        <div className="text-sm text-slate-500">
                            {jobs.length} document{jobs.length !== 1 ? 's' : ''}
                        </div>
                    </div>
                    <JobTable jobs={jobs} onJobUpdate={handleJobUpdate} />
                </section>
            </div>
        </MainLayout>
    );
};

export default Dashboard;
