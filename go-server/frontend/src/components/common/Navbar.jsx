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
import { useNavigate, useLocation } from 'react-router-dom';
import NotificationBell from './NotificationBell';
import '../../styles/hypercard.css';
import { getLocalAuth, isGuestUser } from '../../api/socket';

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
		const id = setInterval(tick, 100);
		return () => clearInterval(id);
	}, []);

	return <span className="hc-menubar__clock">{time}</span>;
};

const Navbar = () =>
{
	const navigate = useNavigate();
	const location = useLocation();
	const [isLogged, setIsLogged] = useState(() => !!localStorage.getItem('authToken'));
	const [username, setUsername] = useState(() => getLocalAuth().username || '');
	const [guest, setGuest] = useState(() => isGuestUser());

	useEffect(() =>
	{
		const checkAuth = () =>
		{
			const token = localStorage.getItem('authToken');
			const auth = getLocalAuth();
			setIsLogged(!!token);
			setUsername(auth.username || '');
			setGuest(isGuestUser());
		};
		window.addEventListener('userDataUpdated', checkAuth);
		checkAuth();
		return () =>
		{
			window.removeEventListener('userDataUpdated', checkAuth);
		};
	}, [location.pathname]);

	return (
		<nav className="hc-menubar" role="menubar" aria-label="System menu">
			<div className="hc-menubar__apple" role="menuitem" aria-label="Apple menu">
				&#63743;
			</div>
			<div
				className="hc-menubar__item_useful"
				role="menuitem"
				onClick={() => navigate('/')}
			>
				Home
			</div>
			<div className="hc-menubar__item" role="menuitem">Edit</div>
			<div className="hc-menubar__item" role="menuitem">Go</div>
			<div className="hc-menubar__item" role="menuitem">Objects</div>
			<div className="hc-menubar__item" role="menuitem">Help</div>

			<div className="hc-menubar__spacer" />
			
			{isLogged && (
				<div className="hc-menubar__username" role="menuitem">
					{username}
				</div>
			)}

			<div className="hc-menubar__spacer" />

			{isLogged ? (
				<div
					className="hc-menubar__item_useful"
					role="menuitem"
					onClick={() => navigate('/logout')}
				>
					Logout
				</div>
			) : (
				<div
					className="hc-menubar__item_useful"
					role="menuitem"
					onClick={() => navigate('/login')}
				>
					Login
				</div>
			)}

			<div
				className="hc-menubar__item_useful"
				role="menuitem"
				onClick={() => { if (username) navigate(`/profile/${username}`); }}
			>
				Profile
			</div>
				{/* FIXME guest when login can see friends first iter */}
			{!guest && (
				<div
					className="hc-menubar__item_useful"
					role="menuitem"
					onClick={() => navigate('/friends')}
				>
					Friends
				</div>
			)}
			<NotificationBell />
			<Clock />
		</nav>
	);
};

export default Navbar;
