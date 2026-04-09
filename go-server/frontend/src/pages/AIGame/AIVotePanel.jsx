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
				<div className="writeprompt__card-header">⭐ Rate each drawing</div>
				<div className="writeprompt__card-body">
					<p className="writeprompt__hint">
						Vote from 1 to 5 for each drawing to validate ({votedCount}/{totalToVote})
					</p>
				</div>
			</div>

			{votableDrawings.map((d) => (
				<div key={d.uniqueId} className="writeprompt__card" style={{marginTop: '10px'}}>
					<div className="writeprompt__card-header">
						👤 {d.name || d.PlayerName || "Anonymous"}
						{d.title && ` - ${d.title}`}
					</div>
					<div className="writeprompt__card-body">
						{d.description && (
							<p className="writeprompt__hint" style={{marginBottom: '8px'}}>
								{d.description}
							</p>
						)}
						<div style={{
							display: 'flex',
							justifyContent: 'center',
							marginBottom: '10px',
							background: '#f8f8f8',
							padding: '8px',
							border: '1px solid #cccccc'
						}}>
							<img
								src={d.drawing}
								alt="drawing"
								style={{
									maxWidth: '100%',
									maxHeight: '200px',
									border: '1px solid #000000',
									imageRendering: 'pixelated'
								}}
							/>
						</div>
						<div style={{
							display: 'flex',
							gap: '6px',
							justifyContent: 'center'
						}}>
							{[1, 2, 3, 4, 5].map((num) => (
								<button
									key={num}
									type="button"
									className={`writeprompt__btn${votes[d.uniqueId] === num ? ' writeprompt__btn--active' : ''}`}
									onClick={() => handleScoreClick(d.uniqueId, num)}
									style={{
										width: '40px',
										padding: '6px',
										fontSize: '12px',
										background: votes[d.uniqueId] === num ? '#000000' : '#ffffff',
										color: votes[d.uniqueId] === num ? '#ffffff' : '#000000'
									}}
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
				style={{
					marginTop: '16px',
					width: '100%',
					opacity: isComplete ? 1 : 0.4,
					cursor: isComplete ? 'pointer' : 'not-allowed'
				}}
			>
				{isComplete ? "✓ Confirm my votes!" : "Rate all drawings first"}
			</button>
		</div>
	);
};

export default AIVotePanel;
