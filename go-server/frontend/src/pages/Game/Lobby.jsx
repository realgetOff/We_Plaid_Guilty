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
import { connect, send, addListener, removeListener } from '../../socket';
import { roomsApi } from '../../api/rooms';
import './Lobby.css';

const DENY_REASONS =
{
  invalid:   'invalid room code format.',
  not_found: 'room not found.',
  started:   'this game has already started.',
  finished:  'this game is already finished.',
  unknown:   'cannot access this room.',
};

const Lobby = () =>
{
  const { code } = useParams();
  const navigate = useNavigate();
  const msgEndRef = useRef(null);

  const [status,    setStatus]    = useState('checking');
  const [deny,      setDeny]      = useState('');
  const [players,   setPlayers]   = useState([]);
  const [messages,  setMessages]  = useState([]);
  const [input,     setInput]     = useState('');
  const [countdown, setCountdown] = useState(null);
  const [isHost,    setIsHost]    = useState(false);
  const [myName,    setMyName]    = useState('');

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

      let room;
      try
      {
        room = await roomsApi.getRoom(normalized);
      }
      catch
      {
        setDeny(DENY_REASONS.not_found);
        setStatus('denied');
        return;
      }

      if (!room || !room.status)
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

      if (Array.isArray(room.players))
        setPlayers(room.players);
      setStatus('ready');
    };

    check();
  }, [code]);

  // websocket: rejoindre le lobby et ecouter les mises a jour
  useEffect(() =>
  {
    const normalized = code?.toUpperCase();
    if (!normalized || !/^[A-Z]{6}$/.test(normalized))
      return;

    connect();

    const playerName = 'guest-' + Math.floor(Math.random() * 10000);

    const handler = (msg) =>
    {
      if (!msg || msg.room !== normalized)
        return;

      if (msg.type === 'lobby_state')
      {
        if (Array.isArray(msg.players))
          setPlayers(msg.players);
        if (Array.isArray(msg.messages))
          setMessages(msg.messages);
        if (msg.me)
        {
          if (typeof msg.me.host === 'boolean')
            setIsHost(msg.me.host);
          if (msg.me.name)
            setMyName(msg.me.name);
        }
      }

      if (msg.type === 'player_joined')
      {
        if (msg.player)
        {
          setPlayers((prev) =>
          {
            const exists = prev.some((p) => p.id === msg.player.id);
            if (exists)
              return prev;
            return [...prev, msg.player];
          });
        }
      }

      if (msg.type === 'player_left' && msg.playerId !== undefined)
      {
        setPlayers((prev) => prev.filter((p) => p.id !== msg.playerId));
      }

      if (msg.type === 'chat_message')
      {
        if (msg.text && msg.user)
        {
          setMessages((prev) => [
            ...prev,
            {
              id: msg.id ?? Date.now(),
              user: msg.user,
              text: msg.text,
            },
          ]);
        }
      }

      if (msg.type === 'start_game')
      {
        navigate(`/game/play/${normalized}`);
      }

      if (msg.type === 'lobby_denied')
      {
        setDeny(msg.reason || DENY_REASONS.unknown);
        setStatus('denied');
      }
    };

    addListener(handler);

    send({
      type: 'join_lobby',
      room: normalized,
      name: playerName,
    });

    return () =>
    {
      removeListener(handler);
      send({
        type: 'leave_lobby',
        room: normalized,
      });
    };
  }, [code, navigate]);

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

  const handleSend = () =>
  {
    const text = input.trim();
    if (!text)
      return;
    setInput('');
    send({
      type: 'chat_message',
      room: code?.toUpperCase(),
      text,
    });
  };

  const handleStart = () =>
  {
    if (players.length < 2)
      return;
    send({
      type: 'start_game',
      room: code?.toUpperCase(),
    });
    setCountdown(5);
  };

  const handleLeave = () =>
  {
    send({
      type: 'leave_lobby',
      room: code?.toUpperCase(),
    });
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
              if (myName && m.user === myName)
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
            <textarea
              className="lobby__chat-input"
              value={input}
              onChange={(e) => setInput(e.target.value.slice(0, 80))}
              onKeyDown={(e) =>
              {
                if (e.key === 'Enter' && !e.shiftKey)
                {
                  e.preventDefault();
                  handleSend();
                }
              }}
              placeholder="say something…"
              maxLength={80}
              rows={2}
              wrap="soft"
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
