/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   rooms.js                                           :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/03/03 03:12:01 by mforest-          #+#    #+#             */
/*   Updated: 2026/03/03 03:12:01 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

export const getApiBaseUrl = () =>
{
	const raw = import.meta.env.VITE_API_URL;
	if (raw && typeof raw === 'string' && raw.trim() !== '')
		return raw.replace(/\/$/, '');

	if (typeof window !== 'undefined' && window.location && window.location.origin)
		return window.location.origin;

	return '';
};

export const roomsApi = {
	async getRoom(code)
	{
		const res = await fetch(getApiBaseUrl() + '/api/rooms/' + encodeURIComponent(code), {
			method: 'GET',
			credentials: 'include',
			cache: 'no-store',
		});

		if (res.status === 404)
			return null;

		const data = await res.json().catch(() => ({}));
		if (!res.ok)
			return null;

		return data;
	},

	async getAIRoom(code)
	{
		const res = await fetch(getApiBaseUrl() + '/api/ai-rooms/' + encodeURIComponent(code), {
			method: 'GET',
			credentials: 'include',
			cache: 'no-store',
		});

		if (res.status === 404)
			return null;

		const data = await res.json().catch(() => ({}));
		if (!res.ok)
			return null;

		return data;
	},
};
