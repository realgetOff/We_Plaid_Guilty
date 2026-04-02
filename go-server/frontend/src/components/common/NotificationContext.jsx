/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   NotificationContext.jsx                            :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/03/04 04:39:58 by mforest-          #+#    #+#             */
/*   Updated: 2026/03/04 04:39:58 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React, { createContext, useContext,
                useState, useEffect, useRef } from 'react';
import { addListener, removeListener }        from '../../api/socket';

const NotificationContext = createContext(null);

export const useNotifications = () => useContext(NotificationContext);

export const NotificationProvider = ({ children }) =>
{
  const [notifs, setNotifs] = useState([]);
  const idRef = useRef(0);
  const cooldowns = useRef({});

  const push = (notif) =>
  {
    const id = ++idRef.current;
    setNotifs((n) => [...n, { ...notif, id }]);
    return id;
  };

  const dismiss = (id) =>
  {
    setNotifs((n) => n.filter((notif) => notif.id !== id));
  };

  useEffect(() =>
  {
    const onMessage = (msg) =>
    {
      if (msg.type === 'room_invite')
      {
        const now = Date.now();
        const last = cooldowns.current[msg.from] || 0;

        if (now - last < 10000)
          return ;
        cooldowns.current[msg.from] = now;
        push({
          kind:  'invite',
          from:  msg.from,
          code:  msg.code,
          timer: 15,
        });
      }
    };

    addListener(onMessage);
    return () => removeListener(onMessage);
  }, []);

  return (
    <NotificationContext.Provider value={{ notifs, push, dismiss }}>
      {children}
    </NotificationContext.Provider>
  );
};
