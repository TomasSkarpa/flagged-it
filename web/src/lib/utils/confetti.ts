/**
 * Simple confetti effect using canvas
 * Creates a celebration effect for game wins
 */

interface ConfettiParticle {
	x: number;
	y: number;
	vx: number;
	vy: number;
	color: string;
	size: number;
	rotation: number;
	rotationSpeed: number;
}

export function triggerConfetti(options?: {
	particleCount?: number;
	spread?: number;
	origin?: { x: number; y: number };
	duration?: number;
}) {
	const {
		particleCount = 50,
		spread = 70,
		origin = { x: 0.5, y: 0.5 },
		duration = 3000
	} = options || {};

	// Create canvas element
	const canvas = document.createElement('canvas');
	canvas.style.position = 'fixed';
	canvas.style.top = '0';
	canvas.style.left = '0';
	canvas.style.width = '100%';
	canvas.style.height = '100%';
	canvas.style.pointerEvents = 'none';
	canvas.style.zIndex = '9999';
	document.body.appendChild(canvas);

	const ctx = canvas.getContext('2d');
	if (!ctx) return;

	// Set canvas size
	canvas.width = window.innerWidth;
	canvas.height = window.innerHeight;

	// Colors for confetti
	const colors = ['#FFD700', '#FF6B6B', '#4ECDC4', '#45B7D1', '#FFA07A', '#98D8C8', '#F7DC6F', '#BB8FCE'];

	// Create particles
	const particles: ConfettiParticle[] = [];
	const originX = origin.x * canvas.width;
	const originY = origin.y * canvas.height;

	for (let i = 0; i < particleCount; i++) {
		const angle = (Math.PI * 2 * i) / particleCount + (Math.random() - 0.5) * spread * (Math.PI / 180);
		const velocity = 2 + Math.random() * 3;
		
		particles.push({
			x: originX,
			y: originY,
			vx: Math.cos(angle) * velocity,
			vy: Math.sin(angle) * velocity - 2,
			color: colors[Math.floor(Math.random() * colors.length)],
			size: 4 + Math.random() * 6,
			rotation: Math.random() * Math.PI * 2,
			rotationSpeed: (Math.random() - 0.5) * 0.2
		});
	}

	// Animation
	let startTime: number | null = null;
	let animationId: number;

	function animate(timestamp: number) {
		if (!startTime) startTime = timestamp;
		const elapsed = timestamp - startTime;

		if (elapsed > duration) {
			document.body.removeChild(canvas);
			return;
		}

		// Clear canvas
		ctx.clearRect(0, 0, canvas.width, canvas.height);

		// Update and draw particles
		particles.forEach(particle => {
			// Update position
			particle.x += particle.vx;
			particle.y += particle.vy;
			particle.vy += 0.1; // Gravity
			particle.rotation += particle.rotationSpeed;

			// Draw particle
			ctx.save();
			ctx.translate(particle.x, particle.y);
			ctx.rotate(particle.rotation);
			ctx.fillStyle = particle.color;
			ctx.fillRect(-particle.size / 2, -particle.size / 2, particle.size, particle.size);
			ctx.restore();
		});

		animationId = requestAnimationFrame(animate);
	}

	animationId = requestAnimationFrame(animate);

	// Cleanup function
	return () => {
		cancelAnimationFrame(animationId);
		if (canvas.parentNode) {
			document.body.removeChild(canvas);
		}
	};
}
