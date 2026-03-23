import React from 'react';
import { useGameStore } from '../hooks/useGameStore';
import './BottomToolbar.css';

const BUTTONS = [
  { id: 'world', icon: '🌍', label: 'World' },
  { id: 'relationships', icon: '❤️', label: 'Relations' },
  { id: 'factions', icon: '🛡️', label: 'Factions' },
  { id: 'economy', icon: '💰', label: 'Economy' },
  { id: 'demographics', icon: '👥', label: 'Demographics' },
  { id: 'creations', icon: '✨', label: 'Creations' },
  { id: 'actions', icon: '⚔️', label: 'Actions' },
  { id: 'leaderboard', icon: '🏆', label: 'Leaderboard' },
  { id: 'timeline', icon: '📜', label: 'Timeline' },
  { id: 'prayer', icon: '🙏', label: 'Prayer' },
  { id: 'combat', icon: '🗡️', label: 'Combat' },
  { id: 'wealth', icon: '📊', label: 'Wealth' },
  { id: 'mood', icon: '😊', label: 'Mood' },
  { id: 'chronicle', icon: '📖', label: 'Chronicle' },
];

export default function BottomToolbar() {
  const activePopup = useGameStore(s => s.activePopup);
  const togglePopup = useGameStore(s => s.togglePopup);
  const minimapVisible = useGameStore(s => s.minimapVisible);
  const toggleMinimap = useGameStore(s => s.toggleMinimap);
  const myAgentNpcId = useGameStore(s => s.myAgentNpcId);
  const selectNpc = useGameStore(s => s.selectNpc);
  const setFollowNpc = useGameStore(s => s.setFollowNpc);
  const isMobile = useGameStore(s => s.isMobile);
  const toggleChronicle = useGameStore(s => s.toggleChronicle);

  return (
    <div className="bottom-toolbar sdv-frame" onPointerDown={e => e.stopPropagation()}>
      <div className="toolbar-buttons">
        {isMobile && (
          <button
            className="sdv-btn toolbar-btn"
            onClick={toggleChronicle}
            title="Chronicle"
          >
            <span className="toolbar-btn-icon">📜</span>
            <span className="toolbar-btn-label">Chronicle</span>
          </button>
        )}
        {myAgentNpcId && (
          <button
            className="sdv-btn toolbar-btn my-agent-btn"
            onClick={() => { selectNpc(myAgentNpcId); setFollowNpc(myAgentNpcId); }}
            title="Follow My Prophet"
          >
            <span className="toolbar-btn-icon">{'\u{1F4CD}'}</span>
            <span className="toolbar-btn-label">My Prophet</span>
          </button>
        )}
        {BUTTONS.map(btn => (
          <button
            key={btn.id}
            className={`sdv-btn toolbar-btn ${activePopup === btn.id ? 'active' : ''}`}
            onClick={() => togglePopup(btn.id)}
            title={btn.label}
          >
            <span className="toolbar-btn-icon">{btn.icon}</span>
            <span className="toolbar-btn-label">{btn.label}</span>
          </button>
        ))}
      </div>
      <div className="toolbar-divider" />
      <div className="toolbar-utility">
        <button
          className={`sdv-btn toolbar-btn ${minimapVisible ? 'active' : ''}`}
          onClick={toggleMinimap}
          title="Toggle Minimap (M)"
        >
          <span className="toolbar-btn-icon">🗺️</span>
          <span className="toolbar-btn-label">Map</span>
        </button>
      </div>
    </div>
  );
}
