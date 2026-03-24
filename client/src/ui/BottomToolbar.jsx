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
  const toolbarOpen = useGameStore(s => s.toolbarOpen);
  const toggleToolbar = useGameStore(s => s.toggleToolbar);
  const closeToolbar = useGameStore(s => s.closeToolbar);

  const allButtons = myAgentNpcId
    ? [{ id: '_myagent', icon: '\u{1F4CD}', label: 'My Prophet' }, ...BUTTONS]
    : BUTTONS;

  const handleButtonClick = (btn) => {
    if (btn.id === '_myagent') {
      selectNpc(myAgentNpcId);
      setFollowNpc(myAgentNpcId);
    } else {
      togglePopup(btn.id);
    }
    if (isMobile) closeToolbar();
  };

  if (isMobile) {
    return (
      <>
        {toolbarOpen && <div className="mobile-toolbar-backdrop" onClick={closeToolbar} />}
        <div className={`mobile-toolbar-panel sdv-frame${toolbarOpen ? ' mobile-toolbar-open' : ''}`} onPointerDown={e => e.stopPropagation()}>
          <div className="mobile-toolbar-grid">
            {allButtons.map(btn => (
              <button
                key={btn.id}
                className={`sdv-btn mobile-toolbar-btn ${activePopup === btn.id ? 'active' : ''}`}
                onClick={() => handleButtonClick(btn)}
              >
                <span className="mobile-toolbar-btn-icon">{btn.icon}</span>
                <span className="mobile-toolbar-btn-label">{btn.label}</span>
              </button>
            ))}
          </div>
        </div>
        <button className="mobile-toolbar-tab" onClick={toggleToolbar}>
          <span className={`mobile-toolbar-arrow${toolbarOpen ? ' open' : ''}`}>‹</span>
        </button>
      </>
    );
  }

  return (
    <div className="bottom-toolbar sdv-frame" onPointerDown={e => e.stopPropagation()}>
      <div className="toolbar-buttons">
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
          className={`sdv-btn toolbar-btn minimap-toggle-btn ${minimapVisible ? 'active' : ''}`}
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
