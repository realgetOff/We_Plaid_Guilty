/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   Home.jsx                                           :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/02/20 03:59:25 by mforest-          #+#    #+#             */
/*   Updated: 2026/02/20 03:59:25 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React, { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { connect } from '../../api/socket';
import './Home.css';

const Home = () =>
{
  const navigate = useNavigate();


	useEffect(() => {
	  const login_check = localStorage.getItem("authToken"); // or whatever your key is
	 
	  console.log("TEST LOGIN_CHECK", login_check);
	 
	  if (login_check) {
	    connect();
	  }
	}, []);

  return (
    <div className="home">

      <div className="home__icon" aria-hidden="true">🎨</div>

      <h1 className="home__title">We Plaid Guilty</h1>

      <hr className="home__divider" />

      <p className="home__subtitle">
        Welcome to ft_transcendence.<br />
        A Gartic Phone experience — draw, guess, laugh.
      </p>

      <div className="home__info-box">
        <strong>How to play:</strong><br />
        ① Click [▶ PLAY NOW]<br />
        ② Select either a normal game or an AI game.<br />
        ③ Login, register or continue as a guest.<br />
        ④ Have fun!
      </div>

      <button
        className="home__play-btn"
        onClick={() => navigate('/game')}
      >
        ▶ Play Now
      </button>

    </div>
  );
};

export default Home;
