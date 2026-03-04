/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   Friends.jsx                                        :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/03/04 04:50:57 by mforest-          #+#    #+#             */
/*   Updated: 2026/03/04 04:50:57 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React, { useState, useEffect }          from 'react';
import { useNavigate, Link }                   from 'react-router-dom';
import { send, addListener, removeListener }   from '../../socket';
import { useNotifications }                    from '../../components/common/NotificationContext';
import './Friends.css';

const API_URL = 'https://<host>/api';

const MOCK_FRIENDS =
[
  { id: 2, username: 'lviravon', online: true  },
  { id: 3, username: 'alice',    online: false },
  { id: 4, username: 'bob',      online: true  },
];

// todo: remplacer par les rooms de l api
const MOCK_ROOMS =
{
  'ABCDEF': { status: 'waiting' },
  'ZZZZZZ': { status: 'started' },
};

const Friends = () =>
{
  const navigate            = useNavigate();
  const { push }            = useNotifications();

  const [friends,    setFriends]    = useState([]);
  const [input,      setInput]      = useState('');
  const [addError,   setAddError]   = useState('');
  const [success,    setSuccess]    = useState('');
  const [loading,    setLoading]    = useState(true);
  const [inviteCode, setInviteCode] = useState('');
  const [inviteError,setInviteError]= useState('');
  const [inviting,   setInviting]   = useState(null);

  useEffect(() =>
  {
    const load = async () =>
    {
      // todo: remplacer par fetch(`${API_URL}/friends`, { credentials: 'include' })
      await new Promise((r) => setTimeout(r, 400));
      setFriends(MOCK_FRIENDS);
      setLoading(false);
    };

    load();
  }, []);

  useEffect(() =>
  {
    const onMessage = (msg) =>
    {
      if (msg.type === 'player_joined')
      {
        setFriends((prev) =>
          prev.map((f) =>
            msg.players.find((p) => p.username === f.username)
              ? { ...f, online: true }
              : f
          )
        );
      }
    };

    addListener(onMessage);
    return () => removeListener(onMessage);
  }, []);

  const handleInviteCodeChange = (e) =>
  {
    setInviteCode(e.target.value.toUpperCase().replace(/[^A-Z]/g, '').slice(0, 6));
    setInviteError('');
  };

  const validateInviteCode = async () =>
  {
    const code = inviteCode.trim();

    if (!code)
    {
      setInviteError('please enter a room code.');
      return false;
    }
    if (!/^[A-Z]{6}$/.test(code))
    {
      setInviteError('room code must be exactly 6 letters.');
      return false;
    }

    // todo: remplacer par fetch(`${API_URL}/rooms/${code}`)
    await new Promise((r) => setTimeout(r, 300));
    const room = MOCK_ROOMS[code];

    if (!room)
    {
      setInviteError('room not found. check the code and try again.');
      return false;
    }
    if (room.status === 'started')
    {
      setInviteError('this game has already started.');
      return false;
    }
    if (room.status !== 'waiting')
    {
      setInviteError('this room is not available.');
      return false;
    }

    setInviteError('');
    return true;
  };

  const handleAdd = async () =>
  {
    const username = input.trim();

    if (!username)
    {
      setAddError('please enter a username.');
      return;
    }
    if (username.length < 2)
    {
      setAddError('username must be at least 2 characters.');
      return;
    }
    if (friends.find((f) => f.username === username))
    {
      setAddError('this user is already your friend.');
      return;
    }

    setAddError('');

    // todo: remplacer par fetch(`${API_URL}/friends`, {
    //   method: 'POST',
    //   credentials: 'include',
    //   headers: { 'Content-Type': 'application/json' },
    //   body: JSON.stringify({ username })
    // })
    await new Promise((r) => setTimeout(r, 400));

    setFriends((f) => [...f, { id: Date.now(), username, online: false }]);
    setSuccess(`${username} added as friend.`);
    setInput('');
    setTimeout(() => setSuccess(''), 3000);
  };

  const handleRemove = async (friend) =>
  {
    if (!window.confirm(`remove ${friend.username} from friends ?`))
      return;

    // todo: remplacer par fetch(`${API_URL}/friends/${friend.id}`, {
    //   method: 'DELETE',
    //   credentials: 'include'
    // })
    await new Promise((r) => setTimeout(r, 300));

    setFriends((f) => f.filter((fr) => fr.id !== friend.id));
  };

  const handleInvite = async (friend) =>
  {
    const valid = await validateInviteCode();
    if (!valid)
      return;

    setInviting(friend.id);

    // todo: websocket send 'invite_friend'
    send({ type: 'invite_friend', to: friend.username, code: inviteCode.trim() });

    setTimeout(() => setInviting(null), 2000);
  };

  const handleMockNotif = () =>
  {
    push({
      kind: 'invite',
      from: 'lviravon',
      code: 'ABCDEF',
    });
  };

  let onlineCount = friends.filter((f) => f.online).length;

  if (loading)
  {
    return (
      <div className="friends__guard">
        <span className="friends__spinner">⧗</span>
        loading friends…
      </div>
    );
  }

  return (
    <div className="friends">

      <div className="friends__card">
        <div className="friends__card-header">test notification</div>
        <div className="friends__card-body">
          <p className="friends__hint">simule une invite de lviravon - ABCDEF.</p>
          <button
            className="friends__btn friends__btn--primary"
            onClick={handleMockNotif}
          >
            ▶ trigger invite notif
          </button>
        </div>
      </div>

      <div className="friends__card">
        <div className="friends__card-header">🎮 invite to a game</div>
        <div className="friends__card-body">
          <p className="friends__hint">
            enter your room code to send an invite to a friend.
          </p>
          <input
            className={`friends__input${inviteError ? ' friends__input--error' : ''}`}
            type="text"
            value={inviteCode}
            onChange={handleInviteCodeChange}
            placeholder="room code (6 letters)"
            maxLength={6}
          />
          {inviteError && <p className="friends__error">⚠ {inviteError}</p>}
        </div>
      </div>

      <div className="friends__card">
        <div className="friends__card-header">➕ add a friend</div>
        <div className="friends__card-body">
          <p className="friends__hint">enter the exact username of the person you want to add.</p>
          <div className="friends__add-row">
            <input
              className={`friends__input${addError ? ' friends__input--error' : ''}`}
              type="text"
              value={input}
              onChange={(e) => { setInput(e.target.value); setAddError(''); }}
              onKeyDown={(e) => e.key === 'Enter' && handleAdd()}
              placeholder="username"
              maxLength={32}
            />
            <button
              className="friends__btn friends__btn--primary"
              onClick={handleAdd}
            >
              + add
            </button>
          </div>
          {addError   && <p className="friends__error">⚠ {addError}</p>}
          {success    && <p className="friends__success">✓ {success}</p>}
        </div>
      </div>

      <div className="friends__card">
        <div className="friends__card-header">
          👥 friends
          <span className="friends__card-count">
            {onlineCount} online · {friends.length} total
          </span>
        </div>

        {friends.length === 0 ? (
          <p className="friends__empty">no friends yet. add someone above.</p>
        ) : (
          <div className="friends__list">
            {friends
              .sort((a, b) => b.online - a.online)
              .map((f) =>
              {
                let inviteLabel = '✉ invite';
                if (inviting === f.id)
                  inviteLabel = '✓ sent!';

                return (
                  <div key={f.id} className="friends__row">
                    <span className={`friends__dot${f.online ? ' friends__dot--online' : ''}`} />
                    <Link to={`/profile/${f.username}`} className="friends__username">
                      {f.username}
                    </Link>
                    <span className="friends__status">
                      {f.online ? 'online' : 'offline'}
                    </span>
                    <div className="friends__actions">
                      <button
                        className="friends__btn friends__btn--invite"
                        onClick={() => handleInvite(f)}
                        disabled={!f.online || inviting === f.id}
                        title={!f.online ? 'friend is offline' : ''}
                      >
                        {inviteLabel}
                      </button>
                      <button
                        className="friends__btn friends__btn--remove"
                        onClick={() => handleRemove(f)}
                      >
                        ✕
                      </button>
                    </div>
                  </div>
                );
              })}
          </div>
        )}
      </div>

      <button className="friends__btn friends__btn--back" onClick={() => navigate(-1)}>
        ← back
      </button>

    </div>
  );
};

export default Friends;
