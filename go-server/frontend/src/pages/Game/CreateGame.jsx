/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   CreateGame.jsx                                     :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: pmilner- <pmilner-@student.42.fr>          +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/02/23 21:11:46 by mforest-          #+#    #+#             */
/*   Updated: 2026/03/04 22:12:01 by pmilner-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React, { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { connect, send, addListener, removeListener } from '../../socket';
import './CreateGame.css';

const CreateGame = () =>
{
	const navigate = useNavigate();

	const [status,    setStatus]    = useState('checking');
	const [roomCode,  setRoomCode]  = useState('');
	const [copied,    setCopied]    = useState(false);
	const [players,   setPlayers]   = useState([]);
	const [createErr, setCreateErr] = useState('');
	const roomCodeRef = useRef('');

	useEffect(() =>
	{
		roomCodeRef.current = roomCode;
	}, [roomCode]);

	useEffect(() =>
	{
		connect();

		const handler = (msg) =>
		{
			if (msg.type === 'room_created')
			{
				if (msg.code)
				{
					setRoomCode(msg.code);
					roomCodeRef.current = msg.code;
				}
				if (Array.isArray(msg.players))
					setPlayers(msg.players);
				setCreateErr('');
				setStatus('ready');
			}

			if (msg.type === 'lobby_state')
			{
				if (Array.isArray(msg.players))
					setPlayers(msg.players);
			}

			if (msg.type === 'create_denied')
			{
				setCreateErr(msg.reason || 'Could not create room.');
				setStatus('ready');
			}
			if (msg.type === 'start_game')
    			navigate(`/game/play/${msg.code || msg.room}`);
		};

		addListener(handler);
		send({ type: 'create_room' });

		return () =>
		{
			removeListener(handler);
			const code = roomCodeRef.current;
			if (code)
				send({ type: 'leave_lobby', code: code });
		};
	}, []);

	const handleCopy = () =>
	{
		navigator.clipboard.writeText(roomCode);
		setCopied(true);
		setTimeout(() => setCopied(false), 2000);
	};

	const handleStart = () =>
	{
		if (players.length < 3 || !roomCode)
			return;
		send({ type: 'start_game', code: roomCode });
	};

	const handleLeave = () =>
	{
		navigate('/game');
	};

	if (status === 'checking')
	{
		return (
			<div className="creategame__guard">
				<span className="creategame__guard-spinner">⧗</span>
				creating your room…
			</div>
		);
	}

	if (createErr && !roomCode)
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

	return (
		<div className="creategame">
			<div className="creategame__card">
				<div className="creategame__card-header">🔑 room code</div>
				<div className="creategame__card-body creategame__card-body--center">
					<p className="creategame__hint">
						share this code with your friends so they can join.
					</p>
					<div className="creategame__code-row">
						<span className="creategame__code">{roomCode}</span>
						<button className="creategame__btn creategame__btn--copy" onClick={handleCopy}>
							{copied ? '✓ copied!' : '⎘ copy'}
						</button>
					</div>
				</div>
			</div>

			<div className="creategame__columns">
				<div className="creategame__card creategame__card--grow">
					<div className="creategame__card-header">
						👥 players
						<span className="creategame__card-header-count">
							{players.length} / 8
						</span>
					</div>
					<div className="creategame__card-body creategame__card-body--list">
						{players.map((p) => (
							<div key={p.id} className="creategame__player-row">
								<span className="creategame__player-dot" />
								<span className="creategame__player-name">{p.name}</span>
								{p.host && <span className="creategame__badge">HOST</span>}
							</div>
						))}
						{players.length < 3 && (
							<p className="creategame__waiting">⧗ waiting for players…</p>
						)}
					</div>
				</div>
			</div>

			<div className="creategame__actions">
				<button className="creategame__btn creategame__btn--leave" onClick={handleLeave}>
					✕ leave room
				</button>
				<button
					className="creategame__btn creategame__btn--start"
					onClick={handleStart}
					disabled={players.length < 3}
					title={players.length < 3 ? 'need at least 2 players' : ''}
				>
					▶ start game
				</button>
			</div>

			{players.length < 3 && (
				<p className="creategame__start-hint">
					⚠ settings will be automatically adjusted based on player count.
				</p>
			)}
		</div>
	);
};

export default CreateGame;
