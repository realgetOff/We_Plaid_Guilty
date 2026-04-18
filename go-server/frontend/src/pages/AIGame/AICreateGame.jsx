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
import { connect, send, addListener, removeListener } from '../../api/socket';
import '../Game/CreateGame.css';

const AICreateGame = () =>
{
	const navigate = useNavigate();
	const msgEndRef = useRef(null);
	const roomCodeRef = useRef('');

	const [status, setStatus] = useState('checking');
	const [createErr, setCreateErr] = useState('');
	const [roomCode, setRoomCode] = useState('');
	const [copied, setCopied] = useState(false);
	const [players, setPlayers] = useState([]);
	const [messages, setMessages] = useState([]);
	const [input, setInput] = useState('');
	const [isStarting, setIsStarting] = useState(false);

	const [showFriends, setShowFriends] = useState(false);
	const [friends, setFriends] = useState([]);
	const [friendsLoading, setFriendsLoading] = useState(false);
	const [inviting, setInviting] = useState(null);

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
				setCreateErr('');
				setRoomCode(msg.code);
				roomCodeRef.current = msg.code;
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
					id: Date.now(),
					user: msg.user,
					text: msg.text
				}]);
			}
			if (msg.type === 'start_ai_game' && isMatch)
				navigate(`/aigame/play/${msg.code}`);
			if (msg.type === 'error' && isMatch)
				setIsStarting(false);

			if (msg.type === 'friends_list')
			{
				setFriends(msg.friends || []);
				setFriendsLoading(false);
			}
			if (msg.type === 'friend_online_status')
			{
				setFriends(prev => prev.map(f => 
					f.username === msg.username 
						? { ...f, online: msg.online }
						: f
				));
			}
			if (msg.type === 'invite_sent')
			{
				if (msg.success)
					setTimeout(() => setInviting(null), 2000);
				else
					setInviting(null);
			}
		};

		addListener(handler);
		send({ type: 'create_ai_room' });

		const timeoutId = setTimeout(() =>
		{
			if (!roomCodeRef.current)
				setCreateErr('No response from server. Check your connection and try again.');
		}, 20000);

		return () =>
		{
			clearTimeout(timeoutId);
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

	const handleStartGame = () =>
	{
		if (isStarting || players.length < 3)
			return;
		setIsStarting(true);
		send({ type: 'start_ai_game', code: roomCode });
		
		setTimeout(() => setIsStarting(false), 10000);
	};

	const handleCopy = () =>
	{
		navigator.clipboard.writeText(roomCode);
		setCopied(true);
		setTimeout(() => setCopied(false), 2000);
	};

	const toggleFriends = () =>
	{
		if (!showFriends && friends.length === 0)
		{
			setFriendsLoading(true);
			send({ type: 'get_friends' });
		}
		setShowFriends(!showFriends);
	};

	const handleInviteFriend = (friend) =>
	{
		if (!roomCode) return;
		setInviting(friend.id);
		send({ 
			type: 'invite_friend', 
			to: friend.username, 
			code: roomCode,
			is_ai: true,
		});
	};

	if (status === 'checking')
	{
		return (
			<div className="creategame__guard">
				<span className="creategame__guard-spinner">⧗</span>
				Creating Neural Room...
				{createErr && (
					<div className="creategame__guard-card" style={{ marginTop: '1.5rem', maxWidth: '420px' }}>
						<p className="creategame__guard-msg">⚠ {createErr}</p>
						<button type="button" className="creategame__guard-btn" onClick={() => window.location.reload()}>
							retry
						</button>
						<button type="button" className="creategame__guard-btn" onClick={() => navigate('/game')} style={{ marginLeft: '0.5rem' }}>
							← back
						</button>
					</div>
				)}
			</div>
		);
	}

	return (
		<div className="creategame">
			<button 
				className="creategame__friends-toggle" 
				onClick={toggleFriends}
			>
				👥 friends ({friends.filter(f => f.online).length} online)
			</button>
			
			<div className="creategame__layout">
				<div className="creategame__main">
					<div className="creategame__card">
						<div className="creategame__card-header">🤖 room code</div>
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
							onClick={handleStartGame}
							disabled={players.length < 3 || isStarting}
						>
							{isStarting ? '🤖 starting...' : '🤖 start AI game'}
						</button>
					</div>

					{players.length < 3 && (
						<p className="creategame__start-hint">
							⚠ need at least 3 players to start.
						</p>
					)}
				</div>

				{showFriends && (
					<div className="creategame__friends-sidebar">
						<div className="creategame__card">
							<div className="creategame__card-header">
								👥 friends online
								<button 
									className="creategame__friends-close"
									onClick={() => setShowFriends(false)}
								>
									✕
								</button>
							</div>
							<div className="creategame__card-body creategame__card-body--list">
								{friendsLoading ? (
									<p className="creategame__waiting">⧗ loading friends...</p>
								) : friends.length === 0 ? (
									<p className="creategame__waiting">no friends yet.</p>
								) : (
									friends
										.filter(f => f.online)
										.map(f => (
											<div key={f.id} className="creategame__player-row">
												<span className="creategame__player-dot" />
												<span className="creategame__player-name">{f.username}</span>
												<button 
													className="creategame__invite-friend"
													onClick={() => handleInviteFriend(f)}
													disabled={inviting === f.id}
													title={`Invite ${f.username} to this room`}
												>
													{inviting === f.id ? '✓' : '✉'}
												</button>
											</div>
										))
								)}
								{friends.filter(f => f.online).length === 0 && friends.length > 0 && (
									<p className="creategame__waiting">no friends online.</p>
								)}
							</div>
						</div>
					</div>
				)}
			</div>
		</div>
	);
};

export default AICreateGame;
