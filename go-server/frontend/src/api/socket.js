/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   socket.js                                          :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/04/18 14:34:30 by mforest-          #+#    #+#             */
/*   Updated: 2026/04/18 14:34:30 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

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
let wsAuthReady = false;

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

export const getIDFromToken = () =>
{
    const token = localStorage.getItem("authToken");
    if (!token)
		return null;
    try
	{
        const payload = JSON.parse(atob(token.split('.')[1]));
        return payload.id;
    }
	catch (e)
	{
        return null;
    }
};

export const updateLocalAuth = (userData) =>
{
    if (userData.username)
        localStorage.setItem('username', userData.username);
    if (userData.token)
        localStorage.setItem('authToken', userData.token);
    if (userData.id)
        localStorage.setItem('userID', userData.id);
	window.dispatchEvent(new CustomEvent('userDataUpdated'));
};

export const isGuestUser = () => localStorage.getItem('isGuest') === '1';

export const getLocalAuth = () =>
{
	const token = localStorage.getItem('authToken');
	const nameFromJwt = getUsernameFromToken();
	const idFromJwt = getIDFromToken();
	return {
		username: nameFromJwt || localStorage.getItem('username'),
		token,
		id: idFromJwt || localStorage.getItem('userID')
	};
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
	const data = JSON.stringify(payload);
	if (payload.type === 'authenticate')
	{
		if (socket && socket.readyState === WebSocket.OPEN)
			socket.send(data);
		return;
	}
	if (!socket || socket.readyState !== WebSocket.OPEN)
	{
		pending.push(data);
		return;
	}
	if (!wsAuthReady)
	{
		pending.push(data);
		return;
	}
	socket.send(data);
};

const flushPendingAfterAuth = () =>
{
	if (!socket || socket.readyState !== WebSocket.OPEN)
		return;
	const batch = pending;
	pending = [];
	for (const data of batch)
	{
		try
		{
			socket.send(data);
		}
		catch (e)
		{
			console.error('ws flush pending send failed', e);
		}
	}
};

const setupSocketHandlers = (token) =>
{
	socket.onopen = () =>
	{
		wsAuthReady = false;
		socket.send(JSON.stringify({ type: 'authenticate', token: token }));
	};

	socket.onmessage = (event) =>
	{
		const msg = JSON.parse(event.data);

		if (msg.type === 'auth_ok')
		{
			wsAuthReady = true;
			flushPendingAfterAuth();
			if (msg.is_guest)
				localStorage.setItem('isGuest', '1');
			else
				localStorage.removeItem('isGuest');
			window.dispatchEvent(new CustomEvent('userDataUpdated'));
		}

		if (msg.type === 'profile_updated' && msg.success && msg.token)
		{
			updateLocalAuth({
				username: msg.user.username,
				token: msg.token
			});
		}

		listeners.forEach((fn) =>
		{
			try
			{
				fn(msg);
			}
			catch (e)
			{
				console.error('socket listener error', e);
			}
		});
	};
	socket.onclose = (event) =>
	{
		if(event.code == 4000)
			window.location.href = "/logout";
		console.warn("event: ", event.msg);
		wsAuthReady = false;
		pending = [];
    	socket = null;
	};
	socket.onerror = (err) =>
	{
		console.error('ws error', err);
	};
};

const connect = async () =>
{
    const token = await getAuthToken();
    if (!token)
        return;
    if (socket && socket.readyState === WebSocket.OPEN)
    {
        if (wsAuthReady)
            return;
        socket.send(JSON.stringify({ type: 'authenticate', token: token }));
        return;
    }
    if (socket && socket.readyState === WebSocket.CONNECTING)
        return;

    socket = new WebSocket(getWsUrl());
    setupSocketHandlers(token);
};

const disconnect = () =>
{
	if (socket)
	{
		socket.close();
	}
	wsAuthReady = false;
	pending = [];
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
