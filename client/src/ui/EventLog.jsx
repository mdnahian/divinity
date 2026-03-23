import React, { useEffect, useRef } from 'react';
import { useGameStore } from '../hooks/useGameStore';
import './EventLog.css';

function classifyNpcEvent(text) {
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

export default function EventLog() {
  const events = useGameStore(s => s.events);
  const selectNpc = useGameStore(s => s.selectNpc);
  const isMobile = useGameStore(s => s.isMobile);
  const chronicleOpen = useGameStore(s => s.chronicleOpen);
  const closeChronicle = useGameStore(s => s.closeChronicle);
  const logRef = useRef(null);

  useEffect(() => {
    if (logRef.current) {
      logRef.current.scrollTop = logRef.current.scrollHeight;
    }
  }, [events]);

  const handleClick = (entry) => {
    if (entry.npcId) {
      selectNpc(entry.npcId);
    }
  };

  const mobileOpen = isMobile && chronicleOpen;

  return (
    <>
      {mobileOpen && <div className="chronicle-backdrop" onClick={closeChronicle} />}
      <div className={`left-panel sdv-frame${mobileOpen ? ' mobile-open' : ''}`} onPointerDown={e => e.stopPropagation()} onWheel={e => e.stopPropagation()}>
        <div className="panel-header">
          <h4 className="sdv-section-title chronicle-title"><span style={{ fontFamily: "'Apple Color Emoji', 'Segoe UI Emoji', 'Noto Color Emoji', sans-serif" }}>📜</span> Chronicle</h4>
        </div>
        <div className="event-log sdv-scroll" ref={logRef}>
          {events.map((entry, i) => {
            const subType = entry.type === 'npc' ? classifyNpcEvent(entry.text) : entry.type;
            return (
              <div
                key={i}
                className={`log-entry log-${subType}${entry.npcId ? ' log-clickable' : ''}`}
                onClick={() => handleClick(entry)}
              >
                <span className="log-time">[{entry.time}]</span> {entry.text}
              </div>
            );
          })}
        </div>
      </div>
    </>
  );
}
