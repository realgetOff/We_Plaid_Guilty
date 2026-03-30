/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   AIJoinGame.jsx                                     :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/03/30 00:06:32 by mforest-          #+#    #+#             */
/*   Updated: 2026/03/30 00:06:32 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';

const getApiBaseUrl = () =>
{
	const raw = import.meta.env.VITE_API_URL;
	if (raw && typeof raw === 'string' && raw.trim() !== '')
		return raw.replace(/\/$/, '');
	if (typeof window !== 'undefined' && window.location && window.location.origin)
		return window.location.origin;
	return '';
};

const AIJoinGame = () =>
{
	const { code } = useParams();
	const navigate  = useNavigate();
	const [error,   setError]   = useState('');
	const [loading, setLoading] = useState(true);

	useEffect(() =>
	{
		const verifyAndRedirect = async () =>
		{
			const normalized = code?.toUpperCase();
			if (!normalized || !/^[A-Z]{6}$/.test(normalized))
			{
				setError('Invalid room code format.');
				setLoading(false);
				return;
			}
			try
			{
				const res  = await fetch(getApiBaseUrl() + '/api/ai-rooms/' + encodeURIComponent(normalized), {
					method: 'GET',
					credentials: 'include',
				});
				if (!res.ok)
				{
					setError('Room not found. Check the code and try again.');
					setLoading(false);
					return;
				}
				const room = await res.json();
				if (room.status !== 'ai_waiting')
				{
					setError(`Game is already ${room.status}.`);
					setLoading(false);
					return;
				}
				navigate(`/aigame/lobby/${normalized}`, { replace: true });
			}
			catch (err)
			{
				setError('Room not found. Check the code and try again.');
				setLoading(false);
			}
		};
		verifyAndRedirect();
	}, [code, navigate]);

	if (loading) return <div className="loader">Checking room...</div>;

	return (
		<div className="error-container">
			<p>⚠ {error}</p>
			<button onClick={() => navigate('/game')}>Back to Home</button>
		</div>
	);
};

export default AIJoinGame;
