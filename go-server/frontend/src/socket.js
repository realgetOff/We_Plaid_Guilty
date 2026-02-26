// ************************************************************************** //
//                                                                            //
//                                                        :::      ::::::::   //
//   socket.js                                          :+:      :+:    :+:   //
//                                                    +:+ +:+         +:+     //
//   By: mforest- <mforest-@student.42angouleme.fr  +#+  +:+       +#+        //
//                                                +#+#+#+#+#+   +#+           //
//   Created: 2026/02/26 00:09:36 by mforest-          #+#    #+#             //
//   Updated: 2026/02/26 00:53:58 by mforest-         ###   ########.fr       //
//                                                                            //
// ************************************************************************** //

const WS_URL = 'wss://<host>/ws'; //TODO: changer l'host par le bon

let socket = null;
let listeners = [];

const connect = () =>
{
  if (socket && socket.readyState === WebSocket.OPEN)
    return;

  socket = new WebSocket(WS_URL);

  socket.onopen = () =>
  {
    console.log('ws connecte');
  };

  socket.onmessage = (event) =>
  {
    const msg = JSON.parse(event.data);
    listeners.forEach((fn) => fn(msg));
  };

  socket.onclose = () =>
  {
    console.log('ws deconnecte');
    socket = null;
  };

  socket.onerror = (err) =>
  {
    console.error('ws error', err);
  };
};

const disconnect = () =>
{
  if (socket)
    socket.close();
  socket = null;
  listeners = [];
};

const send = (payload) =>
{
  if (!socket || socket.readyState !== WebSocket.OPEN)
    return;
  socket.send(JSON.stringify(payload));
};

const addListener = (fn) =>
{
  listeners.push(fn);
};

const removeListener = (fn) =>
{
  listeners = listeners.filter((l) => l !== fn);
};

export { connect, disconnect, send, addListener, removeListener };
