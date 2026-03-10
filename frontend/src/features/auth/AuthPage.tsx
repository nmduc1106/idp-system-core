import React, { useState, useContext, FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import {
    Mail,
    Lock,
    Eye,
    EyeOff,
    Loader2,
    LayoutDashboard,
    Facebook,
} from 'lucide-react';
import { AuthContext, AuthContextType } from '../../contexts/AuthContext';
import * as authService from './authService';

const AuthPage: React.FC = () => {
    const [isLogin, setIsLogin] = useState<boolean>(true);
    const [showPassword, setShowPassword] = useState<boolean>(false);
    const [email, setEmail] = useState<string>('');
    const [password, setPassword] = useState<string>('');
    const [fullName, setFullName] = useState<string>('');
    const [isSubmitting, setIsSubmitting] = useState<boolean>(false);
    const [error, setError] = useState<string>('');
    const [successMsg, setSuccessMsg] = useState<string>('');

    const authContext = useContext<AuthContextType | null>(AuthContext);
    const navigate = useNavigate();

    const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        setError('');
        setSuccessMsg('');
        setIsSubmitting(true);

        const cleanEmail = email.trim();
        const cleanPassword = password.trim();

        try {
            if (isLogin) {
                const data = await authService.login(cleanEmail, cleanPassword);
                if (authContext) {
                    authContext.login(data.user || { email: cleanEmail });
                }
                navigate('/');
            } else {
                await authService.register(cleanEmail, cleanPassword, fullName.trim());
                setSuccessMsg('Đăng ký thành công! Vui lòng đăng nhập.');
                setIsLogin(true);
            }
        } catch (err: any) {
            setError(err.response?.data?.message || err.message || 'Xác thực thất bại');
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <div className="bg-gray-50 font-sans min-h-screen flex transition-colors duration-200">
            {/* Cột bên trái: Form */}
            <div className="w-full lg:w-1/2 flex flex-col justify-center px-8 md:px-24 lg:px-32 bg-white">
                {/* Logo Section */}
                <div className="flex items-center gap-2 mb-12">
                    <div className="w-10 h-10 bg-blue-600 rounded-lg flex items-center justify-center text-white shadow-lg shadow-blue-200">
                        <LayoutDashboard size={20} />
                    </div>
                    <span className="text-2xl font-bold text-slate-800 tracking-tight">
                        IDP <span className="text-blue-600">Core</span>
                    </span>
                </div>

                <div className="mb-8">
                    <h2 className="text-3xl font-bold text-slate-900 mb-2">
                        {isLogin ? 'Sign In' : 'Create Account'}
                    </h2>
                    <p className="text-slate-500">
                        {isLogin ? 'Chào mừng trở lại! Vui lòng nhập thông tin của bạn' : 'Bắt đầu tối ưu hóa quy trình tài liệu của bạn ngay hôm nay'}
                    </p>
                </div>

                {error && <div className="mb-4 p-3 bg-red-50 text-red-600 text-sm rounded-xl border border-red-100">{error}</div>}
                {successMsg && <div className="mb-4 p-3 bg-green-50 text-green-600 text-sm rounded-xl border border-green-100">{successMsg}</div>}

                <form className="space-y-5" onSubmit={handleSubmit}>
                    {!isLogin && (
                        <div>
                            <label className="block text-sm font-medium text-slate-700 mb-1">Full Name</label>
                            <input
                                type="text"
                                placeholder="Nhập họ và tên"
                                value={fullName}
                                onChange={(e) => setFullName(e.target.value)}
                                className="w-full px-4 py-3 border border-slate-200 rounded-xl focus:ring-2 focus:ring-blue-500 focus:outline-none transition"
                                required
                            />
                        </div>
                    )}

                    <div>
                        <label className="block text-sm font-medium text-slate-700 mb-1">Email</label>
                        <div className="relative">
                            <Mail className="absolute left-4 top-3.5 text-slate-400" size={18} />
                            <input
                                type="email"
                                placeholder="name@company.com"
                                value={email}
                                onChange={(e) => setEmail(e.target.value)}
                                className="w-full pl-11 pr-4 py-3 border border-slate-200 rounded-xl focus:ring-2 focus:ring-blue-500 focus:outline-none transition"
                                required
                            />
                        </div>
                    </div>

                    <div>
                        <div className="flex justify-between mb-1">
                            <label className="text-sm font-medium text-slate-700">Password</label>
                            {isLogin && <a href="#" className="text-sm font-semibold text-blue-600 hover:underline">Quên mật khẩu?</a>}
                        </div>
                        <div className="relative">
                            <Lock className="absolute left-4 top-3.5 text-slate-400" size={18} />
                            <input
                                type={showPassword ? "text" : "password"}
                                placeholder="••••••••"
                                value={password}
                                onChange={(e) => setPassword(e.target.value)}
                                className="w-full pl-11 pr-12 py-3 border border-slate-200 rounded-xl focus:ring-2 focus:ring-blue-500 focus:outline-none transition"
                                required
                            />
                            <button
                                type="button"
                                onClick={() => setShowPassword(!showPassword)}
                                className="absolute right-4 top-3.5 text-slate-400 hover:text-blue-600 transition-colors"
                            >
                                {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
                            </button>
                        </div>
                    </div>

                    <div className="flex items-center justify-between">
                        <label className="flex items-center gap-2 cursor-pointer">
                            <input type="checkbox" className="w-4 h-4 rounded text-blue-600 border-slate-300 focus:ring-blue-500" />
                            <span className="text-sm text-slate-600">Ghi nhớ đăng nhập</span>
                        </label>
                    </div>

                    <button
                        type="submit"
                        disabled={isSubmitting}
                        className="w-full bg-blue-600 hover:bg-blue-700 text-white font-semibold py-3 rounded-xl shadow-lg shadow-blue-200 transition flex items-center justify-center disabled:opacity-70"
                    >
                        {isSubmitting ? <Loader2 className="animate-spin mr-2" /> : (isLogin ? 'Sign in' : 'Đăng ký ngay')}
                    </button>

                    <div className="relative py-4">
                        <div className="absolute inset-0 flex items-center"><div className="w-full border-t border-slate-100"></div></div>
                        <div className="relative flex justify-center text-xs uppercase"><span className="bg-white px-2 text-slate-400 font-medium">Hoặc tiếp tục với</span></div>
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                        <button type="button" className="flex items-center justify-center gap-2 border border-slate-200 py-2.5 rounded-xl hover:bg-slate-50 transition font-medium text-sm">
                            <img src="https://www.svgrepo.com/show/355037/google.svg" className="w-5 h-5" alt="Google" /> Google
                        </button>
                        <button type="button" className="flex items-center justify-center gap-2 border border-slate-200 py-2.5 rounded-xl hover:bg-slate-50 transition font-medium text-sm text-slate-700">
                            <Facebook className="text-blue-600" size={18} /> Facebook
                        </button>
                    </div>
                </form>

                <p className="text-center mt-10 text-sm text-slate-500">
                    {isLogin ? "Chưa có tài khoản?" : "Đã có tài khoản?"} {' '}
                    <button
                        onClick={() => setIsLogin(!isLogin)}
                        className="font-bold text-slate-900 hover:underline"
                    >
                        {isLogin ? 'Đăng ký ngay' : 'Đăng nhập'}
                    </button>
                </p>
            </div>

            {/* Cột bên phải: Blue Content (Hidden on Mobile) */}
            <div className="hidden lg:flex lg:w-1/2 bg-blue-600 m-4 rounded-[40px] flex-col items-center justify-center text-white p-12 relative overflow-hidden">
                <div className="max-w-md text-center mb-12 relative z-10">
                    <h1 className="text-5xl font-bold mb-6 leading-tight">Chào mừng đến với <br /> IDP Core System</h1>
                    <p className="text-blue-100 opacity-90 leading-relaxed">
                        Hệ thống xử lý tài liệu thông minh giúp doanh nghiệp tự động hóa 90% quy trình nhập liệu thủ công với độ chính xác tuyệt đối.
                    </p>
                </div>

                {/* Biểu đồ minh họa (Chart UI) */}
                <div className="w-full max-w-lg bg-white rounded-3xl p-8 shadow-2xl relative z-10 border border-white/20">
                    <div className="flex justify-between items-end h-48 gap-3">
                        <div className="flex-1 bg-slate-50 rounded-t-xl relative overflow-hidden group">
                            <div className="absolute bottom-0 w-full bg-blue-400 rounded-t-xl h-1/2 transition-all group-hover:h-3/5"></div>
                        </div>
                        <div className="flex-1 bg-slate-50 rounded-t-xl relative overflow-hidden group">
                            <div className="absolute bottom-0 w-full bg-blue-400 rounded-t-xl h-2/3 transition-all group-hover:h-3/4"></div>
                        </div>
                        <div className="flex-1 bg-slate-50 rounded-t-xl relative overflow-hidden group">
                            <div className="absolute bottom-0 w-full bg-blue-400 rounded-t-xl h-3/4 transition-all group-hover:h-[85%]"></div>
                        </div>
                        <div className="flex-1 bg-slate-50 rounded-t-xl relative overflow-hidden group">
                            <div className="absolute bottom-0 w-full bg-blue-400 rounded-t-xl h-2/5 transition-all group-hover:h-1/2"></div>
                        </div>
                        <div className="flex-1 bg-slate-50 rounded-t-xl relative overflow-hidden group">
                            <div className="absolute bottom-0 w-full bg-blue-600 rounded-t-xl h-[90%] transition-all group-hover:h-full"></div>
                        </div>
                    </div>
                    <div className="mt-4 flex justify-between text-[10px] text-slate-400 uppercase font-bold tracking-widest px-2">
                        <span>T2</span><span>T3</span><span>T4</span><span>T5</span><span>T6</span>
                    </div>

                    {/* Floating Widget */}
                    <div className="absolute -right-6 top-1/2 -translate-y-1/2 bg-white p-4 rounded-2xl shadow-2xl border border-slate-100 w-36 hidden xl:block">
                        <div className="w-16 h-16 border-[6px] border-blue-500 border-t-cyan-400 rounded-full mx-auto flex items-center justify-center">
                            <span className="text-[10px] text-slate-800 font-bold">99.9%</span>
                        </div>
                        <p className="text-[10px] text-center mt-3 font-bold text-slate-400 uppercase tracking-tight">Độ chính xác OCR</p>
                    </div>
                </div>

                {/* Pagination Dots */}
                <div className="flex gap-2 mt-12">
                    <div className="w-6 h-1 bg-white/40 rounded-full"></div>
                    <div className="w-10 h-1 bg-white rounded-full"></div>
                    <div className="w-6 h-1 bg-white/40 rounded-full"></div>
                </div>

                {/* Background Decoration */}
                <div className="absolute top-0 right-0 w-64 h-64 bg-white/5 rounded-full -mr-20 -mt-20 blur-3xl"></div>
                <div className="absolute bottom-0 left-0 w-64 h-64 bg-blue-400/20 rounded-full -ml-20 -mb-20 blur-3xl"></div>
            </div>
        </div>
    );
};

export default AuthPage;