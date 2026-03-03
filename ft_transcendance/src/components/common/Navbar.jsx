/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   Navbar.jsx                                         :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/02/20 04:00:33 by mforest-          #+#    #+#             */
/*   Updated: 2026/02/20 04:00:33 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React, { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import '../../styles/hypercard.css';

const Clock = () =>
{
  const [time, setTime] = useState('');

  useEffect(() =>
	{
    const tick = () =>
		{
      const d  = new Date();
      const h  = d.getHours() % 12 || 12;
      const m  = String(d.getMinutes()).padStart(2, '0');
      const ap = d.getHours() >= 12 ? 'PM' : 'AM';
      setTime(`${h}:${m} ${ap}`);
    	};
    tick();
    const id = setInterval(tick, 10000);
    return () => clearInterval(id);
  }, []);

  return <span className="hc-menubar__clock">{time}</span>;
};

const Navbar = () =>
{
  const navigate = useNavigate();
  const { user, loading, logout } = useAuth();

  const handleLogout = async () =>
  {
    await logout();
    navigate('/');
  };

  return (
    <nav className="hc-menubar" role="menubar" aria-label="System menu">

      <div className="hc-menubar__apple" role="menuitem" aria-label="Apple menu">
        &#63743;
      </div>

      <div className="hc-menubar__item" role="menuitem">File</div>
      <div className="hc-menubar__item" role="menuitem">Edit</div>
      <div className="hc-menubar__item" role="menuitem">Go</div>
      <div className="hc-menubar__item" role="menuitem">Objects</div>
      <div className="hc-menubar__item" role="menuitem">Help</div>

      {!loading && (
        <div className="hc-menubar__auth">
          {user ? (
            <>
              <span className="hc-menubar__user" title={user.login || user.username}>
                {user.login || user.username}
              </span>
              <button
                type="button"
                className="hc-menubar__auth-btn"
                onClick={handleLogout}
              >
                Log out
              </button>
            </>
          ) : (
            <Link to="/login" className="hc-menubar__auth-link">Login</Link>
          )}
        </div>
      )}

      <Clock />
    </nav>
  );
};

export default Navbar;
