/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   AIVotePanel.jsx                                    :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/03/29 23:34:51 by mforest-          #+#    #+#             */
/*   Updated: 2026/03/29 23:34:51 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React, { useState } from 'react';
import './AIVotePanel.css';

const AIVotePanel = ({ drawings, myId, onDone }) =>
{
	const [votes, setVotes] = useState({});

	const votableDrawings = (drawings || [])
		.map((d, index) =>
		{
			const uniqueId = d.player_id ?? `draw-${index}`;
			const name = d.player_name ?? 'Player';
			const drawing = d.drawing ?? '';
			const title = d.title ?? '';
			const description = d.description ?? '';

			return { ...d, uniqueId, name, drawing, title, description };
		})
		.filter(d => d.uniqueId !== myId);

	const handleScoreClick = (targetId, score) =>
	{
		setVotes(prev => ({ ...prev, [targetId]: score }));
	};

	const votedCount  = Object.keys(votes).length;
	const totalToVote = votableDrawings.length;
	const isComplete  = totalToVote > 0 && votedCount === totalToVote;

	return (
		<div className="aivote">
			<div className="aivote__header">
				<h2 className="aivote__title">⭐ Rate each drawing</h2>
				<span className="aivote__hint">{votedCount} / {totalToVote} rated</span>
			</div>

			<div className="aivote__grid">
				{votableDrawings.map((d) => (
					<div
						key={d.uniqueId}
						className={`aivote__card${votes[d.uniqueId] ? ' aivote__card--voted' : ''}`}
					>
						<div className="aivote__card-label">
							👤 {d.name}
							{d.title && <span className="aivote__drawing-title"> — {d.title}</span>}
						</div>

						{d.description && (
							<p className="aivote__description">{d.description}</p>
						)}

						<img
							src={d.drawing}
							alt={`Drawing by ${d.name}`}
							className="aivote__img"
						/>

						<div className="aivote__score-row">
							{[1, 2, 3, 4, 5].map((num) => (
								<button
									key={num}
									type="button"
									className={`aivote__score-btn${votes[d.uniqueId] === num ? ' aivote__score-btn--active' : ''}`}
									onClick={() => handleScoreClick(d.uniqueId, num)}
								>
									{num}
								</button>
							))}
						</div>

						{votes[d.uniqueId] && (
							<div className="aivote__selected-score">
								{'★'.repeat(votes[d.uniqueId])}{'☆'.repeat(5 - votes[d.uniqueId])}
							</div>
						)}
					</div>
				))}
			</div>

			<button
				className={`aivote__submit${isComplete ? '' : ' aivote__submit--disabled'}`}
				disabled={!isComplete}
				onClick={() => onDone(votes)}
			>
				{isComplete ? '✓ Confirm my votes!' : `Rate all drawings first (${votedCount}/${totalToVote})`}
			</button>
		</div>
	);
};

export default AIVotePanel;
