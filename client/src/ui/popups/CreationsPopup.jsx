import React from 'react';
import { useGameStore } from '../../hooks/useGameStore';
import PopupModal from '../PopupModal';

export default function CreationsPopup() {
  const world = useGameStore(s => s.world);
  const events = useGameStore(s => s.events);
  const npcs = (world?.npcs || []).filter(n => n.alive);

  // Unique skills across all NPCs
  const skillMap = {};
  npcs.forEach(n => {
    if (n.skills) {
      Object.entries(n.skills).forEach(([skill, val]) => {
        if (val > 0) {
          if (!skillMap[skill]) skillMap[skill] = { count: 0, totalVal: 0, best: null, bestVal: 0 };
          skillMap[skill].count++;
          skillMap[skill].totalVal += val;
          if (val > skillMap[skill].bestVal) {
            skillMap[skill].bestVal = val;
            skillMap[skill].best = n.name;
          }
        }
      });
    }
  });
  const skillEntries = Object.entries(skillMap).sort(([, a], [, b]) => b.count - a.count);

  // Unique professions
  const profMap = {};
  npcs.forEach(n => {
    const prof = n.profession || 'None';
    if (!profMap[prof]) profMap[prof] = [];
    profMap[prof].push(n.name);
  });
  const profEntries = Object.entries(profMap).sort(([, a], [, b]) => b.length - a.length);

  // Crafting events
  const craftEvents = events.filter(e => {
    const t = e.text?.toLowerCase() || '';
    return /\b(forged|smelted|brewed|crafted|tanned|baked|milled|shaped clay|copied a manuscript|wrote|invented|discovered|created)\b/.test(t);
  });

  // Building events
  const buildEvents = events.filter(e => {
    const t = e.text?.toLowerCase() || '';
    return /\b(began constructing|finished building|built|commission)\b/.test(t);
  });

  // God events
  const godEvents = events.filter(e => e.type === 'god');

  // Teaching events
  const teachEvents = events.filter(e => {
    const t = e.text?.toLowerCase() || '';
    return /\b(taught|lesson|literacy|learned)\b/.test(t);
  });

  return (
    <PopupModal icon="✨" title="Creations & Discoveries">
      <div className="popup-section">
        <div className="popup-grid" style={{ gridTemplateColumns: 'repeat(4, 1fr)' }}>
          <div className="popup-card" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)' }}>Unique Skills</div>
            <div className="popup-stat-value" style={{ fontSize: 18 }}>{skillEntries.length}</div>
          </div>
          <div className="popup-card" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)' }}>Professions</div>
            <div className="popup-stat-value" style={{ fontSize: 18 }}>{profEntries.length}</div>
          </div>
          <div className="popup-card" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)' }}>Items Crafted</div>
            <div className="popup-stat-value" style={{ fontSize: 18, color: 'var(--sdv-orange)' }}>{craftEvents.length}</div>
          </div>
          <div className="popup-card" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)' }}>Divine Acts</div>
            <div className="popup-stat-value" style={{ fontSize: 18, color: 'var(--sdv-gold)' }}>{godEvents.length}</div>
          </div>
        </div>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 14 }}>
        <div className="popup-section">
          <h4>Known Skills ({skillEntries.length})</h4>
          <table className="popup-table">
            <thead>
              <tr><th>Skill</th><th>Known By</th><th>Best</th></tr>
            </thead>
            <tbody>
              {skillEntries.map(([skill, data]) => (
                <tr key={skill}>
                  <td style={{ textTransform: 'capitalize', fontWeight: 600 }}>{skill}</td>
                  <td>{data.count} NPCs</td>
                  <td style={{ color: 'var(--sdv-gold)' }}>{data.best} ({Math.round(data.bestVal)})</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        <div className="popup-section">
          <h4>Active Professions ({profEntries.length})</h4>
          {profEntries.map(([prof, names]) => {
            const shown = names.slice(0, 5);
            const extra = names.length - shown.length;
            return (
              <div key={prof} style={{ marginBottom: 6 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 11 }}>
                  <span style={{ fontWeight: 600, color: 'var(--sdv-text-bright)' }}>{prof}</span>
                  <span style={{ color: 'var(--sdv-text-dim)' }}>{names.length}</span>
                </div>
                <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)' }}>
                  {shown.join(', ')}{extra > 0 ? ` +${extra} more` : ''}
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {buildEvents.length > 0 && (
        <div className="popup-section">
          <h4>Construction History</h4>
          <div style={{ maxHeight: 120, overflowY: 'auto' }}>
            {buildEvents.slice(-15).reverse().map((e, i) => (
              <div key={i} style={{ fontSize: 11, padding: '3px 0', borderBottom: '1px solid var(--sdv-parchment-deep)' }}>
                <span style={{ color: 'var(--sdv-text-dim)', fontSize: 10 }}>[{e.time}]</span> {e.text}
              </div>
            ))}
          </div>
        </div>
      )}

      {teachEvents.length > 0 && (
        <div className="popup-section">
          <h4>Knowledge Transfer</h4>
          <div style={{ maxHeight: 120, overflowY: 'auto' }}>
            {teachEvents.slice(-15).reverse().map((e, i) => (
              <div key={i} style={{ fontSize: 11, padding: '3px 0', borderBottom: '1px solid var(--sdv-parchment-deep)' }}>
                <span style={{ color: 'var(--sdv-text-dim)', fontSize: 10 }}>[{e.time}]</span> {e.text}
              </div>
            ))}
          </div>
        </div>
      )}
    </PopupModal>
  );
}
