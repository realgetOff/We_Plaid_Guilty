/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   MacWindow.jsx                                      :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/02/20 04:00:21 by mforest-          #+#    #+#             */
/*   Updated: 2026/02/20 04:00:21 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React       from 'react';
import { Link,
         useLocation } from 'react-router-dom';
import '../../styles/hypercard.css';

const ROUTE_META =
{
  '/':        { title: 'ft_transcendence — HyperCard Gartic Edition', card: '🎨 Home' },
  '/privacy': { title: 'ft_transcendence — Privacy Policy',           card: '🔒 Privacy Policy' },
  '/tos':     { title: 'ft_transcendence — Terms of Service',         card: '📋 Terms of Service' },
};

const MacWindow = ({ children }) =>
{
  const location = useLocation();
  const meta = ROUTE_META[location.pathname] || ROUTE_META['/'];

  return (
    <div className="hc-window" role="main">

      <div className="hc-titlebar">
        <div className="hc-titlebar__btn" aria-hidden="true" />
        <div style={{ width: 4 }} />
        <div className="hc-titlebar__btn" aria-hidden="true" />
        <span className="hc-titlebar__title">{meta.title}</span>
      </div>

      <div className="hc-card-header">
        {meta.card}
        <span className="hc-card-header__right">ft_transcendence · 42 School</span>
      </div>

      <div className="hc-card-content">
        {children}
      </div>

      <footer className="hc-footer">
        <span className="hc-footer__left">
          ft_transcendence v1.0 · HyperCard Edition
        </span>
        <div className="hc-footer__links">
          <Link to="/privacy" className="hc-footer__link">Privacy Policy</Link>
          <Link to="/tos"     className="hc-footer__link">Terms of Service</Link>
        </div>
      </footer>

    </div>
  );
};

export default MacWindow;
