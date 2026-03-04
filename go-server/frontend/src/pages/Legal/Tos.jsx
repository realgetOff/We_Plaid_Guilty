/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   Tos.jsx                                            :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
<<<<<<< HEAD:go-server/frontend/src/pages/Legal/Tos.jsx
<<<<<<<< HEAD:go-server/frontend/src/pages/Legal/Tos.jsx
/*   Created: 2026/02/20 04:01:44 by mforest-          #+#    #+#             */
/*   Updated: 2026/02/20 04:01:44 by mforest-         ###   ########.fr       */
========
/*   Created: 2026/02/20 03:59:25 by mforest-          #+#    #+#             */
/*   Updated: 2026/02/20 03:59:25 by mforest-         ###   ########.fr       */
>>>>>>>> dev:ft_transcendance/src/pages/Home/Home.jsx
=======
/*   Created: 2026/02/20 04:01:44 by mforest-          #+#    #+#             */
/*   Updated: 2026/02/20 04:01:44 by mforest-         ###   ########.fr       */
>>>>>>> dev:ft_transcendance/src/pages/Legal/Tos.jsx
/*                                                                            */
/* ************************************************************************** */

import React from 'react';
<<<<<<< HEAD:go-server/frontend/src/pages/Legal/Tos.jsx
<<<<<<<< HEAD:go-server/frontend/src/pages/Legal/Tos.jsx
import { Link } from 'react-router-dom';
import './Legal.css';
========
import { useNavigate } from 'react-router-dom';
import './Home.css';
>>>>>>>> dev:ft_transcendance/src/pages/Home/Home.jsx

const Tos = () =>
{
<<<<<<<< HEAD:go-server/frontend/src/pages/Legal/Tos.jsx
=======
import { Link } from 'react-router-dom';
import './Legal.css';

const Tos = () =>
{
>>>>>>> dev:ft_transcendance/src/pages/Legal/Tos.jsx
  return (
    <div className="legal">
      <h2 className="legal__title">Terms of Service</h2>
      <p  className="legal__date">Last updated: February 2026</p>

      <section className="legal__section">
        <h3>1. x</h3>
        <p>to do</p>
      </section>

      <Link to="/" className="legal__back">← Back to Home</Link>
<<<<<<< HEAD:go-server/frontend/src/pages/Legal/Tos.jsx
========
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

>>>>>>>> dev:ft_transcendance/src/pages/Home/Home.jsx
=======
>>>>>>> dev:ft_transcendance/src/pages/Legal/Tos.jsx
    </div>
  );
};

export default Tos;
