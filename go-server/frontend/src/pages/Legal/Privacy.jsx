/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   Privacy.jsx                                        :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/03/04 02:59:53 by mforest-          #+#    #+#             */
/*   Updated: 2026/03/04 02:59:53 by mforest-         ###   ########.fr       */
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
			<p className="legal__date">Last updated: February 2026</p>

			<section className="legal__section">
				<h3>1. Data Protection</h3>
				<p>
					We are committed to protecting your personal data and 
					respecting your privacy according to school standards.
				</p>
			</section>

			<section className="legal__section">
				<h3>2. Usage</h3>
				<p>
					Your data is only used for the purpose of the 
					ft_transcendence project and is not shared with third parties.
				</p>
			</section>

			<Link to="/" className="legal__back">← Back to Home</Link>
		</div>
	);
};

export default Privacy;
