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

import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import './Auth.css';

function Login()
{
	const navigate = useNavigate();
	const [authToken, setAuthToken] = useState(localStorage.getItem("authToken") || '');

	const handlePlay = async () =>
	{
		if (localStorage.getItem("authToken"))
		{
			navigate("/");
			return;
		}
		try
		{
			const res = await fetch("/api/auth/player", {
				method: "POST",
				headers: {
					"Content-Type": "application/json",
				},
			});

			if (!res.ok)
				throw new Error("Server error");

			const data = await res.json();

			if (data.token)
			{
				localStorage.setItem("authToken", data.token);
				setAuthToken(data.token);
				navigate("/game");
			}
		}
		catch (error)
		{
			console.error("Login failed:", error);
		}
	};

	return (
		<div className="login-container">
			<h1>Click here if ur are not loris</h1>
			<button className="auth__btn auth__btn--primary" onClick={handlePlay}>
				Login
			</button>
			{authToken && (
				<div className="token-display">
					<p>Connected</p>
				</div>
			)}
		</div>
	);
}

export default Login;
