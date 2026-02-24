/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   CreateGame.jsx                                     :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/02/23 21:11:46 by mforest-          #+#    #+#             */
/*   Updated: 2026/02/23 21:11:46 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React, { useState, useEffect } from 'react';
import { useNavigate }                from 'react-router-dom';
import './CreateGame.css';

const generateCode = () => {
  const chars = 'ABCDEFGHJKLMNPQRSTUVWXYZ';
  return Array.from({ length: 6 }, () =>
    chars[Math.floor(Math.random() * chars.length)]
  ).join('');
};

/* TODO: Joueurs mock, a remplacer par les donnees WebSocket */
const MOCK_PLAYERS = [
  { id: 1, name: 'mforest-', host: true },
];

const CreateGame = () => {
  const navigate = useNavigate();

  const [roomCode, setRoomCode]   = useState('');
  const [copied,   setCopied]     = useState(false);
  const [players,  setPlayers]    = useState(MOCK_PLAYERS);
  const [rounds,   setRounds]     = useState(3);
  const [timer,    setTimer]      = useState(60);

  useEffect(() => {
    setRoomCode(generateCode());
    /* TODO: appeler l'API backend pour creer la room et recevoir le vrai code */
  }, []);

  const handleCopy = () => {
    navigator.clipboard.writeText(roomCode);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleStart = () => {
    if (players.length < 2) return;
    /* TODO: envoyer l'evenement WebSocket "start_game" */
    navigate('/game/play');
  };

  const handleLeave = () => {
    /* TODO: quitter la room cote WebSocket */
    navigate('/game');
  };

  return (
    <div className="creategame">

      {/* ── Room code ── */}
      <div className="creategame__card">
        <div className="creategame__card-header">🔑 Room Code</div>
        <div className="creategame__card-body creategame__card-body--center">
          <p className="creategame__hint">
            Share this code with your friends so they can join.
          </p>
          <div className="creategame__code-row">
            <span className="creategame__code">{roomCode}</span>
            <button className="creategame__btn creategame__btn--copy" onClick={handleCopy}>
              {copied ? '✓ Copied!' : '⎘ Copy'}
            </button>
          </div>
        </div>
      </div>

      <div className="creategame__columns">

        {/* ── Players ── */}
        <div className="creategame__card creategame__card--grow">
          <div className="creategame__card-header">
            👥 Players
            <span className="creategame__card-header-count">
              {players.length} / 8
            </span>
          </div>
          <div className="creategame__card-body creategame__card-body--list">
            {players.map((p) => (
              <div key={p.id} className="creategame__player-row">
                <span className="creategame__player-dot" />
                <span className="creategame__player-name">{p.name}</span>
                {p.host && (
                  <span className="creategame__badge">HOST</span>
                )}
              </div>
            ))}
            {players.length < 2 && (
              <p className="creategame__waiting">
                ⧗ Waiting for players…
              </p>
            )}
          </div>
        </div>

        {/* ── Settings ── */}
        <div className="creategame__card creategame__card--grow">
          <div className="creategame__card-header">⚙ Settings</div>
          <div className="creategame__card-body">

            <label className="creategame__label">
              Rounds
              <select
                className="creategame__select"
                value={rounds}
                onChange={(e) => setRounds(Number(e.target.value))}
              >
                {[2, 3, 4, 5, 6].map((n) => (
                  <option key={n} value={n}>{n} rounds</option>
                ))}
              </select>
            </label>

            <label className="creategame__label">
              Draw timer
              <select
                className="creategame__select"
                value={timer}
                onChange={(e) => setTimer(Number(e.target.value))}
              >
                {[30, 45, 60, 90, 120].map((n) => (
                  <option key={n} value={n}>{n} seconds</option>
                ))}
              </select>
            </label>

          </div>
        </div>

      </div>

      {/* ── Actions ── */}
      <div className="creategame__actions">
        <button className="creategame__btn creategame__btn--leave" onClick={handleLeave}>
          ✕ Leave Room
        </button>
        <button
          className="creategame__btn creategame__btn--start"
          onClick={handleStart}
          disabled={players.length < 2}
          title={players.length < 2 ? 'Need at least 2 players' : ''}
        >
          ▶ Start Game
        </button>
      </div>

      {players.length < 2 && (
        <p className="creategame__start-hint">
          ⚠ At least 2 players are required to start.
        </p>
      )}

    </div>
  );
};

export default CreateGame;
