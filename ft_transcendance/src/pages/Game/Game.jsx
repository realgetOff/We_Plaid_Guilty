/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   Game.jsx                                           :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/02/23 23:31:46 by mforest-          #+#    #+#             */
/*   Updated: 2026/02/23 23:31:46 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React, { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import WritePrompt from './WritePrompt';
import DrawBoard from './DrawBoard';
import GuessPrompt from './GuessPrompt';
import './Game.css';

// TODO: remplacer par les donnees de l'api
const MOCK_ROOMS =
{
  'ABCDEF': { status: 'started'  },
  'ZZZZZZ': { status: 'waiting'  },
  'FINISH': { status: 'finished' },
};

const DENY_REASONS =
{
  invalid:   'invalid room code format.',
  not_found: 'room not found.',
  waiting:   'this game has not started yet.',
  finished:  'this game is already finished.',
  unknown:   'cannot join this room.',
};

const Game = () =>
{
  const { code } = useParams();
  const navigate = useNavigate();

  const [status,  setStatus]  = useState('checking');
  const [deny,    setDeny]    = useState('');
  const [phase,   setPhase]   = useState('write');
  const [prompt,  setPrompt]  = useState('');
  const [drawing, setDrawing] = useState(null);

  useEffect(() =>
  {
    const check = async () =>
    {
      const normalized = code?.toUpperCase();

      if (!normalized || !/^[A-Z]{6}$/.test(normalized))
      {
        setDeny(DENY_REASONS.invalid);
        setStatus('denied');
        return;
      }

      // TODO: remplacer par fetch(`/api/rooms/${normalized}`)
      await new Promise((r) => setTimeout(r, 400));
      const room = MOCK_ROOMS[normalized];

      if (!room)
      {
        setDeny(DENY_REASONS.not_found);
        setStatus('denied');
        return;
      }
      if (room.status === 'waiting')
      {
        setDeny(DENY_REASONS.waiting);
        setStatus('denied');
        return;
      }
      if (room.status === 'finished')
      {
        setDeny(DENY_REASONS.finished);
        setStatus('denied');
        return;
      }
      if (room.status !== 'started')
      {
        setDeny(DENY_REASONS.unknown);
        setStatus('denied');
        return;
      }

      // TODO: verifier que l'user connecte est bien dans cette room
      setStatus('playing');
    };

    check();
  }, [code]);

  const handlePromptDone = (text) =>
  {
    setPrompt(text);
    setPhase('draw');
  };

  const handleDrawDone = (dataURL) =>
  {
    setDrawing(dataURL);
    setPhase('guess');
  };

  const handleGuessDone = (_guess) =>
  {
    // TODO: envoyer les resultats au backend, passer au round suivant
    setPhase('write');
  };

  let phaseLabel = '';
  if (phase === 'write') phaseLabel = '✏ write a prompt';
  if (phase === 'draw')  phaseLabel = '🎨 draw it!';
  if (phase === 'guess') phaseLabel = '🔍 what is it?';

  if (status === 'checking')
  {
    return (
      <div className="game__guard">
        <span className="game__guard-spinner">⧗</span>
        verifying room <strong>{code?.toUpperCase()}</strong>…
      </div>
    );
  }

  if (status === 'denied')
  {
    return (
      <div className="game__guard">
        <div className="game__guard-card">
          <div className="game__guard-icon">✕</div>
          <p className="game__guard-msg">⚠ {deny}</p>
          <div className="game__guard-actions">
            <button
              className="game__guard-btn"
              onClick={() => navigate('/game')}
            >
              ← back to home
            </button>
            {deny === DENY_REASONS.waiting && (
              <button
                className="game__guard-btn game__guard-btn--primary"
                onClick={() => navigate(`/game/lobby/${code?.toUpperCase()}`)}
              >
                → go to lobby
              </button>
            )}
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="game">
      <div className="game__phase-bar">
        <span className={`game__phase-dot${phase === 'write' ? ' game__phase-dot--on' : ''}`} />
        <span className={`game__phase-dot${phase === 'draw'  ? ' game__phase-dot--on' : ''}`} />
        <span className={`game__phase-dot${phase === 'guess' ? ' game__phase-dot--on' : ''}`} />
        <span className="game__phase-label">{phaseLabel}</span>
        <span className="game__room-code">#{code?.toUpperCase()}</span>
      </div>

      {phase === 'write' && <WritePrompt onDone={handlePromptDone} />}
      {phase === 'draw'  && <DrawBoard   prompt={prompt} onDone={handleDrawDone} />}
      {phase === 'guess' && <GuessPrompt drawing={drawing} onDone={handleGuessDone} />}
    </div>
  );
};

export default Game;
