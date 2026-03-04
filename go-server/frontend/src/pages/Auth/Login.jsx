/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   Login.jsx                                          :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/02/26 01:19:24 by mforest-          #+#    #+#             */
/*   Updated: 2026/02/26 01:19:24 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React, { useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { authApi } from '../../api/auth';
import { useAuth } from '../../context/AuthContext';
import './Auth.css';

const Login = () =>
{
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const redirect = searchParams.get('redirect') || '/game';
  const { user, loading, refresh } = useAuth();

  useEffect(() =>
  {
    if (loading)
      return;
    if (user)
      navigate(redirect, { replace: true });
  }, [user, loading, redirect, navigate]);

  useEffect(() =>
  {
    refresh();
  }, []);

  const oauthUrl = authApi.oauth42Url(redirect);

  return (
    <div className="auth">
      <div className="auth__card">
        <div className="auth__card-header">🔐 Sign In</div>
        <div className="auth__body">
          <p className="auth__hint">
            Use your 42 account.
	  </p>
          <a className="auth__btn auth__btn--primary" href={oauthUrl}>
            ▶ Sign in with 42
          </a>
        </div>
      </div>
    </div>
  );
};

export default Login;

