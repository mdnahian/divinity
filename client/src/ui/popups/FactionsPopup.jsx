import React from 'react';
import { useGameStore } from '../../hooks/useGameStore';
import PopupModal from '../PopupModal';

export default function FactionsPopup() {
  const world = useGameStore(s => s.world);
  const factions = world?.factions || [];
  const npcs = world?.npcs || [];

  if (factions.length === 0) {
    return (
      <PopupModal icon="🛡️" title="Factions">
        <div className="popup-empty">No factions have formed yet.</div>
      </PopupModal>
    );
  }

  return (
    <PopupModal icon="🛡️" title="Factions">
      <div className="popup-grid" style={{ gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))' }}>
        {factions.map(faction => {
          const leader = npcs.find(n => n.id === faction.leaderId);
          const members = faction.memberIds.map(id => npcs.find(n => n.id === id)).filter(Boolean);
          const aliveMembers = members.filter(m => m.alive);

          return (
            <div key={faction.id} className="popup-card">
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 6 }}>
                <span style={{ fontFamily: 'var(--sdv-font-pixel)', fontSize: 9, color: 'var(--sdv-text-bright)' }}>
                  {faction.name}
                </span>
                <span className="popup-tag">{faction.type}</span>
              </div>

              {leader && (
                <div style={{ fontSize: 11, marginBottom: 4 }}>
                  <span style={{ color: 'var(--sdv-text-dim)' }}>Leader:</span>{' '}
                  <span style={{ fontWeight: 600, color: 'var(--sdv-gold)' }}>{leader.name}</span>
                </div>
              )}

              <div style={{ fontSize: 11, marginBottom: 6 }}>
                <span style={{ color: 'var(--sdv-text-dim)' }}>Members:</span>{' '}
                <span className="popup-stat-value">{aliveMembers.length}</span>
                {aliveMembers.length < members.length && (
                  <span style={{ color: 'var(--sdv-red)', fontSize: 10 }}> ({members.length - aliveMembers.length} dead)</span>
                )}
              </div>

              <div style={{ display: 'flex', flexWrap: 'wrap', gap: 3 }}>
                {aliveMembers.map(m => (
                  <span key={m.id} className="popup-tag" style={{
                    fontSize: 9,
                    background: m.id === faction.leaderId ? 'var(--sdv-gold-bright)' : undefined,
                    color: m.id === faction.leaderId ? 'var(--sdv-text-bright)' : undefined,
                  }}>
                    {m.name} ({m.profession})
                  </span>
                ))}
              </div>

              {faction.contracts && faction.contracts.length > 0 && (
                <div style={{ marginTop: 6 }}>
                  <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)', marginBottom: 2 }}>Contracts:</div>
                  {faction.contracts.map((c, i) => (
                    <div key={i} style={{ fontSize: 10, padding: '2px 0', color: 'var(--sdv-text)' }}>
                      {c.description || c.type || JSON.stringify(c)}
                    </div>
                  ))}
                </div>
              )}
            </div>
          );
        })}
      </div>
    </PopupModal>
  );
}
