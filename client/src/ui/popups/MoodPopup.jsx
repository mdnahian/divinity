import React from 'react';
import { useGameStore } from '../../hooks/useGameStore';
import PopupModal from '../PopupModal';

const MOOD_COLORS = {
  happy: '#4A9A30',
  joyful: '#5CAD4A',
  content: '#6AB050',
  calm: '#3AAA8A',
  neutral: '#8B6914',
  worried: '#D08A20',
  anxious: '#CC8822',
  sad: '#4A7ACC',
  angry: '#CC4444',
  fearful: '#AA3333',
  depressed: '#6A5A8A',
  stressed: '#CC6622',
  excited: '#D4A020',
  hopeful: '#30AAB0',
  lonely: '#7A4A9F',
  grieving: '#5A4A7A',
  unknown: '#8A8A7A',
};

const MOOD_EMOJI = {
  happy: '😊', joyful: '😄', content: '🙂', calm: '😌',
  neutral: '😐', worried: '😟', anxious: '😰', sad: '😢',
  angry: '😠', fearful: '😨', depressed: '😞', stressed: '😤',
  excited: '🤩', hopeful: '🌟', lonely: '😔', grieving: '💔',
  unknown: '❓',
};

export default function MoodPopup() {
  const world = useGameStore(s => s.world);
  const selectNpc = useGameStore(s => s.selectNpc);
  const closePopup = useGameStore(s => s.closePopup);
  const npcs = (world?.npcs || []).filter(n => n.alive);

  // Mood counts
  const moodCounts = {};
  npcs.forEach(n => {
    const mood = (n.mood || 'unknown').toLowerCase();
    moodCounts[mood] = (moodCounts[mood] || 0) + 1;
  });
  const moodEntries = Object.entries(moodCounts).sort(([, a], [, b]) => b - a);

  // Average happiness and stress
  const avgHappiness = npcs.length > 0 ? npcs.reduce((s, n) => s + (n.happiness || 0), 0) / npcs.length : 0;
  const avgStress = npcs.length > 0 ? npcs.reduce((s, n) => s + (n.stress || 0), 0) / npcs.length : 0;

  // Positive vs negative mood ratio
  const positiveMoods = ['happy', 'joyful', 'content', 'calm', 'excited', 'hopeful'];
  const negativeMoods = ['sad', 'angry', 'fearful', 'depressed', 'stressed', 'worried', 'anxious', 'lonely', 'grieving'];
  const positiveCount = npcs.filter(n => positiveMoods.includes((n.mood || '').toLowerCase())).length;
  const negativeCount = npcs.filter(n => negativeMoods.includes((n.mood || '').toLowerCase())).length;
  const neutralCount = npcs.length - positiveCount - negativeCount;

  // Group NPCs by mood for heatmap
  const moodGroups = {};
  npcs.forEach(n => {
    const mood = (n.mood || 'unknown').toLowerCase();
    if (!moodGroups[mood]) moodGroups[mood] = [];
    moodGroups[mood].push(n);
  });

  // Happiest and saddest NPCs
  const happiest = [...npcs].sort((a, b) => (b.happiness || 0) - (a.happiness || 0)).slice(0, 5);
  const saddest = [...npcs].sort((a, b) => (a.happiness || 0) - (b.happiness || 0)).slice(0, 5);
  const maxHappy = Math.max(...npcs.map(n => n.happiness || 0), 1);
  const mostStressed = [...npcs].sort((a, b) => (b.stress || 0) - (a.stress || 0)).slice(0, 5);
  const maxStress = Math.max(...npcs.map(n => n.stress || 0), 1);

  const handleNpcClick = (npcId) => {
    selectNpc(npcId);
    closePopup();
  };

  return (
    <PopupModal icon="😊" title="Mood Heatmap">
      <div className="popup-section">
        <div className="popup-grid" style={{ gridTemplateColumns: 'repeat(4, 1fr)' }}>
          <div className="popup-card" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)' }}>Avg Happiness</div>
            <div className="popup-stat-value" style={{ fontSize: 18, color: avgHappiness > 50 ? 'var(--sdv-green)' : 'var(--sdv-red)' }}>{avgHappiness.toFixed(1)}</div>
          </div>
          <div className="popup-card" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)' }}>Avg Stress</div>
            <div className="popup-stat-value" style={{ fontSize: 18, color: avgStress > 50 ? 'var(--sdv-red)' : 'var(--sdv-green)' }}>{avgStress.toFixed(1)}</div>
          </div>
          <div className="popup-card" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)' }}>Positive</div>
            <div className="popup-stat-value" style={{ fontSize: 18, color: 'var(--sdv-green)' }}>{positiveCount}</div>
          </div>
          <div className="popup-card" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)' }}>Negative</div>
            <div className="popup-stat-value" style={{ fontSize: 18, color: 'var(--sdv-red)' }}>{negativeCount}</div>
          </div>
        </div>
      </div>

      <div className="popup-section">
        <h4>Mood Heatmap</h4>
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 3 }}>
          {npcs.map(n => {
            const mood = (n.mood || 'unknown').toLowerCase();
            const color = MOOD_COLORS[mood] || MOOD_COLORS.unknown;
            const emoji = MOOD_EMOJI[mood] || MOOD_EMOJI.unknown;
            return (
              <div
                key={n.id}
                title={`${n.name}: ${mood} (H:${n.happiness || 0} S:${n.stress || 0})`}
                onClick={() => handleNpcClick(n.id)}
                style={{
                  width: 36,
                  height: 36,
                  borderRadius: 4,
                  background: color,
                  display: 'flex',
                  flexDirection: 'column',
                  alignItems: 'center',
                  justifyContent: 'center',
                  cursor: 'pointer',
                  border: '1px solid rgba(0,0,0,0.15)',
                  fontSize: 14,
                  position: 'relative',
                  transition: 'transform 0.1s',
                }}
                onMouseEnter={(e) => e.currentTarget.style.transform = 'scale(1.15)'}
                onMouseLeave={(e) => e.currentTarget.style.transform = 'scale(1)'}
              >
                <span>{emoji}</span>
                <span style={{ fontSize: 5, color: '#fff', fontFamily: 'var(--sdv-font-pixel)', textShadow: '0 1px 0 rgba(0,0,0,0.5)', marginTop: 1 }}>
                  {n.name.split(' ')[0].slice(0, 5)}
                </span>
              </div>
            );
          })}
        </div>
      </div>

      <div className="popup-section">
        <h4>Mood Distribution</h4>
        {moodEntries.map(([mood, count]) => (
          <div key={mood} className="popup-bar-row">
            <span className="popup-bar-label" style={{ textTransform: 'capitalize' }}>
              {MOOD_EMOJI[mood] || ''} {mood}
            </span>
            <div className="popup-bar-track">
              <div className="popup-bar-fill" style={{ width: `${(count / npcs.length) * 100}%`, background: MOOD_COLORS[mood] || MOOD_COLORS.unknown }} />
            </div>
            <span className="popup-bar-val">{count}</span>
          </div>
        ))}
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 14 }}>
        <div className="popup-section">
          <h4>Happiest NPCs</h4>
          {happiest.map(n => (
            <div key={n.id} className="popup-bar-row" style={{ cursor: 'pointer' }} onClick={() => handleNpcClick(n.id)}>
              <span className="popup-bar-label">{n.name}</span>
              <div className="popup-bar-track">
                <div className="popup-bar-fill" style={{ width: `${((n.happiness || 0) / maxHappy) * 100}%`, background: '#4A9A30' }} />
              </div>
              <span className="popup-bar-val">{n.happiness || 0}</span>
            </div>
          ))}
        </div>

        <div className="popup-section">
          <h4>Most Stressed NPCs</h4>
          {mostStressed.map(n => (
            <div key={n.id} className="popup-bar-row" style={{ cursor: 'pointer' }} onClick={() => handleNpcClick(n.id)}>
              <span className="popup-bar-label">{n.name}</span>
              <div className="popup-bar-track">
                <div className="popup-bar-fill" style={{ width: `${((n.stress || 0) / maxStress) * 100}%`, background: '#CC6622' }} />
              </div>
              <span className="popup-bar-val">{n.stress || 0}</span>
            </div>
          ))}
        </div>
      </div>

      {/* Sentiment pie summary */}
      <div className="popup-section">
        <h4>Overall Sentiment</h4>
        <div style={{ display: 'flex', gap: 8 }}>
          <div className="popup-card" style={{ flex: positiveCount, background: 'rgba(74, 154, 48, 0.3)', textAlign: 'center', minWidth: 30 }}>
            <div style={{ fontSize: 14 }}>😊</div>
            <div style={{ fontSize: 10, fontWeight: 700, color: '#4A9A30' }}>{positiveCount}</div>
          </div>
          {neutralCount > 0 && (
            <div className="popup-card" style={{ flex: neutralCount, background: 'rgba(139, 105, 20, 0.2)', textAlign: 'center', minWidth: 30 }}>
              <div style={{ fontSize: 14 }}>😐</div>
              <div style={{ fontSize: 10, fontWeight: 700, color: '#8B6914' }}>{neutralCount}</div>
            </div>
          )}
          <div className="popup-card" style={{ flex: negativeCount || 1, background: 'rgba(204, 68, 68, 0.2)', textAlign: 'center', minWidth: 30 }}>
            <div style={{ fontSize: 14 }}>😟</div>
            <div style={{ fontSize: 10, fontWeight: 700, color: '#CC4444' }}>{negativeCount}</div>
          </div>
        </div>
      </div>
    </PopupModal>
  );
}
