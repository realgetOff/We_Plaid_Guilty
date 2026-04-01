/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   AIJoinGame.jsx                                     :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/03/30 03:44:53 by mforest-          #+#    #+#             */
/*   Updated: 2026/03/30 03:44:53 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { roomsApi } from '../../api/rooms';

const AIJoinGame = () =>
{
	const { code } = useParams();
	const navigate = useNavigate();
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
				const room = await roomsApi.getAIRoom(normalized);
				
				if (!room)
				{
					setError('AI Room not found.');
					setLoading(false);
					return;
				}

				if (room.status !== 'ai_waiting')
				{
					setError(`Game is already in progress (${room.status}).`);
					setLoading(false);
					return;
				}

				navigate(`/aigame/lobby/${normalized}`, { replace: true });
			}
			catch (err)
			{
				setError('Failed to connect to the neural server.');
				setLoading(false);
			}
		};
		verifyAndRedirect();
	}, [code, navigate]);

	if (loading)
		return <div className="loader">Synchronizing Neural Link...</div>;

	return (
		<div className="error-container">
			<p>⚠ {error}</p>
			<button onClick={() => navigate('/game')}>Back to Selection</button>
		</div>
	);
};

export default AIJoinGame;
