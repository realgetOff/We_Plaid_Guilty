/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   HomeGame.jsx                                       :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/02/23 20:11:08 by mforest-          #+#    #+#             */
/*   Updated: 2026/02/23 20:11:08 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { roomsApi, getApiBaseUrl } from '../../api/rooms';
import { connect } from '../../api/socket'
import './HomeGame.css';

const VALID_CODE_RE = /^[A-Z]{6}$/;

const HomeGame = () =>
{
	const navigate = useNavigate();

	useEffect(() => {
	    connect();
    }, []);


	const [joinCode,     setJoinCode]     = useState('');
	const [joinError,    setJoinError]    = useState('');
	const [isChecking,   setIsChecking]   = useState(false);
	const [aiJoinCode,   setAiJoinCode]   = useState('');
	const [aiJoinError,  setAiJoinError]  = useState('');
	const [aiIsChecking, setAiIsChecking] = useState(false);

	const handleJoin = async () =>
	{
		const code = joinCode.trim().toUpperCase();
		if (!code)
		{
			setJoinError('Please enter a room code.');
			return;
		}
		if (!VALID_CODE_RE.test(code))
		{
			setJoinError('Room code must be exactly 6 letters (A–Z).');
			return;
		}
		setIsChecking(true);
		setJoinError('');
		let room;
		try
		{
			room = await roomsApi.getRoom(code);
		}
		catch
		{
			setIsChecking(false);
			setJoinError('Room not found. Check the code and try again.');
			return;
		}
		setIsChecking(false);
		if (!room || !room.status)
		{
			setJoinError('Room not found. Check the code and try again.');
			return;
		}
		if (room.status === 'started')
		{
			setJoinError('This game has already started. You cannot join.');
			return;
		}
		navigate('/game/join/' + code);
	};

	const handleAiJoin = async () =>
	{
		const code = aiJoinCode.trim().toUpperCase();
		if (!code)
		{
			setAiJoinError('Please enter a room code.');
			return;
		}
		if (!VALID_CODE_RE.test(code))
		{
			setAiJoinError('Room code must be exactly 6 letters (A–Z).');
			return;
		}
		setAiIsChecking(true);
		setAiJoinError('');
		try
		{
			const res = await fetch(getApiBaseUrl() + '/api/ai-rooms/' + encodeURIComponent(code), {
				method: 'GET',
				credentials: 'include',
			});
			if (!res.ok)
			{
				setAiIsChecking(false);
				setAiJoinError('Room not found. Check the code and try again.');
				return;
			}
			const room = await res.json();
			setAiIsChecking(false);
			if (room.status !== 'ai_waiting')
			{
				setAiJoinError('This game has already started.');
				return;
			}
			navigate('/aigame/join/' + code);
		}
		catch
		{
			setAiIsChecking(false);
			setAiJoinError('Room not found. Check the code and try again.');
		}
	};

	const handleCodeChange = (e) =>
	{
		setJoinCode(
			e.target.value
				.toUpperCase()
				.replace(/[^A-Z]/g, '')
				.slice(0, 6)
		);
		setJoinError('');
	};

	const handleAiCodeChange = (e) =>
	{
		setAiJoinCode(
			e.target.value
				.toUpperCase()
				.replace(/[^A-Z]/g, '')
				.slice(0, 6)
		);
		setAiJoinError('');
	};

	return (
		<div className="homegame">

			<div className="homegame__section-title">🎨 Gartic Phone</div>

			<div className="homegame__card">
				<div className="homegame__card-header">Create a Game</div>
				<div className="homegame__card-body">
					<p className="homegame__card-desc">
						Start a new room and invite your friends with a 6-letter code.
					</p>
					<button
						className="homegame__btn homegame__btn--primary"
						onClick={() => navigate('/game/create')}
					>
						▶ Create Room
					</button>
				</div>
			</div>

			<div className="homegame__separator">— or —</div>

			<div className="homegame__card">
				<div className="homegame__card-header">🔑 Join a Game</div>
				<div className="homegame__card-body">
					<p className="homegame__card-desc">
						Enter the 6-letter room code shared by the host.
					</p>
					<div className="homegame__join-row">
						<input
							className={`homegame__input${joinError ? ' homegame__input--error' : ''}`}
							type="text"
							value={joinCode}
							onChange={handleCodeChange}
							onKeyDown={(e) => { if (e.key === 'Enter' && !isChecking) handleJoin(); }}
							placeholder="ABCDEF"
							maxLength={6}
							disabled={isChecking}
						/>
						<button
							className="homegame__btn homegame__btn--secondary"
							onClick={handleJoin}
							disabled={isChecking}
						>
							{isChecking ? '⧗' : '→ Join'}
						</button>
					</div>
					{joinError && <p className="homegame__error">⚠ {joinError}</p>}
				</div>
			</div>

			<div className="homegame__section-title" style={{ marginTop: '2rem' }}>🤖 AI Game</div>

			<div className="homegame__card">
				<div className="homegame__card-header">Create an AI Game</div>
				<div className="homegame__card-body">
					<p className="homegame__card-desc">
						Let the AI generate a funny prompt. Everyone draws, then you vote!
					</p>
					<button
						className="homegame__btn homegame__btn--primary"
						onClick={() => navigate('/aigame/create')}
					>
						▶ Create AI Room
					</button>
				</div>
			</div>

			<div className="homegame__separator">— or —</div>

			<div className="homegame__card">
				<div className="homegame__card-header">🔑 Join an AI Game</div>
				<div className="homegame__card-body">
					<p className="homegame__card-desc">
						Enter the 6-letter room code shared by the host.
					</p>
					<div className="homegame__join-row">
						<input
							className={`homegame__input${aiJoinError ? ' homegame__input--error' : ''}`}
							type="text"
							value={aiJoinCode}
							onChange={handleAiCodeChange}
							onKeyDown={(e) => { if (e.key === 'Enter' && !aiIsChecking) handleAiJoin(); }}
							placeholder="ABCDEF"
							maxLength={6}
							disabled={aiIsChecking}
						/>
						<button
							className="homegame__btn homegame__btn--secondary"
							onClick={handleAiJoin}
							disabled={aiIsChecking}
						>
							{aiIsChecking ? '⧗' : '→ Join'}
						</button>
					</div>
					{aiJoinError && <p className="homegame__error">⚠ {aiJoinError}</p>}
				</div>
			</div>

		</div>
	);
};

export default HomeGame;
