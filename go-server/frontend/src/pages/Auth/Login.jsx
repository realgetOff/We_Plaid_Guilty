/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   Login.jsx                                          :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/02/26 01:19:24 by mforest-          #+#    #+#             */
/*   Updated: 2026/02/26 01:19:24 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React, { useState, useEffect } from 'react';
import { useNavigate, useSearchParams, Link } from 'react-router-dom';
import './Auth.css';

const Login = () =>
{
	const navigate = useNavigate();
	const [searchParams] = useSearchParams();
	const redirect = searchParams.get('redirect') || '/game';

	const [loading, setLoading] = useState(false);
	const [error, setError] = useState('');

	const hasToken = !!localStorage.getItem('authToken');

	useEffect(() =>
	{
		const token = searchParams.get('token');
		const authError = searchParams.get('error');

		if (token)
		{
			localStorage.setItem('authToken', token);
			window.dispatchEvent(new CustomEvent('userDataUpdated'));
			navigate(redirect, { replace: true });
		}
		else if (authError)
		{
			setError('Authentication failed: ' + authError);
		}
	}, [searchParams, navigate, redirect]);

	const handleGuest = async () =>
	{
		setError('');
		if (hasToken)
		{
			navigate(redirect, { replace: true });
			return;
		}
		setLoading(true);
		try
		{
			const res = await fetch('/api/auth/player', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
			});
			if (!res.ok)
				throw new Error('Server error');
			const data = await res.json();
			if (!data.token)
				throw new Error('No token');
			localStorage.setItem('authToken', data.token);
			window.dispatchEvent(new CustomEvent('userDataUpdated'));
			navigate(redirect, { replace: true });
		}
		catch (e)
		{
			setError('Could not start a guest session.');
		}
		finally
		{
			setLoading(false);
		}
	};

	const handleIntra = () =>
	{
		setError('');
		if (hasToken)
		{
			navigate(redirect, { replace: true });
			return;
		}
		setLoading(true);
		window.location.href = "http://localhost:8080/login/42";
	};

	const handleSignOut = () =>
	{
		localStorage.removeItem('authToken');
		localStorage.removeItem('username');
		localStorage.removeItem('userID');
		localStorage.removeItem('isGuest');
		window.dispatchEvent(new CustomEvent('userDataUpdated'));
		window.location.reload();
	};

	return (
		<div className="login-page">
			<div className="login-page__window">
				<div className="login-page__titlebar">
					<span className="login-page__titlebar-dot login-page__titlebar-dot--r" />
					<span className="login-page__titlebar-dot login-page__titlebar-dot--y" />
					<span className="login-page__titlebar-dot login-page__titlebar-dot--g" />
					<span className="login-page__title">ft_transcendence — sign in</span>
				</div>
				<div className="login-page__body">
					<p className="login-page__lead">Choose how you want to play.</p>

					{error && <p className="login-page__error" role="alert">{error}</p>}

					<div className="login-page__actions">
						<button
							type="button"
							className="login-page__btn login-page__btn--guest"
							onClick={handleGuest}
							disabled={loading}
						>
							{loading ? '⧗ connecting…' : hasToken ? '→ continue' : '◇ continue as guest'}
						</button>

						<button
							type="button"
							className="login-page__btn login-page__btn--intra"
							onClick={handleIntra}
							disabled={loading}
						>
							<span className="login-page__btn-label">sign in with</span>
							<span className="login-page__btn-intra-mark">42</span>
						</button>
					</div>

					{hasToken && (
						<div className="login-page__signed">
							<p className="login-page__signed-text">You already have a session.</p>
							<button type="button" className="login-page__linkish" onClick={handleSignOut}>
								sign out and use another account
							</button>
						</div>
					)}

					<div className="login-page__footer">
						<Link to="/" className="login-page__footer-link">← back</Link>
					</div>
				</div>
			</div>
		</div>
	);
};

export default Login;
