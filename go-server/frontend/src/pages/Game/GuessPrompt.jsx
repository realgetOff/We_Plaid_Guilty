/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   GuessPrompt.jsx                                    :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/02/23 23:45:02 by mforest-          #+#    #+#             */
/*   Updated: 2026/02/23 23:45:02 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React, { useState, useEffect, useRef } from 'react';
import './GuessPrompt.css';

const TIMER_SEC = 44;

const GuessPrompt = ({ drawing, onDone }) =>
{
	const [guess,   setGuess]   = useState('');
	const [error,   setError]   = useState('');
	const [seconds, setSeconds] = useState(TIMER_SEC);

	const onDoneRef = useRef(onDone);
	useEffect(() => { onDoneRef.current = onDone; }, [onDone]);

	useEffect(() =>
	{
		if (seconds <= 0)
		{
			onDoneRef.current(guess.trim() || '');
			return;
		}
		const id = setTimeout(() => setSeconds((s) => s - 1), 1000);
		return () => clearTimeout(id);
	}, [seconds, guess]);

	const handleSubmit = () =>
	{
		if (!guess.trim())
		{
			setError('Please write your guess before submitting.');
			return;
		}
		if (guess.trim().length < 2)
		{
			setError('Guess must be at least 2 characters.');
			return;
		}
		onDone(guess.trim());
	};

	const pct             = (seconds / TIMER_SEC) * 100;
	const timerBackground = seconds <= 10 ? '#aa0000' : '#000000';

	return (
		<div className="guessprompt">
			<div className="guessprompt__timer-track">
				<div className="guessprompt__timer-fill" style={{ width: `${pct}%`, background: timerBackground }} />
				<span className="guessprompt__timer-label">{seconds}s</span>
			</div>
			<div className="guessprompt__card">
				<div className="guessprompt__card-header">🔍 What is this drawing?</div>
				<div className="guessprompt__drawing-area">
					{drawing
						? <img src={drawing} alt="Drawing to guess" className="guessprompt__img" />
						: <span className="guessprompt__placeholder">No drawing received.</span>
					}
				</div>
			</div>
			<div className="guessprompt__card">
				<div className="guessprompt__card-header">✏ Your Guess</div>
				<div className="guessprompt__card-body">
					<input
						className={`guessprompt__input${error ? ' guessprompt__input--error' : ''}`}
						type="text"
						value={guess}
						onChange={(e) => { setGuess(e.target.value); setError(''); }}
						onKeyDown={(e) => { if (e.key === 'Enter') handleSubmit(); }}
						placeholder="What do you see?"
						maxLength={60}
						autoFocus
					/>
					{error && <p className="guessprompt__error">⚠ {error}</p>}
					<div className="guessprompt__footer-row">
						<span className="guessprompt__char-count">{guess.length} / 60</span>
						<button className="guessprompt__btn" onClick={handleSubmit}>
							✓ Submit Guess
						</button>
					</div>
				</div>
			</div>
		</div>
	);
};

export default GuessPrompt;
