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
import '../Game/CreateGame.css';

const AICreateGame = () =>
{
	const navigate = useNavigate();
	const msgEndRef = useRef(null);
	const roomCodeRef = useRef('');

	const [status,    setStatus]    = useState('checking');
	const [roomCode,  setRoomCode]  = useState('');
	const [copied,    setCopied]    = useState(false);
	const [players,   setPlayers]   = useState([]);
	const [messages,  setMessages]  = useState([]);
	const [input,     setInput]     = useState('');
	const [createErr, setCreateErr] = useState('');

	useEffect(() =>
	{
		roomCodeRef.current = roomCode;
	}, [roomCode]);

	useEffect(() =>
	{
		connect();
		const handler = (msg) =>
		{
			const currentCode = roomCodeRef.current;
			const msgRoom = msg.room || msg.code;
			const isMatch = msgRoom === currentCode;

			if (msg.type === 'ai_room_created')
			{
				setRoomCode(msg.code);
				roomCodeRef.current = msg.code;
				setCreateErr('');
				setStatus('ready');
			}
			if (msg.type === 'lobby_state' && isMatch)
			{
				if (Array.isArray(msg.players))
					setPlayers(msg.players);
			}
			if (msg.type === 'ai_chat_message' && isMatch)
			{
				setMessages(prev => [...prev, {
					id:   Date.now(),
					user: msg.user,
					text: msg.text
				}]);
			}
			if (msg.type === 'start_ai_game' && isMatch)
			{
				navigate(`/aigame/play/${msg.code}`);
			}
		};

		addListener(handler);
		send({ type: 'create_ai_room' });

		return () =>
		{
			removeListener(handler);
			if (roomCodeRef.current)
				send({ type: 'leave_ai_room', code: roomCodeRef.current });
		};
	}, [navigate]);

	useEffect(() =>
	{
		msgEndRef.current?.scrollIntoView({ behavior: 'smooth' });
	}, [messages]);

	const handleSend = () =>
	{
		if (!input.trim() || !roomCode)
			return;
		send({ type: 'ai_chat_message', code: roomCode, text: input.trim() });
		setInput('');
	};

	const handleCopy = () =>
	{
		navigator.clipboard.writeText(roomCode);
		setCopied(true);
		setTimeout(() => setCopied(false), 2000);
	};

	if (status === 'checking')
	{
		return (
			<div className="lobby__guard">
				<span className="lobby__guard-spinner">⧗</span>
				Creating Neural Room...
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
                    {players.length < 3 && (
                        <p className="creategame__waiting">⧗ waiting for players…</p>
                    )}
                </div>
            </div>

            <div className="creategame__card creategame__card--chat">
                <div className="creategame__card-header">💬 chat</div>
                <div className="creategame__chat-messages">
                    {messages.map((m) => (
                        <div key={m.id} className="creategame__msg">
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
                    />
                    <button onClick={handleSend}>→</button>
                </div>
            </div>
        </div>

        <div className="creategame__actions">
            <button className="creategame__btn creategame__btn--leave" onClick={() => navigate('/game')}>
                ✕ leave room
            </button>
            <button
                className="creategame__btn creategame__btn--start"
                onClick={() => send({ type: 'start_ai_game', code: roomCode })}
                disabled={players.length < 3}
            >
                🤖 start AI game
            </button>
        </div>

        {players.length < 3 && (
            <p className="creategame__start-hint">
                ⚠ need at least 3 players to start.
            </p>
        )}
    </div>
	);
};

export default AICreateGame;
