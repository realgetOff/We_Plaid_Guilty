/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   Privacy.jsx                                        :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: pmilner- <pmilner-@student.42.fr>          +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/03/04 02:59:53 by mforest-          #+#    #+#             */
/*   Updated: 2026/03/04 22:12:15 by pmilner-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React from 'react';
import { Link } from 'react-router-dom';
import './Legal.css';

const Navbar = () =>
{
	return (
		<div className="legal">
			<h2 className="legal__title">Privacy? What Privacy?</h2>
			<p className="legal__date">Last updated: The moment you clicked</p>

			<section className="legal__section">
				<h3>1. Data Harvesting</h3>
				<p>
					We reserve the right to monitor your keystrokes, your
					unfiltered thoughts, and the exact brand of snacks you
					consume while coding. If you blink, we log it.
				</p>
			</section>

			<section className="legal__section">
				<h3>2. Deep Dark Web Sharing</h3>
				<p>
					Your personal data, including your most embarrassing
					git commit messages, will be sold to the highest bidder,
					or simply traded for a lukewarm cup of cluster coffee.
				</p>
			</section>

			<section className="legal__section">
				<h3>3. Biometric Surrender</h3>
				<p>
					By staying on this page for more than five seconds, you
					legally grant us ownership of your digital soul and
					your first-born's future GitHub username.
				</p>
			</section>

			<Link to="/" className="legal__back">← Escape while you can</Link>
		</div>
	);
};

export default Navbar;
