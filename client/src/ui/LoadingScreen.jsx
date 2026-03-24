import React, { useEffect, useState } from 'react';
import { useGameStore } from '../hooks/useGameStore';

const LORE_LINES = [
  'The gods shape the world from primordial chaos...',
  'Mountains rise from the bones of ancient titans...',
  'Rivers carve paths through untamed wilderness...',
  'Villages take root in fertile valleys...',
  'The first fires are lit against the darkness...',
  'Kingdoms are forged by iron will and divine mandate...',
  'Forests whisper secrets of ages past...',
  'The forge of creation burns eternal...',
  'Mortal souls awaken to their destinies...',
  'The celestial tapestry unfolds across the land...',
  'Ancient roads connect the corners of the realm...',
  'Crops are sown in blessed soil...',
  'Walls of stone guard the dreams of the faithful...',
  'Stars align to herald a new age...',
  'The balance between light and shadow is set...',
];

export default function LoadingScreen() {
  const progress = useGameStore(s => s.genesisProgress);
  const [dots, setDots] = useState('');
  const [loreIdx, setLoreIdx] = useState(0);
  const [loreFade, setLoreFade] = useState(true);

  useEffect(() => {
    const iv = setInterval(() => {
      setDots(d => d.length >= 3 ? '' : d + '.');
    }, 500);
    return () => clearInterval(iv);
  }, []);

  useEffect(() => {
    const iv = setInterval(() => {
      setLoreFade(false);
      setTimeout(() => {
        setLoreIdx(i => (i + 1) % LORE_LINES.length);
        setLoreFade(true);
      }, 400);
    }, 4000);
    return () => clearInterval(iv);
  }, []);

  const phase = progress?.phase || 'Awakening';
  const detail = progress?.detail || 'Connecting to the divine realm...';
  const current = progress?.current || 0;
  const total = progress?.total || 1;
  const pct = Math.min(100, Math.round((current / total) * 100));

  return (
    <div style={styles.overlay}>
      <div style={styles.vignette} />
      <div style={styles.content}>
        <div style={styles.ornamentTop}>&#9776; &#10040; &#9776;</div>

        <h1 style={styles.title}>DIVINITY</h1>
        <div style={styles.subtitle}>The Gods Are Shaping The World</div>

        <div style={styles.emblemWrap}>
          <div style={styles.emblem}>
            <div style={styles.emblemInner}>☥</div>
          </div>
        </div>

        {/* Progress bar — driven by real backend progress */}
        <div style={styles.barOuter}>
          <div style={{
            ...styles.barInner,
            width: `${Math.max(5, pct)}%`,
            transition: 'width 0.8s ease-out',
          }}>
            <div style={styles.barShimmer} />
          </div>
        </div>

        {/* Phase name */}
        <div style={styles.phase}>{phase}{dots}</div>

        {/* Detail text */}
        <div style={styles.detail}>{detail}</div>

        {/* Step counter */}
        {total > 1 && (
          <div style={styles.counter}>{current} / {total}</div>
        )}

        {/* Lore text */}
        <div style={{
          ...styles.lore,
          opacity: loreFade ? 1 : 0,
        }}>
          {LORE_LINES[loreIdx]}
        </div>

        <div style={styles.ornamentBottom}>&#9776; &#10040; &#9776;</div>
      </div>

      <style>{keyframes}</style>
    </div>
  );
}

const keyframes = `
  @keyframes pulse-glow {
    0%, 100% { box-shadow: 0 0 20px rgba(212,160,32,0.3), 0 0 60px rgba(212,160,32,0.1); }
    50% { box-shadow: 0 0 30px rgba(212,160,32,0.5), 0 0 80px rgba(212,160,32,0.2); }
  }
  @keyframes spin-slow {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
  }
  @keyframes shimmer {
    0% { transform: translateX(-100%); }
    100% { transform: translateX(200%); }
  }
  @keyframes float {
    0%, 100% { transform: translateY(0); }
    50% { transform: translateY(-6px); }
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
    padding: '40px 60px',
    animation: 'float 4s ease-in-out infinite',
  },
  ornamentTop: {
    fontFamily: 'serif',
    fontSize: '18px',
    color: '#8B6914',
    letterSpacing: '12px',
    marginBottom: '20px',
    opacity: 0.6,
  },
  title: {
    fontFamily: "'Press Start 2P', monospace",
    fontSize: '20px',
    color: '#D4A020',
    textShadow: '0 0 20px rgba(212,160,32,0.4), 0 2px 0 #6B4E0A',
    letterSpacing: '6px',
    marginBottom: '12px',
  },
  subtitle: {
    fontFamily: "'Press Start 2P', monospace",
    fontSize: '8px',
    color: '#A8841E',
    letterSpacing: '3px',
    textTransform: 'uppercase',
    marginBottom: '36px',
    opacity: 0.8,
  },
  emblemWrap: {
    display: 'flex',
    justifyContent: 'center',
    marginBottom: '36px',
  },
  emblem: {
    width: '80px',
    height: '80px',
    borderRadius: '50%',
    border: '3px solid #8B6914',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    background: 'radial-gradient(circle, rgba(212,160,32,0.15) 0%, transparent 70%)',
    animation: 'pulse-glow 3s ease-in-out infinite',
  },
  emblemInner: {
    fontSize: '36px',
    color: '#D4A020',
    animation: 'spin-slow 8s linear infinite',
    textShadow: '0 0 12px rgba(212,160,32,0.6)',
  },
  barOuter: {
    width: '300px',
    height: '12px',
    margin: '0 auto 20px',
    background: '#1a1a14',
    borderRadius: '6px',
    border: '2px solid #6B4E0A',
    overflow: 'hidden',
    boxShadow: 'inset 0 2px 4px rgba(0,0,0,0.5)',
  },
  barInner: {
    height: '100%',
    background: 'linear-gradient(90deg, #8B6914, #D4A020, #8B6914)',
    borderRadius: '4px',
    position: 'relative',
    overflow: 'hidden',
  },
  barShimmer: {
    position: 'absolute',
    top: 0,
    left: 0,
    width: '50%',
    height: '100%',
    background: 'linear-gradient(90deg, transparent, rgba(255,255,255,0.25), transparent)',
    animation: 'shimmer 2s ease-in-out infinite',
  },
  phase: {
    fontFamily: "'Press Start 2P', monospace",
    fontSize: '9px',
    color: '#C4A030',
    letterSpacing: '2px',
    marginBottom: '8px',
  },
  detail: {
    fontFamily: "'Courier New', 'Consolas', monospace",
    fontSize: '12px',
    color: '#A8841E',
    marginBottom: '8px',
    minHeight: '18px',
  },
  counter: {
    fontFamily: "'Press Start 2P', monospace",
    fontSize: '7px',
    color: '#6B4E0A',
    letterSpacing: '1px',
    marginBottom: '20px',
  },
  lore: {
    fontFamily: "'Courier New', 'Consolas', monospace",
    fontSize: '13px',
    color: '#8A7056',
    fontStyle: 'italic',
    maxWidth: '400px',
    margin: '0 auto 24px',
    lineHeight: '1.6',
    minHeight: '42px',
    transition: 'opacity 0.4s ease',
  },
  ornamentBottom: {
    fontFamily: 'serif',
    fontSize: '18px',
    color: '#8B6914',
    letterSpacing: '12px',
    opacity: 0.6,
  },
};
