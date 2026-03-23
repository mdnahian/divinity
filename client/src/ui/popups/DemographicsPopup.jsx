import React from 'react';
import { useGameStore } from '../../hooks/useGameStore';
import PopupModal from '../PopupModal';

function bucketize(values, bucketSize, labelFn) {
  const buckets = {};
  values.forEach(v => {
    const key = Math.floor(v / bucketSize) * bucketSize;
    const label = labelFn ? labelFn(key) : `${key}-${key + bucketSize - 1}`;
    buckets[label] = (buckets[label] || 0) + 1;
  });
  return Object.entries(buckets).sort(([a], [b]) => parseInt(a) - parseInt(b));
}

export default function DemographicsPopup() {
  const world = useGameStore(s => s.world);
  const allNpcs = world?.npcs || [];
  const alive = allNpcs.filter(n => n.alive);
  const dead = allNpcs.filter(n => !n.alive);

  // Age distribution
  const ageBuckets = bucketize(alive.map(n => n.age || 0), 10, k => `${k}-${k + 9}`);
  const maxAgeBucket = Math.max(...ageBuckets.map(([, c]) => c), 1);

  // Life stage distribution
  const lifeStageCounts = {};
  alive.forEach(n => {
    const stage = n.lifeStage || 'unknown';
    lifeStageCounts[stage] = (lifeStageCounts[stage] || 0) + 1;
  });
  const lifeStageEntries = Object.entries(lifeStageCounts).sort(([, a], [, b]) => b - a);

  // Mood distribution
  const moodCounts = {};
  alive.forEach(n => {
    const mood = n.mood || 'unknown';
    moodCounts[mood] = (moodCounts[mood] || 0) + 1;
  });
  const moodEntries = Object.entries(moodCounts).sort(([, a], [, b]) => b - a);

  // Literacy distribution
  const litCounts = {};
  alive.forEach(n => {
    const lit = n.literacyLevel || 'unknown';
    litCounts[lit] = (litCounts[lit] || 0) + 1;
  });
  const litEntries = Object.entries(litCounts).sort(([, a], [, b]) => b - a);

  // Avg stats
  const avgHappiness = alive.length > 0 ? (alive.reduce((s, n) => s + (n.happiness || 0), 0) / alive.length).toFixed(1) : 0;
  const avgStress = alive.length > 0 ? (alive.reduce((s, n) => s + (n.stress || 0), 0) / alive.length).toFixed(1) : 0;
  const avgHp = alive.length > 0 ? (alive.reduce((s, n) => s + (n.hp || 0), 0) / alive.length).toFixed(1) : 0;

  // Causes of death
  const deathCauses = {};
  dead.forEach(n => {
    const cause = n.causeOfDeath || 'Unknown';
    deathCauses[cause] = (deathCauses[cause] || 0) + 1;
  });
  const deathEntries = Object.entries(deathCauses).sort(([, a], [, b]) => b - a);

  return (
    <PopupModal icon="👥" title="Demographics">
      <div className="popup-section">
        <div className="popup-grid" style={{ gridTemplateColumns: 'repeat(5, 1fr)' }}>
          <div className="popup-card" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)' }}>Alive</div>
            <div className="popup-stat-value" style={{ fontSize: 18, color: 'var(--sdv-green)' }}>{alive.length}</div>
          </div>
          <div className="popup-card" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)' }}>Dead</div>
            <div className="popup-stat-value" style={{ fontSize: 18, color: 'var(--sdv-red)' }}>{dead.length}</div>
          </div>
          <div className="popup-card" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)' }}>Avg HP</div>
            <div className="popup-stat-value" style={{ fontSize: 18 }}>{avgHp}</div>
          </div>
          <div className="popup-card" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)' }}>Avg Happy</div>
            <div className="popup-stat-value" style={{ fontSize: 18, color: 'var(--sdv-green)' }}>{avgHappiness}</div>
          </div>
          <div className="popup-card" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)' }}>Avg Stress</div>
            <div className="popup-stat-value" style={{ fontSize: 18, color: 'var(--sdv-orange)' }}>{avgStress}</div>
          </div>
        </div>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 14 }}>
        <div className="popup-section">
          <h4>Age Distribution</h4>
          {ageBuckets.map(([label, count]) => (
            <div key={label} className="popup-bar-row">
              <span className="popup-bar-label">{label}</span>
              <div className="popup-bar-track">
                <div className="popup-bar-fill" style={{ width: `${(count / maxAgeBucket) * 100}%`, background: '#4A7ACC' }} />
              </div>
              <span className="popup-bar-val">{count}</span>
            </div>
          ))}
        </div>

        <div className="popup-section">
          <h4>Life Stages</h4>
          {lifeStageEntries.map(([stage, count]) => (
            <div key={stage} className="popup-bar-row">
              <span className="popup-bar-label">{stage}</span>
              <div className="popup-bar-track">
                <div className="popup-bar-fill" style={{ width: `${(count / alive.length) * 100}%`, background: '#8A5AB0' }} />
              </div>
              <span className="popup-bar-val">{count}</span>
            </div>
          ))}
        </div>

        <div className="popup-section">
          <h4>Mood Distribution</h4>
          {moodEntries.map(([mood, count]) => (
            <div key={mood} className="popup-bar-row">
              <span className="popup-bar-label">{mood}</span>
              <div className="popup-bar-track">
                <div className="popup-bar-fill" style={{ width: `${(count / alive.length) * 100}%`, background: '#5CAD4A' }} />
              </div>
              <span className="popup-bar-val">{count}</span>
            </div>
          ))}
        </div>

        <div className="popup-section">
          <h4>Literacy Levels</h4>
          {litEntries.map(([lit, count]) => (
            <div key={lit} className="popup-bar-row">
              <span className="popup-bar-label">{lit}</span>
              <div className="popup-bar-track">
                <div className="popup-bar-fill" style={{ width: `${(count / alive.length) * 100}%`, background: '#D4A020' }} />
              </div>
              <span className="popup-bar-val">{count}</span>
            </div>
          ))}
        </div>
      </div>

      {dead.length > 0 && (
        <div className="popup-section">
          <h4>Causes of Death</h4>
          {deathEntries.map(([cause, count]) => (
            <div key={cause} className="popup-bar-row">
              <span className="popup-bar-label">{cause}</span>
              <div className="popup-bar-track">
                <div className="popup-bar-fill" style={{ width: `${(count / dead.length) * 100}%`, background: '#CC4444' }} />
              </div>
              <span className="popup-bar-val">{count}</span>
            </div>
          ))}
        </div>
      )}
    </PopupModal>
  );
}
