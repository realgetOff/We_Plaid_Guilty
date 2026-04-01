// // ************************************************************************** //
// //                                                                            //
// //                                                        :::      ::::::::   //
// //   socket.js                                          :+:      :+:    :+:   //
// //                                                    +:+ +:+         +:+     //
// //   By: mforest- <mforest-@student.42angouleme.fr  +#+  +:+       +#+        //
// //                                                +#+#+#+#+#+   +#+           //
// //   Created: 2026/02/26 00:09:36 by mforest-          #+#    #+#             //
// //   Updated: 2026/02/26 00:53:58 by mforest-         ###   ########.fr       //
// //                                                                            //
// // ************************************************************************** //

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

export const getUsernameFromToken = () =>
{
    const token = localStorage.getItem("authToken");
    if (!token)
		return null;
    try
	{
        const payload = JSON.parse(atob(token.split('.')[1]));
        return payload.username;
    }
	catch (e)
	{
        return null;
    }
};

const getAuthToken = async() =>
{
    const localToken = localStorage.getItem("authToken");

    if (localToken)
        return localToken;

    console.log("[getAuthToken] no token found");
    window.location.href = "/login";
    return null;
};

const send = (payload) =>
{
    const    data = JSON.stringify(payload);

    if (socket && socket.readyState === WebSocket.OPEN)
    {
        socket.send(data);
        return ;
    }
    pending.push(data);
};

const setupSocketHandlers = (token) =>
{
	socket.onopen = () =>
	{
		send({type: 'authenticate', token: token});
		pending.forEach((data) =>
			{
				try
				{
					socket.send(data);
				}
				catch (e) {}
			});
		pending = [];
	};
	socket.onmessage = (event) =>
	{
		const    msg = JSON.parse(event.data);
		listeners.forEach((fn) =>
		{
			fn(msg);
		});
	};
	socket.onclose = (event) =>
	{
		if(event.code == 4000)
			window.location.href = "/logout";
		console.warn("event: ", event.msg);
    	socket = null;
	};
	socket.onerror = (err) =>
	{
		console.error('ws error', err);
	};
};

const connect = async () =>
{
	const    token = await getAuthToken();

	if (socket && (socket.readyState === WebSocket.OPEN
		|| socket.readyState === WebSocket.CONNECTING))
	{
		return ;
	}
	if (!token)
	{
		return ;
	}
	socket = new WebSocket(getWsUrl());
	setupSocketHandlers(token);
};

const disconnect = () =>
{
	if (socket)
	{
		socket.close();
	}
	socket = null;
	listeners = [];
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
