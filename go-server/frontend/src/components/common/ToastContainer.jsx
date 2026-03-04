/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   ToastContainer.jsx                                 :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/03/04 04:30:18 by mforest-          #+#    #+#             */
/*   Updated: 2026/03/04 04:30:18 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React, { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useNotifications } from './NotificationContext';
import './ToastContainer.css';

const TOAST_DURATION = 15000;

const Toast = ({ notif, onAccept, onRefuse }) =>
{
  const { dismiss } = useNotifications();

  useEffect(() =>
  {
    const id = setTimeout(() => dismiss(notif.id), TOAST_DURATION);
    return () => clearTimeout(id);
  }, [notif.id, dismiss]);

  return (
    <div className="toast">
      <div className="toast__header">
        🎮 game invite
        <button className="toast__close" onClick={() => dismiss(notif.id)}>✕</button>
      </div>
      <p className="toast__msg">
        <strong>{notif.from}</strong> invited you to join room <strong>{notif.code}</strong>
      </p>
      <div className="toast__actions">
        <button
          className="toast__btn toast__btn--accept"
          onClick={() => onAccept(notif)}
        >
          ✓ join
        </button>
        <button
          className="toast__btn toast__btn--refuse"
          onClick={() => onRefuse(notif)}
        >
          ✕ refuse
        </button>
      </div>
      <div className="toast__bar" />
    </div>
  );
};

const ToastContainer = () =>
{
  const { notifs, dismiss } = useNotifications();
  const navigate            = useNavigate();

  const handleAccept = (notif) =>
  {
    dismiss(notif.id);
    navigate(`/game/join/${notif.code}`);
  };

  const handleRefuse = (notif) =>
  {
    dismiss(notif.id);
  };

  let invites = notifs.filter((n) => n.kind === 'invite');

  return (
    <div className="toast-container">
      {invites.map((n) =>
      {
        return (
          <Toast
            key={n.id}
            notif={n}
            onAccept={handleAccept}
            onRefuse={handleRefuse}
          />
        );
      })}
    </div>
  );
};

export default ToastContainer;
