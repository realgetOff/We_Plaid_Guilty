/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   ProtectedRoute.jsx                                 :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/03/03 01:32:49 by mforest-          #+#    #+#             */
/*   Updated: 2026/03/03 01:32:49 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import '../../pages/Auth/Auth.css';

const ProtectedRoute = ({ children }) =>
{
  const { user, loading } = useAuth();
  const location = useLocation();

  if (loading)
  {
    return (
      <div className="auth auth__loading">
        <span className="auth__loading-spinner">⧗</span>
        <span>loading…</span>
      </div>
    );
  }

  if (!user)
  {
    const from = location.pathname + location.search;

    return <Navigate to={`/login?redirect=${encodeURIComponent(from)}`} replace />;
  }

  return children;
};

export default ProtectedRoute;

