import React, { useContext } from 'react';
import { AuthContext, AuthContextType } from '../../contexts/AuthContext';
import { useNavigate, useLocation, Link } from 'react-router-dom';
import { LogOut, FileText, LayoutDashboard, Briefcase, Users, ShieldCheck } from 'lucide-react';

interface MainLayoutProps {
    children: React.ReactNode;
}

const MainLayout: React.FC<MainLayoutProps> = ({ children }) => {
    const authContext = useContext<AuthContextType | null>(AuthContext);
    const navigate = useNavigate();
    const location = useLocation();

    const handleLogout = async () => {
        if (authContext) {
            await authContext.logout();
            navigate('/login');
        }
    };

    const userName = authContext?.user?.full_name || authContext?.user?.email || 'User';
    const isAdmin = authContext?.user?.role === 'ADMIN';

    const navLink = (path: string, label: string, Icon: React.FC<{ className?: string }>) => {
        const active = location.pathname === path;
        return (
            <Link
                to={path}
                className={`inline-flex items-center px-3 py-2 text-sm font-medium rounded-lg transition-colors ${
                    active
                        ? 'bg-primary/10 text-primary'
                        : 'text-slate-600 hover:bg-slate-100 hover:text-slate-900'
                }`}
            >
                <Icon className="h-4 w-4 mr-2" />
                {label}
            </Link>
        );
    };

    return (
        <div className="min-h-screen bg-slate-50 transition-colors duration-200 flex flex-col">
            <nav className="bg-white border-b border-slate-200 shadow-sm z-10">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                    <div className="flex justify-between h-16">
                        <div className="flex items-center space-x-3">
                            <Link to="/" className="flex items-center space-x-3">
                                <div className="w-10 h-10 bg-primary rounded-lg flex items-center justify-center shadow-inner">
                                    <FileText className="h-6 w-6 text-white" />
                                </div>
                                <span className="font-bold text-xl tracking-tight text-slate-900">IDP Dashboard</span>
                            </Link>
                        </div>

                        <div className="flex items-center space-x-4">
                            <span className="text-sm font-medium text-slate-600">
                                Hi, {userName}
                            </span>
                            {isAdmin && (
                                <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold bg-violet-100 text-violet-700 border border-violet-200">
                                    <ShieldCheck className="h-3 w-3 mr-1" />
                                    ADMIN
                                </span>
                            )}
                            <button
                                onClick={handleLogout}
                                className="inline-flex items-center px-3 py-2 border border-transparent text-sm font-medium rounded-md text-slate-700 bg-slate-100 hover:bg-slate-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary transition-colors"
                                title="Logout"
                            >
                                <LogOut className="h-4 w-4 mr-2" />
                                Logout
                            </button>
                        </div>
                    </div>
                </div>

                {/* Sub-navigation */}
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                    <div className="flex items-center space-x-1 pb-3 overflow-x-auto">
                        {navLink('/', 'Documents', FileText)}

                        {isAdmin && (
                            <>
                                <div className="w-px h-5 bg-slate-200 mx-2"></div>
                                {navLink('/admin/dashboard', 'Admin Panel', LayoutDashboard)}
                                {navLink('/admin/jobs', 'All Jobs', Briefcase)}
                                {navLink('/admin/users', 'All Users', Users)}
                            </>
                        )}
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
