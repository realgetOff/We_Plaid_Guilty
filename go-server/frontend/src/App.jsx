
/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   App.jsx                                            :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: pmilner- <pmilner-@student.42.fr>          +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/02/20 03:51:04 by mforest-          #+#    #+#             */
/*   Updated: 2026/03/04 20:44:42 by pmilner-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React from 'react';
import { BrowserRouter as Router,
         Routes, Route } from 'react-router-dom';
import { Toaster } from 'react-hot-toast';
import { AuthProvider } from './context/AuthContext';
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
import ProtectedRoute from './pages/Legal/ProtectedRoute.jsx'
import './styles/global.css';
import './styles/hypercard.css';

const App = () =>
{
  return (
    <Router>
      <AuthProvider>
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
 			 	<Route path="/login" element={<Login />} />

 			 	<Route
 			   		path="/game"
 			   		element={<ProtectedRoute><HomeGame /></ProtectedRoute>}
 			 	/>
 			 	<Route
 			   		path="/game/create"
 			  		element={<ProtectedRoute><CreateGame /></ProtectedRoute>}
 			 	/>
 			 	<Route
 			   		path="/game/join/:code"
 			   		element={<ProtectedRoute><JoinGame /></ProtectedRoute>}
 			 	/>
 			 	<Route
 			   		path="/game/lobby/:code"
 			    	element={<ProtectedRoute><Lobby /></ProtectedRoute>}
 			 	/>
 			 	<Route
 			    	path="/game/play/:code"
 			    	element={<ProtectedRoute><Game /></ProtectedRoute>}
 			    />
 			    <Route
  			  		path="/profile/:username"
   			 		element={<ProtectedRoute><Profile /></ProtectedRoute>}
  			    />
  				<Route
  					path="/friends"
  			  		element={<ProtectedRoute><Friends /></ProtectedRoute>}
				/>
			  	<Route path="*" element={<NotFound />} />
			</Routes>
            </MacWindow>
          </main>
        </NotificationProvider>
      </AuthProvider>
    </Router>
  );
};

export default App;
