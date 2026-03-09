import React, { useContext } from 'react';
import { AuthContext, AuthContextType } from '../../contexts/AuthContext';
import { useNavigate } from 'react-router-dom';
import { LogOut, FileText } from 'lucide-react';

interface MainLayoutProps {
    children: React.ReactNode;
}

const MainLayout: React.FC<MainLayoutProps> = ({ children }) => {
    const authContext = useContext<AuthContextType | null>(AuthContext);
    const navigate = useNavigate();

    const handleLogout = () => {
        if (authContext) {
            authContext.logout();
            navigate('/login');
        }
    };

    const userName = authContext?.user?.full_name || authContext?.user?.email || 'User';

    return (
        <div className="min-h-screen bg-slate-50 dark:bg-slate-900 transition-colors duration-200 flex flex-col">
            <nav className="bg-white dark:bg-slate-800 border-b border-slate-200 dark:border-slate-700 shadow-sm z-10">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                    <div className="flex justify-between h-16">
                        <div className="flex items-center space-x-3">
                            <div className="w-10 h-10 bg-indigo-600 rounded-lg flex items-center justify-center shadow-inner">
                                <FileText className="h-6 w-6 text-white" />
                            </div>
                            <span className="font-bold text-xl tracking-tight text-slate-900 dark:text-white">IDP Dashboard</span>
                        </div>

                        <div className="flex items-center space-x-4">
                            <span className="text-sm font-medium text-slate-600 dark:text-slate-300">
                                Hi, {userName}
                            </span>
                            <button
                                onClick={handleLogout}
                                className="inline-flex items-center px-3 py-2 border border-transparent text-sm font-medium rounded-md text-slate-700 dark:text-slate-200 bg-slate-100 dark:bg-slate-700 hover:bg-slate-200 dark:hover:bg-slate-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 transition-colors"
                                title="Logout"
                            >
                                <LogOut className="h-4 w-4 mr-2" />
                                Logout
                            </button>
                        </div>
                    </div>
                </div>
            </nav>

            <main className="flex-1 max-w-7xl w-full mx-auto p-4 sm:p-6 lg:p-8">
                {children}
            </main>
        </div>
    );
};

export default MainLayout;
