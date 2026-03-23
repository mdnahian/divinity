import React, { useRef, useEffect } from 'react';
import { useGameStore } from '../../hooks/useGameStore';
import PopupModal from '../PopupModal';

const TYPE_COLORS = {
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
  dialogue: '#B0A0D0',
  system: '#D08A20',
  world: '#9A6ABF',
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
  if (/\b(began constructing|finished building|commission)\b/.test(t)) return 'build';
  return 'npc';
}

export default function TimelinePopup() {
  const events = useGameStore(s => s.events);
  const selectNpc = useGameStore(s => s.selectNpc);
  const closePopup = useGameStore(s => s.closePopup);
  const endRef = useRef(null);

  // Group events by time/day
  const grouped = {};
  events.forEach(e => {
    const time = e.time || 'Unknown';
    if (!grouped[time]) grouped[time] = [];
    grouped[time].push(e);
  });
  const timeGroups = Object.entries(grouped);

  useEffect(() => {
    if (endRef.current) {
      endRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [events.length]);

  const handleNpcClick = (npcId) => {
    if (npcId) {
      selectNpc(npcId);
      closePopup();
    }
  };

  return (
    <PopupModal icon="📜" title="World Timeline">
      {timeGroups.length === 0 ? (
        <div className="popup-empty">No events recorded yet.</div>
      ) : (
        <div>
          {timeGroups.map(([time, evts], gi) => (
            <div key={gi} style={{ marginBottom: 12 }}>
              <div style={{
                fontFamily: 'var(--sdv-font-pixel)',
                fontSize: 8,
                color: 'var(--sdv-wood)',
                textTransform: 'uppercase',
                letterSpacing: 1,
                padding: '4px 0',
                borderBottom: '2px solid var(--sdv-parchment-deep)',
                marginBottom: 4,
                position: 'sticky',
                top: 0,
                background: 'var(--sdv-parchment)',
                zIndex: 1,
              }}>
                {time}
              </div>
              {evts.map((e, i) => {
                const subType = e.type === 'npc' ? classifyEvent(e.text || '') : e.type;
                const color = TYPE_COLORS[subType] || TYPE_COLORS[e.type] || '#8B6914';
                return (
                  <div
                    key={i}
                    style={{
                      display: 'flex',
                      alignItems: 'flex-start',
                      gap: 8,
                      padding: '3px 6px',
                      borderBottom: '1px solid var(--sdv-parchment-deep)',
                      cursor: e.npcId ? 'pointer' : 'default',
                      fontSize: 11,
                    }}
                    onClick={() => handleNpcClick(e.npcId)}
                  >
                    <span style={{
                      width: 8,
                      height: 8,
                      borderRadius: '50%',
                      background: color,
                      flexShrink: 0,
                      marginTop: 4,
                      border: '1px solid rgba(0,0,0,0.15)',
                    }} />
                    <span style={{ color: 'var(--sdv-text)' }}>{e.text}</span>
                  </div>
                );
              })}
            </div>
          ))}
          <div ref={endRef} />
        </div>
      )}
    </PopupModal>
  );
}
