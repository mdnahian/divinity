import React, { useEffect, useState } from 'react';

export default function WorldLoadingSpinner() {
  const [dots, setDots] = useState('');

  useEffect(() => {
    const iv = setInterval(() => {
      setDots(d => d.length >= 3 ? '' : d + '.');
    }, 500);
    return () => clearInterval(iv);
  }, []);

  return (
    <div style={styles.overlay}>
      <div style={styles.vignette} />
      <div style={styles.content}>
        <div style={styles.emblemWrap}>
          <div style={styles.emblem}>
            <div style={styles.emblemInner}>☥</div>
          </div>
        </div>
        <div style={styles.text}>Loading world{dots}</div>
      </div>
      <style>{keyframes}</style>
    </div>
  );
}

const keyframes = `
  @keyframes wls-pulse-glow {
    0%, 100% { box-shadow: 0 0 20px rgba(212,160,32,0.3), 0 0 60px rgba(212,160,32,0.1); }
    50% { box-shadow: 0 0 30px rgba(212,160,32,0.5), 0 0 80px rgba(212,160,32,0.2); }
  }
  @keyframes wls-spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
  }
`;

const styles = {
  overlay: {
    position: 'fixed',
    inset: 0,
    zIndex: 99999,
    background: '#0e0e0a',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
  },
  vignette: {
    position: 'absolute',
    inset: 0,
    background: 'radial-gradient(ellipse at center, transparent 40%, rgba(0,0,0,0.6) 100%)',
    pointerEvents: 'none',
  },
  content: {
    position: 'relative',
    textAlign: 'center',
  },
  emblemWrap: {
    display: 'flex',
    justifyContent: 'center',
    marginBottom: '24px',
  },
  emblem: {
    width: '72px',
    height: '72px',
    borderRadius: '50%',
    border: '3px solid #8B6914',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    background: 'radial-gradient(circle, rgba(212,160,32,0.15) 0%, transparent 70%)',
    animation: 'wls-pulse-glow 3s ease-in-out infinite',
  },
  emblemInner: {
    fontSize: '32px',
    color: '#D4A020',
    animation: 'wls-spin 6s linear infinite',
    textShadow: '0 0 12px rgba(212,160,32,0.6)',
  },
  text: {
    fontFamily: "'Press Start 2P', monospace",
    fontSize: '10px',
    color: '#A8841E',
    letterSpacing: '2px',
    minWidth: '200px',
  },
};
