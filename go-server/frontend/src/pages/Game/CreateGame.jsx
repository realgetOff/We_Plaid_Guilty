/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   CreateGame.jsx                                     :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/04/18 14:32:29 by mforest-          #+#    #+#             */
/*   Updated: 2026/04/18 14:32:29 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React, { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { connect, send, addListener, removeListener } from '../../api/socket';
import './CreateGame.css';

const CreateGame = () =>
{
	const navigate = useNavigate();
	const roomCodeRef = useRef('');

	const [status,    setStatus]    = useState('checking');
	const [createErr, setCreateErr] = useState('');

	useEffect(() =>
	{
		connect();
		const handler = (msg) =>
		{
			if (msg.type === 'room_created')
			{
				if (msg.code)
				{
					roomCodeRef.current = msg.code;
					setStatus('ready');
					navigate(`/game/lobby/${msg.code}`, { replace: true });
				}
				setCreateErr('');
			}

			if (msg.type === 'create_denied')
			{
				setCreateErr(msg.reason || 'Could not create room.');
				setStatus('error');
			}
		};

		addListener(handler);
		send({ type: 'create_room' });

		return () =>
		{
			removeListener(handler);
		};
	}, [navigate]);

	if (status === 'checking')
	{
		return (
			<div className="creategame__guard">
				<span className="creategame__guard-spinner">⧗</span>
				creating your room…
			</div>
		);
	}

	if (status === 'error' || createErr)
	{
		return (
			<div className="creategame__guard">
				<div className="creategame__guard-card">
					<p className="creategame__guard-msg">⚠ {createErr}</p>
					<button className="creategame__guard-btn" onClick={() => navigate('/game')}>
						← back to game
					</button>
				</div>
			</div>
		);
	}

	return null;
};

export default CreateGame;
