/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   HomeGame.jsx                                       :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/02/23 20:11:08 by mforest-          #+#    #+#             */
/*   Updated: 2026/02/23 20:11:08 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React, { useState } from 'react';
import { useNavigate }     from 'react-router-dom';
import './HomeGame.css';

/* simule room tant que backend pas fait*/
const MOCK_ROOMS = {
  'ABCDEF': { status: 'waiting' },   // room valide, pas encore demarrer
  'ZZZZZZ': { status: 'started' },   // room deja en cours
};

const VALID_CODE_RE = /^[A-Z]{6}$/;

const HomeGame = () => {
  const navigate = useNavigate();

  const [joinCode,    setJoinCode]    = useState('');
  const [joinError,   setJoinError]   = useState('');
  const [isChecking,  setIsChecking]  = useState(false);

  const handleJoin = async () => {
    const code = joinCode.trim().toUpperCase();

    if (!code) {
      setJoinError('Please enter a room code.');
      return;
    }
    if (!VALID_CODE_RE.test(code)) {
      setJoinError('Room code must be exactly 6 letters (A–Z).');
      return;
    }

    setIsChecking(true);
    setJoinError('');

    /* TODO: faire l'appel reel a l'api */
       // const res  = await fetch(`/api/rooms/${code}`);
       // const data = await res.json();
    const room = MOCK_ROOMS[code];

    setIsChecking(false);

    if (!room) {
      setJoinError('Room not found. Check the code and try again.');
      return;
    }
	/* NOTE: a voir si mode spectator ou pas?? */
    if (room.status === 'started') {
      setJoinError('This game has already started. You cannot join.');
      return;
    }

    navigate('/game/join/' + code);
  };

  const handleCodeChange = (e) => {
    setJoinCode(e.target.value.toUpperCase().replace(/[^A-Z]/g, '').slice(0, 6));
    setJoinError('');
  };

  return (
    <div className="homegame">

      <div className="homegame__card">
        <div className="homegame__card-header">🎨 Create a Game</div>
        <div className="homegame__card-body">
          <p className="homegame__card-desc">
            Start a new room and invite your friends with a 6-letter code.
          </p>
          <button
            className="homegame__btn homegame__btn--primary"
            onClick={() => navigate('/game/create')}
          >
            ▶ Create Room
          </button>
        </div>
      </div>

      <div className="homegame__separator">— or —</div>

      <div className="homegame__card">
        <div className="homegame__card-header">🔑 Join a Game</div>
        <div className="homegame__card-body">
          <p className="homegame__card-desc">
            Enter the 6-letter room code shared by the host.
          </p>
          <div className="homegame__join-row">
            <input
              className={`homegame__input${joinError ? ' homegame__input--error' : ''}`}
              type="text"
              value={joinCode}
              onChange={handleCodeChange}
              onKeyDown={(e) => e.key === 'Enter' && !isChecking && handleJoin()}
              placeholder="ABCDEF"
              maxLength={6}
              aria-label="Room code"
              disabled={isChecking}
            />
            <button
              className="homegame__btn homegame__btn--secondary"
              onClick={handleJoin}
              disabled={isChecking}
            >
              {isChecking ? '⧗' : '→ Join'}
            </button>
          </div>

          {joinError && (
            <p className="homegame__error">⚠ {joinError}</p>
          )}
        </div>
      </div>

    </div>
  );
};

export default HomeGame;
