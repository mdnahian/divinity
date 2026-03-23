import React from 'react';
import { useGameStore } from '../../hooks/useGameStore';
import PopupModal from '../PopupModal';

export default function EconomyPopup() {
  const world = useGameStore(s => s.world);
  const npcs = (world?.npcs || []).filter(n => n.alive);
  const locations = (world?.locations || []).filter(l => l.type !== 'home');

  // Wealth stats
  const totalGold = npcs.reduce((sum, n) => sum + (n.goldCount || 0), 0);
  const avgGold = npcs.length > 0 ? totalGold / npcs.length : 0;
  const wealthiest = [...npcs].sort((a, b) => (b.goldCount || 0) - (a.goldCount || 0)).slice(0, 5);
  const poorest = [...npcs].sort((a, b) => (a.goldCount || 0) - (b.goldCount || 0)).slice(0, 5);

  // Profession distribution
  const profCounts = {};
  npcs.forEach(n => {
    const prof = n.profession || 'None';
    profCounts[prof] = (profCounts[prof] || 0) + 1;
  });
  const profEntries = Object.entries(profCounts).sort(([, a], [, b]) => b - a);

  // Employment stats
  const employed = npcs.filter(n => n.employerId || n.isBusinessOwner).length;
  const businessOwners = npcs.filter(n => n.isBusinessOwner).length;

  // Resource summary across all locations
  const resourceTotals = {};
  locations.forEach(loc => {
    if (loc.resources) {
      Object.entries(loc.resources).forEach(([key, val]) => {
        if (!resourceTotals[key]) resourceTotals[key] = { current: 0, max: 0 };
        resourceTotals[key].current += val;
        resourceTotals[key].max += (loc.maxResources?.[key] || val);
      });
    }
  });
  const resourceEntries = Object.entries(resourceTotals).sort(([, a], [, b]) => b.current - a.current);

  const maxGold = Math.max(...npcs.map(n => n.goldCount || 0), 1);

  return (
    <PopupModal icon="💰" title="Economy">
      <div className="popup-section">
        <div className="popup-grid" style={{ gridTemplateColumns: 'repeat(4, 1fr)' }}>
          <div className="popup-card" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)' }}>Total Gold</div>
            <div className="popup-stat-value" style={{ fontSize: 18, color: 'var(--sdv-gold)' }}>{totalGold}</div>
          </div>
          <div className="popup-card" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)' }}>Avg Gold</div>
            <div className="popup-stat-value" style={{ fontSize: 18 }}>{avgGold.toFixed(1)}</div>
          </div>
          <div className="popup-card" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)' }}>Employed</div>
            <div className="popup-stat-value" style={{ fontSize: 18, color: 'var(--sdv-green)' }}>{employed}/{npcs.length}</div>
          </div>
          <div className="popup-card" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)' }}>Businesses</div>
            <div className="popup-stat-value" style={{ fontSize: 18 }}>{businessOwners}</div>
          </div>
        </div>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 14 }}>
        <div className="popup-section">
          <h4>Wealthiest NPCs</h4>
          {wealthiest.map(n => (
            <div key={n.id} className="popup-bar-row">
              <span className="popup-bar-label">{n.name}</span>
              <div className="popup-bar-track">
                <div className="popup-bar-fill" style={{ width: `${((n.goldCount || 0) / maxGold) * 100}%`, background: '#D4A020' }} />
              </div>
              <span className="popup-bar-val">{n.goldCount || 0}g</span>
            </div>
          ))}
        </div>

        <div className="popup-section">
          <h4>Poorest NPCs</h4>
          {poorest.map(n => (
            <div key={n.id} className="popup-bar-row">
              <span className="popup-bar-label">{n.name}</span>
              <div className="popup-bar-track">
                <div className="popup-bar-fill" style={{ width: `${((n.goldCount || 0) / maxGold) * 100}%`, background: '#CC4444' }} />
              </div>
              <span className="popup-bar-val">{n.goldCount || 0}g</span>
            </div>
          ))}
        </div>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 14 }}>
        <div className="popup-section">
          <h4>Professions ({profEntries.length})</h4>
          {profEntries.map(([prof, count]) => (
            <div key={prof} className="popup-bar-row">
              <span className="popup-bar-label">{prof}</span>
              <div className="popup-bar-track">
                <div className="popup-bar-fill" style={{ width: `${(count / npcs.length) * 100}%`, background: '#8B6914' }} />
              </div>
              <span className="popup-bar-val">{count}</span>
            </div>
          ))}
        </div>

        <div className="popup-section">
          <h4>World Resources</h4>
          {resourceEntries.length === 0 ? (
            <div className="popup-empty">No tracked resources.</div>
          ) : resourceEntries.map(([name, data]) => {
            const pct = data.max > 0 ? (data.current / data.max) * 100 : 0;
            const color = pct > 60 ? '#4A8A30' : pct > 30 ? '#D08A20' : '#CC4444';
            return (
              <div key={name} className="popup-bar-row">
                <span className="popup-bar-label">{name}</span>
                <div className="popup-bar-track">
                  <div className="popup-bar-fill" style={{ width: `${pct}%`, background: color }} />
                </div>
                <span className="popup-bar-val">{data.current}/{data.max}</span>
              </div>
            );
          })}
        </div>
      </div>
    </PopupModal>
  );
}
