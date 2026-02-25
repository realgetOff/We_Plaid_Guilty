/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   DrawBoard.jsx                                      :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2026/02/23 23:33:59 by mforest-          #+#    #+#             */
/*   Updated: 2026/02/23 23:33:59 by mforest-         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */

import React, { useRef, useState, useEffect, useCallback } from 'react';
import './DrawBoard.css';

const COLORS =
[
  '#000000', '#ffffff', '#808080', '#c0c0c0',
  '#aa0000', '#ff4444', '#ff8800', '#ffcc00',
  '#008800', '#00aa44', '#0000aa', '#4488ff',
  '#aa00aa', '#ff44ff', '#884400', '#ffaaaa',
];

const SIZES   = [2, 5, 10, 18];
const TIMER_SEC = 80;
const MAX_UNDO  = 30;

const DrawBoard = ({ prompt, onDone }) =>
{
  const canvasRef   = useRef(null);
  const isDrawing   = useRef(false);
  const lastPos     = useRef(null);
  const historyRef  = useRef([]);

  const [tool,    setTool]    = useState('pen');
  const [color,   setColor]   = useState('#000000');
  const [size,    setSize]    = useState(5);
  const [seconds, setSeconds] = useState(TIMER_SEC);

  const saveHistory = useCallback(() =>
  {
    const cv  = canvasRef.current;
    const ctx = cv.getContext('2d');
    const data = ctx.getImageData(0, 0, cv.width, cv.height);
    historyRef.current.push(data);
    if (historyRef.current.length > MAX_UNDO)
      historyRef.current.shift();
  }, []);

  useEffect(() =>
  {
    const cv  = canvasRef.current;
    const ctx = cv.getContext('2d');
    ctx.fillStyle = '#ffffff';
    ctx.fillRect(0, 0, cv.width, cv.height);
    saveHistory();
  }, [saveHistory]);

  const onDoneRef = useRef(onDone);
  useEffect(() =>
  {
    onDoneRef.current = onDone;
  }, [onDone]);

  useEffect(() =>
  {
    if (seconds <= 0)
    {
      const cv = canvasRef.current;
      onDoneRef.current(cv.toDataURL('image/png'));
      return;
    }
    const id = setTimeout(() =>
    {
      setSeconds((s) => s - 1);
    }, 1000);
    return () =>
    {
      clearTimeout(id);
    };
  }, [seconds]);

  const saveHistoryLocal = () =>
  {
    const cv  = canvasRef.current;
    const ctx = cv.getContext('2d');
    const data = ctx.getImageData(0, 0, cv.width, cv.height);
    historyRef.current.push(data);
    if (historyRef.current.length > MAX_UNDO)
      historyRef.current.shift();
  };

  const undo = () =>
  {
    if (historyRef.current.length <= 1)
      return;
    historyRef.current.pop();
    const cv  = canvasRef.current;
    const ctx = cv.getContext('2d');
    ctx.putImageData(historyRef.current[historyRef.current.length - 1], 0, 0);
  };

  const getPos = (e) =>
  {
    const cv   = canvasRef.current;
    const rect = cv.getBoundingClientRect();
    const scaleX = cv.width  / rect.width;
    const scaleY = cv.height / rect.height;
    const src  = e.touches ? e.touches[0] : e;
    return {
      x: (src.clientX - rect.left) * scaleX,
      y: (src.clientY - rect.top)  * scaleY,
    };
  };

  const startDraw = useCallback((e) =>
  {
    e.preventDefault();
    isDrawing.current = true;
    lastPos.current   = getPos(e);
    saveHistoryLocal();
  }, []);

  const draw = useCallback((e) =>
  {
    if (!isDrawing.current)
      return;
    e.preventDefault();
    const cv  = canvasRef.current;
    const ctx = cv.getContext('2d');
    const pos = getPos(e);

    let lineWidth = size;
    if (tool === 'eraser')
      lineWidth = size * 3;

    let strokeColor = color;
    if (tool === 'eraser')
      strokeColor = '#ffffff';

    ctx.lineCap     = 'round';
    ctx.lineJoin    = 'round';
    ctx.lineWidth   = lineWidth;
    ctx.strokeStyle = strokeColor;

    ctx.beginPath();
    ctx.moveTo(lastPos.current.x, lastPos.current.y);
    ctx.lineTo(pos.x, pos.y);
    ctx.stroke();

    lastPos.current = pos;
  }, [tool, color, size]);

  const stopDraw = useCallback(() =>
  {
    isDrawing.current = false;
    lastPos.current   = null;
  }, []);

  const clearCanvas = () =>
  {
    const cv  = canvasRef.current;
    const ctx = cv.getContext('2d');
    ctx.fillStyle = '#ffffff';
    ctx.fillRect(0, 0, cv.width, cv.height);
    saveHistoryLocal();
  };

  const handleDone = () =>
  {
    const cv = canvasRef.current;
    onDone(cv.toDataURL('image/png'));
  };

  const pct = (seconds / TIMER_SEC) * 100;

  let timerBackground = '#000000';
  if (seconds <= 15)
    timerBackground = '#aa0000';

  let penBtnClass = 'drawboard__tool-btn';
  if (tool === 'pen')
    penBtnClass = penBtnClass + ' drawboard__tool-btn--active';

  let eraserBtnClass = 'drawboard__tool-btn';
  if (tool === 'eraser')
    eraserBtnClass = eraserBtnClass + ' drawboard__tool-btn--active';

  return (
    <div className="drawboard">

      <div className="drawboard__timer-track">
        <div
          className="drawboard__timer-fill"
          style={{
            width: `${pct}%`,
            background: timerBackground,
          }}
        />
        <span className="drawboard__timer-label">{seconds}s</span>
      </div>

      <div className="drawboard__prompt-band">
        Draw : <strong>{prompt}</strong>
      </div>

      <div className="drawboard__toolbar">

        <div className="drawboard__toolbar-group">
          <button
            className={penBtnClass}
            onClick={() => setTool('pen')}
            title="Pen"
          >
            ✏
          </button>
          <button
            className={eraserBtnClass}
            onClick={() => setTool('eraser')}
            title="Eraser"
          >
            ⌫
          </button>
        </div>

        <div className="drawboard__toolbar-sep" />

        <div className="drawboard__toolbar-group">
          {SIZES.map((s) =>
          {
            let sizeBtnClass = 'drawboard__size-btn';
            if (size === s)
              sizeBtnClass = sizeBtnClass + ' drawboard__size-btn--active';

            return (
              <button
                key={s}
                className={sizeBtnClass}
                onClick={() => setSize(s)}
                title={`Size ${s}`}
              >
                <span
                  className="drawboard__size-dot"
                  style={{ width: s + 4, height: s + 4 }}
                />
              </button>
            );
          })}
        </div>

        <div className="drawboard__toolbar-sep" />

        <div className="drawboard__palette">
          {COLORS.map((c) =>
          {
            let swatchClass = 'drawboard__swatch';
            if (color === c)
              swatchClass = swatchClass + ' drawboard__swatch--active';

            return (
              <button
                key={c}
                className={swatchClass}
                style={{ background: c }}
                onClick={() =>
                {
                  setColor(c);
                  setTool('pen');
                }}
                title={c}
              />
            );
          })}
        </div>

        <div className="drawboard__toolbar-sep" />

        <div className="drawboard__toolbar-group">
          <button className="drawboard__tool-btn" onClick={undo}        title="Undo">↩</button>
          <button className="drawboard__tool-btn" onClick={clearCanvas} title="Clear">🗑</button>
        </div>

      </div>

      <canvas
        ref={canvasRef}
        className="drawboard__canvas"
        width={700}
        height={380}
        onMouseDown={startDraw}
        onMouseMove={draw}
        onMouseUp={stopDraw}
        onMouseLeave={stopDraw}
        onTouchStart={startDraw}
        onTouchMove={draw}
        onTouchEnd={stopDraw}
      />

      <button className="drawboard__submit-btn" onClick={handleDone}>
        ✓ Done Drawing
      </button>

    </div>
  );
};

export default DrawBoard;
