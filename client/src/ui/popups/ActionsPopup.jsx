import React from 'react';
import { useGameStore } from '../../hooks/useGameStore';
import PopupModal from '../PopupModal';

const ACTION_COLORS = {
  combat: '#CC6622',
  death: '#CC4444',
  theft: '#AA3333',
  teaching: '#30AAB0',
  gift: '#4A9A30',
  social: '#B0A0D0',
  trade: '#D0A030',
  craft: '#CC7722',
  gather: '#4A9A40',
  wellbeing: '#3AAA8A',
  survival: '#8A8A7A',
  equip: '#9A9A8A',
  build: '#D4A020',
  god: '#F0D060',
  birth: '#D04070',
  faction: '#3AAA8A',
  world_event: '#9A6ABF',
  npc: '#8B6914',
};

function classifyEvent(text) {
  const t = text.toLowerCase();
  if (/\b(fought|slew|attack|fled from|party.+attack)\b/.test(t)) return 'combat';
  if (/\b(died|has died|perished|killed)\b/.test(t)) return 'death';
  if (/\b(stole|steal|caught .+ stealing)\b/.test(t)) return 'theft';
  if (/\b(taught|lesson|literacy)\b/.test(t)) return 'teaching';
  if (/\b(gave .+ to|shared .+ with|comforted|recruited)\b/.test(t)) return 'gift';
  if (/\b(chatted|flirted|shared a journal|shared ale)\b/.test(t)) return 'social';
  if (/\b(sold|bought|traded|served .+ to|hired|work at)\b/.test(t)) return 'trade';
  if (/\b(forged|smelted|brewed|crafted|tanned|baked|milled|shaped clay|copied a manuscript|wrote)\b/.test(t)) return 'craft';
  if (/\b(foraged|harvested|mined|caught \d+ fish|chopped|gathered|hunted|picked up loot)\b/.test(t)) return 'gather';
  if (/\b(drank a healing|healing potion|prayed|rested|slept|read a)\b/.test(t)) return 'wellbeing';
  if (/\b(ate |drank |bathed)\b/.test(t)) return 'survival';
  if (/\b(equipped|unequipped|repair)\b/.test(t)) return 'equip';
  if (/\b(began constructing|finished building|commission)\b/.test(t)) return 'build';
  return 'npc';
}

export default function ActionsPopup() {
  const events = useGameStore(s => s.events);
  const world = useGameStore(s => s.world);
  const npcs = (world?.npcs || []).filter(n => n.alive);

  // Count events by type
  const typeCounts = {};
  events.forEach(e => {
    const type = e.type === 'npc' ? classifyEvent(e.text || '') : e.type;
    typeCounts[type] = (typeCounts[type] || 0) + 1;
  });
  const typeEntries = Object.entries(typeCounts).sort(([, a], [, b]) => b - a);
  const maxTypeCount = Math.max(...typeEntries.map(([, c]) => c), 1);

  // Most active NPCs (by event count)
  const npcEventCounts = {};
  events.forEach(e => {
    if (e.npcId) {
      npcEventCounts[e.npcId] = (npcEventCounts[e.npcId] || 0) + 1;
    }
  });
  const npcActivityEntries = Object.entries(npcEventCounts)
    .sort(([, a], [, b]) => b - a)
    .slice(0, 10);
  const maxNpcEvents = Math.max(...npcActivityEntries.map(([, c]) => c), 1);

  // Current activities
  const currentActivities = npcs
    .filter(n => n.pendingActionId || n.lastAction)
    .map(n => ({
      name: n.name,
      action: n.pendingActionId ? n.pendingActionId.replace(/_/g, ' ') : n.lastAction,
      isPending: !!n.pendingActionId,
    }));

  return (
    <PopupModal icon="⚔️" title="Actions & Activity">
      <div className="popup-section">
        <div className="popup-grid" style={{ gridTemplateColumns: 'repeat(3, 1fr)' }}>
          <div className="popup-card" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)' }}>Total Events</div>
            <div className="popup-stat-value" style={{ fontSize: 18 }}>{events.length}</div>
          </div>
          <div className="popup-card" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)' }}>Event Types</div>
            <div className="popup-stat-value" style={{ fontSize: 18 }}>{typeEntries.length}</div>
          </div>
          <div className="popup-card" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)' }}>Active Now</div>
            <div className="popup-stat-value" style={{ fontSize: 18, color: 'var(--sdv-green)' }}>{currentActivities.filter(a => a.isPending).length}</div>
          </div>
        </div>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 14 }}>
        <div className="popup-section">
          <h4>Event Breakdown</h4>
          {typeEntries.map(([type, count]) => (
            <div key={type} className="popup-bar-row">
              <span className="popup-bar-label" style={{ textTransform: 'capitalize' }}>{type}</span>
              <div className="popup-bar-track">
                <div className="popup-bar-fill" style={{ width: `${(count / maxTypeCount) * 100}%`, background: ACTION_COLORS[type] || '#8B6914' }} />
              </div>
              <span className="popup-bar-val">{count}</span>
            </div>
          ))}
        </div>

        <div className="popup-section">
          <h4>Most Active NPCs</h4>
          {npcActivityEntries.map(([id, count]) => {
            const npc = npcs.find(n => n.id === id);
            return (
              <div key={id} className="popup-bar-row">
                <span className="popup-bar-label">{npc?.name || 'Unknown'}</span>
                <div className="popup-bar-track">
                  <div className="popup-bar-fill" style={{ width: `${(count / maxNpcEvents) * 100}%`, background: '#4A7ACC' }} />
                </div>
                <span className="popup-bar-val">{count}</span>
              </div>
            );
          })}
        </div>
      </div>

      {currentActivities.length > 0 && (
        <div className="popup-section">
          <h4>Current Activities</h4>
          <table className="popup-table">
            <thead>
              <tr><th>NPC</th><th>Activity</th><th>Status</th></tr>
            </thead>
            <tbody>
              {currentActivities.map((a, i) => (
                <tr key={i}>
                  <td style={{ fontWeight: 600 }}>{a.name}</td>
                  <td style={{ textTransform: 'capitalize' }}>{a.action}</td>
                  <td style={{ color: a.isPending ? 'var(--sdv-green)' : 'var(--sdv-text-dim)' }}>
                    {a.isPending ? 'In Progress' : 'Completed'}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </PopupModal>
  );
}
