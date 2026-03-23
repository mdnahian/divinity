import React, { useRef, useEffect } from 'react';
import { useGameStore } from '../../hooks/useGameStore';
import PopupModal from '../PopupModal';

function WealthChart({ npcs }) {
  const canvasRef = useRef(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas || !npcs.length) return;
    const ctx = canvas.getContext('2d');
    const W = canvas.width;
    const H = canvas.height;
    const padding = { top: 30, right: 20, bottom: 50, left: 50 };

    ctx.clearRect(0, 0, W, H);

    const sorted = [...npcs].sort((a, b) => (a.goldCount || 0) - (b.goldCount || 0));
    const maxGold = Math.max(...sorted.map(n => n.goldCount || 0), 1);
    const barWidth = Math.max(4, Math.min(30, (W - padding.left - padding.right) / sorted.length - 2));
    const chartW = sorted.length * (barWidth + 2);
    const offsetX = padding.left + ((W - padding.left - padding.right - chartW) / 2);

    // Background grid
    ctx.strokeStyle = 'rgba(139, 105, 20, 0.15)';
    ctx.lineWidth = 1;
    for (let i = 0; i <= 4; i++) {
      const y = padding.top + ((H - padding.top - padding.bottom) * i / 4);
      ctx.beginPath();
      ctx.moveTo(padding.left, y);
      ctx.lineTo(W - padding.right, y);
      ctx.stroke();
      ctx.font = '8px "Press Start 2P", monospace';
      ctx.fillStyle = '#8B6914';
      ctx.textAlign = 'right';
      ctx.fillText(Math.round(maxGold * (4 - i) / 4) + 'g', padding.left - 6, y + 3);
    }

    // Bars
    sorted.forEach((npc, i) => {
      const gold = npc.goldCount || 0;
      const barH = (gold / maxGold) * (H - padding.top - padding.bottom);
      const x = offsetX + i * (barWidth + 2);
      const y = H - padding.bottom - barH;

      // Color by wealth percentile
      const pct = i / sorted.length;
      const color = pct > 0.8 ? '#D4A020' : pct > 0.5 ? '#8B6914' : pct > 0.2 ? '#AA6644' : '#CC6644';

      ctx.fillStyle = color;
      ctx.fillRect(x, y, barWidth, barH);
      ctx.strokeStyle = 'rgba(0,0,0,0.2)';
      ctx.strokeRect(x, y, barWidth, barH);

      // Highlight inner
      ctx.fillStyle = 'rgba(255,255,255,0.15)';
      ctx.fillRect(x + 1, y + 1, barWidth / 2 - 1, barH - 2);
    });

    // Lorenz curve overlay
    const totalGold = sorted.reduce((s, n) => s + (n.goldCount || 0), 0);
    if (totalGold > 0) {
      const chartH = H - padding.top - padding.bottom;
      const chartW2 = W - padding.left - padding.right;

      // Perfect equality line
      ctx.beginPath();
      ctx.moveTo(padding.left, H - padding.bottom);
      ctx.lineTo(W - padding.right, padding.top);
      ctx.strokeStyle = 'rgba(74, 154, 48, 0.4)';
      ctx.lineWidth = 1;
      ctx.setLineDash([4, 4]);
      ctx.stroke();
      ctx.setLineDash([]);

      // Lorenz curve
      ctx.beginPath();
      ctx.moveTo(padding.left, H - padding.bottom);
      let cumGold = 0;
      sorted.forEach((npc, i) => {
        cumGold += npc.goldCount || 0;
        const x = padding.left + ((i + 1) / sorted.length) * chartW2;
        const y = H - padding.bottom - (cumGold / totalGold) * chartH;
        ctx.lineTo(x, y);
      });
      ctx.strokeStyle = '#CC4444';
      ctx.lineWidth = 2;
      ctx.stroke();
    }

    // Title
    ctx.font = '8px "Press Start 2P", monospace';
    ctx.fillStyle = '#5D410C';
    ctx.textAlign = 'center';
    ctx.fillText('Wealth Distribution (sorted low to high)', W / 2, 16);

    // Legend
    ctx.font = '7px "Press Start 2P", monospace';
    ctx.fillStyle = 'rgba(74, 154, 48, 0.7)';
    ctx.textAlign = 'left';
    ctx.fillText('--- Perfect Equality', W - 180, 16);
    ctx.fillStyle = '#CC4444';
    ctx.fillText('--- Lorenz Curve', W - 180, 28);

  }, [npcs]);

  return (
    <canvas
      ref={canvasRef}
      width={680}
      height={300}
      style={{ width: '100%', height: 'auto', borderRadius: 4, background: 'rgba(245, 230, 200, 0.5)', border: '1px solid var(--sdv-parchment-deep)' }}
    />
  );
}

