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
import './AILobby.css';

const AICreateGame = () =>
{
	const navigate = useNavigate();
	const msgEndRef = useRef(null);

	const [status,    setStatus]    = useState('checking');
	const [roomCode,  setRoomCode]  = useState('');
	const [copied,    setCopied]    = useState(false);
	const [players,   setPlayers]   = useState([]);
	const [messages,  setMessages]  = useState([]);
	const [input,     setInput]     = useState('');
	const [createErr, setCreateErr] = useState('');
	const roomCodeRef = useRef('');

	useEffect(() => { roomCodeRef.current = roomCode; }, [roomCode]);

	useEffect(() =>
	{
		connect();
		const handler = (msg) =>
		{
			const currentCode = roomCodeRef.current;
			const isMine = !currentCode || msg.code === currentCode || msg.room === currentCode;

			if (msg.type === 'ai_room_created')
			{
				setRoomCode(msg.code);
				setCreateErr('');
				setStatus('ready');
			}
			if (msg.type === 'lobby_state' && isMine)
			{
				if (Array.isArray(msg.players)) setPlayers(msg.players);
			}
			if (msg.type === 'chat_message' && isMine)
			{
				setMessages(prev => [...prev, { id: Date.now(), user: msg.user, text: msg.text }]);
			}
			if (msg.type === 'start_ai_game' && isMine)
			{
				navigate(`/aigame/play/${msg.code}`);
			}
		};

		addListener(handler);
		send({ type: 'create_ai_room' });
		return () => {
			removeListener(handler);
			if (roomCodeRef.current) send({ type: 'leave_ai_room', code: roomCodeRef.current });
		};
	}, [navigate]);

	useEffect(() => { msgEndRef.current?.scrollIntoView({ behavior: 'smooth' }); }, [messages]);

	const handleSend = () => {
		if (!input.trim() || !roomCode) return;
		send({ type: 'chat_message', code: roomCode, text: input.trim() });
		setInput('');
	};

	if (status === 'checking') return <div className="lobby__guard">⧗ Creating Neural Room...</div>;

	return (
		<div className="lobby">
			<div className="lobby__code-band">
				🤖 AI HOST — CODE: <span className="lobby__code">{roomCode}</span>
				<button className="lobby__code-copy" onClick={() => {
					navigator.clipboard.writeText(roomCode);
					setCopied(true);
					setTimeout(() => setCopied(false), 2000);
				}}>
					{copied ? '✓' : '⎘'}
				</button>
			</div>

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
						<button 
							className="lobby__btn--start" 
							onClick={() => send({ type: 'start_ai_game', code: roomCode })}
							disabled={players.length < 3}
						>
							▶ START
						</button>
					</div>
				</div>

				<div className="lobby__card lobby__card--chat">
					<div className="lobby__card-header">💬 Chat</div>
					<div className="lobby__chat-messages">
						{messages.map(m => (
							<div key={m.id} className="lobby__msg"><strong>{m.user}:</strong> {m.text}</div>
						))}
						<div ref={msgEndRef} />
					</div>
					<div className="lobby__chat-input-row">
						<input value={input} onChange={e => setInput(e.target.value)} onKeyDown={e => e.key === 'Enter' && handleSend()} placeholder="Type here..." />
						<button onClick={handleSend}>→</button>
					</div>
				</div>
			</div>
		</div>
	);
};

export default AICreateGame;
