/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   Profile.jsx                                        :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/03/04 04:25:37 by mforest-          #+#    #+#             */
/*   Updated: 2026/03/04 04:25:37 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React, { useState, useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { connect, send, addListener, removeListener } from '../../api/socket';
import './Profile.css';

const FONT_STYLES = ['normal', 'bold', 'italic'];
const COLORS      =
[
  '#000000', '#aa0000', '#0000aa', '#008800',
  '#884400', '#aa00aa', '#008888', '#555555',
];

const Profile = () =>
{
  const { username } = useParams();
  const navigate     = useNavigate();

  const [status,     setStatus]     = useState('loading');
  const [user,       setUser]       = useState(null);
  const [isMe,       setIsMe]       = useState(false);
  const [error,      setError]      = useState('');
  const [editName,   setEditName]   = useState('');
  const [nameError,  setNameError]  = useState('');
  const [color,      setColor]      = useState('#000000');
  const [font,       setFont]       = useState('normal');
  const [saved,      setSaved]      = useState(false);

  useEffect(() =>
  {
    connect();
    const handler = (msg) =>
    {
      if (msg.type === 'profile_data')
      {
        if (msg.success)
        {
          setUser(msg.user);
          setIsMe(msg.is_me);
          setEditName(msg.user.username);
          setColor(msg.user.style?.color || '#000000');
          setFont(msg.user.style?.font || 'normal');
          setStatus('ready');
        }
        else
        {
          setError(msg.error || 'User not found.');
          setStatus('error');
        }
      }
      else if (msg.type === 'profile_updated')
      {
        if (msg.success)
        {
          setUser(prev => ({
            ...prev,
            username: msg.user.username,
            style: msg.user.style,
          }));
          setSaved(true);
          setTimeout(() => setSaved(false), 2500);
          if (msg.user.username !== username)
            navigate(`/profile/${msg.user.username}`, { replace: true });
        }
        else
          setNameError(msg.error || 'Failed to update profile.');
      }
    };

    addListener(handler);

    send({
      type: 'get_profile',
      username: username
    });

    return () => removeListener(handler);
  }, [username]);

  const handleSave = async () =>
  {
    const name = editName.trim();

    if (!name)
    {
      setNameError('Username cannot be empty.');
      return;
    }
    if (name.length < 2)
    {
      setNameError('Username must be at least 2 characters.');
      return;
    }
    if (name.length > 32)
    {
      setNameError('Username must be under 32 characters.');
      return;
    }

    setNameError('');

    send({
      type: 'update_profile',
      username: name,
      style: {
        color: color,
        font: font
      },
    });
  };

  if (status === 'loading')
  {
    return (
      <div className="profile__guard">
        <span className="profile__spinner">⧗</span>
        loading profile…
      </div>
    );
  }

  if (status === 'error')
  {
    return (
      <div className="profile__guard">
        <div className="profile__guard-card">
          <div className="profile__guard-icon">✕</div>
          <p className="profile__guard-msg">⚠ {error}</p>
          <button className="profile__btn" onClick={() => navigate('/')}>
            ← back to home
          </button>
        </div>
      </div>
    );
  }

  const initials = user.username.slice(0, 2).toUpperCase();

  const usernameStyle =
  {
    color:      user.style?.color || '#000000',
    fontWeight: user.style?.font === 'bold'   ? 'bold'   : 'normal',
    fontStyle:  user.style?.font === 'italic' ? 'italic' : 'normal',
  };

  const isGuestProfile = !!user.is_guest;
  const canEditProfile = isMe && !isGuestProfile;

  return (
    <div className="profile">

      {isGuestProfile && (
        <p className="profile__guest-banner" style={{
          margin: '0 0 1rem',
          padding: '0.75rem 1rem',
          background: '#f5f0e0',
          border: '1px solid #ccb',
          borderRadius: '8px',
          fontSize: '0.95rem',
        }}>
          {isMe
            ? 'You are on a guest account. Profile editing is disabled.'
            : 'This is a guest account.'}
        </p>
      )}

      <div className="profile__header">

        <div className="profile__avatar">
          <span>{initials}</span>
        </div>

        <div className="profile__info">
          <span className="profile__username" style={usernameStyle}>
            {user.username}
          </span>
          <span className="profile__email">{user.email}</span>
          <span className={`profile__status${user.online ? ' profile__status--online' : ''}`}>
            {user.online ? '● online' : '○ offline'}
          </span>
        </div>

        {!isGuestProfile && (
          <Link to="/friends" className="profile__friends-btn">
            👥 friends
          </Link>
        )}

      </div>

      {canEditProfile && (
        <div className="profile__card">
          <div className="profile__card-header">✎ edit profile</div>
          <div className="profile__edit-body">

            <label className="profile__label">
              username
              <input
                className={`profile__edit-input${nameError ? ' profile__edit-input--error' : ''}`}
                type="text"
                value={editName}
                onChange={(e) => { setEditName(e.target.value); setNameError(''); }}
                maxLength={32}
              />
              {nameError && <span className="profile__field-error">⚠ {nameError}</span>}
            </label>

            <label className="profile__label">
              username color
              <div className="profile__colors">
                {COLORS.map((c) =>
                {
                  let cls = 'profile__color-swatch';
                  if (color === c)
                    cls += ' profile__color-swatch--active';
                  return (
                    <button
                      key={c}
                      className={cls}
                      style={{ background: c }}
                      onClick={() => setColor(c)}
                    />
                  );
                })}
              </div>
            </label>

            <label className="profile__label">
              username style
              <div className="profile__fonts">
                {FONT_STYLES.map((f) =>
                {
                  let cls = 'profile__font-btn';
                  if (font === f)
                    cls += ' profile__font-btn--active';
                  return (
                    <button
                      key={f}
                      className={cls}
                      onClick={() => setFont(f)}
                    >
                      <span style={{
                        fontWeight: f === 'bold'   ? 'bold'   : 'normal',
                        fontStyle:  f === 'italic' ? 'italic' : 'normal',
                      }}>
                        {f}
                      </span>
                    </button>
                  );
                })}
              </div>
            </label>

            <div className="profile__edit-footer">
              <span className="profile__preview-label">preview :</span>
              <span className="profile__preview" style={{
                color,
                fontWeight: font === 'bold'   ? 'bold'   : 'normal',
                fontStyle:  font === 'italic' ? 'italic' : 'normal',
              }}>
                {editName || user.username}
              </span>
              <button className="profile__save-btn" onClick={handleSave}>
                {saved ? '✓ saved!' : '💾 save'}
              </button>
            </div>

          </div>
        </div>
      )}

      <button className="profile__btn profile__btn--back" onClick={() => navigate(-1)}>
        ← back
      </button>

    </div>
  );
};

export default Profile;
