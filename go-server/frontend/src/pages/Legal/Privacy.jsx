/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   Privacy.jsx                                        :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/04/18 14:28:53 by mforest-          #+#    #+#             */
/*   Updated: 2026/04/18 14:28:53 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React from 'react';
import { Link } from 'react-router-dom';
import './Legal.css';

const Privacy = () => {
    return (
        <div className="legal">
            <h2 className="legal__title">Privacy Policy</h2>
            <p className="legal__date">Last updated: April 18, 2026</p>

            <section className="legal__section">
                <h3>1. Data Collection</h3>
                <p>
                    To provide you with the best experience, we collect essential information such as your username.
                </p>
            </section>

            <section className="legal__section">
                <h3>2. Data Usage</h3>
                <p>
                    Your data is used strictly to manage your account, track your game statistics, and enable social features like friend lists and real-time chat. Your security is a priority, and all tokens are properly hashed and secured.
                </p>
            </section>

            <section className="legal__section">
                <h3>3. User Rights</h3>
                <p>
                    You have the right to access, update, or request the deletion of your personal data at any time. Please note that this platform is built as an educational project for the 42 curriculum.
                </p>
            </section>

            <Link to="/" className="legal__back">← Back to Home</Link>
        </div>
    );
};

export default Privacy;
