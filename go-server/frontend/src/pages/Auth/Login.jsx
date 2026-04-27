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
import { authApi } from '../../api/auth';
import './Auth.css';

const AuthModal = ({ type, onClose, onSubmit, loading, error }) =>
{
	const [formData, setFormData] = useState({
		username: '',
		password: '',
		email: ''
	});
	const [localError, setLocalError] = useState('');

	const handleChange = (e) =>
	{
		setFormData({ ...formData, [e.target.name]: e.target.value });
		if (localError)
			setLocalError('');
	};

	const validateAndSubmit = (e) =>
	{
		e.preventDefault();
		if (!formData.username || !formData.password || (type === 'register' && !formData.email))
		{
			setLocalError('⚠ missing field: access denied');
			return;
		}
		onSubmit(formData);
	};

	return (
		<div className="login-modal-overlay">
			<div className="login-page__window modal-content">
				<div className="login-page__titlebar">
					<span className="login-page__title" style={{ paddingLeft: '10px' }}>
						{type === 'login' ? 'system_auth.exe' : 'new_user_reg.sys'}
					</span>
					<button onClick={onClose} className="login-page__titlebar-dot login-page__titlebar-dot--r" style={{ border: 'none', cursor: 'pointer', marginLeft: 'auto', marginRight: '5px' }} />
				</div>
				<form className="login-page__body" onSubmit={validateAndSubmit} noValidate>
					{(error || localError) && (
						<p className="login-page__error" style={{ textTransform: 'lowercase', fontStyle: 'italic' }}>
							{localError || error}
						</p>
					)}
					<div className="form-group">
						<label className="login-page__lead" style={{ fontSize: '0.8rem', marginBottom: '5px', display: 'block' }}>
							{'>'} user_id
						</label>
						<input name="username" type="text" className="login-page__input" onChange={handleChange} autoComplete="off" />
					</div>
					{type === 'register' && (
						<div className="form-group" style={{ marginTop: '10px' }}>
							<label className="login-page__lead" style={{ fontSize: '0.8rem', marginBottom: '5px', display: 'block' }}>
								{'>'} email_addr
							</label>
							<input name="email" type="email" className="login-page__input" onChange={handleChange} autoComplete="off" />
						</div>
					)}
					<div className="form-group" style={{ marginTop: '10px' }}>
						<label className="login-page__lead" style={{ fontSize: '0.8rem', marginBottom: '5px', display: 'block' }}>
							{'>'} access_key
						</label>
						<input name="password" type="password" className="login-page__input" onChange={handleChange} />
					</div>
					<div className="login-page__actions" style={{ marginTop: '25px' }}>
						<button type="submit" className="login-page__btn" disabled={loading}>
							{loading ? '⧗ executing…' : `[ ${type} ]`}
						</button>
					</div>
				</form>
			</div>
		</div>
	);
};

const Login = () =>
{
	const navigate = useNavigate();
	const [searchParams] = useSearchParams();
	const redirect = searchParams.get('redirect') || '/game';

	const [loading, setLoading] = useState(false);
	const [error, setError] = useState('');
	const [modal, setModal] = useState(null);

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
			setError('authentication failed: ' + authError);
		}
	}, [searchParams, navigate, redirect]);

	const finalizeAuth = (token) =>
	{
		localStorage.setItem('authToken', token);
		window.dispatchEvent(new CustomEvent('userDataUpdated'));
		navigate(redirect, { replace: true });
	};

	const safeFetch = async (url, options) =>
	{
		const res = await fetch(url, options);
		const contentType = res.headers.get("content-type");
		if (contentType && contentType.includes("application/json"))
			return await res.json();
		throw new Error(`expected json, received ${res.status}`);
	};

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
			const data = await safeFetch('/api/auth/player', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
			});
			if (!data.token)
				throw new Error('null_token_err');
			finalizeAuth(data.token);
		}
		catch (e)
		{
			setError('guest session failed');
		}
		finally
		{
			setLoading(false);
		}
	};

	const handleRegister = async (formData) =>
	{
		setLoading(true);
		setError('');
		try
		{
			const data = await safeFetch('/api/auth/register', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(formData)
			});
			if (!data.token)
				throw new Error(data.message || 'reg_denied');
			finalizeAuth(data.token);
		}
		catch (e)
		{
			setError(e.message);
		}
		finally
		{
			setLoading(false);
		}
	};

	const handleLogin = async (formData) =>
	{
		setLoading(true);
		setError('');
		try
		{
			const data = await safeFetch('/api/auth/login', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(formData)
			});
			if (!data.token)
				throw new Error(data.message || 'auth_denied');
			finalizeAuth(data.token);
		}
		catch (e)
		{
			setError(e.message);
		}
		finally
		{
			setLoading(false);
		}
	};

	const handleIntra = async () =>
	{
		setError('');
		setLoading(true);
		try
		{
			const url = await authApi.oauth42Url();

			if (url.startsWith("https://api.intra.42.fr/"))
			{
				window.location.href = url;
			}
		}
		catch
		{
			setError(' 42 gateway unreachable');
			setLoading(false);
		}
	};

	const handleSignOut = () =>
	{
		localStorage.clear();
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
					<p className="login-page__lead">Select connection protocol...</p>
					{error && <p className="login-page__error" role="alert">{error}</p>}
					<div className="login-page__actions">
						<button type="button" className="login-page__btn login-page__btn--intra" onClick={handleIntra} disabled={loading}>
							<span className="login-page__btn-label">link_with_intra</span>
							<span className="login-page__btn-intra-mark">42</span>
						</button>
						<button type="button" className="login-page__btn" onClick={() => setModal('login')} disabled={loading}>
							{'>'} login
						</button>
						<button type="button" className="login-page__btn" onClick={() => setModal('register')} disabled={loading}>
							{'>'} register
						</button>
						<button type="button" className="login-page__btn login-page__btn--guest" onClick={handleGuest} disabled={loading}>
							{hasToken ? '→ resume_session' : '◇ bypass_auth (guest)'}
						</button>
					</div>
					{hasToken && (
						<div className="login-page__signed">
							<button type="button" className="login-page__linkish" onClick={handleSignOut}>
								/terminate_current_session
							</button>
						</div>
					)}
					<div className="login-page__footer">
						<Link to="/" className="login-page__footer-link">{'<'} abort_return_home</Link>
					</div>
				</div>
			</div>
			{modal === 'login' && (
				<AuthModal type="login" onClose={() => setModal(null)} onSubmit={handleLogin} loading={loading} error={error} />
			)}
			{modal === 'register' && (
				<AuthModal type="register" onClose={() => setModal(null)} onSubmit={handleRegister} loading={loading} error={error} />
			)}
		</div>
	);
};

export default Login;
