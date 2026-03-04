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

import React, { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import './Profile.css';

const API_URL    = 'https://<host>/api';
const MOCK_ME    = 'mforest-'; // todo: remplacer par GET /api/auth/me

const FONT_STYLES = ['normal', 'bold', 'italic'];
const COLORS      =
[
  '#000000', '#aa0000', '#0000aa', '#008800',
  '#884400', '#aa00aa', '#008888', '#555555',
];

// todo: remplacer par les donnees de l api
const MOCK_USERS =
{
  'mforest-':
  {
    id:         1,
    username:   'mforest-',
    email:      'mforest@42.fr',
    online:     true,
    avatar_url: null,
    style:      { color: '#000000', font: 'normal' },
  },
  'lviravon':
  {
    id:         2,
    username:   'lviravon',
    email:      'lviravon@42.fr',
    online:     true,
    avatar_url: null,
    style:      { color: '#aa0000', font: 'bold' },
  },
};

const Profile = () =>
{
  const { username } = useParams();
  const navigate     = useNavigate();
  const fileRef      = useRef(null);

  const isMe = username === MOCK_ME; // todo: comparer avec GET /api/auth/me

  const [status,     setStatus]     = useState('loading');
  const [user,       setUser]       = useState(null);
  const [error,      setError]      = useState('');
  const [editName,   setEditName]   = useState('');
  const [nameError,  setNameError]  = useState('');
  const [color,      setColor]      = useState('#000000');
  const [font,       setFont]       = useState('normal');
  const [avatarSrc,  setAvatarSrc]  = useState(null);
  const [saved,      setSaved]      = useState(false);

  useEffect(() =>
  {
    const load = async () =>
    {
      // todo: remplacer par fetch(`${API_URL}/users/${username}`)
      await new Promise((r) => setTimeout(r, 400));
      const found = MOCK_USERS[username];

      if (!found)
      {
        setError('user not found.');
        setStatus('error');
        return;
      }

      setUser(found);
      setEditName(found.username);
      setColor(found.style.color);
      setFont(found.style.font);
      setAvatarSrc(found.avatar_url);
      setStatus('ready');
    };

    load();
  }, [username]);

  const handleAvatarClick = () =>
  {
    if (!isMe)
      return;
    fileRef.current?.click();
  };

  const handleFileChange = (e) =>
  {
    const file = e.target.files[0];
    if (!file)
      return;

    const reader = new FileReader();
    reader.onload = (ev) =>
    {
      setAvatarSrc(ev.target.result);
    };
    reader.readAsDataURL(file);
  };

  const handleSave = async () =>
  {
    const name = editName.trim();

    if (!name)
    {
      setNameError('username cannot be empty.');
      return;
    }
    if (name.length < 2)
    {
      setNameError('username must be at least 2 characters.');
      return;
    }
    if (name.length > 32)
    {
      setNameError('username must be under 32 characters.');
      return;
    }

    setNameError('');

    // todo: remplacer par fetch(`${API_URL}/users/me`, {
    //   method: 'PATCH',
    //   body: JSON.stringify({ username: name, color, font, avatar_url: avatarSrc })
    // })
    await new Promise((r) => setTimeout(r, 400));

    setUser((u) => ({ ...u, username: name, style: { color, font } }));
    setSaved(true);
    setTimeout(() => setSaved(false), 2500);
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

  let initials = user.username.slice(0, 2).toUpperCase();

  let usernameStyle = { color: user.style.color };
  if (user.style.font === 'bold')   usernameStyle.fontWeight = 'bold';
  if (user.style.font === 'italic') usernameStyle.fontStyle  = 'italic';

  return (
    <div className="profile">

      <div className="profile__header">

        <div
          className={`profile__avatar${isMe ? ' profile__avatar--editable' : ''}`}
          onClick={handleAvatarClick}
          title={isMe ? 'click to change avatar' : ''}
        >
          {avatarSrc
            ? <img src={avatarSrc} alt="avatar" className="profile__avatar-img" />
            : <span>{initials}</span>
          }
          {isMe && <span className="profile__avatar-overlay">✎</span>}
        </div>

        {isMe && (
          <input
            ref={fileRef}
            type="file"
            accept="image/*"
            className="profile__file-input"
            onChange={handleFileChange}
          />
        )}

        <div className="profile__info">
          <span className="profile__username" style={usernameStyle}>
            {user.username}
          </span>
          <span className="profile__email">{user.email}</span>
          <span className={`profile__status${user.online ? ' profile__status--online' : ''}`}>
            {user.online ? '● online' : '○ offline'}
          </span>
        </div>

        <Link to="/friends" className="profile__friends-btn">
          👥 friends
        </Link>

      </div>

      {isMe && (
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
              <span className="profile__preview" style={
              {
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
