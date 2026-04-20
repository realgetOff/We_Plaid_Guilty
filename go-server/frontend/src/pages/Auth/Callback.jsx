/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   Callback.jsx                                       :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/04/18 14:15:53 by mforest-          #+#    #+#             */
/*   Updated: 2026/04/18 14:15:53 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import { useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';

const AuthCallback = () =>
{
    const [searchParams] = useSearchParams();
    const navigate = useNavigate();
    const redirect = '/';

    useEffect(() =>
	{
        const token = searchParams.get('token');
        
        if (token)
		{
            localStorage.setItem('authToken', token);
            window.dispatchEvent(new CustomEvent('userDataUpdated'));
            navigate(redirect, { replace: true });
        }
		else
		{
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
