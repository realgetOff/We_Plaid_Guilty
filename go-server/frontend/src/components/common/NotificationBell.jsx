/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   NotificationBell.jsx                               :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/03/04 04:35:14 by mforest-          #+#    #+#             */
/*   Updated: 2026/03/04 04:35:14 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useNotifications } from './NotificationContext';
import './NotificationBell.css';

const NotificationBell = () =>
{
  const { notifs, dismiss } = useNotifications();
  const navigate            = useNavigate();
  const [open, setOpen]     = useState(false);

  const count = notifs.length;

  const handleAccept = (notif) =>
  {
    dismiss(notif.id);
    setOpen(false);
    const path = notif.isAI ? `/aigame/join/${notif.code}` : `/game/join/${notif.code}`;
    navigate(path);
  };

  const handleRefuse = (notif) =>
  {
    dismiss(notif.id);
  };

  return (
    <div className="bell">
      <button
        className={`bell__btn${count > 0 ? ' bell__btn--active' : ''}`}
        onClick={() => setOpen((o) => !o)}
        title="notifications"
      >
        🔔
        {count > 0 && (
          <span className="bell__count">{count}</span>
        )}
      </button>

      {open && (
        <div className="bell__dropdown">
          <div className="bell__dropdown-header">
            notifications
            {count > 0 && (
              <button
                className="bell__clear"
                onClick={() => { notifs.forEach((n) => dismiss(n.id)); }}
              >
                clear all
              </button>
            )}
          </div>

          {count === 0 ? (
            <p className="bell__empty">no notifications.</p>
          ) : (
            notifs.map((n) =>
            {
              if (n.kind === 'invite')
              {
                return (
                  <div key={n.id} className="bell__notif">
                    <p className="bell__notif-msg">
                      <strong>{n.from}</strong> invited you to room <strong>{n.code}</strong>
                    </p>
                    <div className="bell__notif-actions">
                      <button
                        className="bell__notif-btn bell__notif-btn--accept"
                        onClick={() => handleAccept(n)}
                      >
                        ✓ join
                      </button>
                      <button
                        className="bell__notif-btn bell__notif-btn--refuse"
                        onClick={() => handleRefuse(n)}
                      >
                        ✕ refuse
                      </button>
                    </div>
                  </div>
                );
              }
              return null;
            })
          )}
        </div>
      )}

      {open && (
        <div className="bell__backdrop" onClick={() => setOpen(false)} />
      )}
    </div>
  );
};

export default NotificationBell;
