/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   App.jsx                                            :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/04/18 14:31:57 by mforest-          #+#    #+#             */
/*   Updated: 2026/04/18 14:31:57 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React, { useState, useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { Toaster } from 'react-hot-toast';
import { NotificationProvider } from './components/common/NotificationContext';
import Navbar from './components/common/Navbar';
import MacWindow from './components/common/MacWindow';
import ToastContainer from './components/common/ToastContainer';
import Home from './pages/Home/Home';
import Privacy from './pages/Legal/Privacy';
import Tos from './pages/Legal/Tos';
import NotFound from './pages/NotFound/NotFound';
import Login from './pages/Auth/Login';
import HomeGame from './pages/Game/HomeGame';
import CreateGame from './pages/Game/CreateGame';
import JoinGame from './pages/Game/JoinGame';
import Lobby from './pages/Game/Lobby';
import Game from './pages/Game/Game';
import Profile from './pages/Profile/Profile';
import Friends from './pages/Friends/Friends';
import Credits from './pages/Legal/Credits';
import AICreateGame from './pages/AIGame/AICreateGame';
import AIJoinGame from './pages/AIGame/AIJoinGame';
import AILobby from './pages/AIGame/AILobby';
import AIGame from './pages/AIGame/AIGame';
import AuthCallback from './pages/Auth/Callback';
import './styles/global.css';
import './styles/hypercard.css';

const App = () =>
{
	const [isMobile, setIsMobile] = useState(false);

	useEffect(() =>
	{
		const checkMobile = () =>
		{
			const userAgent = navigator.userAgent;
			const isMobileDevice = /android|iphone|ipad/i.test(userAgent);
			setIsMobile(isMobileDevice);
		};

		checkMobile();
		window.addEventListener('resize', checkMobile);

		return () =>
		{
			window.removeEventListener('resize', checkMobile);
		};
	}, []);

	if (isMobile)
	{
		return (
			<div className="mobile-blocked">
				<h1>Desktop Only</h1>
				<p>This application is not available on mobile devices.</p>
			</div>
		);
	}

	return (
		<Router>
			<NotificationProvider>
				<Toaster position="top-right" />
				<ToastContainer />
				<Navbar />
				<main className="hc-main-container">
					<MacWindow>
						<Routes>
							<Route path="/" element={<Home />} />
							<Route path="/privacy" element={<Privacy />} />
							<Route path="/tos" element={<Tos />} />
							<Route path="/credits" element={<Credits />} />
							<Route path="/login" element={<Login />} />
							<Route path="/game" element={<HomeGame />} />
							<Route path="/game/create" element={<CreateGame />} />
							<Route path="/game/join/:code" element={<JoinGame />} />
							<Route path="/game/lobby/:code" element={<Lobby />} />
							<Route path="/game/play/:code" element={<Game />} />
							<Route path="/aigame/create" element={<AICreateGame />} />
							<Route path="/aigame/join/:code" element={<AIJoinGame />} />
							<Route path="/aigame/lobby/:code" element={<AILobby />} />
							<Route path="/aigame/play/:code" element={<AIGame />} />
							<Route path="/profile/:username" element={<Profile />} />
							<Route path="/friends" element={<Friends />} />
							<Route path="/callback" element={<AuthCallback />} />
							<Route path="*" element={<NotFound />} />
						</Routes>
					</MacWindow>
				</main>
			</NotificationProvider>
		</Router>
	);
};

export default App;
