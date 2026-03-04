/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   CreateGame.jsx                                     :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: pmilner- <pmilner-@student.42.fr>          +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/02/23 21:11:46 by mforest-          #+#    #+#             */
/*   Updated: 2026/03/04 20:42:31 by pmilner-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React, { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { connect, send, addListener, removeListener } from '../../socket';
import './CreateGame.css';

//   useEffect(() =>
//   {
//     const init = async () =>
//     {
//       // TODO: verifier que l'user est connecte
//       // TODO: remplacer par fetch('/api/rooms', { method: 'POST' }) et recevoir le vrai code
      
// 	  // did the todo myself
// 		const lobbyData =
// 		{
// 			hostId: "mforest-",
// 			settings: {
// 				rounds: 3,
// 				timer: 60
// 			}
// 		};

// 		try {
// 			const response = await fetch('/api/rooms', {
// 				method: 'POST',
// 				headers: {'Content-Type': 'application/json'},
// 				body: JSON.stringify(lobbyData)
// 			});
// 			if (!response.ok) {
// 				throw new Error(`Couldn't get the room code :: ${response.status}`);
// 			}

// 			const data = await response.json();
// 			console.log("room code =", data.lobbyCode);
// 			setRoomCode(data.lobbyCode);
// 			setStatus('ready');
// 			return ;
// 		} catch (error) {
// 			console.error("A back-end error occurred: ", error);
// 		}
// 	  //await new Promise((r) => setTimeout(r, 400));
//       //setRoomCode(generateCode());
//       //setStatus('ready');
//     };

//     init();
//   }, []);

const CreateGame = () =>
{
  const navigate = useNavigate();

  const [status,   setStatus]   = useState('checking');
  const [roomCode, setRoomCode] = useState('');
  const [copied,   setCopied]   = useState(false);
  const [players,  setPlayers]  = useState([]);
  const [rounds,   setRounds]   = useState(3);
  const [timer,    setTimer]    = useState(60);
  const [createErr, setCreateErr] = useState('');
  const roomCodeRef = useRef('');

  useEffect(() =>
  {
    roomCodeRef.current = roomCode;
  }, [roomCode]);

  useEffect(() =>
  {
    connect();

    const handler = (msg) =>
    {
      if (msg.type === 'room_created')
      {
        if (msg.code)
        {
          setRoomCode(msg.code);
          roomCodeRef.current = msg.code;
        }
        if (Array.isArray(msg.players))
          setPlayers(msg.players);
        setCreateErr('');
        setStatus('ready');
      }

      if (msg.type === 'create_denied')
      {
        setCreateErr(msg.reason || 'Could not create room.');
        setStatus('ready');
      }

      if (msg.type === 'player_joined' && msg.player)
      {
        setPlayers((prev) =>
        {
          const exists = prev.some((p) => p.id === msg.player.id);
          if (exists)
            return prev;
          return [...prev, msg.player];
        });
      }

      if (msg.type === 'player_left' && msg.playerId !== undefined)
      {
        setPlayers((prev) => prev.filter((p) => p.id !== msg.playerId));
      }
    };

    addListener(handler);

    send({
      type: 'create_room',
      rounds,
      timer,
    });

    return () =>
    {
      removeListener(handler);
      const code = roomCodeRef.current;
      if (code)
        send({ type: 'leave_lobby', room: code });
    };
  }, []);

  const handleCopy = () =>
  {
    navigator.clipboard.writeText(roomCode);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleStart = () =>
  {
    if (players.length < 2 || !roomCode)
      return;
    send({ type: 'start_game', room: roomCode });
    navigate(`/game/play/${roomCode}`);
  };

  const handleLeave = () =>
  {
    if (roomCode)
      send({ type: 'leave_lobby', room: roomCode });
    navigate('/game');
  };

  let copyLabel = '⎘ copy';
  if (copied)
    copyLabel = '✓ copied!';

  let startTitle = '';
  if (players.length < 2)
    startTitle = 'need at least 2 players';

  if (status === 'checking')
  {
    return (
      <div className="creategame__guard">
        <span className="creategame__guard-spinner">⧗</span>
        creating your room…
      </div>
    );
  }

  if (createErr && !roomCode)
  {
    return (
      <div className="creategame__guard">
        <div className="creategame__guard-card">
          <p className="creategame__guard-msg">⚠ {createErr}</p>
          <button
            className="creategame__guard-btn"
            onClick={() => navigate('/game')}
          >
            ← back to game
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="creategame">

      <div className="creategame__card">
        <div className="creategame__card-header">🔑 room code</div>
        <div className="creategame__card-body creategame__card-body--center">
          <p className="creategame__hint">
            share this code with your friends so they can join.
          </p>
          <div className="creategame__code-row">
            <span className="creategame__code">{roomCode}</span>
            <button
              className="creategame__btn creategame__btn--copy"
              onClick={handleCopy}
            >
              {copyLabel}
            </button>
          </div>
        </div>
      </div>

      <div className="creategame__columns">

        <div className="creategame__card creategame__card--grow">
          <div className="creategame__card-header">
            👥 players
            <span className="creategame__card-header-count">
              {players.length} / 8
            </span>
          </div>
          <div className="creategame__card-body creategame__card-body--list">
            {players.map((p) =>
            {
              return (
                <div key={p.id} className="creategame__player-row">
                  <span className="creategame__player-dot" />
                  <span className="creategame__player-name">{p.name}</span>
                  {p.host && (
                    <span className="creategame__badge">HOST</span>
                  )}
                </div>
              );
            })}
            {players.length < 2 && (
              <p className="creategame__waiting">
                ⧗ waiting for players…
              </p>
            )}
          </div>
        </div>

        <div className="creategame__card creategame__card--grow">
          <div className="creategame__card-header">⚙ settings</div>
          <div className="creategame__card-body">

            <label className="creategame__label">
              rounds
              <select
                className="creategame__select"
                value={rounds}
                onChange={(e) => setRounds(Number(e.target.value))}
              >
                {[2, 3, 4, 5, 6].map((n) =>
                {
                  return (
                    <option key={n} value={n}>
                      {n} rounds
                    </option>
                  );
                })}
              </select>
            </label>

            <label className="creategame__label">
              draw timer
              <select
                className="creategame__select"
                value={timer}
                onChange={(e) => setTimer(Number(e.target.value))}
              >
                {[30, 45, 60, 90, 120].map((n) =>
                {
                  return (
                    <option key={n} value={n}>
                      {n} seconds
                    </option>
                  );
                })}
              </select>
            </label>

          </div>
        </div>

      </div>

      <div className="creategame__actions">
        <button
          className="creategame__btn creategame__btn--leave"
          onClick={handleLeave}
        >
          ✕ leave room
        </button>
        <button
          className="creategame__btn creategame__btn--start"
          onClick={handleStart}
          disabled={players.length < 2}
          title={startTitle}
        >
          ▶ start game
        </button>
      </div>

      {players.length < 2 && (
        <p className="creategame__start-hint">
          ⚠ at least 2 players are required to start.
        </p>
      )}

    </div>
  );
};

export default CreateGame;
