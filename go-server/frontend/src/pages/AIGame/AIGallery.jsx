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

const normalizeResult = (result, index) =>
{
	const playerId = result.player_id ?? `p-${index}`;
	const playerName = result.player_name ?? 'Unknown';
	const drawing = result.drawing ?? '';
	const title = result.title ?? '';
	const description = result.description ?? '';
	const rawScore = result.score ?? 0;

	const score = typeof rawScore === 'number' && !Number.isNaN(rawScore) ? rawScore : parseFloat(rawScore) || 0;

	return { playerId, playerName, drawing, title, description, score };
};

const getRankDisplay = (index) =>
{
	const medals = ['🥇', '🥈', '🥉'];
	return medals[index] || `#${index + 1}`;
};

const renderStars = (score) =>
{
	const clamped      = Math.min(5, Math.max(0, score));
	const full         = Math.floor(clamped);
	const hasHalf      = clamped - full >= 0.5;
	const empty        = 5 - full - (hasHalf ? 1 : 0);

	return (
		<span className="aigallery__stars">
			{'★'.repeat(full)}
			{hasHalf ? '½' : ''}
			{'☆'.repeat(empty)}
			<span className="aigallery__score-num">{clamped.toFixed(1)} / 5</span>
		</span>
	);
};

const AIGallery = ({ results, onBack }) =>
{
	if (!Array.isArray(results) || results.length === 0)
	{
		return (
			<div className="aigallery">
				<p className="aigallery__empty">No results to display.</p>
				<button className="aigallery__btn" onClick={onBack}>← Back</button>
			</div>
		);
	}

	const sortedResults = results
		.map(normalizeResult)
		.sort((a, b) => b.score - a.score);

	return (
		<div className="aigallery">
			<h2 className="aigallery__title">🏆 Results</h2>

			<div className="aigallery__grid">
				{sortedResults.map((result, index) => (
					<div
						key={result.playerId}
						className={`aigallery__card${index === 0 ? ' aigallery__card--winner' : ''}`}
					>
						<div className="aigallery__rank">
							{getRankDisplay(index)}
						</div>

						<div className="aigallery__player">
							{result.playerName}
						</div>

						{result.title && (
							<div className="aigallery__drawing-title">
								{result.title}
							</div>
						)}

						{result.drawing
							? (
								<img
									src={result.drawing}
									alt={`Drawing by ${result.playerName}`}
									className="aigallery__img"
								/>
							)
							: (
								<div className="aigallery__img-placeholder">
									🖼 No drawing
								</div>
							)
						}

						{result.description && (
							<p className="aigallery__description">{result.description}</p>
						)}

						<div className="aigallery__score">
							{renderStars(result.score)}
						</div>
					</div>
				))}
			</div>

			<button className="aigallery__btn" onClick={onBack}>
				← Back to home
			</button>
		</div>
	);
};

export default AIGallery;
