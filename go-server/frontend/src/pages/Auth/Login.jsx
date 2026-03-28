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
	const [authToken, setAuthToken] = useState(
		localStorage.getItem("authToken") || ''
	);

    const handlePlay = async () =>
	{
        const token = localStorage.getItem("authToken");
        if (token)
		{
			setAuthToken(token);
        }
        try
		{
            const res = await fetch("/api/player",
			{
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
            });

            if (!res.ok)
			{
                throw new Error("Server error");
            }

            const data = await res.json();

			localStorage.setItem("authToken", data.token);
			setAuthToken(data.token);

            console.log("Player registered:", data);

            navigate("/game");
        }
		catch (error)
		{
            console.error("Request failed:", error);
        }
    };

    return (
		<div className="login-container">
			<h1>Click on the button if u are gay</h1>
			<button
				className="auth__btn auth__btn--primary"
				onClick={handlePlay}
			>
				I am!
			</button>
			<h2>
			{authToken && (
				<p><strong>{authToken}</strong></p>
			)}
			</h2>
		</div>
    );
}

export default Login;
