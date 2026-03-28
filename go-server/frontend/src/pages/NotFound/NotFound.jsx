/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   NotFound.jsx                                       :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/03/14 17:12:48 by mforest-          #+#    #+#             */
/*   Updated: 2026/03/14 17:12:48 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import { useState, useRef, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import './NotFound.css';

const THEME_SKI = 'https://i.imgur.com/6U5Ia1d.jpeg';
const THEME_BORDEL = 'https://i.imgur.com/q1Oc3ml.jpeg';
const THEME_APP = 'https://i.imgur.com/LvfglKJ.jpeg';
const THEME_PORTAIL = 'https://i.imgur.com/LkSidoO.jpeg';
const THEME_DOMINOS = 'https://i.imgur.com/hwtTy0y.jpeg';

const LVIRAVON_SKI = 'https://i.imgur.com/m6TunQh.jpeg';
const LVIRAVON_CHARLINE = 'https://i.imgur.com/n8zbDrV.jpeg';
const LVIRAVON_PORTAIL = 'https://i.imgur.com/1CUD4JE.jpeg';
const LVIRAVON_DOMINOS = 'https://i.imgur.com/Ss8LOAb.jpeg';
const LVIRAVON_BORDEL = 'https://i.imgur.com/ctagzbN.jpeg';

const TOLERANCE = 3;

const ENDING_GIF_URL = 'https://i.makeagif.com/media/3-16-2026/a4BZaB.gif';
const GIF_DURATION_MS = 7850;

const LEVEL_DATA = [
	{
		theme: THEME_SKI,
		img: LVIRAVON_SKI,
		text: "you fool, you've found lviravon clone !",
		quote: "<lviravon> Mais les gars, c'est pas de ma faute c'est impossible de travailler on capte pas !",
		sub: "<Slaves of lviravon> bullshit...",
		targetX: 78.31,
		targetY: 28.69
	},
	{
		theme: THEME_APP,
		img: LVIRAVON_CHARLINE,
		text: "you fool, you've found lviravon clone !",
		quote: "<lviravon> Je suis trop fatiguer, je vais me reposer ce soir, je viens demain",
		sub: "<Slaves of lviravon> bullshit...",
		targetX: 52.65,
		targetY: 59.06
	},
	{
		theme: THEME_PORTAIL,
		img: LVIRAVON_PORTAIL,
		text: "you fool, you've found lviravon clone !",
		quote: "<lviravon> Mais euhhh c'est pas de ma faute Nathan, le portail est bloquer, a un moment faut arreter !",
		sub: "<Slaves of lviravon> bullshit...",
		targetX: 49.07,
		targetY: 73.39
	},
	{
		theme: THEME_DOMINOS,
		img: LVIRAVON_DOMINOS,
		text: "you fool, you've found lviravon clone !",
		quote: "<lviravon> Ahhh, ca risque d'etre tendu faut que je bosse ce soir, mais je viens apres, sans faute !",
		sub: "<Slaves of lviravon> bullshit...",
		targetX: 36.51,
		targetY: 58.49
	},
	{
		theme: THEME_BORDEL,
		img: LVIRAVON_BORDEL,
		text: "you fool, you've found lviravon clone !",
		quote: "<lviravon> Je range la maison mais apres je viens.",
		sub: "<Slaves of lviravon> bullshit...",
		targetX: 55.29,
		targetY: 24.47
	}
];

const NotFound = () =>
{
	const imgRef = useRef(null);
	const timerRef = useRef(null);

	const [found, setFound] = useState(false);
	const [elapsed, setElapsed] = useState(0);
	const [started, setStarted] = useState(false);
	const [miss, setMiss] = useState(null);
	const [showHint, setShowHint] = useState(false);
	const [showEndingGif, setShowEndingGif] = useState(false);

	const [winCount, setWinCount] = useState(() =>
	{
		const savedWins = localStorage.getItem('lviravon_wins');
		if (savedWins !== null)
			return (parseInt(savedWins, 10));
		return (0);
	});

	const currentLevelIndex = winCount % LEVEL_DATA.length;
	const currentLevel = LEVEL_DATA[currentLevelIndex];

	useEffect(() =>
	{
		const hasCredit = localStorage.getItem('loris_credit');
		if (hasCredit === null)
			localStorage.setItem('loris_credit', 'true');
	}, []);

	useEffect(() =>
	{
		if (started)
		{
			if (!found)
				timerRef.current = setInterval(() => setElapsed((e) => e + 1), 1000);
		}
		return (() => clearInterval(timerRef.current));
	}, [started, found]);

	const navigate = useNavigate();
	useEffect(() =>
	{
		if (showEndingGif)
		{
			const timer = setTimeout(() =>
			{
				setShowEndingGif(false);
				setWinCount(0);
				localStorage.setItem('lviravon_wins', 0);
				localStorage.setItem('loris_credit', 'null');
				navigate('/credits');
			}, GIF_DURATION_MS);

			return (() => clearTimeout(timer));
		}
	}, [showEndingGif]);

	const handleClick = (e) =>
	{
		if (found)
			return;
		if (!started)
			setStarted(true);

		const rect = imgRef.current.getBoundingClientRect();
		const clickX = ((e.clientX - rect.left) / rect.width) * 100;
		const clickY = ((e.clientY - rect.top) / rect.height) * 100;

		const dx = clickX - currentLevel.targetX;
		const dy = clickY - currentLevel.targetY;
		const dist = Math.sqrt(dx * dx + dy * dy);

		if (dist <= TOLERANCE)
		{
			clearInterval(timerRef.current);
			setFound(true);
			setMiss(null);
		}
		else
		{
			setMiss({
				x: e.clientX - rect.left,
				y: e.clientY - rect.top,
				id: Date.now()
			});
		}
	};

	const handleRestart = () =>
	{
		clearInterval(timerRef.current);
		setFound(false);
		setElapsed(0);
		setStarted(false);
		setMiss(null);
		setShowHint(false);
	};

	const handleNextPhoto = () =>
	{
		handleRestart();
		setWinCount((prev) =>
		{
			const next = prev + 1;
			if (next >= LEVEL_DATA.length)
			{
				setShowEndingGif(true);
				return (prev);
			}
			localStorage.setItem('lviravon_wins', next);
			return (next);
		});
	};

	let minutes = String(Math.floor(elapsed / 60)).padStart(2, '0');
	let seconds = String(elapsed % 60).padStart(2, '0');

	let timerClass = "find__timer";
	let timerPrefix = "";
	if (found)
	{
		timerClass = "find__timer find__timer--found";
		timerPrefix = "✓ ";
	}

	let zoneClass = "find__zone";
	if (!found)
	{
		zoneClass = "find__zone find__zone--hunting";
	}

	let timerElement = null;
	if (started)
	{
		timerElement = (
			<span className={timerClass}>
				{timerPrefix}{minutes}:{seconds}
			</span>
		);
	}

	let missElement = null;
	if (miss)
	{
		missElement = (
			<div
				key={miss.id}
				className="find__miss"
				style={{ left: miss.x - 14, top: miss.y - 14 }}
			/>
		);
	}

	let winElement = null;
	if (found)
	{
		winElement = (
			<div className="find__win">
				<img
					src={currentLevel.img}
					alt="lviravon"
					className="find__win-avatar"
				/>
				<div className="find__win-text">
					<p className="find__win-title">
						<b>[{currentLevel.text}]</b>
					</p>

					<div className="find__win-body">
						<p className="find__win-quote">
							{currentLevel.quote}
						</p>
						<p className="find__win-sub">
							{currentLevel.sub}
						</p>
					</div>

					<p className="find__win-stats">
						time : {minutes}:{seconds}
					</p>
				</div>
				<div className="find__win-actions">
					<button className="find__btn" onClick={() => setShowHint(true)}>
						Laplace's demon
					</button>
					<button className="find__btn" onClick={handleNextPhoto}>
						try again
					</button>
				</div>
			</div>
		);
	}

	let hintElement = null;
	if (!started)
	{
		if (!found)
		{
			hintElement = (
				<p className="find__hint">click anywhere on the image to start</p>
			);
		}
	}

	let popupElement = null;
	if (showHint)
	{
		popupElement = (
			<div className="find__popup-overlay" onClick={() => setShowHint(false)}>
				<div className="find__popup" onClick={(e) => e.stopPropagation()}>
					<p className="find__popup-text">lviravon won't come, again... 😔</p>
					<button className="find__btn" onClick={() => setShowHint(false)}>
						fermer
					</button>
				</div>
			</div>
		);
	}

	let gifOverlayElement = null;
	if (showEndingGif)
	{
		gifOverlayElement = (
			<div className="find__gif-overlay">
				<img src={ENDING_GIF_URL} alt="the end lol" className="find__gif" />
			</div>
		);
	}

	return (
		<div className="find">
			<div className="find__header">
				<span className="find__title">🔍 find lviravon</span>
				<div className="find__header-right">
					{timerElement}
					<button className="find__btn" onClick={handleRestart}>
						↺ restart timer
					</button>
				</div>
			</div>

			<div className={zoneClass} onClick={handleClick}>
				<img
					ref={imgRef}
					src={currentLevel.theme}
					alt="foule"
					className="find__img"
					draggable={false}
				/>
				{missElement}
			</div>

			{winElement}
			{hintElement}
			{popupElement}
			{gifOverlayElement}
		</div>
	);
};

export default NotFound;
