/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   auth.js                                            :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/03/03 01:31:40 by mforest-          #+#    #+#             */
/*   Updated: 2026/03/03 01:31:40 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

const getApiBaseUrl = () =>
{
  const raw = import.meta.env.VITE_API_URL;
  if (raw && typeof raw === 'string' && raw.trim() !== '')
    return raw.replace(/\/$/, '');

  if (typeof window !== 'undefined' && window.location && window.location.origin)
    return window.location.origin;

  return '';
};

export const authApi = {
  oauth42Url (redirectPath = '/game')
  {
    const base = getApiBaseUrl();
    const url = new URL('/api/auth/42', base || 'http://localhost');
    url.searchParams.set('redirect', redirectPath);
    return url.toString();
  },

  async me ()
  {
    try
    {
      const res = await fetch(getApiBaseUrl() + '/api/auth/me', {
        method: 'GET',
        credentials: 'include',
      });
      if (!res.ok || res.status === 401)
        return null;

      let data;
      try
      {
        data = await res.json();
      }
      catch
      {
        return null;
      }

      const user = data.user ?? data;
      if (!user || typeof user !== 'object')
        return null;
      return user;
    }
    catch
    {
      return null;
    }
  },

  async logout ()
  {
    await fetch(getApiBaseUrl() + '/api/auth/logout', {
      method: 'POST',
      credentials: 'include',
    }).catch(() => {});
  },
};