export default function WealthPopup() {
  const world = useGameStore(s => s.world);
  const npcs = (world?.npcs || []).filter(n => n.alive);

  const golds = npcs.map(n => n.goldCount || 0).sort((a, b) => a - b);
  const totalGold = golds.reduce((s, v) => s + v, 0);
  const avgGold = npcs.length > 0 ? totalGold / npcs.length : 0;
  const medianGold = golds.length > 0 ? golds[Math.floor(golds.length / 2)] : 0;

  // Gini coefficient calculation
  let gini = 0;
  if (npcs.length > 1 && totalGold > 0) {
    let sumDiff = 0;
    for (let i = 0; i < golds.length; i++) {
      for (let j = 0; j < golds.length; j++) {
        sumDiff += Math.abs(golds[i] - golds[j]);
      }
    }
    gini = sumDiff / (2 * golds.length * totalGold);
  }

  const giniColor = gini < 0.3 ? 'var(--sdv-green)' : gini < 0.5 ? 'var(--sdv-orange)' : 'var(--sdv-red)';
  const giniLabel = gini < 0.2 ? 'Very Equal' : gini < 0.3 ? 'Fairly Equal' : gini < 0.4 ? 'Moderate' : gini < 0.5 ? 'Unequal' : 'Very Unequal';

  // Wealth brackets
  const brackets = { 'Destitute (0)': 0, 'Poor (1-10)': 0, 'Modest (11-30)': 0, 'Comfortable (31-60)': 0, 'Wealthy (61-100)': 0, 'Rich (100+)': 0 };
  npcs.forEach(n => {
    const g = n.goldCount || 0;
    if (g === 0) brackets['Destitute (0)']++;
    else if (g <= 10) brackets['Poor (1-10)']++;
    else if (g <= 30) brackets['Modest (11-30)']++;
    else if (g <= 60) brackets['Comfortable (31-60)']++;
    else if (g <= 100) brackets['Wealthy (61-100)']++;
    else brackets['Rich (100+)']++;
  });
  const bracketColors = {
    'Destitute (0)': '#CC4444',
    'Poor (1-10)': '#CC6644',
    'Modest (11-30)': '#D08A20',
    'Comfortable (31-60)': '#8B6914',
    'Wealthy (61-100)': '#4A9A30',
    'Rich (100+)': '#D4A020',
  };

  // Top 10% vs Bottom 50%
  const top10Count = Math.max(1, Math.ceil(npcs.length * 0.1));
  const bottom50Count = Math.max(1, Math.floor(npcs.length * 0.5));
  const top10Gold = golds.slice(-top10Count).reduce((s, v) => s + v, 0);
  const bottom50Gold = golds.slice(0, bottom50Count).reduce((s, v) => s + v, 0);
  const top10Pct = totalGold > 0 ? ((top10Gold / totalGold) * 100).toFixed(1) : 0;
  const bottom50Pct = totalGold > 0 ? ((bottom50Gold / totalGold) * 100).toFixed(1) : 0;

  return (
    <PopupModal icon="📊" title="Wealth Distribution">
      <div className="popup-section">
        <div className="popup-grid" style={{ gridTemplateColumns: 'repeat(5, 1fr)' }}>
          <div className="popup-card" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)' }}>Total Gold</div>
            <div className="popup-stat-value" style={{ fontSize: 18, color: 'var(--sdv-gold)' }}>{totalGold}</div>
          </div>
          <div className="popup-card" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)' }}>Average</div>
            <div className="popup-stat-value" style={{ fontSize: 18 }}>{avgGold.toFixed(1)}</div>
          </div>
          <div className="popup-card" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)' }}>Median</div>
            <div className="popup-stat-value" style={{ fontSize: 18 }}>{medianGold}</div>
          </div>
          <div className="popup-card" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)' }}>Gini Index</div>
            <div className="popup-stat-value" style={{ fontSize: 18, color: giniColor }}>{gini.toFixed(3)}</div>
            <div style={{ fontSize: 9, color: giniColor }}>{giniLabel}</div>
          </div>
          <div className="popup-card" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)' }}>Top 10% Own</div>
            <div className="popup-stat-value" style={{ fontSize: 18, color: top10Pct > 50 ? 'var(--sdv-red)' : 'var(--sdv-text-bright)' }}>{top10Pct}%</div>
          </div>
        </div>
      </div>

      <div className="popup-section">
        <h4>Wealth Distribution Chart</h4>
        <WealthChart npcs={npcs} />
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 14 }}>
        <div className="popup-section">
          <h4>Wealth Brackets</h4>
          {Object.entries(brackets).map(([label, count]) => (
            <div key={label} className="popup-bar-row">
              <span className="popup-bar-label">{label}</span>
              <div className="popup-bar-track">
                <div className="popup-bar-fill" style={{ width: `${npcs.length > 0 ? (count / npcs.length) * 100 : 0}%`, background: bracketColors[label] }} />
              </div>
              <span className="popup-bar-val">{count}</span>
            </div>
          ))}
        </div>

        <div className="popup-section">
          <h4>Inequality Metrics</h4>
          <div className="popup-card" style={{ marginBottom: 8 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 11 }}>
              <span style={{ color: 'var(--sdv-text-dim)' }}>Top 10% share</span>
              <span style={{ fontWeight: 700, color: top10Pct > 50 ? 'var(--sdv-red)' : 'var(--sdv-text-bright)' }}>{top10Pct}% of all gold</span>
            </div>
          </div>
          <div className="popup-card" style={{ marginBottom: 8 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 11 }}>
              <span style={{ color: 'var(--sdv-text-dim)' }}>Bottom 50% share</span>
              <span style={{ fontWeight: 700, color: bottom50Pct < 15 ? 'var(--sdv-red)' : 'var(--sdv-text-bright)' }}>{bottom50Pct}% of all gold</span>
            </div>
          </div>
          <div className="popup-card" style={{ marginBottom: 8 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 11 }}>
              <span style={{ color: 'var(--sdv-text-dim)' }}>Mean / Median ratio</span>
              <span style={{ fontWeight: 700 }}>{medianGold > 0 ? (avgGold / medianGold).toFixed(2) : 'N/A'}</span>
            </div>
          </div>
          <div className="popup-card">
            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 11 }}>
              <span style={{ color: 'var(--sdv-text-dim)' }}>Max / Min spread</span>
              <span style={{ fontWeight: 700 }}>{golds.length > 0 ? `${golds[golds.length - 1]} - ${golds[0]}` : 'N/A'}</span>
            </div>
          </div>
        </div>
      </div>
    </PopupModal>
  );
}
