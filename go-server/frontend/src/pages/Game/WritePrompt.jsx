/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   WritePrompt.jsx                                    :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/02/23 23:33:00 by mforest-          #+#    #+#             */
/*   Updated: 2026/02/23 23:33:00 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React, { useState, useEffect, useRef } from 'react';
import './WritePrompt.css';

const TIMER_SEC = 44;

const WritePrompt = ({ onDone }) =>
{
	const [text,    setText]    = useState('');
	const [error,   setError]   = useState('');
	const [seconds, setSeconds] = useState(TIMER_SEC);

	const onDoneRef = useRef(onDone);
	useEffect(() => { onDoneRef.current = onDone; }, [onDone]);

	useEffect(() =>
	{
		if (seconds <= 0)
		{
			onDoneRef.current(text.trim());
			return;
		}
		const id = setTimeout(() => setSeconds((s) => s - 1), 1000);
		return () => clearTimeout(id);
	}, [seconds, text]);

	const handleSubmit = () =>
	{
		if (!text.trim())
		{
			setError('Please write something before submitting.');
			return;
		}
		if (text.trim().length < 3)
		{
			setError('Prompt must be at least 3 characters.');
			return;
		}
		onDone(text.trim());
	};

	const pct             = (seconds / TIMER_SEC) * 100;
	const timerBackground = seconds <= 10 ? '#aa0000' : '#000000';

	return (
		<div className="writeprompt">
			<div className="writeprompt__timer-track">
				<div className="writeprompt__timer-fill" style={{ width: `${pct}%`, background: timerBackground }} />
				<span className="writeprompt__timer-label">{seconds}s</span>
			</div>
			<div className="writeprompt__card">
				<div className="writeprompt__card-header">✏ Write a Prompt</div>
				<div className="writeprompt__card-body">
					<p className="writeprompt__hint">
						Write a word or short phrase. The next player will have to draw it!
					</p>
					<input
						className={`writeprompt__input${error ? ' writeprompt__input--error' : ''}`}
						type="text"
						value={text}
						onChange={(e) => { setText(e.target.value); setError(''); }}
						onKeyDown={(e) => { if (e.key === 'Enter') handleSubmit(); }}
						placeholder="e.g. a cat riding a skateboard"
						maxLength={60}
						autoFocus
					/>
					{error && <p className="writeprompt__error">⚠ {error}</p>}
					<div className="writeprompt__footer-row">
						<span className="writeprompt__char-count">{text.length} / 60</span>
						<button className="writeprompt__btn" onClick={handleSubmit}>
							✓ Submit Prompt
						</button>
					</div>
				</div>
			</div>
		</div>
	);
};

export default WritePrompt;
