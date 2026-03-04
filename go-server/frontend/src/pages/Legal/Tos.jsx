/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   Tos.jsx                                            :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/03/04 03:00:23 by mforest-          #+#    #+#             */
/*   Updated: 2026/03/04 03:00:23 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React from 'react';
import { Link } from 'react-router-dom';
import './Legal.css';

const Tos = () =>
{
	return (
		<div className="legal">
			<h2 className="legal__title">Terms of Service</h2>
			<p className="legal__date">Last updated: February 2026</p>

			<section className="legal__section">
				<h3>1. Acceptance of Terms</h3>
				<p>
					By accessing "We Plaid Guilty", you agree to be bound by 
					these terms and all applicable school regulations.
				</p>
			</section>

			<section className="legal__section">
				<h3>2. Code of Conduct</h3>
				<p>
					Users must respect others. Any inappropriate drawings or 
					behavior will result in an immediate ban from the platform.
				</p>
			</section>

			<section className="legal__section">
				<h3>3. Disclaimer</h3>
				<p>
					This project is for educational purposes under the 
					42 curriculum (ft_transcendence).
				</p>
			</section>

			<Link to="/" className="legal__back">← Back to Home</Link>
		</div>
	);
};

export default Tos;
