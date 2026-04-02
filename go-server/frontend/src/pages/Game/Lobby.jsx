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
import { connect, send, addListener, removeListener } from '../../api/socket';
import { roomsApi } from '../../api/rooms';
import '../Game/CreateGame.css';

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

	const [showFriends, setShowFriends] = useState(false);
	const [friends, setFriends] = useState([]);
	const [friendsLoading, setFriendsLoading] = useState(false);
	const [inviting, setInviting] = useState(null);

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
		if (!normalized) return;
		setInviting(friend.id);
		send({ 
			type: 'invite_friend', 
			to: friend.username, 
			code: normalized 
		});
	};

	if (status === 'checking')
	{
		return (
			<div className="creategame__guard">
				<span className="creategame__guard-spinner">⧗</span>
				checking room…
			</div>
		);
	}

	if (status === 'denied')
	{
		return (
			<div className="creategame__guard">
				<div className="creategame__guard-card">
					<p className="creategame__guard-msg">⚠ {deny}</p>
					<button className="creategame__guard-btn" onClick={handleLeave}>
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
								you have joined this room. waiting for host to start.
							</p>
							<div className="creategame__code-row">
								<span className="creategame__code">{normalized}</span>
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
						{isHost ? (
							<button
								className="creategame__btn creategame__btn--start"
								onClick={handleStart}
								disabled={players.length < 3}
								title={players.length < 3 ? 'need at least 3 players' : ''}
							>
								▶ start game
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

export default Lobby;
