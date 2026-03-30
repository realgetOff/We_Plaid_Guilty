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
import './AILobby.css';

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
			if (msg.type === 'chat_message' && roomMatch)
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

	const handleSend = () => {
		if (!input.trim()) return;
		send({ type: 'chat_message', code: normalized, text: input.trim() });
		setInput('');
	};

	if (deny) return <div className="lobby__guard">⚠ {deny} <button onClick={() => navigate('/game')}>Back</button></div>;

	return (
		<div className="lobby">
			<div className="lobby__code-band">🤖 AI LOBBY — ROOM: <span className="lobby__code">{normalized}</span></div>

			<div className="lobby__columns">
				<div className="lobby__card lobby__card--players">
					<div className="lobby__card-header">👥 Players ({players.length}/8)</div>
					<div className="lobby__player-list">
						{players.map(p => (
							<div key={p.id} className="lobby__player-row">
								<span className="lobby__player-name">{p.name}</span>
								{p.host && <span className="lobby__badge">HOST</span>}
							</div>
						))}
					</div>
					<div className="lobby__actions-sub">
						<button className="lobby__btn--leave" onClick={() => navigate('/game')}>✕ LEAVE</button>
						{isHost && (
							<button 
								className="lobby__btn--start" 
								onClick={() => send({ type: 'start_ai_game', code: normalized })}
								disabled={players.length < 3}
							>
								▶ START
							</button>
						)}
					</div>
				</div>

				<div className="lobby__card lobby__card--chat">
					<div className="lobby__card-header">💬 Chat</div>
					<div className="lobby__chat-messages">
						{messages.map(m => (
							<div key={m.id} className={`lobby__msg ${m.user === myName ? 'lobby__msg--me' : ''}`}>
								<strong>{m.user}:</strong> {m.text}
							</div>
						))}
						<div ref={msgEndRef} />
					</div>
					<div className="lobby__chat-input-row">
						<input value={input} onChange={e => setInput(e.target.value)} onKeyDown={e => e.key === 'Enter' && handleSend()} placeholder="Type a message..." />
						<button onClick={handleSend}>→</button>
					</div>
				</div>
			</div>
		</div>
	);
};

export default AILobby;
