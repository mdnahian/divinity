import React from 'react';
import { useGameStore } from '../../hooks/useGameStore';
import PopupModal from '../PopupModal';

export default function CombatPopup() {
  const events = useGameStore(s => s.events);
  const world = useGameStore(s => s.world);
  const npcs = world?.npcs || [];
  const alive = npcs.filter(n => n.alive);

  // Combat events
  const combatEvents = events.filter(e => {
    const t = (e.text || '').toLowerCase();
    return /\b(fought|slew|attack|fled from|party.+attack|battle|duel|struck|wounded)\b/.test(t);
  });

  // Death events
  const deathEvents = events.filter(e => {
    const t = (e.text || '').toLowerCase();
    return /\b(died|has died|perished|killed|slain)\b/.test(t);
  });

  // Combat participation by NPC
  const combatCount = {};
  combatEvents.forEach(e => {
    if (e.npcId) {
      combatCount[e.npcId] = (combatCount[e.npcId] || 0) + 1;
    }
  });
  const warriors = Object.entries(combatCount)
    .map(([id, count]) => {
      const npc = npcs.find(n => n.id === id);
      return { id, name: npc?.name || 'Unknown', count, alive: npc?.alive ?? false, hp: npc?.hp || 0 };
    })
    .sort((a, b) => b.count - a.count)
    .slice(0, 10);
  const maxCombat = Math.max(...warriors.map(w => w.count), 1);

  // Kill tracking (NPCs mentioned in death events as killers)
  const kills = {};
  deathEvents.forEach(e => {
    const t = e.text || '';
    // Try to extract killer name patterns like "killed by X", "slain by X"
    const match = t.match(/(?:killed by|slain by|slew|murdered by)\s+([A-Z][a-z]+ [A-Z][a-z]+)/i);
    if (match) {
      const killerName = match[1];
      kills[killerName] = (kills[killerName] || 0) + 1;
    }
  });
  const killEntries = Object.entries(kills).sort(([, a], [, b]) => b - a).slice(0, 8);
  const maxKills = Math.max(...killEntries.map(([, c]) => c), 1);

  // HP distribution of alive NPCs
  const hpBuckets = { 'Critical (0-20)': 0, 'Wounded (21-50)': 0, 'Healthy (51-80)': 0, 'Full (81-100)': 0 };
  alive.forEach(n => {
    const hp = n.hp || 0;
    if (hp <= 20) hpBuckets['Critical (0-20)']++;
    else if (hp <= 50) hpBuckets['Wounded (21-50)']++;
    else if (hp <= 80) hpBuckets['Healthy (51-80)']++;
    else hpBuckets['Full (81-100)']++;
  });
  const hpColors = {
    'Critical (0-20)': '#CC4444',
    'Wounded (21-50)': '#D08A20',
    'Healthy (51-80)': '#4A9A30',
    'Full (81-100)': '#2A8A6A',
  };

  // Combat event types breakdown
  const combatTypes = { 'Attacks': 0, 'Fled': 0, 'Kills': 0, 'Party Raids': 0 };
  combatEvents.forEach(e => {
    const t = (e.text || '').toLowerCase();
    if (/fled/.test(t)) combatTypes['Fled']++;
    else if (/slew|killed/.test(t)) combatTypes['Kills']++;
    else if (/party/.test(t)) combatTypes['Party Raids']++;
    else combatTypes['Attacks']++;
  });

  // Average combat stats
  const avgStrength = alive.length > 0 ? (alive.reduce((s, n) => s + (n.stats?.strength || 0), 0) / alive.length).toFixed(1) : 0;
  const avgCourage = alive.length > 0 ? (alive.reduce((s, n) => s + (n.stats?.courage || 0), 0) / alive.length).toFixed(1) : 0;

  return (
    <PopupModal icon="🗡️" title="Combat & Warfare">
      <div className="popup-section">
        <div className="popup-grid" style={{ gridTemplateColumns: 'repeat(5, 1fr)' }}>
          <div className="popup-card" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)' }}>Battles</div>
            <div className="popup-stat-value" style={{ fontSize: 18, color: '#CC6622' }}>{combatEvents.length}</div>
          </div>
          <div className="popup-card" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)' }}>Deaths</div>
            <div className="popup-stat-value" style={{ fontSize: 18, color: '#CC4444' }}>{deathEvents.length}</div>
          </div>
          <div className="popup-card" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)' }}>Avg STR</div>
            <div className="popup-stat-value" style={{ fontSize: 18 }}>{avgStrength}</div>
          </div>
          <div className="popup-card" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)' }}>Avg CRG</div>
            <div className="popup-stat-value" style={{ fontSize: 18 }}>{avgCourage}</div>
          </div>
          <div className="popup-card" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)' }}>Casualties</div>
            <div className="popup-stat-value" style={{ fontSize: 18, color: '#CC4444' }}>{npcs.length - alive.length}</div>
          </div>
        </div>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 14 }}>
        <div className="popup-section">
          <h4>Top Warriors</h4>
          {warriors.length === 0 ? (
            <div className="popup-empty">No combat recorded yet.</div>
          ) : warriors.map(w => (
            <div key={w.id} className="popup-bar-row">
              <span className="popup-bar-label" style={{ color: w.alive ? 'var(--sdv-text)' : 'var(--sdv-red)' }}>
                {w.name} {!w.alive ? '(dead)' : ''}
              </span>
              <div className="popup-bar-track">
                <div className="popup-bar-fill" style={{ width: `${(w.count / maxCombat) * 100}%`, background: '#CC6622' }} />
              </div>
              <span className="popup-bar-val">{w.count}</span>
            </div>
          ))}
        </div>

        <div className="popup-section">
          <h4>HP Distribution</h4>
          {Object.entries(hpBuckets).map(([label, count]) => (
            <div key={label} className="popup-bar-row">
              <span className="popup-bar-label">{label}</span>
              <div className="popup-bar-track">
                <div className="popup-bar-fill" style={{ width: `${alive.length > 0 ? (count / alive.length) * 100 : 0}%`, background: hpColors[label] }} />
              </div>
              <span className="popup-bar-val">{count}</span>
            </div>
          ))}
        </div>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 14 }}>
        {killEntries.length > 0 && (
          <div className="popup-section">
            <h4>Most Dangerous</h4>
            {killEntries.map(([name, count]) => (
              <div key={name} className="popup-bar-row">
                <span className="popup-bar-label">{name}</span>
                <div className="popup-bar-track">
                  <div className="popup-bar-fill" style={{ width: `${(count / maxKills) * 100}%`, background: '#CC4444' }} />
                </div>
                <span className="popup-bar-val">{count} kills</span>
              </div>
            ))}
          </div>
        )}

        <div className="popup-section">
          <h4>Combat Breakdown</h4>
          {Object.entries(combatTypes).filter(([, c]) => c > 0).map(([type, count]) => (
            <div key={type} className="popup-bar-row">
              <span className="popup-bar-label">{type}</span>
              <div className="popup-bar-track">
                <div className="popup-bar-fill" style={{ width: `${(count / Math.max(combatEvents.length, 1)) * 100}%`, background: '#8B6914' }} />
              </div>
              <span className="popup-bar-val">{count}</span>
            </div>
          ))}
        </div>
      </div>

      {combatEvents.length > 0 && (
        <div className="popup-section">
          <h4>Recent Combat</h4>
          <div style={{ maxHeight: 150, overflowY: 'auto' }}>
            {combatEvents.slice(-15).reverse().map((e, i) => (
              <div key={i} style={{ fontSize: 11, padding: '4px 6px', borderBottom: '1px solid var(--sdv-parchment-deep)', color: '#CC6622' }}>
                <span style={{ color: 'var(--sdv-text-dim)', fontSize: 10 }}>[{e.time}]</span> {e.text}
              </div>
            ))}
          </div>
        </div>
      )}
    </PopupModal>
  );
}
