/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   AIGame.jsx                                         :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/03/29 23:34:36 by mforest-          #+#    #+#             */
/*   Updated: 2026/03/30 00:00:00 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import DrawBoard from './AIDrawBoard';
import AIVotePanel from './AIVotePanel';
import AIGallery from './AIGallery';
import { connect, send, addListener, removeListener } from '../../socket';
import '../Game/Game.css';

const AIGame = () =>
{
	const { code } = useParams();
	const navigate = useNavigate();
	const normalized = code?.toUpperCase();

	const [status, setStatus] = useState('waiting');
	const [phase, setPhase] = useState(null);
	const [prompt, setPrompt] = useState('');
	const [drawings, setDrawings] = useState([]);
	const [results, setResults] = useState([]);
	const [myId, setMyId] = useState('');

	useEffect(() =>
	{
		if (!normalized) return;

		connect();

		const handler = (msg) =>
		{

			if (!msg || msg.room?.toUpperCase() !== normalized)
				return;

			if (msg.type === 'ai_game_state')
			{
				setStatus('playing');
				setPhase(msg.phase);
				if (msg.prompt)
					setPrompt(msg.prompt);
				if (msg.my_id)
					setMyId(msg.my_id);
				if (Array.isArray(msg.drawings))
					setDrawings(msg.drawings);
			}

			if (msg.type === 'ai_vote_state')
			{
				if (Array.isArray(msg.drawings)) setDrawings(msg.drawings);
				setPhase('vote');
			}

			if (msg.type === 'ai_results')
			{
				if (Array.isArray(msg.results)) setResults(msg.results);
				setPhase('gallery');
			}
		};

		addListener(handler);

		send({ type: 'join_ai_game', code: normalized });

		return () =>
		{
			removeListener(handler);
			send({ type: 'leave_ai_game', code: normalized });
		};
	}, [normalized]);

	const handleDrawDone = (dataURL, title, description) =>
	{
    	send({
    	    type: 'ai_drawing_submitted',
    		    code: code?.toUpperCase(),
       		drawing: dataURL,
        	title: title || '',
        	description: description || '',
    	});
    	setPhase('waiting');
	};

	const handleVoteDone = (votes) =>
	{
		send({
			type: 'ai_votes_submitted',
			code: normalized,
			votes: votes,
		});
		setPhase('waiting');
	};

	let phaseLabel = '';
	if (phase === 'draw') phaseLabel = '🎨 Draw your answer!';
	if (phase === 'vote') phaseLabel = '⭐ Rate the drawings';
	if (phase === 'waiting') phaseLabel = '⧗ Waiting for others…';
	if (phase === 'gallery') phaseLabel = '🏆 Results';

	if (status === 'waiting')
	{
		return (
			<div className="aigame__guard">
				<span className="aigame__spinner">⧗</span>
				<p>Connecting to room <strong>{normalized}</strong>…</p>
			</div>
		);
	}


return (
    <div className="game">
        <div className="game__phase-bar">
            <span className={`game__phase-dot${phase === 'draw' ? ' game__phase-dot--on' : ''}`} />
            <span className={`game__phase-dot${phase === 'vote' ? ' game__phase-dot--on' : ''}`} />
            <span className={`game__phase-dot${phase === 'gallery' ? ' game__phase-dot--on' : ''}`} />
            <span className="game__phase-label">{phaseLabel}</span>
            <span className="game__room-code">#{code?.toUpperCase()}</span>
        </div>

        {phase === 'draw' && (
            <DrawBoard prompt={prompt} onDone={handleDrawDone} />
        )}
        {phase === 'vote' && (
            <AIVotePanel drawings={drawings} myId={myId} onDone={handleVoteDone} />
        )}
        {phase === 'waiting' && (
            <div className="game__waiting">
                <span className="game__waiting-spinner">⧗</span>
                <p>Waiting for other players to finish…</p>
            </div>
        )}
        {phase === 'gallery' && (
            <AIGallery results={results} onBack={() => navigate('/game')} />
        )}
        {!phase && (
            <div className="game__waiting">
                <span className="game__waiting-spinner">⧗</span>
                <p>Connecting…</p>
            </div>
        )}
    </div>
);
};

export default AIGame;
