/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   AuthContext.jsx                                    :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/03/03 02:51:56 by mforest-          #+#    #+#             */
/*   Updated: 2026/03/03 02:51:56 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React, { createContext, useContext, useEffect, useMemo, useState } from 'react';
import { authApi } from '../api/auth';

const AuthContext = createContext(null);

export const useAuth = () =>
{
  const ctx = useContext(AuthContext);
  if (!ctx)
    throw new Error('useAuth must be used within AuthProvider');
  return ctx;
};

export const AuthProvider = ({ children }) =>
{
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  const refresh = async () =>
  {
    setLoading(true);
    try
    {
      const me = await authApi.me();
      setUser(me);
    }
    finally
    {
      setLoading(false);
    }
  };

  useEffect(() =>
  {
    refresh();
  }, []);

  const logout = async () =>
  {
    await authApi.logout();
    setUser(null);
  };

  const value = useMemo(() =>
  {
    return {
      user,
      loading,
      isAuthenticated: !!user,
      refresh,
      logout,
    };
  }, [user, loading]);

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};

