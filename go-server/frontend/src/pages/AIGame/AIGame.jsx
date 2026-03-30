/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   AIGame.jsx                                         :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/03/29 23:34:36 by mforest-          #+#    #+#             */
/*   Updated: 2026/03/29 23:34:36 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import DrawBoard from '../Game/DrawBoard';
import AIVotePanel from './AIVotePanel';
import AIGallery from './AIGallery';
import { connect, send, addListener, removeListener } from '../../socket';
import './AIGame.css';

const AIGame = () => {
	const { code } = useParams();
	const navigate = useNavigate();

	const [status, setStatus] = useState('waiting');
	const [phase, setPhase] = useState(null);
	const [prompt, setPrompt] = useState('');
	const [drawings, setDrawings] = useState([]);
	const [results, setResults] = useState([]);
	const [myId, setMyId] = useState('');

	useEffect(() => {
		const normalized = code?.toUpperCase();
		connect();

		const handler = (msg) => {
			console.log("DEBUG WS RECEIVE:", msg); // Vérifie ici ce que le serveur envoie

			// Vérification du code de la room (insensible à la casse)
			if (!msg || msg.room?.toUpperCase() !== normalized) return;

			// État de jeu global (utilisé pour la transition DRAW -> VOTE)
			if (msg.type === 'ai_game_state') {
				setStatus('playing');
				setPhase(msg.phase);
				if (msg.prompt) setPrompt(msg.prompt);
				if (msg.my_id) setMyId(msg.my_id);
				// Si on reçoit des dessins dans ce message, on les met à jour
				if (Array.isArray(msg.drawings)) setDrawings(msg.drawings);
			}

			// Message spécifique pour le vote
			if (msg.type === 'ai_vote_state') {
				console.log("SWITCHING TO VOTE PHASE");
				setPhase('vote');
				if (Array.isArray(msg.drawings)) setDrawings(msg.drawings);
			}

			// Message spécifique pour les résultats
			if (msg.type === 'ai_results') {
				setPhase('gallery');
				if (Array.isArray(msg.results)) setResults(msg.results);
			}
		};

		addListener(handler);
		
		// On demande l'état actuel de la partie au join
		send({ type: 'join_ai_game', code: normalized });

		return () => {
			removeListener(handler);
			send({ type: 'leave_ai_game', code: normalized });
		};
	}, [code]);

	const handleDrawDone = (dataURL) => {
		console.log("SENDING DRAWING...");
		send({
			type: 'ai_drawing_submitted',
			code: code?.toUpperCase(),
			drawing: dataURL,
		});
		setPhase('waiting'); // On passe en attente locale
	};

	const handleVoteDone = (votes) => {
		console.log("SENDING VOTES:", votes);
		send({
			type: 'ai_votes_submitted',
			code: code?.toUpperCase(),
			votes: votes,
		});
		setPhase('waiting');
	};

	let phaseLabel = '';
	if (phase === 'draw') phaseLabel = '🎨 Draw your answer!';
	if (phase === 'vote') phaseLabel = '⭐ Rate the drawings';
	if (phase === 'waiting') phaseLabel = '⧗ Waiting for others…';
	if (phase === 'gallery') phaseLabel = '🏆 Results';

	if (status === 'waiting') {
		return (
			<div className="aigame__guard">
				<span className="aigame__spinner">⧗</span>
				<p>Connecting to room <strong>{code?.toUpperCase()}</strong>…</p>
			</div>
		);
	}

	return (
		<div className="aigame">
			<div className="aigame__header">
				<span className="aigame__phase-label">{phaseLabel}</span>
				<span className="aigame__room-code">#{code?.toUpperCase()}</span>
			</div>

			{phase === 'draw' && (
				<>
					<div className="aigame__prompt-banner">
						🤖 <strong>{prompt}</strong>
					</div>
					<DrawBoard prompt={prompt} onDone={handleDrawDone} />
				</>
			)}

			{phase === 'vote' && (
				<AIVotePanel drawings={drawings} myId={myId} onDone={handleVoteDone} />
			)}

			{phase === 'waiting' && (
				<div className="aigame__waiting">
					<span className="aigame__spinner">⧗</span>
					<p>Waiting for other players to finish…</p>
				</div>
			)}

			{phase === 'gallery' && (
				<AIGallery results={results} onBack={() => navigate('/game')} />
			)}
		</div>
	);
};

export default AIGame;
