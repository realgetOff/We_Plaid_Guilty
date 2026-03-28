/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   Credits.jsx                                        :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/03/16 21:07:43 by mforest-          #+#    #+#             */
/*   Updated: 2026/03/16 21:07:43 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React, { useState, useEffect } from 'react';
import './Credits.css';

const ALL_CREATORS =
[
  {
    id:       1,
    username: 'namichel',
    tasks:    ['DevOps Network Architecture', 'Docker Config', 'CI/CD'],
  },
  {
    id:       2,
    username: 'pmilner-',
    tasks:    ['Backend Game Logic (Go)', 'Database Management', 'REST API'],
  },
  {
    id:       3,
    username: 'lviravon',
    tasks:    ['Backend Authentication (Go)', 'User Management', 'REST API'],
  },
  {
    id:       4,
    username: 'mforest-',
    tasks:    ['Frontend User Interface (React)', 'UI/UX', 'API Integration'],
  },
];

const Credits = () =>
{
  const [creators, setCreators] = useState([]);

  useEffect(() =>
  {
    const lorisCredit = localStorage.getItem('loris_credit');
    let list = [...ALL_CREATORS];

    if (lorisCredit === 'null')
    {
      list = ALL_CREATORS.filter((c) => c.username !== 'lviravon');
      list = list.map((c) =>
      {
        if (c.username !== 'pmilner-')
          return c;
        return {
          ...c,
          tasks: ['Backend Game Logic (Go)', 'Backend Authentication (Go)', 'User Management', 'Database Management', 'REST API'],
        };
      });
    }

    setCreators(list);
  }, []);

  let rows = creators.map((c) =>
  {
    let initials = c.username.slice(0, 2).toUpperCase();

    let tags = c.tasks.map((t) =>
    {
      return (
        <span key={t} className="credits__tag">{t}</span>
      );
    });

    return (
      <div key={c.id} className="credits__row">
        <div className="credits__avatar">{initials}</div>
        <div className="credits__info">
          <span className="credits__username">{c.username}</span>
          <div className="credits__tags">{tags}</div>
        </div>
      </div>
    );
  });

    return (
      <div className="credits">
        <div className="credits__card-body">
          {rows}
        </div>
        <p className="credits__footer">
      made with <b>COFFEE</b> and suffering at 42 school
        </p>
      </div>
    );
};

export default Credits;
