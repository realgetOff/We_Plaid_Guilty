/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   Tos.jsx                                            :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/04/18 14:29:04 by mforest-          #+#    #+#             */
/*   Updated: 2026/04/18 14:29:04 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React from 'react';
import { Link } from 'react-router-dom';
import './Legal.css';

const Tos = () => {
    return (
        <div className="legal">
            <h2 className="legal__title">Terms of Service</h2>
            <p className="legal__date">Last updated: April 18, 2026</p>

            <section className="legal__section">
                <h3>1. Acceptance of Terms</h3>
                <p>
                    By accessing this web application, you agree to comply with these terms. This platform supports multiple users simultaneously, and you agree to interact respectfully with others without causing conflicts or data corruption.
                </p>
            </section>

            <section className="legal__section">
                <h3>2. User Conduct</h3>
                <p>
                    Users are expected to engage fairly in our multiplayer games. Any exploitation of bugs, inappropriate behavior in the real-time chat, or attempts to compromise the application's security will result in account suspension.
                </p>
            </section>

            <section className="legal__section">
                <h3>3. Educational Project Disclaimer</h3>
                <p>
                    This application is developed strictly as part of the ft_transcendence project within the 42 curriculum. The service is provided "as is" for evaluation and learning purposes.
                </p>
            </section>

            <Link to="/" className="legal__back">← Accept and Continue</Link>
        </div>
    );
};

export default Tos;
