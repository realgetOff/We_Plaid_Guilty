/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   AILobby.jsx                                        :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/03/30 00:06:44 by mforest-          #+#    #+#             */
/*   Updated: 2026/03/30 00:06:44 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React, { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { connect, send, addListener, removeListener } from '../../socket';
import '../Game/CreateGame.css';

const AILobby = () =>
{
	const { code } = useParams();
	const navigate = useNavigate();
	const msgEndRef = useRef(null);
	const normalized = code?.toUpperCase();

	const [players, setPlayers] = useState([]);
	const [messages, setMessages] = useState([]);
	const [input, setInput] = useState('');
	const [isHost, setIsHost] = useState(false);
	const [myName, setMyName] = useState('');
	const [deny, setDeny] = useState('');

	useEffect(() =>
	{
		connect();
		const handler = (msg) =>
		{
			const roomMatch = msg.room === normalized || msg.code === normalized;

			if (msg.type === 'lobby_state' && roomMatch)
			{
				if (Array.isArray(msg.players))
					setPlayers(msg.players);
				if (msg.me)
				{
					setIsHost(!!msg.me.host);
					setMyName(msg.me.name);
				}
			}
			if (msg.type === 'ai_chat_message' && roomMatch)
				setMessages(prev => [...prev, { id: Date.now(), user: msg.user, text: msg.text }]);
			if (msg.type === 'start_ai_game' && roomMatch)
				navigate(`/aigame/play/${normalized}`);
			if (msg.type === 'join_denied')
				setDeny(msg.reason);
		};

		addListener(handler);
		send({ type: 'join_ai_room', code: normalized });
		return () =>
		{
			removeListener(handler);
			send({ type: 'leave_ai_room', code: normalized });
		};
	}, [normalized, navigate]);

	useEffect(() => { msgEndRef.current?.scrollIntoView({ behavior: 'smooth' }); }, [messages]);

	const handleSend = () =>
	{
		if (!input.trim())
			return;
		send({ type: 'ai_chat_message', code: normalized, text: input.trim() });
		setInput('');
	};

	if (deny) return (
		<div className="creategame__guard">
			<div className="creategame__guard-card">
				<p className="creategame__guard-msg">⚠ {deny}</p>
				<button className="creategame__guard-btn" onClick={() => navigate('/game')}>
					← back to game
				</button>
			</div>
		</div>
	);

	return (
		<div className="creategame">
			<div className="creategame__card">
				<div className="creategame__card-header">🤖 AI Lobby — room code</div>
				<div className="creategame__card-body creategame__card-body--center">
					<p className="creategame__hint">
						you have joined this room. waiting for host to start.
					</p>
					<div className="creategame__code-row">
						<span className="creategame__code">{normalized}</span>
						<button className="creategame__btn creategame__btn--copy" style={{opacity: 0.3}} disabled>
							⎘ copy
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

				<div className="creategame__card creategame__card--chat">
					<div className="creategame__card-header">💬 chat</div>
					<div className="creategame__chat-messages">
						{messages.map((m) => (
							<div key={m.id} className={`creategame__msg ${m.user === myName ? 'creategame__msg--me' : ''}`}>
								<strong>{m.user}:</strong> {m.text}
							</div>
						))}
						<div ref={msgEndRef} />
					</div>
					<div className="creategame__chat-input-row">
						<input
							value={input}
							onChange={(e) => setInput(e.target.value)}
							onKeyDown={(e) => e.key === 'Enter' && handleSend()}
							placeholder="Type a message..."
						/>
						<button onClick={handleSend}>→</button>
					</div>
				</div>
			</div>

			<div className="creategame__actions">
				<button className="creategame__btn creategame__btn--leave" onClick={() => navigate('/game')}>
					✕ leave room
				</button>
				{isHost ? (
					<button
						className="creategame__btn creategame__btn--start"
						onClick={() => send({ type: 'start_ai_game', code: normalized })}
						disabled={players.length < 3}
					>
						🤖 start AI game
					</button>
				) : (
					<button
						className="creategame__btn creategame__btn--start"
						disabled={true}
					>
						⧗ waiting for host
					</button>
				)}
			</div>

			{players.length < 3 && (
				<p className="creategame__start-hint">
					⚠ need at least 3 players to start.
				</p>
			)}
		</div>
	);
};

export default AILobby;
