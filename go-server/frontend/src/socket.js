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

const getWsUrl = () =>
{
	const env = import.meta.env.VITE_WS_URL;

	if (env && typeof env === 'string' && env.trim() !== '')
		return env;

	if (typeof window !== 'undefined' && window.location)
	{
		const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
		return `${proto}//${window.location.host}/ws`;
	}

	return 'ws://localhost:8080/ws';
};

let socket = null;
let listeners = [];
let pending = [];

const getAuthToken = () =>
{
	const token = localStorage.getItem("authToken");

	if (!token)
	{
		window.location.href = "/login";
		return null;
	}

	return token;
};

const connect = () =>
{
	if (socket && (socket.readyState === WebSocket.OPEN
		|| socket.readyState === WebSocket.CONNECTING))
		return;

	const token = getAuthToken();
	if (!token)
		return;

	socket = new WebSocket(getWsUrl());

	socket.onopen = () =>
	{
		console.log('ws connected');
		send({ type: "authenticate", token });

		if (pending.length > 0)
		{
			pending.forEach((data) =>
			{
				try
				{
					socket.send(data);
				}
				catch
				{}
			});
			pending = [];
		}
	};

	socket.onmessage = (event) =>
	{
		const msg = JSON.parse(event.data);
		listeners.forEach((fn) => fn(msg));
	};

	socket.onclose = () =>
	{
		console.log('ws disconnected');
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
	const data = JSON.stringify(payload);

	if (socket && socket.readyState === WebSocket.OPEN)
	{
		socket.send(data);
		return;
	}

	pending.push(data);
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
