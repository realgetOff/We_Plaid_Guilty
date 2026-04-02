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
import { connect, send, addListener, removeListener } from '../../api/socket';
import './CreateGame.css';

const CreateGame = () =>
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
	const [myName,    setMyName]    = useState('');
	const [createErr, setCreateErr] = useState('');

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

			if (msg.type === 'room_created')
			{
				if (msg.code)
				{
					setRoomCode(msg.code);
					roomCodeRef.current = msg.code;
				}
				if (Array.isArray(msg.players))
					setPlayers(msg.players);
				if (msg.me)
					setMyName(msg.me.name || '');
				setCreateErr('');
				setStatus('ready');
			}

			if (msg.type === 'lobby_state' && isMatch)
			{
				if (Array.isArray(msg.players))
					setPlayers(msg.players);
				if (msg.me)
					setMyName(msg.me.name || '');
			}

			if (msg.type === 'chat_message' && isMatch)
			{
				setMessages((prev) => [...prev,
				{
					id:   msg.id || Date.now(),
					user: msg.user,
					text: msg.text,
				}]);
			}

			if (msg.type === 'create_denied')
			{
				setCreateErr(msg.reason || 'Could not create room.');
				setStatus('ready');
			}

			if (msg.type === 'start_game' && isMatch)
				navigate(`/game/play/${msg.code || msg.room}`);

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
		send({ type: 'create_room' });

		return () =>
		{
			removeListener(handler);
			const code = roomCodeRef.current;
			if (code)
				send({ type: 'leave_lobby', code: code });
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
		send({ type: 'chat_message', code: roomCode, text: input.trim() });
		setInput('');
	};

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
			code: roomCode 
		});
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
			<button 
				className="creategame__friends-toggle" 
				onClick={toggleFriends}
			>
				👥 friends ({friends.filter(f => f.online).length} online)
			</button>
			
			<div className="creategame__layout">
				<div className="creategame__main">
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
								/>
								<button onClick={handleSend}>→</button>
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

					{players.length < 3 && (
						<p className="creategame__start-hint">
							⚠ settings will be automatically adjusted based on player count.
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

export default CreateGame;
