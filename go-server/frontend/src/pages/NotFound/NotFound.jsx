/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   NotFound.jsx                                       :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/02/24 22:58:48 by mforest-          #+#    #+#             */
/*   Updated: 2026/02/24 22:58:48 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React from 'react';
import { useNavigate } from 'react-router-dom';
import './NotFound.css';

const NotFound = () =>
{
  const navigate = useNavigate();

  return (
    <div className="notfound">
      <div className="notfound__card">
        <div className="notfound__card-header">⚠ error 404</div>
        <div className="notfound__card-body">
          <div className="notfound__code">404</div>
          <hr className="notfound__divider" />
          <p className="notfound__msg">
            this page does not exist.
          </p>
          <p className="notfound__url">
            {window.location.pathname}
          </p>
          <button
            className="notfound__btn"
            onClick={() => navigate('/')}
          >
            ← back to home
          </button>
        </div>
      </div>
    </div>
  );
};

export default NotFound;
