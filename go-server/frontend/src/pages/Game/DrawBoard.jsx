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
	'#aa00aa', '#ff44ff', '#884400', '#00cccc',
];

const TIMER_SEC = 89;
const MAX_UNDO  = 30;
const TOLERANCE = 30;

const colorMatch = (data, idx, r, g, b, a) =>
	Math.abs(data[idx]     - r) <= TOLERANCE &&
	Math.abs(data[idx + 1] - g) <= TOLERANCE &&
	Math.abs(data[idx + 2] - b) <= TOLERANCE &&
	Math.abs(data[idx + 3] - a) <= TOLERANCE;

const floodFill = (ctx, startX, startY, fillColor, canvas) =>
{
	const imageData = ctx.getImageData(0, 0, canvas.width, canvas.height);
	const data      = imageData.data;
	const w         = canvas.width;
	const h         = canvas.height;
	const toIndex   = (x, y) => (y * w + x) * 4;
	const sx        = Math.floor(startX);
	const sy        = Math.floor(startY);
	if (sx < 0 || sx >= w || sy < 0 || sy >= h)
		return;

	const targetIdx = toIndex(sx, sy);
	const targetR   = data[targetIdx];
	const targetG   = data[targetIdx + 1];
	const targetB   = data[targetIdx + 2];
	const targetA   = data[targetIdx + 3];

	const tmp        = document.createElement('canvas');
	tmp.width        = 1;
	tmp.height       = 1;
	const tmpCtx     = tmp.getContext('2d');
	tmpCtx.fillStyle = fillColor;
	tmpCtx.fillRect(0, 0, 1, 1);
	const fd    = tmpCtx.getImageData(0, 0, 1, 1).data;
	const fillR = fd[0], fillG = fd[1], fillB = fd[2], fillA = fd[3];

	if (colorMatch(data, targetIdx, fillR, fillG, fillB, fillA))
		return;

	const visited = new Uint8Array(w * h);
	const stack   = [sx + sy * w];

	while (stack.length > 0)
	{
		const pos = stack.pop();
		const x   = pos % w;
		const y   = Math.floor(pos / w);
		if (x < 0 || x >= w || y < 0 || y >= h)
			continue;
		if (visited[pos])
			continue;
		visited[pos] = 1;
		const idx = pos * 4;
		if (!colorMatch(data, idx, targetR, targetG, targetB, targetA))
			continue;
		data[idx]     = fillR;
		data[idx + 1] = fillG;
		data[idx + 2] = fillB;
		data[idx + 3] = fillA;
		stack.push(pos + 1, pos - 1, pos + w, pos - w);
	}
	ctx.putImageData(imageData, 0, 0);
};

const TOOLS =
[
	{ id: 'pen',         label: '✏',  title: 'Pen',              group: 'draw' },
	{ id: 'eraser',      label: '⌫',  title: 'Eraser',           group: 'draw' },
	{ id: 'bucket',      label: '🪣', title: 'Fill',             group: 'draw' },
	{ id: 'pipette',     label: '💉', title: 'Color picker',     group: 'draw' },
	{ id: 'line',        label: '╱',  title: 'Line',             group: 'shape' },
	{ id: 'rect',        label: '▭',  title: 'Rectangle',        group: 'shape' },
	{ id: 'rect_fill',   label: '▬',  title: 'Rectangle filled', group: 'shape' },
	{ id: 'circle',      label: '◯',  title: 'Circle',           group: 'shape' },
	{ id: 'circle_fill', label: '●',  title: 'Circle filled',    group: 'shape' },
];

const isShapeTool = (t) => ['line', 'rect', 'rect_fill', 'circle', 'circle_fill'].includes(t);

