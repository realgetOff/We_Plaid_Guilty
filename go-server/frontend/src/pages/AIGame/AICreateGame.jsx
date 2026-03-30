/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   AICreateGame.jsx                                   :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/03/30 00:06:16 by mforest-          #+#    #+#             */
/*   Updated: 2026/03/30 00:06:16 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React, { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { connect, send, addListener, removeListener } from '../../socket';

const AICreateGame = () =>
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
			if (msg.type === 'ai_room_created')
			{
				if (msg.code)
				{
					setRoomCode(msg.code);
					roomCodeRef.current = msg.code;
				}
				setCreateErr('');
				setStatus('ready');
			}

			if (msg.type === 'lobby_state')
			{
				if (Array.isArray(msg.players))
					setPlayers(msg.players);
			}

			if (msg.type === 'start_ai_game')
			{
				navigate(`/aigame/play/${msg.room || msg.code}`);
			}

			if (msg.type === 'create_denied')
			{
				setCreateErr(msg.reason || 'Could not create room.');
				setStatus('ready');
			}
		};

		addListener(handler);
		send({ type: 'create_ai_room' });

		return () =>
		{
			removeListener(handler);
			const code = roomCodeRef.current;
			if (code)
				send({ type: 'leave_ai_room', code: code });
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
		if (players.length < 3 || !roomCode) return;
		send({ type: 'start_ai_game', code: roomCode });
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
				<div className="creategame__card-header">🤖 AI Game — room code</div>
				<div className="creategame__card-body creategame__card-body--center">
					<p className="creategame__hint">
						Share this code with your friends. The AI will generate the prompt!
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
						{players.length < 3 &&
							<p className="creategame__waiting">⧗ waiting for players…</p>
						}
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
					title={players.length < 3 ? 'need at least 3 players' : ''}
				>
					▶ start game
				</button>
			</div>

			{players.length < 3 &&
				<p className="creategame__start-hint">
					⚠ need at least 3 players to start.
				</p>
			}
		</div>
	);
};

export default AICreateGame;
