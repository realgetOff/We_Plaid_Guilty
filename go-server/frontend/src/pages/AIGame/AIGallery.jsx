/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   AIGallery.jsx                                      :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/03/29 23:35:05 by mforest-          #+#    #+#             */
/*   Updated: 2026/03/29 23:35:05 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React from 'react';
import './AIGallery.css';

const normalizeResult = (r, index) =>
{
	const playerId = r.player_id ?? r.playerId ?? r.PlayerID ?? `p-${index}`;
	const playerName = r.player_name ?? r.playerName ?? r.name ?? 'Player';
	const drawing = r.drawing ?? r.Drawing ?? '';
	const rawScore = r.score ?? r.Score ?? 0;
	const score = typeof rawScore === 'number' && !Number.isNaN(rawScore) ? rawScore : 0;
	return { playerId, playerName, drawing, score };
};

const AIGallery = ({ results, onBack }) =>
{
	if (!Array.isArray(results) || results.length === 0)
	{
		return (
			<div className="aigallery">
				<p className="aigallery__empty">No results.</p>
				<button className="aigallery__btn" onClick={onBack}>← Back</button>
			</div>
		);
	}

	const normalized = results.map(normalizeResult);
	const sorted = [...normalized].sort((a, b) => b.score - a.score);

	const medals = ['🥇', '🥈', '🥉'];

	return (
		<div className="aigallery">
			<h2 className="aigallery__title">🏆 Results</h2>
			<div className="aigallery__grid">
				{sorted.map((r, idx) => (
					<div
						key={r.playerId}
						className={`aigallery__card${idx === 0 ? ' aigallery__card--winner' : ''}`}
					>
						<div className="aigallery__rank">
							{medals[idx] || `#${idx + 1}`}
						</div>
						<div className="aigallery__player">{r.playerName}</div>
						<img
							src={r.drawing}
							alt={r.playerName}
							className="aigallery__img"
						/>
						<div className="aigallery__score">
							{'★'.repeat(Math.min(5, Math.max(0, Math.round(r.score))))}
							{'☆'.repeat(5 - Math.min(5, Math.max(0, Math.round(r.score))))}
							<span className="aigallery__score-num">{r.score.toFixed(1)} / 5</span>
						</div>
					</div>
				))}
			</div>
			<button className="aigallery__btn" onClick={onBack}>← Back to home</button>
		</div>
	);
};

export default AIGallery;
