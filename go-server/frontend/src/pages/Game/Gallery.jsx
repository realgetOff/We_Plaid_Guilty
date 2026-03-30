/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   Gallery.jsx                                        :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/03/01 23:06:13 by mforest-          #+#    #+#             */
/*   Updated: 2026/03/01 23:06:13 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React, { useRef, useEffect } from 'react';
import './Gallery.css';

const Gallery = ({ chains, onBack }) =>
{
  const scrollRef = useRef(null);

  useEffect(() =>
  {
    scrollRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, []);

  if (!Array.isArray(chains) || chains.length === 0)
  {
    return (
      <div className="gallery">
        <p className="gallery__empty">No chains to display yet.</p>
        <button className="gallery__btn" onClick={onBack}>
          ← Back to game
        </button>
      </div>
    );
  }

  return (
    <div className="gallery">
      <div className="gallery__scroll" ref={scrollRef}>
        {chains.map((chain, idx) =>
        {
          return (
            <div key={chain.id ?? idx} className="gallery__chain">
              <div className="gallery__chain-header">
                Chain #{idx + 1}
              </div>
              <div className="gallery__chain-body">
                <div className="gallery__step gallery__step--prompt">
                  <span className="gallery__step-label">Prompt</span>
                  <p className="gallery__step-content">{chain.prompt ?? '(empty)'}</p>
                </div>
                {Array.isArray(chain.steps) && chain.steps.map((step, si) =>
                {
                  return (
                    <div key={si} className={`gallery__step gallery__step--${step.type}`}>
                      <span className="gallery__step-label">
                        {step.type === 'drawing' ? '🎨 Drawing' : '🔍 Guess'}
                      </span>
                      {step.type === 'drawing' && step.content ? (
                        <img
                          src={step.content}
                          alt=""
                          className="gallery__step-img"
                        />
                      ) : (
                        <p className="gallery__step-content">{step.content ?? '(empty)'}</p>
                      )}
                    </div>
                  );
                })}
              </div>
            </div>
          );
        })}
      </div>
      <button className="gallery__btn" onClick={onBack}>
        ← Back to game
      </button>
    </div>
  );
};

export default Gallery;
