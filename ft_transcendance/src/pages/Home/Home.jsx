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

import React from 'react';
import { useNavigate } from 'react-router-dom';
import './Home.css';

const Home = () => {
  const navigate = useNavigate();

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
        ① Buy a gun.<br />
        ② Call local Domino's Pizza.<br />
        ③ Ask for 20 giant pizza, 42 Angouleme.<br />
        ④ Force lviravon to work (and abuse him btw).
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
