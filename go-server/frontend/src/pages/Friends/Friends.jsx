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

import React, { useState, useEffect } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { connect, send, addListener, removeListener, getIDFromToken } from '../../api/socket';
import './Friends.css';

const Friends = () =>
{
  const navigate = useNavigate();

  const [friends,     setFriends]     = useState([]);
  const [pendingIn,   setPendingIn]   = useState([]);
  const [pendingOut,  setPendingOut]  = useState([]);
  const [input,       setInput]       = useState('');
  const [addError,    setAddError]    = useState('');
  const [success,     setSuccess]     = useState('');
  const [loading,     setLoading]     = useState(true);
  const [guestBlock,  setGuestBlock]  = useState(false);
  const [inviteCode,  setInviteCode]  = useState('');
  const [inviteError, setInviteError] = useState('');
  const [inviting,    setInviting]    = useState(null);

  useEffect(() =>
  {
    connect();
    const onMessage = (msg) =>
    {
      if (msg.type === 'friends_list')
      {
        setFriends(msg.friends || []);
        setPendingIn(msg.pending_in || []);
        setPendingOut(msg.pending_out || []);
        if (msg.guest_no_friends)
          setGuestBlock(true);
        setLoading(false);
      }
      else if (msg.type === 'friend_added')
      {
        if (msg.success && msg.friend)
        {
          setFriends((prev) =>
          {
            if (prev.some((f) => f.id === msg.friend.id))
              return prev;
            return [...prev, msg.friend];
          });
          setPendingIn((prev) => prev.filter((p) => p.id !== msg.friend.id));
          setPendingOut((prev) => prev.filter((p) => p.id !== msg.friend.id));
          setSuccess(`${msg.friend.username} is now your friend.`);
          setInput('');
          setTimeout(() => setSuccess(''), 3000);
        }
        else
          setAddError(msg.error || 'Failed to add friend.');
      }
      else if (msg.type === 'friend_removed')
      {
        setFriends((prev) => prev.filter((f) => f.id !== msg.friend_id));
        setPendingIn((prev) => prev.filter((p) => p.id !== msg.friend_id));
        setPendingOut((prev) => prev.filter((p) => p.id !== msg.friend_id));
      }
      else if (msg.type === 'friend_online_status')
      {
        const patch = (list) =>
          list.map((f) =>
            f.username === msg.username ? { ...f, online: msg.online } : f
          );
        setFriends(patch);
        setPendingIn(patch);
        setPendingOut(patch);
      }
      else if (msg.type === 'invite_sent')
      {
        if (msg.success)
          setTimeout(() => setInviting(null), 2000);
        else
        {
          setInviteError(msg.error || 'Failed to send invite.');
          setInviting(null);
        }
      }
      else if (msg.type === 'friend_request' && msg.user)
      {
        setPendingIn((prev) =>
        {
          if (prev.some((p) => p.id === msg.user.id))
            return prev;
          return [...prev, msg.user];
        });
      }
      else if (msg.type === 'friend_request_sent' && msg.user)
      {
        setPendingOut((prev) =>
        {
          if (prev.some((p) => p.id === msg.user.id))
            return prev;
          return [...prev, msg.user];
        });
      }
      else if (msg.type === 'friend_add_failed' || msg.type === 'friend_accept_failed')
      {
        setAddError(msg.error || 'Friend request failed.');
      }
    };

    addListener(onMessage);

    send({
      type: 'get_friends',
      id: getIDFromToken()
    });

    return () => removeListener(onMessage);
  }, []);

  const handleInviteCodeChange = (e) =>
  {
    const code = e.target.value.toUpperCase().replace(/[^A-Z]/g, '').slice(0, 6);
    setInviteCode(code);
    setInviteError('');
  };

  const handleAdd = () =>
  {
    const username = input.trim();

    if (!username)
    {
      setAddError('Please enter a username.');
      return;
    }
    if (username.length < 2)
    {
      setAddError('Username must be at least 2 characters.');
      return;
    }
    if (friends.find((f) => f.username === username))
    {
      setAddError('This user is already your friend.');
      return;
    }
    if (pendingOut.find((p) => p.username === username))
    {
      setAddError('You already have a pending request to this user.');
      return;
    }
    if (pendingIn.find((p) => p.username === username))
    {
      setAddError('This user already sent you a request — accept it below.');
      return;
    }

    setAddError('');

    send({
      type: 'add_friend',
      username: username
    });
  };

  const handleAccept = (p) =>
  {
    setAddError('');
    send({ type: 'accept_friend', username: p.username });
  };

  const handleRejectIncoming = (p) =>
  {
    send({
      type: 'remove_friend',
      id: getIDFromToken(),
      username: p.username
    });
  };

  const handleCancelOutgoing = (p) =>
  {
    send({
      type: 'remove_friend',
      id: getIDFromToken(),
      username: p.username
    });
  };

  const handleRemove = async (friend) =>
  {
    if (!window.confirm(`Remove ${friend.username} from friends?`))
      return;
    send({
      type: 'remove_friend',
      id: getIDFromToken(),
      username: friend.username
    });
  };

  const handleInvite = async (friend) =>
  {
    const code = inviteCode.trim();

    if (!code)
    {
      setInviteError('Please enter a room code.');
      return;
    }
    if (!/^[A-Z]{6}$/.test(code))
    {
      setInviteError('Room code must be exactly 6 letters.');
      return;
    }

    setInviting(friend.id);

    send({
      type: 'invite_friend',
      to: friend.username,
      code: code
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

  if (guestBlock)
  {
    return (
      <div className="friends">
        <div className="friends__guard-card" style={{ margin: '2rem auto', maxWidth: '420px', textAlign: 'center' }}>
          <p className="friends__guard-msg">Guest accounts cannot use friends or invites.</p>
          <button className="friends__btn friends__btn--primary" onClick={() => navigate('/')}>
            ← back to home
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="friends">

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

      {(pendingIn.length > 0 || pendingOut.length > 0) && (
        <div className="friends__card">
          <div className="friends__card-header">⏳ pending requests</div>
          <div className="friends__card-body">
            {pendingIn.length > 0 && (
              <>
                <p className="friends__hint">wants to be your friend:</p>
                <div className="friends__list">
                  {pendingIn.map((p) => (
                    <div key={p.id} className="friends__row">
                      <span className={`friends__dot${p.online ? ' friends__dot--online' : ''}`} />
                      <Link to={`/profile/${p.username}`} className="friends__username">{p.username}</Link>
                      <div className="friends__actions">
                        <button type="button" className="friends__btn friends__btn--primary" onClick={() => handleAccept(p)}>✓ accept</button>
                        <button type="button" className="friends__btn friends__btn--remove" onClick={() => handleRejectIncoming(p)}>✕</button>
                      </div>
                    </div>
                  ))}
                </div>
              </>
            )}
            {pendingOut.length > 0 && (
              <>
                <p className="friends__hint" style={{ marginTop: pendingIn.length ? '1rem' : 0 }}>waiting for response:</p>
                <div className="friends__list">
                  {pendingOut.map((p) => (
                    <div key={p.id} className="friends__row">
                      <span className={`friends__dot${p.online ? ' friends__dot--online' : ''}`} />
                      <span className="friends__username">{p.username}</span>
                      <button type="button" className="friends__btn friends__btn--remove" onClick={() => handleCancelOutgoing(p)}>cancel</button>
                    </div>
                  ))}
                </div>
              </>
            )}
          </div>
        </div>
      )}

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
