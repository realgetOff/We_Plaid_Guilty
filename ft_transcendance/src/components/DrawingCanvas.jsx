/* ************************************************************************** */
/* */
/* :::      ::::::::   */
/* DrawingCanvas.jsx                                  :+:      :+:    :+:   */
/* +:+ +:+         +:+     */
/* By: mforest- <marvin@d42.fr>                   +#+  +:+       +#+        */
/* +#+#+#+#+#+   +#+           */
/* Created: 2026/02/04 04:06:33 by mforest-          #+#    #+#             */
/* Updated: 2026/02/04 05:55:00 by mforest-         ###   ########.fr       */
/* */
/* ************************************************************************** */

import React, { useRef, useState, useEffect } from 'react';

const DrawingCanvas = () =>
{
	const canvasRef = useRef(null);
	const ctxRef = useRef(null);
	const [isDrawing, setIsDrawing] = useState(false);
	const [lineWidth, setLineWidth] = useState(5);
	const [color, setColor] = useState('#646cff');
	const [isErasing, setIsErasing] = useState(false);
	const [mousePos, setMousePos] = useState({ x: 0, y: 0 });
	const [showCursor, setShowCursor] = useState(false);

	useEffect(() =>
	{
		const canvas = canvasRef.current;
		const ctx = canvas.getContext('2d');
		ctx.lineCap = "round";
		ctx.lineJoin = "round";
		ctxRef.current = ctx;
	}, []);

	useEffect(() =>
	{
		if (ctxRef.current)
		{
			ctxRef.current.lineWidth = lineWidth;
			ctxRef.current.strokeStyle = isErasing ? '#FFFFFF' : color;
		}
	}, [lineWidth, color, isErasing]);

	const getCoordinates = (e) =>
	{
		const canvas = canvasRef.current;
		const rect = canvas.getBoundingClientRect();
		
		// mobile + desktop support
		const clientX = e.touches ? e.touches[0].clientX : e.clientX;
		const clientY = e.touches ? e.touches[0].clientY : e.clientY;

		// relative pos 
		const relX = clientX - rect.left;
		const relY = clientY - rect.top;

		const scaleX = canvas.width / rect.width;
		const scaleY = canvas.height / rect.height;

		return {
			canvasX: relX * scaleX,
			canvasY: relY * scaleY,
			displayX: relX,
			displayY: relY
		};
	};

	const startDrawing = (e) =>
	{
		const coords = getCoordinates(e);
		ctxRef.current.beginPath();
		ctxRef.current.moveTo(coords.canvasX, coords.canvasY);
		setIsDrawing(true);
		setMousePos({ x: coords.displayX, y: coords.displayY });
	};

	const handleMouseMove = (e) =>
	{
		const coords = getCoordinates(e);
		setMousePos({ x: coords.displayX, y: coords.displayY });
		
		if (isDrawing)
		{
			ctxRef.current.lineTo(coords.canvasX, coords.canvasY);
			ctxRef.current.stroke();
		}
	};

	const stopDrawing = () =>
	{
		if (isDrawing)
		{
			ctxRef.current.closePath();
			setIsDrawing(false);
		}
	};

	return (
		<div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '20px', width: '100%' }}>
			<div 
				style={{ 
					position: 'relative',
					width: '100%', 
					maxWidth: '600px', 
					aspectRatio: '6/4',
					borderRadius: '12px', 
					overflow: 'hidden',
					boxShadow: '0 10px 30px rgba(0,0,0,0.5)',
					touchAction: 'none',
					cursor: 'none',
					background: '#fff'
				}}
				onMouseEnter={() => setShowCursor(true)}
				onMouseLeave={() => { setShowCursor(false); stopDrawing(); }}
			>
				<canvas
					ref={canvasRef}
					onMouseDown={startDrawing}
					onMouseMove={handleMouseMove}
					onMouseUp={stopDrawing}
					onTouchStart={startDrawing}
					onTouchMove={handleMouseMove}
					onTouchEnd={stopDrawing}
					width={600}
					height={400}
					style={{ width: '100%', height: '100%', display: 'block' }}
				/>

				{showCursor && (
					<div style={{
						position: 'absolute',
						left: mousePos.x,
						top: mousePos.y,
						width: `${lineWidth}px`,
						height: `${lineWidth}px`,
						border: '1px solid rgba(0,0,0,0.5)',
						backgroundColor: isErasing ? '#ffffff' : color,
						borderRadius: '50%',
						pointerEvents: 'none',
						transform: 'translate(-50%, -50%)',
						zIndex: 10,
						opacity: 0.7
					}} />
				)}
			</div>

			{/* Tools bar */}
			<div style={{ 
				display: 'flex', flexWrap: 'wrap', justifyContent: 'center', alignItems: 'center', 
				gap: '15px', background: 'var(--bg-card)', padding: '12px 25px', 
				borderRadius: '50px', border: '1px solid #333' 
			}}>
				
				{/* Preview + Color Picker */}
				<div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
					<div style={{ 
						width: '35px', height: '35px', display: 'flex', alignItems: 'center', justifyContent: 'center',
						background: '#1a1a1a', borderRadius: '50%'
					}}>
						<div style={{
							width: `${Math.min(lineWidth, 25)}px`, height: `${Math.min(lineWidth, 25)}px`,
							backgroundColor: isErasing ? '#fff' : color, borderRadius: '50%',
							border: isErasing ? '1px solid #444' : 'none'
						}} />
					</div>
					<input 
						type="color" 
						value={color}
						onChange={(e) => { setColor(e.target.value); setIsErasing(false); }}
						style={{ width: '30px', height: '30px', cursor: 'pointer', border: 'none', background: 'transparent' }}
					/>
				</div>

				<div style={{ width: '1px', height: '25px', background: '#444' }} />

				{/* Action button */}
				<button 
					onClick={() => setIsErasing(!isErasing)}
					style={{
						background: isErasing ? 'var(--accent)' : '#222',
						color: 'white', border: '1px solid #444', borderRadius: '12px', padding: '6px 15px', cursor: 'pointer', fontSize: '0.85rem'
					}}
				>
					{isErasing ? '✏️  Draw' : '🧽 Erase'}
				</button>

				<button 
					onClick={() => ctxRef.current.clearRect(0,0,600,400)} 
					style={{ background: '#332222', color: '#ff5555', border: '1px solid #553333', borderRadius: '12px', padding: '6px 15px', cursor: 'pointer', fontSize: '0.85rem' }}
				>
					🗑️ Trash
				</button>

				<div style={{ width: '1px', height: '25px', background: '#444' }} />

				{/* Size selector */}
				<div style={{ display: 'flex', gap: '8px' }}>
		    		{[5, 10, 20, 30].map((size) => (
		        	<button
    		    	    key={size}
    		    	    onClick={() => setLineWidth(size)}
        			    style={{
		        	        width: '32px',
        			        height: '32px',
	            		    borderRadius: '50%',
			                border: lineWidth === size ? '2px solid var(--accent)' : '1px solid #444',
    			            background: lineWidth === size ? '#333' : '#222',
        	    		    color: 'white',
		            	    cursor: 'pointer',
	    		            fontSize: '0.75rem',
    	        		    fontWeight: '500',
		            	    display: 'flex',
        		        	alignItems: 'center',
			                justifyContent: 'center',
    			            padding: 0, 
        	    		    lineHeight: 1 
 		     			    }}
    					    >
        			    {size}
      				  </button>
			    	))}
				</div>
			</div>
		</div>
	);
};

export default DrawingCanvas;
