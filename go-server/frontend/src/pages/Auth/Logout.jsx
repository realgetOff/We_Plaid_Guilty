/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   Logout.jsx                                         :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/03/11 05:12:21 by mforest-          #+#    #+#             */
/*   Updated: 2026/03/11 05:12:21 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { disconnect } from '../../api/socket';

const Logout = () =>
{
	const navigate = useNavigate();

	useEffect(() =>
	{
		const performLogout = () =>
		{
			disconnect();

			localStorage.removeItem("authToken");
			localStorage.removeItem("isGuest");

			setTimeout(() =>
			{
				navigate('/login');
			}, 250);
		};

		performLogout();
	}, [navigate]);

	return (
		<div className="logout-screen">
			<p>Logout...</p>
		</div>
	);
};

export default Logout;
