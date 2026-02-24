/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   Privacy.jsx                                        :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/02/20 04:01:08 by mforest-          #+#    #+#             */
/*   Updated: 2026/02/20 04:01:08 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React from 'react';
import { Link } from 'react-router-dom';
import './Legal.css';

const Privacy = () =>
{
  return (
    <div className="legal">
      <h2 className="legal__title">Privacy Policy</h2>
      <p  className="legal__date">Last updated: February 2026</p>

      <section className="legal__section">
        <h3>1. lazy asshole get back to work</h3>
        <p>todo</p>
      </section>

      <Link to="/" className="legal__back">← Back to Home</Link>
    </div>
  );
};

export default Privacy;
