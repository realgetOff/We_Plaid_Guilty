/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   JoinGame.jsx                                       :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/02/23 23:29:36 by mforest-          #+#    #+#             */
/*   Updated: 2026/02/23 23:29:36 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { roomsApi } from '../../api/rooms';

const JoinGame = () =>
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
				const room = await roomsApi.getRoom(normalized);
				if (room.status !== 'waiting')
				{
					setError(`Game is already ${room.status}.`);
					setLoading(false);
					return;
				}
				navigate(`/game/lobby/${normalized}`, { replace: true });
			}
			catch (err)
			{
				setError('Room not found. Check the code and try again.');
				setLoading(false);
			}
		};
		verifyAndRedirect();
	}, [code, navigate]);

	if (loading)
		return <div className="loader">Checking room...</div>;

	return (
		<div className="error-container">
			<p>⚠ {error}</p>
			<button onClick={() => navigate('/game')}>Back to Home</button>
		</div>
	);
};

export default JoinGame;
