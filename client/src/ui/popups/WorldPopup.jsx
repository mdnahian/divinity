import React, { useState } from 'react';
import { useGameStore } from '../../hooks/useGameStore';
import PopupModal from '../PopupModal';

export default function WorldPopup() {
  const world = useGameStore(s => s.world);
  const selectNpc = useGameStore(s => s.selectNpc);
  const selectLocation = useGameStore(s => s.selectLocation);
  const closePopup = useGameStore(s => s.closePopup);
  const [search, setSearch] = useState('');
  const [locSearch, setLocSearch] = useState('');

  const allNpcs = world?.npcs || [];
  const aliveCount = allNpcs.filter(n => n.alive).length;
  const locs = (world?.locations || []).filter(l => l.type !== 'home');

  const sorted = [...allNpcs].sort((a, b) => {
    if (a.alive !== b.alive) return a.alive ? -1 : 1;
    return (a.name || '').localeCompare(b.name || '');
  });

  const filtered = search
    ? sorted.filter(n => n.name?.toLowerCase().includes(search.toLowerCase()))
    : sorted;

  const handleNpcClick = (npcId) => {
    selectNpc(npcId);
    closePopup();
  };

  const handleLocClick = (locId) => {
    selectLocation(locId);
    closePopup();
  };

  return (
    <PopupModal icon="🌍" title="World Directory">
      <div className="popup-section">
        <h4>NPCs ({aliveCount} alive / {allNpcs.length} total)</h4>
        <input
          className="world-popup-search"
          type="text"
          placeholder="Search by name..."
          value={search}
          onChange={e => setSearch(e.target.value)}
          onClick={e => e.stopPropagation()}
        />
        <div className="world-popup-npc-list">
          {filtered.length === 0 && <div className="popup-empty">No NPCs found.</div>}
          {filtered.map(n => (
            <div
              key={n.id}
              className={`world-popup-npc-row ${!n.alive ? 'world-popup-npc-dead' : ''}`}
              onClick={() => handleNpcClick(n.id)}
            >
              <span className="world-popup-dot" style={{ background: n.alive ? '#2ecc71' : '#e74c3c' }} />
              <span className="world-popup-npc-name">{n.name}</span>
              <span className="world-popup-npc-prof">{n.profession}</span>
              {!n.alive && <span className="world-popup-dead-tag">(dead)</span>}
            </div>
          ))}
        </div>
      </div>

      <div className="popup-section">
        <h4>Resources &amp; Locations ({locs.length})</h4>
        <input
          className="world-popup-search"
          type="text"
          placeholder="Search locations or resources..."
          value={locSearch}
          onChange={e => setLocSearch(e.target.value)}
          onClick={e => e.stopPropagation()}
        />
        {locs.length === 0 ? (
          <div className="popup-empty">No locations yet.</div>
        ) : (
          <div className="world-popup-loc-list">
            {locs.filter(loc => {
              if (!locSearch) return true;
              const q = locSearch.toLowerCase();
              if (loc.name?.toLowerCase().includes(q)) return true;
              if (loc.type?.toLowerCase().includes(q)) return true;
              if (loc.resources && Object.keys(loc.resources).some(k => k.toLowerCase().includes(q))) return true;
              return false;
            }).map(loc => {
              const owner = loc.ownerId ? allNpcs.find(n => n.id === loc.ownerId) : null;
              const empCount = allNpcs.filter(n => n.alive && n.workplaceId === loc.id && n.employerId).length;
              return (
                <div
                  key={loc.id}
                  className="res-loc"
                  style={{ borderLeftColor: loc.color || '#8B6914', cursor: 'pointer' }}
                  onClick={() => handleLocClick(loc.id)}
                >
                  <span className="res-loc-name">{loc.name}</span>
                  <span className="res-loc-type">({loc.type})</span>
                  {empCount > 0 && <span className="res-loc-type">({empCount} emp)</span>}
                  {owner && <span className="res-loc-owner"> Owner: {owner.name}</span>}
                  {loc.resources && Object.keys(loc.resources).length > 0 && (
                    <div className="res-items">
                      {Object.entries(loc.resources).map(([key, val]) => {
                        const max = loc.maxResources?.[key] || val;
                        const pct = max > 0 ? val / max : 0;
                        const cls = pct <= 0.2 ? 'res-tag-low' : pct <= 0.5 ? 'res-tag-mid' : 'res-tag-ok';
                        return <span key={key} className={`res-tag ${cls}`}>{key}: {val}/{max}</span>;
                      })}
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        )}
      </div>
    </PopupModal>
  );
}
