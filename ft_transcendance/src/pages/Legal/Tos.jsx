/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   Tos.jsx                                            :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/02/20 04:01:44 by mforest-          #+#    #+#             */
/*   Updated: 2026/02/20 04:01:44 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React         from 'react';
import { Link }      from 'react-router-dom';
import './Legal.css';

const Tos = () => {
  return (
    <div className="legal">
      <h2 className="legal__title">Terms of Service</h2>
      <p  className="legal__date">Last updated: February 2026</p>

      <section className="legal__section">
        <h3>1. x</h3>
        <p>to do</p>
      </section>

      <Link to="/" className="legal__back">← Back to Home</Link>
    </div>
  );
};

export default Tos;
