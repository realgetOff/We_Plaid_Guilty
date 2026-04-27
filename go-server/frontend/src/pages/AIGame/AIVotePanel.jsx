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
import '../Game/WritePrompt.css';

const AIVotePanel = ({ drawings, myId, onDone }) =>
{
	const [votes, setVotes] = useState({});

	const votableDrawings = (drawings || [])
		.map((d, index) => ({
			...d,
			uniqueId: d.player_id || d.playerId || d.PlayerID || `draw-${index}`
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
		<div className="writeprompt">
			<div className="writeprompt__card">
				<div className="writeprompt__card-header">
					⭐ Rate each drawing
				</div>
				<div className="writeprompt__card-body">
					<p className="writeprompt__hint">
						Vote from 1 to 5 for each drawing to validate ({votedCount}/{totalToVote})
					</p>
				</div>
			</div>

			{votableDrawings.map((d) => (
				<div key={d.uniqueId} className="writeprompt__card">
					<div className="writeprompt__card-header">
						👤 {d.name || d.PlayerName || "Anonymous"}
						{d.title && ` - ${d.title}`}
					</div>
					<div className="writeprompt__card-body">
						{d.description && (
							<p className="writeprompt__hint">
								{d.description}
							</p>
						)}
						<div className="writeprompt__drawing-container">
							<img
								src={d.drawing}
								alt="drawing"
								className="writeprompt__image"
							/>
						</div>
						<div className="writeprompt__vote-group">
							{[1, 2, 3, 4, 5].map((num) => (
								<button
									key={num}
									type="button"
									className={`writeprompt__btn${votes[d.uniqueId] === num ? ' writeprompt__btn--active' : ''}`}
									onClick={() => handleScoreClick(d.uniqueId, num)}
								>
									{num}
								</button>
							))}
						</div>
					</div>
				</div>
			))}

			<button
				className="writeprompt__btn"
				disabled={!isComplete}
				onClick={() => onDone(votes)}
			>
				{isComplete ? "✓ Confirm my votes!" : "Rate all drawings first"}
			</button>
		</div>
	);
};

export default AIVotePanel;
