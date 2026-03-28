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

const getAuthToken = async() =>
{
	try
<<<<<<< HEAD
	{
			
		const res = await fetch('api/auth/player');
		const data = await res.json();
		if(!data.token || !res.ok)
		{
			window.location.href = "/login";
			return null;
		}
=======
	{		
		const res = await fetch('/api/auth/player');
		const data = await res.json();
		if(!data.token || !res.ok)
		{
			console.log("test");
			// window.location.href = "/login";
			return null;
		}
		console.log("token: ", data.token);
>>>>>>> 5fe6cb6f876601e10f69acdbe2579727f8c9fe60
		return(data.token)
	}
	catch(err)
	{
<<<<<<< HEAD
		window.location.href = "/login";
=======
		console.log("AAAAAAAAAAAAAAA");
		// window.location.href = "/login";
>>>>>>> 5fe6cb6f876601e10f69acdbe2579727f8c9fe60
		return null;
	}
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
		send({type: 'authenticate', token});
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
	socket.onclose = () =>
	{
		socket = null;
	};
	socket.onerror = (err) =>
	{
		console.error('ws error', err);
	};
};

const connect = async () =>
{
<<<<<<< HEAD
	const    token = await getAuthToken();

	if (socket && (socket.readyState === WebSocket.OPEN
		|| socket.readyState === WebSocket.CONNECTING))
	{
=======
	console.log("connect test 1");
	const    token = await getAuthToken();

	console.log("connect test 2");
	if (socket && (socket.readyState === WebSocket.OPEN
		|| socket.readyState === WebSocket.CONNECTING))
	{
		console.log("connect test 3");
>>>>>>> 5fe6cb6f876601e10f69acdbe2579727f8c9fe60
		return ;
	}
	if (!token)
	{
<<<<<<< HEAD
		return ;
	}
	socket = new WebSocket(getWsUrl());
	setupSocketHandlers(token);
=======
		console.log("connect test 4");
		return ;
	}
	console.log("connect test 5");
	socket = new WebSocket(getWsUrl());
	setupSocketHandlers(token);
	console.log("connect test 6");
>>>>>>> 5fe6cb6f876601e10f69acdbe2579727f8c9fe60
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
