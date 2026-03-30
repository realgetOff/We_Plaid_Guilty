/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   Lobby.jsx                                          :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/03/28 19:52:46 by mforest-          #+#    #+#             */
/*   Updated: 2026/03/28 19:52:46 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React, { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { connect, send, addListener, removeListener } from '../../socket';
import { roomsApi } from '../../api/rooms';
import './Lobby.css';

const DENY_REASONS = {
	invalid:   'invalid room code format.',
	not_found: 'room not found.',
	started:   'this game has already started.',
	finished:  'this game is already finished.',
	unknown:   'cannot access this room.',
};

const Lobby = () =>
{
	const { code } = useParams();
	const navigate = useNavigate();
	const msgEndRef = useRef(null);

	const [status,   setStatus]   = useState('checking');
	const [deny,     setDeny]     = useState('');
	const [players,  setPlayers]  = useState([]);
	const [messages, setMessages] = useState([]);
	const [input,    setInput]    = useState('');
	const [isHost,   setIsHost]   = useState(false);
	const [myName,   setMyName]   = useState('');

	const normalized = code?.toUpperCase();

	useEffect(() =>
	{
		const checkRoom = async () =>
		{
			if (!normalized || !/^[A-Z]{6}$/.test(normalized))
			{
				setDeny(DENY_REASONS.invalid);
				setStatus('denied');
				return;
			}
			try
			{
				const room = await roomsApi.getRoom(normalized);
				if (!room || room.status === 'started')
				{
					setDeny(room?.status === 'started' ? DENY_REASONS.started : DENY_REASONS.not_found);
					setStatus('denied');
					return;
				}
				setStatus('ready');
			}
			catch (err)
			{
				setDeny(DENY_REASONS.not_found);
				setStatus('denied');
			}
		};
		checkRoom();
	}, [normalized]);

	useEffect(() =>
	{
		if (status !== 'ready')
			return;

		connect();

		const handler = (msg) =>
		{
			if (!msg)
				return;
			const roomMatch = msg.room === normalized || msg.code === normalized;

			if (msg.type === 'lobby_state' && roomMatch)
			{
				if (Array.isArray(msg.players))
					setPlayers(msg.players);
				if (msg.me)
				{
					setIsHost(!!msg.me.host);
					setMyName(msg.me.name || '');
				}
			}

			if (msg.type === 'chat_message' && roomMatch)
			{
				console.log("msg.type modal received: ", msg.text);
				setMessages((prev) => [...prev,
				{
					id:   msg.id || Date.now(),
					user: msg.user,
					text: msg.text,
				}]);
			}

			if (msg.type === 'start_game' && roomMatch)
				navigate(`/game/play/${normalized}`);

			if (msg.type === 'join_denied')
			{
				setDeny(msg.reason || DENY_REASONS.unknown);
				setStatus('denied');
			}
		};

		addListener(handler);
		send({ type: 'join_room', code: normalized });

		return () =>
		{
			removeListener(handler);
			send({ type: 'leave_lobby', code: normalized });
		};
	}, [status, normalized, navigate]);

	useEffect(() =>
	{
		msgEndRef.current?.scrollIntoView({ behavior: 'smooth' });
	}, [messages]);

	const handleSend = () =>
	{
		if (!input.trim() || !normalized)
			return;
		send({ type: 'chat_message', code: normalized, text: input.trim() });
		setInput('');
	};

	const handleStart = () =>
	{
		if (players.length < 3 || !normalized)
			return;
		send({ type: 'start_game', code: normalized });
	};

	const handleLeave = () =>
	{
		navigate('/game');
	};

	if (status === 'checking')
	{
		return (
			<div className="lobby__guard">
				<span className="lobby__guard-spinner">⧗</span>
				checking room…
			</div>
		);
	}

	if (status === 'denied')
	{
		return (
			<div className="lobby__guard">
				<div className="lobby__guard-card">
					<p className="lobby__guard-msg">⚠ {deny}</p>
					<button className="lobby__guard-btn" onClick={handleLeave}>
						← back to game
					</button>
				</div>
			</div>
		);
	}

	return (
		<div className="lobby">
			<div className="lobby__card">
				<div className="lobby__card-header">🔑 room code</div>
				<div className="lobby__card-body lobby__card-body--center">
					<p className="lobby__hint">
						you have joined this room. waiting for host to start.
					</p>
					<div className="lobby__code-row">
						<span className="lobby__code">{normalized}</span>
					</div>
				</div>
			</div>

			<div className="lobby__columns">
				<div className="lobby__card lobby__card--grow">
					<div className="lobby__card-header">
						👥 players
						<span className="lobby__card-header-count">
							{players.length} / 8
						</span>
					</div>
					<div className="lobby__card-body lobby__card-body--list">
						{players.map((p) => (
							<div key={p.id} className="lobby__player-row">
								<span className="lobby__player-dot" />
								<span className="lobby__player-name">{p.name}</span>
								{p.host && <span className="lobby__badge">HOST</span>}
							</div>
						))}
						{players.length < 3 && (
							<p className="lobby__waiting">⧗ waiting for players…</p>
						)}
					</div>
				</div>

				<div className="lobby__card lobby__card--chat">
					<div className="lobby__card-header">💬 chat</div>
					<div className="lobby__chat-messages">
						{messages.map((m) => (
							<div key={m.id} className={`lobby__msg ${m.user === myName ? 'lobby__msg--me' : ''}`}>
								<strong>{m.user}:</strong> {m.text}
							</div>
						))}
						<div ref={msgEndRef} />
					</div>
					<div className="lobby__chat-input-row">
						<input
							value={input}
							onChange={(e) => setInput(e.target.value)}
							onKeyDown={(e) => e.key === 'Enter' && handleSend()}
						/>
						<button onClick={handleSend}>→</button>
					</div>
				</div>
			</div>

			<div className="lobby__actions">
				<button className="lobby__btn lobby__btn--leave" onClick={handleLeave}>
					✕ leave room
				</button>
				{isHost ? (
					<button
						className="lobby__btn lobby__btn--start"
						onClick={handleStart}
						disabled={players.length < 3}
						title={players.length < 3 ? 'need at least 3 players' : ''}
					>
						▶ start game
					</button>
				) : (
					<button
						className="lobby__btn lobby__btn--start"
						disabled={true}
					>
						⧗ waiting for host
					</button>
				)}
			</div>

			{players.length < 3 && (
				<p className="lobby__start-hint">
					⚠ settings will be automatically adjusted based on player count.
				</p>
			)}
		</div>
	);
};

export default Lobby;
