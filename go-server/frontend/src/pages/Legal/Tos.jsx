/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   Tos.jsx                                            :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: pmilner- <pmilner-@student.42.fr>          +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/03/04 03:00:23 by mforest-          #+#    #+#             */
/*   Updated: 2026/03/04 22:12:23 by pmilner-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React from 'react';
import { Link } from 'react-router-dom';
import './Legal.css';

const Home = () =>
{
	return (
		<div className="legal">
			<h2 className="legal__title">Terms of Eternal Servitude</h2>
			<p className="legal__date">Last updated: Too late now</p>

			<section className="legal__section">
				<h3>1. Absolute Submission</h3>
				<p>
					By clicking "Accept", you acknowledge that "We Plaid Guilty"
					is now the legal guardian of your free time. You agree to
					debug C++ at 3:00 AM if we feel like it.
				</p>
			</section>

			<section className="legal__section">
				<h3>2. The "No Complaining" Clause</h3>
				<p>
					Any attempt to complain about Segfaults or leaks will result
					in your account being replaced by a script that only types
					"Norm Error" in your terminal forever.
				</p>
			</section>

			<section className="legal__section">
				<h3>3. Mandatory Donations</h3>
				<p>
					We reserve the right to use your GPU to mine obscure
					cryptocurrencies while you are trying to finish your
					ft_transcendence features.
				</p>
			</section>

			<section className="legal__section">
				<h3>4. Style Over Substance</h3>
				<p>
					If your CSS is ugly, we are legally allowed to delete
					your entire home directory. Plaid patterns are mandatory
					on Tuesdays.
				</p>
			</section>

			<Link to="/" className="legal__back">← I Agree to Everything</Link>
		</div>
	);
};

export default Home;
