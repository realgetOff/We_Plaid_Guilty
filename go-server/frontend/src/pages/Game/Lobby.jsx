/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   Lobby.jsx                                          :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/02/23 23:30:46 by mforest-          #+#    #+#             */
/*   Updated: 2026/02/23 23:30:46 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React, { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import './Lobby.css';

// TODO: remplacer par les donnees de l'api
const MOCK_ROOMS =
{
  'ABCDEF': { status: 'waiting',  players: [{ id: 1, name: 'mforest-', host: true }, { id: 2, name: 'lviravon', host: false }] },
  'ZZZZZZ': { status: 'started',  players: [] },
  'FINISH': { status: 'finished', players: [] },
};

const DENY_REASONS =
{
  invalid:   'invalid room code format.',
  not_found: 'room not found.',
  started:   'this game has already started.',
  finished:  'this game is already finished.',
  unknown:   'cannot access this room.',
};

const MOCK_MESSAGES = [
  { id: 1, user: 'lviravon', text: 'ready when you are' },
];

const Lobby = () =>
{
  const { code } = useParams();
  const navigate = useNavigate();
  const msgEndRef = useRef(null);

  const [status,    setStatus]    = useState('checking');
  const [deny,      setDeny]      = useState('');
  const [players,   setPlayers]   = useState([]);
  const [messages,  setMessages]  = useState(MOCK_MESSAGES);
  const [input,     setInput]     = useState('');
  const [countdown, setCountdown] = useState(null);
  const isHost = true; // TODO: comparer avec l'user connecte

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
      if (room.status === 'started')
      {
        setDeny(DENY_REASONS.started);
        setStatus('denied');
        return;
      }
      if (room.status === 'finished')
      {
        setDeny(DENY_REASONS.finished);
        setStatus('denied');
        return;
      }
      if (room.status !== 'waiting')
      {
        setDeny(DENY_REASONS.unknown);
        setStatus('denied');
        return;
      }

      // TODO: verifier que l'user connecte est bien dans cette room
      setPlayers(room.players);
      setStatus('ready');
    };

    check();
  }, [code]);

  useEffect(() =>
  {
    msgEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  useEffect(() =>
  {
    if (countdown === null)
      return;
    if (countdown === 0)
    {
      navigate(`/game/play/${code?.toUpperCase()}`);
      return;
    }
    const id = setTimeout(() => setCountdown((c) => c - 1), 1000);
    return () => clearTimeout(id);
  }, [countdown, code, navigate]);

  // FIXME: mettre une barre scroll pour eviter que la fenetre augmente en taille lors d'envoie de msg
  const handleSend = () =>
  {
    const text = input.trim();
    if (!text)
      return;
    setMessages((m) => [...m, { id: Date.now(), user: 'mforest-', text }]);
    setInput('');
    // TODO: websocket send 'chat_message'
  };

  const handleStart = () =>
  {
    if (players.length < 2)
      return;
    // TODO: websocket send 'start_game'
    setCountdown(5);
  };

  const handleLeave = () =>
  {
    // TODO: websocket send 'leave_room'
    navigate('/game');
  };

  if (status === 'checking')
  {
    return (
      <div className="lobby__guard">
        <span className="lobby__guard-spinner">⧗</span>
        verifying room <strong>{code?.toUpperCase()}</strong>…
      </div>
    );
  }

  if (status === 'denied')
  {
    let extraBtn = null;
    if (deny === DENY_REASONS.started)
      extraBtn = (
        <button
          className="lobby__guard-btn lobby__guard-btn--primary"
          onClick={() => navigate(`/game/play/${code?.toUpperCase()}`)}
        >
          → go to game
        </button>
      );

    return (
      <div className="lobby__guard">
        <div className="lobby__guard-card">
          <div className="lobby__guard-icon">✕</div>
          <p className="lobby__guard-msg">⚠ {deny}</p>
          <div className="lobby__guard-actions">
            <button
              className="lobby__guard-btn"
              onClick={() => navigate('/game')}
            >
              ← back to home
            </button>
            {extraBtn}
          </div>
        </div>
      </div>
    );
  }

  let startTitle = '';
  if (players.length < 2)
    startTitle = 'need at least 2 players';

  return (
    <div className="lobby">

      <div className="lobby__code-band">
        room code : <span className="lobby__code">{code?.toUpperCase()}</span>
      </div>

      <div className="lobby__columns">

        <div className="lobby__card lobby__card--players">
          <div className="lobby__card-header">
            👥 players
            <span className="lobby__header-count">{players.length} / 8</span>
          </div>
          <div className="lobby__player-list">
            {players.map((p) =>
            {
              return (
                <div key={p.id} className="lobby__player-row">
                  <span className="lobby__dot" />
                  <span className="lobby__player-name">{p.name}</span>
                  {p.host && <span className="lobby__badge">HOST</span>}
                </div>
              );
            })}
            {players.length < 2 && (
              <p className="lobby__waiting">⧗ waiting for more players…</p>
            )}
          </div>
        </div>

        <div className="lobby__card lobby__card--chat">
          <div className="lobby__card-header">💬 chat</div>
          <div className="lobby__chat-messages">
            {messages.map((m) =>
            {
              let cls = 'lobby__msg';
              if (m.user === 'mforest-')
                cls += ' lobby__msg--me';
              return (
                <div key={m.id} className={cls}>
                  <span className="lobby__msg-user">{m.user}:</span> {m.text}
                </div>
              );
            })}
            <div ref={msgEndRef} />
          </div>
          <div className="lobby__chat-input-row">
            <input
              className="lobby__chat-input"
              type="text"
              value={input}
              onChange={(e) => setInput(e.target.value.slice(0, 120))}
              onKeyDown={(e) => e.key === 'Enter' && handleSend()}
              placeholder="say something…"
              maxLength={120}
            />
            <button className="lobby__chat-send" onClick={handleSend}>→</button>
          </div>
        </div>

      </div>

      {countdown !== null && (
        <div className="lobby__countdown">
          game starts in <span className="lobby__countdown-num">{countdown}</span>
        </div>
      )}

      <div className="lobby__actions">
        <button
          className="lobby__btn lobby__btn--leave"
          onClick={handleLeave}
        >
          ✕ leave
        </button>
        {isHost ? (
          <button
            className="lobby__btn lobby__btn--start"
            onClick={handleStart}
            disabled={players.length < 2 || countdown !== null}
            title={startTitle}
          >
            ▶ start game
          </button>
        ) : (
          <p className="lobby__host-hint">waiting for the host to start…</p>
        )}
      </div>

      {players.length < 2 && (
        <p className="lobby__min-hint">⚠ at least 2 players required.</p>
      )}

    </div>
  );
};

export default Lobby;
