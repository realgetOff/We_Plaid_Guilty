import React from 'react';
import Swal from 'sweetalert2';
import toast from 'react-hot-toast';
import DrawingCanvas from '../../components/DrawingCanvas.jsx';

const Game = () =>
{
	const handleFinish = () =>
	{
		Swal.fire({
			title: 'Finish ?',
			text: "Your drawing will be sent !",
			icon: 'question',
			background: '#1e1e1e',
			color: '#ffffff',
			confirmButtonColor: '#646cff',
			cancelButtonColor: '#333',
			showCancelButton: true,
			confirmButtonText: 'Yes, send.',
			cancelButtonText: 'Wait !'
		}).then((result) =>
		{
			if (result.isConfirmed)
			{
				toast.success('Drawing sent!', {
					style: {
						background: '#1e1e1e',
						color: '#4caf50',
						border: '1px solid #4caf50'
					}
				});
			}
		});
	};

	return (
		<div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', width: '100%', padding: '20px', boxSizing: 'border-box' }}>
			<h1 style={{ color: 'var(--accent)', marginBottom: '20px' }}>Gartic Phone</h1>
			
			<DrawingCanvas />
			
			<button 
				onClick={handleFinish} 
				style={{ 
					marginTop: '30px',
					padding: '12px 40px',
					fontSize: '1.1rem',
					backgroundColor: '#646cff',
					color: 'white',
					border: 'none',
					borderRadius: '50px',
					cursor: 'pointer',
					fontWeight: 'bold',
					boxShadow: '0 4px 15px rgba(100, 108, 255, 0.3)',
					transition: 'transform 0.2s'
				}}
				onMouseEnter={(e) => e.target.style.transform = 'scale(1.05)'}
				onMouseLeave={(e) => e.target.style.transform = 'scale(1)'}
			>
				Finish!
			</button>
		</div>
	);
};

export default Game;
