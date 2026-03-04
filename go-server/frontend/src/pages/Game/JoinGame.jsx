/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   JoinGame.jsx                                       :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/02/23 23:29:36 by mforest-          #+#    #+#             */
/*   Updated: 2026/02/23 23:29:36 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { roomsApi } from '../../api/rooms';
import './JoinGame.css';

const DENY_REASONS =
{
  invalid:   'invalid room code format.',
  not_found: 'room not found. check the code and try again.',
  started:   'this game has already started. you cannot join.',
  finished:  'this game is already finished.',
  unknown:   'cannot join this room.',
};

const JoinGame = () =>
{
  const { code } = useParams();
  const navigate = useNavigate();

  const [error,   setError]   = useState('');
  const [loading, setLoading] = useState(true);

  useEffect(() =>
  {
    const check = async () =>
    {
      const normalized = code?.toUpperCase();

      if (!normalized || !/^[A-Z]{6}$/.test(normalized))
      {
        setError(DENY_REASONS.invalid);
        setLoading(false);
        return;
      }

      let room;
      try
      {
        room = await roomsApi.getRoom(normalized);
      }
      catch
      {
        setError(DENY_REASONS.not_found);
        setLoading(false);
        return;
      }

      if (!room || !room.status)
      {
        setError(DENY_REASONS.not_found);
        setLoading(false);
        return;
      }
      if (room.status === 'started')
      {
        setError(DENY_REASONS.started);
        setLoading(false);
        return;
      }
      if (room.status === 'finished')
      {
        setError(DENY_REASONS.finished);
        setLoading(false);
        return;
      }
      if (room.status !== 'waiting')
      {
        setError(DENY_REASONS.unknown);
        setLoading(false);
        return;
      }

      navigate(`/game/lobby/${normalized}`, { replace: true });
    };

    check();
  }, [code, navigate]);

  if (loading)
  {
    return (
      <div className="joingame">
        <div className="joingame__checking">
          <span className="joingame__spinner">⧗</span>
          checking room <strong>{code?.toUpperCase()}</strong>…
        </div>
      </div>
    );
  }

  let extraBtn = null;
  if (error === DENY_REASONS.started)
    extraBtn = (
      <button
        className="joingame__btn joingame__btn--primary"
        onClick={() => navigate(`/game/play/${code?.toUpperCase()}`)}
      >
        → go to game
      </button>
    );

  return (
    <div className="joingame">
      <div className="joingame__error-card">
        <div className="joingame__error-icon">✕</div>
        <p className="joingame__error-msg">⚠ {error}</p>
        <div className="joingame__error-actions">
          <button
            className="joingame__btn"
            onClick={() => navigate('/game')}
          >
            ← back to home
          </button>
          {extraBtn}
        </div>
      </div>
    </div>
  );
};

export default JoinGame;