const DrawBoard = ({ prompt, onDone }) =>
{
	const canvasRef     = useRef(null);
	const isDrawing     = useRef(false);
	const lastPos       = useRef(null);
	const startPosRef   = useRef(null);
	const historyRef    = useRef([]);
	const snapshotRef   = useRef(null);

	const [tool,        setTool]        = useState('pen');
	const [color,       setColor]       = useState('#000000');
	const [size,        setSize]        = useState(5);
	const [seconds,     setSeconds]     = useState(TIMER_SEC);
	const [customColor, setCustomColor] = useState('#000000');

	const saveHistory = useCallback(() =>
	{
		const cv   = canvasRef.current;
		const ctx  = cv.getContext('2d');
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
	useEffect(() => { onDoneRef.current = onDone; }, [onDone]);

	useEffect(() =>
	{
		if (seconds <= 0)
		{
			const cv = canvasRef.current;
			onDoneRef.current(cv.toDataURL('image/png'));
			return;
		}
		const id = setTimeout(() => setSeconds((s) => s - 1), 1000);
		return () => clearTimeout(id);
	}, [seconds]);

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
		const cv     = canvasRef.current;
		const rect   = cv.getBoundingClientRect();
		const scaleX = cv.width  / rect.width;
		const scaleY = cv.height / rect.height;
		const src    = e.touches ? e.touches[0] : e;
		return {
			x: (src.clientX - rect.left) * scaleX,
			y: (src.clientY - rect.top)  * scaleY,
		};
	};

	const applyStrokeStyle = (ctx) =>
	{
		ctx.strokeStyle = tool === 'eraser' ? '#ffffff' : color;
		ctx.fillStyle   = tool === 'eraser' ? '#ffffff' : color;
		ctx.lineCap     = 'round';
		ctx.lineJoin    = 'round';
		ctx.lineWidth   = tool === 'eraser' ? size * 3 : size;
	};

	const drawShape = useCallback((ctx, start, end) =>
	{
		applyStrokeStyle(ctx);

		if (tool === 'line')
		{
			ctx.beginPath();
			ctx.moveTo(start.x, start.y);
			ctx.lineTo(end.x, end.y);
			ctx.stroke();
		}
		else if (tool === 'rect')
		{
			ctx.beginPath();
			ctx.strokeRect(start.x, start.y, end.x - start.x, end.y - start.y);
		}
		else if (tool === 'rect_fill')
		{
			ctx.fillRect(start.x, start.y, end.x - start.x, end.y - start.y);
		}
		else if (tool === 'circle')
		{
			const rx = (end.x - start.x) / 2;
			const ry = (end.y - start.y) / 2;
			ctx.beginPath();
			ctx.ellipse(start.x + rx, start.y + ry, Math.abs(rx), Math.abs(ry), 0, 0, Math.PI * 2);
			ctx.stroke();
		}
		else if (tool === 'circle_fill')
		{
			const rx = (end.x - start.x) / 2;
			const ry = (end.y - start.y) / 2;
			ctx.beginPath();
			ctx.ellipse(start.x + rx, start.y + ry, Math.abs(rx), Math.abs(ry), 0, 0, Math.PI * 2);
			ctx.fill();
		}

	}, [tool, color, size]);

	const startDraw = useCallback((e) =>
	{
		e.preventDefault();
		const cv  = canvasRef.current;
		const ctx = cv.getContext('2d');
		const pos = getPos(e);

		if (tool === 'bucket')
		{
			saveHistory();
			floodFill(ctx, pos.x, pos.y, color, cv);
			saveHistory();
			return;
		}

		if (tool === 'pipette')
		{
			const pixel = ctx.getImageData(Math.floor(pos.x), Math.floor(pos.y), 1, 1).data;
			const hex   = '#' + [pixel[0], pixel[1], pixel[2]]
				.map(v => v.toString(16).padStart(2, '0')).join('');
			setColor(hex);
			setCustomColor(hex);
			setTool('pen');
			return;
		}

		isDrawing.current   = true;
		lastPos.current     = pos;
		startPosRef.current = pos;

		if (isShapeTool(tool))
			snapshotRef.current = ctx.getImageData(0, 0, cv.width, cv.height);
		else
			saveHistory();
	}, [tool, color, saveHistory]);


	const draw = useCallback((e) =>
	{
	    if (!isDrawing.current)
	        return;
	    e.preventDefault();
	    const cv  = canvasRef.current;
	    const ctx = cv.getContext('2d');
	    const pos = getPos(e);

	    if (isShapeTool(tool))
	    {
	        ctx.putImageData(snapshotRef.current, 0, 0);
	        drawShape(ctx, startPosRef.current, pos);
	        return;
	    }

	    applyStrokeStyle(ctx);
    
	    if (lastPos.current === startPosRef.current)
		{
	        ctx.beginPath();
	        ctx.moveTo(lastPos.current.x, lastPos.current.y);
	    }
    
	    ctx.lineTo(pos.x, pos.y);
	    ctx.stroke();
    
	    lastPos.current = pos;
	}, [tool, color, size, drawShape]);

	const stopDraw = useCallback(() =>
	{
		if (!isDrawing.current)
			return;
		isDrawing.current = false;
		lastPos.current   = null;

		if (isShapeTool(tool))
			saveHistory();
	}, [tool, saveHistory]);

	const clearCanvas = () =>
	{
		const cv  = canvasRef.current;
		const ctx = cv.getContext('2d');
		ctx.fillStyle = '#ffffff';
		ctx.fillRect(0, 0, cv.width, cv.height);
		saveHistory();
	};

	const handleDone = () =>
	{
		const cv = canvasRef.current;
		onDone(cv.toDataURL('image/png'));
	};

	const pct             = (seconds / TIMER_SEC) * 100;
	const timerColor      = seconds <= 15 ? '#e53935' : seconds <= 30 ? '#ff8800' : '#4488ff';

	return (
		<div className="drawboard">

			{/* timer */}
			<div className="drawboard__timer-track">
				<div className="drawboard__timer-fill" style={{ width: `${pct}%`, background: timerColor }} />
				<span className="drawboard__timer-label">{seconds}s</span>
			</div>

			{/* prompt */}
			<div className="drawboard__prompt-band">
				✏ Draw : <strong>{prompt || '…'}</strong>
			</div>

			{/* toolbar */}
			<div className="drawboard__toolbar">

				{/* draw tools */}
				<div className="drawboard__section">
					<span className="drawboard__section-label">Tools</span>
					<div className="drawboard__toolbar-group">
						{TOOLS.filter(t => t.group === 'draw').map(({ id, label, title }) => (
							<button
								key={id}
								className={`drawboard__tool-btn${tool === id ? ' drawboard__tool-btn--active' : ''}`}
								onClick={() => setTool(id)}
								title={title}
							>
								{label}
							</button>
						))}
					</div>
				</div>

				<div className="drawboard__toolbar-sep" />

				{/* shapes */}
				<div className="drawboard__section">
					<span className="drawboard__section-label">Shapes</span>
					<div className="drawboard__toolbar-group">
						{TOOLS.filter(t => t.group === 'shape').map(({ id, label, title }) => (
							<button
								key={id}
								className={`drawboard__tool-btn${tool === id ? ' drawboard__tool-btn--active' : ''}`}
								onClick={() => setTool(id)}
								title={title}
							>
								{label}
							</button>
						))}
					</div>
				</div>

				<div className="drawboard__toolbar-sep" />

				{/* size+prev */}
				<div className="drawboard__section">
					<span className="drawboard__section-label">Size</span>
					<div className="drawboard__size-row">
						<div
							className="drawboard__size-preview"
							style={{
								width:  Math.min(size, 28) + 4,
								height: Math.min(size, 28) + 4,
								background: tool === 'eraser' ? '#ccc' : color,
								borderRadius: '50%',
							}}
						/>
						<input
							type="range"
							min="1"
							max="40"
							value={size}
							onChange={(e) => setSize(Number(e.target.value))}
							className="drawboard__size-slider"
						/>
						<span className="drawboard__size-value">{size}px</span>
					</div>
				</div>

				<div className="drawboard__toolbar-sep" />

				{/* palette */}
				<div className="drawboard__section">
					<span className="drawboard__section-label">Color</span>
					<div className="drawboard__palette">
						{COLORS.map((c) => (
							<button
								key={c}
								className={`drawboard__swatch${color === c ? ' drawboard__swatch--active' : ''}`}
								style={{ background: c }}
								onClick={() => { setColor(c); setCustomColor(c); if (tool === 'eraser' || tool === 'pipette') setTool('pen'); }}
								title={c}
							/>
						))}
						<input
							type="color"
							className="drawboard__custom-color"
							value={customColor}
							onChange={(e) =>
							{
								setCustomColor(e.target.value);
								setColor(e.target.value);
								if (tool === 'eraser' || tool === 'pipette') setTool('pen');
							}}
							title="Custom color"
						/>
					</div>
				</div>

				<div className="drawboard__toolbar-sep" />

				{/* actions */}
				<div className="drawboard__section">
					<span className="drawboard__section-label">Actions</span>
					<div className="drawboard__toolbar-group">
						<button className="drawboard__tool-btn" onClick={undo}        title="Undo">↩</button>
						<button className="drawboard__tool-btn" onClick={clearCanvas} title="Clear">🗑</button>
					</div>
				</div>

			</div>

			{/* canvas */}
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
