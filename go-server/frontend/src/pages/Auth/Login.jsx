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

function Login()
{
    const navigate = useNavigate();

    const handlePlay = async () =>
	{
        let playerName = localStorage.getItem("playerName");

        if (!playerName)
		{
            playerName = "player_" + Math.random().toString(36).slice(2, 8);
            localStorage.setItem("playerName", playerName);
        }

        try
		{
            const res = await fetch("/api/player",
			{
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({ playerName }),
            });

            if (!res.ok)
			{
                throw new Error("Server error");
            }

            const data = await res.json();
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
		</div>
		
    );
}

export default Login;
