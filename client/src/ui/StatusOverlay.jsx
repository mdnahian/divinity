import React from 'react';
import { useGameStore } from '../hooks/useGameStore';
import './StatusOverlay.css';

const WEATHER_ICONS = {
  clear: '☀', sunny: '☀', cloudy: '☁',
  rain: '🌧️', storm: '⛈️',
  snow: '❄️',
};

export default function StatusOverlay() {
  const world = useGameStore(s => s.world);

  const timeString = world?.timeString || '';
  const weather = world?.weather || '';
  const weatherIcon = WEATHER_ICONS[weather] || '';
  const npcs = world?.npcs || [];
  const alive = npcs.filter(n => n.alive).length;
  const total = npcs.length;

  return (
    <div className="status-overlay sdv-frame" onPointerDown={e => e.stopPropagation()}>
      <span className="status-item">
        <span className="status-icon">📅</span>
        <span className="status-text">{timeString}</span>
      </span>
      <span className="status-divider" />
      <span className="status-item">
        <span className="status-icon">{weatherIcon}</span>
        <span className="status-text">{weather}</span>
      </span>
      <span className="status-divider" />
      <span className="status-item">
        <span className="status-icon">👥</span>
        <span className="status-text" style={{ color: alive < total ? 'var(--sdv-red)' : 'var(--sdv-green)' }}>
          {alive}/{total}
        </span>
      </span>
    </div>
  );
}
