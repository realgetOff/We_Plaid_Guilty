import { useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';

const AuthCallback = () => {
    const [searchParams] = useSearchParams();
    const navigate = useNavigate();
    const redirect = '/'; // Or wherever your default landing is

    useEffect(() => {
        const token = searchParams.get('token');
        
        if (token) {
            // 1. Persist the session
            localStorage.setItem('authToken', token);
            
            // 2. Notify other components (like Navbars) to re-fetch user data
            window.dispatchEvent(new CustomEvent('userDataUpdated'));
            
            // 3. Clean redirect
            navigate(redirect, { replace: true });
        } else {
            const errorReason = searchParams.get('error') || 'auth_failed';
            navigate(`/login?error=${errorReason}`);
        }
    }, [searchParams, navigate]);

    return (
        <div className="flex items-center justify-center h-screen">
            <p className="animate-pulse">Verifying your Intra login...</p>
        </div>
    );
};

export default AuthCallback;