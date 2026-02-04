/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   Navbar.jsx                                         :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/02/04 04:06:43 by mforest-          #+#    #+#             */
/*   Updated: 2026/02/04 04:06:43 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React from 'react';
import { Link } from 'react-router-dom';

const Navbar = () =>
{
	return (
		<nav style={{
			width: '100%',
			backgroundColor: '#1a1a1a',
			padding: '1rem 2rem',
			display: 'flex',
			alignItems: 'center',
			boxSizing: 'border-box',
			borderBottom: '1px solid #333'
		}}>
			<div style={{ display: 'flex', gap: '20px', color: 'white' }}>
				<Link to="/" style={{ color: 'white', textDecoration: 'none' }}>Home</Link>
				<span style={{ color: '#444' }}>|</span>
				<Link to="/game" style={{ color: 'white', textDecoration: 'none' }}>Play to Gartic Phone</Link>
			</div>
		</nav>
	);
};

export default Navbar;
