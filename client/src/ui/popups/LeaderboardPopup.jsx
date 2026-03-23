import React, { useState } from 'react';
import { useGameStore } from '../../hooks/useGameStore';
import PopupModal from '../PopupModal';

const CATEGORIES = [
  { id: 'wealth', label: 'Wealthiest', fn: n => n.goldCount || 0, unit: 'g', color: '#D4A020' },
  { id: 'hp', label: 'Healthiest', fn: n => n.hp || 0, unit: 'HP', color: '#CC4444' },
  { id: 'happiness', label: 'Happiest', fn: n => n.happiness || 0, unit: '', color: '#5CAD4A' },
  { id: 'stress', label: 'Most Stressed', fn: n => n.stress || 0, unit: '', color: '#D08A20' },
  { id: 'skills', label: 'Most Skilled', fn: n => Object.values(n.skills || {}).reduce((s, v) => s + v, 0), unit: 'total', color: '#4A7ACC' },
  { id: 'connections', label: 'Most Connected', fn: n => Object.keys(n.relationships || {}).length, unit: 'bonds', color: '#8A5AB0' },
  { id: 'popularity', label: 'Most Popular', fn: (n, allNpcs) => {
    return allNpcs.reduce((sum, other) => sum + Math.max(0, (other.relationships || {})[n.id] || 0), 0);
  }, unit: 'rep', color: '#5CAD4A' },
  { id: 'hated', label: 'Most Disliked', fn: (n, allNpcs) => {
    return allNpcs.reduce((sum, other) => sum + Math.abs(Math.min(0, (other.relationships || {})[n.id] || 0)), 0);
  }, unit: 'hate', color: '#CC4444' },
  { id: 'strength', label: 'Strongest', fn: n => n.stats?.strength || 0, unit: 'STR', color: '#CC6622' },
  { id: 'intelligence', label: 'Smartest', fn: n => n.stats?.intelligence || 0, unit: 'INT', color: '#4A7ACC' },
  { id: 'charisma', label: 'Most Charismatic', fn: n => n.stats?.charisma || 0, unit: 'CHA', color: '#8A5AB0' },
  { id: 'courage', label: 'Most Courageous', fn: n => n.stats?.courage || 0, unit: 'CRG', color: '#D4A020' },
];

export default function LeaderboardPopup() {
  const world = useGameStore(s => s.world);
  const [category, setCategory] = useState('wealth');
  const npcs = (world?.npcs || []).filter(n => n.alive);

  const cat = CATEGORIES.find(c => c.id === category) || CATEGORIES[0];
  const ranked = [...npcs]
    .map(n => ({ ...n, score: cat.fn(n, npcs) }))
    .sort((a, b) => b.score - a.score)
    .slice(0, 15);
  const maxScore = Math.max(ranked[0]?.score || 1, 1);

  return (
    <PopupModal icon="🏆" title="Leaderboard">
      <div className="popup-section">
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4, marginBottom: 12 }}>
          {CATEGORIES.map(c => (
            <button
              key={c.id}
              className={`sdv-btn ${category === c.id ? 'active' : ''}`}
              style={{ fontSize: 7, padding: '4px 8px' }}
              onClick={() => setCategory(c.id)}
            >
              {c.label}
            </button>
          ))}
        </div>
      </div>

      <div className="popup-section">
        <h4>{cat.label} - Top {ranked.length}</h4>
        {ranked.map((n, i) => (
          <div key={n.id} className="popup-bar-row" style={{ marginBottom: 4 }}>
            <span style={{ width: 24, fontSize: 12, fontWeight: 700, textAlign: 'center', color: i < 3 ? 'var(--sdv-gold)' : 'var(--sdv-text-dim)' }}>
              {i === 0 ? '🥇' : i === 1 ? '🥈' : i === 2 ? '🥉' : `${i + 1}.`}
            </span>
            <span className="popup-bar-label" style={{ width: 100 }}>{n.name}</span>
            <div className="popup-bar-track">
              <div className="popup-bar-fill" style={{ width: `${(n.score / maxScore) * 100}%`, background: cat.color }} />
            </div>
            <span className="popup-bar-val">{Math.round(n.score)} {cat.unit}</span>
          </div>
        ))}
      </div>
    </PopupModal>
  );
}
