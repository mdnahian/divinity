import React from 'react';
import { useGameStore } from '../../hooks/useGameStore';
import PopupModal from '../PopupModal';

export default function PrayerPopup() {
  const events = useGameStore(s => s.events);
  const world = useGameStore(s => s.world);
  const npcs = (world?.npcs || []).filter(n => n.alive);

  // Prayer events
  const prayerEvents = events.filter(e => {
    const t = (e.text || '').toLowerCase();
    return /\b(prayed|prayer|pray|divine|blessing|miracle|worship|offering)\b/.test(t);
  });

  // God events (divine interventions)
  const godEvents = events.filter(e => e.type === 'god');

  // Count prayers by NPC
  const prayersByNpc = {};
  prayerEvents.forEach(e => {
    if (e.npcId) {
      const npc = npcs.find(n => n.id === e.npcId);
      const name = npc?.name || 'Unknown';
      prayersByNpc[name] = (prayersByNpc[name] || 0) + 1;
    }
  });
  const prayerRanking = Object.entries(prayersByNpc).sort(([, a], [, b]) => b - a);
  const maxPrayers = Math.max(...prayerRanking.map(([, c]) => c), 1);

  // Extract prayer themes from text
  const themes = {};
  prayerEvents.forEach(e => {
    const t = (e.text || '').toLowerCase();
    if (/health|heal|cure|sick/.test(t)) themes['Health & Healing'] = (themes['Health & Healing'] || 0) + 1;
    else if (/wealth|gold|prosper|fortune/.test(t)) themes['Wealth & Fortune'] = (themes['Wealth & Fortune'] || 0) + 1;
    else if (/protect|safe|guard|danger/.test(t)) themes['Protection'] = (themes['Protection'] || 0) + 1;
    else if (/love|friend|companion/.test(t)) themes['Love & Friendship'] = (themes['Love & Friendship'] || 0) + 1;
    else if (/strength|power|courage/.test(t)) themes['Strength & Power'] = (themes['Strength & Power'] || 0) + 1;
    else if (/wisdom|knowledge|learn/.test(t)) themes['Wisdom'] = (themes['Wisdom'] || 0) + 1;
    else themes['General'] = (themes['General'] || 0) + 1;
  });
  const themeEntries = Object.entries(themes).sort(([, a], [, b]) => b - a);
  const maxTheme = Math.max(...themeEntries.map(([, c]) => c), 1);

  const themeColors = {
    'Health & Healing': '#4A9A30',
    'Wealth & Fortune': '#D4A020',
    'Protection': '#4A7ACC',
    'Love & Friendship': '#CC4477',
    'Strength & Power': '#CC6622',
    'Wisdom': '#8A5AB0',
    'General': '#8B6914',
  };

  return (
    <PopupModal icon="🙏" title="Prayer & Divine">
      <div className="popup-section">
        <div className="popup-grid" style={{ gridTemplateColumns: 'repeat(3, 1fr)' }}>
          <div className="popup-card" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)' }}>Total Prayers</div>
            <div className="popup-stat-value" style={{ fontSize: 18, color: 'var(--sdv-gold)' }}>{prayerEvents.length}</div>
          </div>
          <div className="popup-card" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)' }}>Divine Acts</div>
            <div className="popup-stat-value" style={{ fontSize: 18, color: '#F0D060' }}>{godEvents.length}</div>
          </div>
          <div className="popup-card" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)' }}>Devout NPCs</div>
            <div className="popup-stat-value" style={{ fontSize: 18 }}>{prayerRanking.length}</div>
          </div>
        </div>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 14 }}>
        <div className="popup-section">
          <h4>Most Devout NPCs</h4>
          {prayerRanking.length === 0 ? (
            <div className="popup-empty">No prayers recorded yet.</div>
          ) : prayerRanking.slice(0, 10).map(([name, count]) => (
            <div key={name} className="popup-bar-row">
              <span className="popup-bar-label">{name}</span>
              <div className="popup-bar-track">
                <div className="popup-bar-fill" style={{ width: `${(count / maxPrayers) * 100}%`, background: '#D4A020' }} />
              </div>
              <span className="popup-bar-val">{count}</span>
            </div>
          ))}
        </div>

        <div className="popup-section">
          <h4>Prayer Themes</h4>
          {themeEntries.length === 0 ? (
            <div className="popup-empty">No themes identified.</div>
          ) : themeEntries.map(([theme, count]) => (
            <div key={theme} className="popup-bar-row">
              <span className="popup-bar-label">{theme}</span>
              <div className="popup-bar-track">
                <div className="popup-bar-fill" style={{ width: `${(count / maxTheme) * 100}%`, background: themeColors[theme] || '#8B6914' }} />
              </div>
              <span className="popup-bar-val">{count}</span>
            </div>
          ))}
        </div>
      </div>

      {godEvents.length > 0 && (
        <div className="popup-section">
          <h4>Recent Divine Interventions</h4>
          <div style={{ maxHeight: 180, overflowY: 'auto' }}>
            {godEvents.slice(-20).reverse().map((e, i) => (
              <div key={i} style={{ fontSize: 11, padding: '4px 6px', borderBottom: '1px solid var(--sdv-parchment-deep)', color: '#B88A10' }}>
                <span style={{ color: 'var(--sdv-text-dim)', fontSize: 10 }}>[{e.time}]</span> {e.text}
              </div>
            ))}
          </div>
        </div>
      )}

      {prayerEvents.length > 0 && (
        <div className="popup-section">
          <h4>Recent Prayers</h4>
          <div style={{ maxHeight: 150, overflowY: 'auto' }}>
            {prayerEvents.slice(-15).reverse().map((e, i) => (
              <div key={i} style={{ fontSize: 11, padding: '4px 6px', borderBottom: '1px solid var(--sdv-parchment-deep)' }}>
                <span style={{ color: 'var(--sdv-text-dim)', fontSize: 10 }}>[{e.time}]</span> {e.text}
              </div>
            ))}
          </div>
        </div>
      )}
    </PopupModal>
  );
}
