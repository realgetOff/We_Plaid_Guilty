/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   App.jsx                                            :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/02/04 04:06:57 by mforest-          #+#    #+#             */
/*   Updated: 2026/02/04 04:06:57 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { Toaster } from 'react-hot-toast';
import Navbar from './components/common/Navbar';
import Home from './pages/Home/Home';
import Game from './pages/Game/Game';

const App = () =>
{
	return (
		<Router>
			<Toaster position="top-right" />
			<div style={{ 
				minHeight: '100vh', 
				display: 'flex', 
				flexDirection: 'column', 
				backgroundColor: 'var(--bg-dark)',
				width: '100vw',
				overflowX: 'hidden'
			}}>
				<Navbar /> 
				<main style={{ 
					flex: 1, 
					display: 'flex', 
					flexDirection: 'column',
					alignItems: 'center', 
					justifyContent: 'center',
					width: '100%' 
				}}>
					<Routes>
						<Route path="/" element={<Home />} />
						<Route path="/game" element={<Game />} />
					</Routes>
				</main>
			</div>
		</Router>
	);
};

export default App;
