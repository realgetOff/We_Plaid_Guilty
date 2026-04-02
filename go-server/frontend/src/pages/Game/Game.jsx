/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   Game.jsx                                           :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/02/23 23:31:46 by mforest-          #+#    #+#             */
/*   Updated: 2026/02/23 23:31:46 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import WritePrompt from './WritePrompt';
import DrawBoard from './DrawBoard';
import GuessPrompt from './GuessPrompt';
import Gallery from './Gallery';
import { connect, send, addListener, removeListener } from '../../api/socket';
import './Game.css';

const DENY_REASONS =
{
	invalid:   'invalid room code format.',
	not_found: 'room not found.',
	waiting:   'this game has not started yet.',
	finished:  'this game is already finished.',
	unknown:   'cannot join this room.',
};

const Game = () =>
{
	const { code } = useParams();
	const navigate = useNavigate();

	const [status,  setStatus]  = useState('checking');
	const [deny,    setDeny]    = useState('');
	const [phase,   setPhase]   = useState(null);
	const [prompt,  setPrompt]  = useState('');
	const [drawing, setDrawing] = useState(null);
	const [chains,  setChains]  = useState([]);

	useEffect(() =>
	{
		const normalized = code?.toUpperCase();
		if (!normalized || !/^[A-Z]{6}$/.test(normalized))
		{
			setDeny(DENY_REASONS.invalid);
			setStatus('denied');
			return;
		}

		connect();

		const handler = (msg) =>
		{
			if (!msg || msg.room !== normalized)
				return;

			if (msg.type === 'game_state')
			{
				setPhase(msg.phase || null);
				setPrompt(msg.prompt || '');
				setDrawing(msg.drawing || null);
				if (Array.isArray(msg.chains))
					setChains(msg.chains);
				setStatus('playing');
				return;
			}

			if (msg.type === 'game_denied')
			{
				setDeny(msg.reason || DENY_REASONS.unknown);
				setStatus('denied');
				return;
			}
		};

		addListener(handler);

		send({
			type: 'join_game',
			code: normalized,
		});

		return () =>
		{
			removeListener(handler);
			send({
				type: 'leave_game',
				code: normalized,
			});
		};
	}, [code]);

	const handlePromptDone = (text) =>
	{
		send({
			type:   'prompt_submitted',
			code:   code?.toUpperCase(),
			prompt: text,
		});
		setPhase('waiting');
	};

	const handleDrawDone = (dataURL) =>
	{
		send({
			type:    'drawing_submitted',
			code:    code?.toUpperCase(),
			drawing: dataURL,
		});
		setPhase('waiting');
	};

	const handleGuessDone = (guess) =>
	{
		send({
			type:  'guess_submitted',
			code:  code?.toUpperCase(),
			guess: guess,
		});
		setPhase('waiting');
	};

	let phaseLabel = '';
	if (phase === 'write')   phaseLabel = '✏ write a prompt';
	if (phase === 'draw')    phaseLabel = '🎨 draw it!';
	if (phase === 'guess')   phaseLabel = '🔍 what is it?';
	if (phase === 'waiting') phaseLabel = '⧗ waiting for others…';
	if (phase === 'gallery') phaseLabel = '📜 gallery';

	if (status === 'checking')
	{
		return (
			<div className="game__guard">
				<span className="game__guard-spinner">⧗</span>
				verifying room <strong>{code?.toUpperCase()}</strong>…
			</div>
		);
	}

	if (status === 'denied')
	{
		return (
			<div className="game__guard">
				<div className="game__guard-card">
					<div className="game__guard-icon">✕</div>
					<p className="game__guard-msg">⚠ {deny}</p>
					<div className="game__guard-actions">
						<button className="game__guard-btn" onClick={() => navigate('/game')}>
							← back to home
						</button>
						{deny === DENY_REASONS.waiting &&
							<button
								className="game__guard-btn game__guard-btn--primary"
								onClick={() => navigate(`/game/lobby/${code?.toUpperCase()}`)}
							>
								→ go to lobby
							</button>
						}
					</div>
				</div>
			</div>
		);
	}

	return (
		<div className="game">
			<div className="game__phase-bar">
				<span className={`game__phase-dot${phase === 'write'   ? ' game__phase-dot--on' : ''}`} />
				<span className={`game__phase-dot${phase === 'draw'    ? ' game__phase-dot--on' : ''}`} />
				<span className={`game__phase-dot${phase === 'guess'   ? ' game__phase-dot--on' : ''}`} />
				<span className={`game__phase-dot${phase === 'gallery' ? ' game__phase-dot--on' : ''}`} />
				<span className="game__phase-label">{phaseLabel}</span>
				<span className="game__room-code">#{code?.toUpperCase()}</span>
			</div>

			{phase === 'write' &&
				<WritePrompt onDone={handlePromptDone} />
			}
			{phase === 'draw' &&
				<DrawBoard prompt={prompt} onDone={handleDrawDone} />
			}
			{phase === 'guess' &&
				<GuessPrompt drawing={drawing} onDone={handleGuessDone} />
			}
			{phase === 'waiting' &&
				<div className="game__waiting">
					<span className="game__waiting-spinner">⧗</span>
					<p>Waiting for other players to finish…</p>
				</div>
			}
			{phase === 'gallery' &&
				<Gallery chains={chains} onBack={() => navigate('/game')} />
			}
			{!phase &&
				<div className="game__waiting">
					<span className="game__waiting-spinner">⧗</span>
					<p>Connecting…</p>
				</div>
			}
		</div>
	);
};

export default Game;
