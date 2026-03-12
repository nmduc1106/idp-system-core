import React, { useState, useEffect } from 'react';
import MainLayout from '../../components/layout/MainLayout';
import { getSystemStats, SystemStats } from './adminService';
import { Users, Briefcase, CheckCircle2, XCircle, Clock, Activity, Loader2 } from 'lucide-react';

const AdminDashboard: React.FC = () => {
    const [stats, setStats] = useState<SystemStats | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');

    useEffect(() => {
        const fetchStats = async () => {
            try {
                const data = await getSystemStats();
                setStats(data);
            } catch (err: any) {
                setError(err.response?.data?.error || 'Failed to load statistics');
            } finally {
                setLoading(false);
            }
        };
        fetchStats();
    }, []);

    if (loading) {
        return (
            <MainLayout>
                <div className="flex items-center justify-center h-64">
                    <Loader2 className="h-8 w-8 animate-spin text-primary" />
                </div>
            </MainLayout>
        );
    }

    if (error) {
        return (
            <MainLayout>
                <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg">{error}</div>
            </MainLayout>
        );
    }

    const cards = [
        {
            title: 'Total Users',
            value: stats?.total_users ?? 0,
            icon: Users,
            gradient: 'from-blue-500 to-cyan-400',
            bg: 'bg-blue-50',
            text: 'text-blue-600',
        },
        {
            title: 'Total Jobs',
            value: stats?.total_jobs ?? 0,
            icon: Briefcase,
            gradient: 'from-violet-500 to-purple-400',
            bg: 'bg-violet-50',
            text: 'text-violet-600',
        },
        {
            title: 'Completed',
            value: stats?.jobs_by_state?.['COMPLETED'] ?? 0,
            icon: CheckCircle2,
            gradient: 'from-emerald-500 to-teal-400',
            bg: 'bg-emerald-50',
            text: 'text-emerald-600',
        },
        {
            title: 'Failed',
            value: stats?.jobs_by_state?.['FAILED'] ?? 0,
            icon: XCircle,
            gradient: 'from-rose-500 to-pink-400',
            bg: 'bg-rose-50',
            text: 'text-rose-600',
        },
        {
            title: 'Pending',
            value: stats?.jobs_by_state?.['PENDING'] ?? 0,
            icon: Clock,
            gradient: 'from-amber-500 to-orange-400',
            bg: 'bg-amber-50',
            text: 'text-amber-600',
        },
        {
            title: 'Extracting',
            value: stats?.jobs_by_state?.['EXTRACTING'] ?? 0,
            icon: Activity,
            gradient: 'from-sky-500 to-indigo-400',
            bg: 'bg-sky-50',
            text: 'text-sky-600',
        },
    ];

    return (
        <MainLayout>
            <div className="space-y-8">
                <header>
                    <h1 className="text-2xl font-bold text-slate-900">Admin Dashboard</h1>
                    <p className="mt-1 text-sm text-slate-500">System-wide statistics and performance overview.</p>
                </header>

                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
                    {cards.map((card) => (
                        <div
                            key={card.title}
                            className="relative bg-white rounded-xl border border-slate-200 p-6 shadow-sm hover:shadow-md transition-shadow overflow-hidden group"
                        >
                            {/* Gradient accent bar */}
                            <div className={`absolute top-0 left-0 right-0 h-1 bg-gradient-to-r ${card.gradient}`}></div>

                            <div className="flex items-center justify-between">
                                <div>
                                    <p className="text-sm font-medium text-slate-500">{card.title}</p>
                                    <p className="mt-2 text-3xl font-extrabold text-slate-900">{card.value.toLocaleString()}</p>
                                </div>
                                <div className={`${card.bg} p-3 rounded-xl group-hover:scale-110 transition-transform`}>
                                    <card.icon className={`h-6 w-6 ${card.text}`} />
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            </div>
        </MainLayout>
    );
};

export default AdminDashboard;
