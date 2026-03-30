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
		.map((d, index) => ({
			...d,
			uniqueId: d.playerId || d.PlayerID || `draw-${index}`
		}))
		.filter(d => d.uniqueId !== myId);

	const handleScoreClick = (targetId, score) =>
	{
		setVotes(prev => ({
			...prev,
			[targetId]: score
		}));
	};

	const votedCount = Object.keys(votes).length;
	const totalToVote = votableDrawings.length;
	const isComplete = totalToVote > 0 && votedCount === totalToVote;

	return (
		<div className="aivote">
			<div className="aivote__hint">
				Notez chaque dessin de 1 à 5 pour valider (<strong>{votedCount}/{totalToVote}</strong>)
			</div>

			<div className="aivote__grid">
				{votableDrawings.map((d) => (
					<div key={d.uniqueId} className="aivote__card">
						<span className="aivote__card-label">{d.name || d.PlayerName || "Anonyme"}</span>
						<img src={d.drawing} alt="drawing" className="aivote__img" />

						<div className="aivote__score-row">
							{[1, 2, 3, 4, 5].map((num) => (
								<button
									key={num}
									type="button"
									className={`aivote__score-btn ${votes[d.uniqueId] === num ? 'aivote__score-btn--active' : ''}`}
									onClick={() => handleScoreClick(d.uniqueId, num)}
								>
									{num}
								</button>
							))}
						</div>
					</div>
				))}
			</div>

			<button
				className="aivote__submit"
				disabled={!isComplete}
				onClick={() => onDone(votes)}
			>
				{isComplete ? "Confirm my votes!" : "Choose one vote for each draw"}
			</button>
		</div>
	);
};

export default AIVotePanel;
