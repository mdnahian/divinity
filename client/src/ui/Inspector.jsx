import React, { useEffect, useRef, useState } from 'react';
import { useGameStore } from '../hooks/useGameStore';
import { generateNpcPreviewCanvas, generateBuildingPreviewCanvas } from '../game/systems/spriteGenerator';
import './Inspector.css';

function moodClass(mood) {
  if (['desperate', 'miserable', 'furious'].includes(mood)) return 'danger';
  if (['anxious', 'exhausted', 'starving', 'parched', 'lonely', 'melancholy', 'depressed'].includes(mood)) return 'warn';
  if (['content', 'cheerful'].includes(mood)) return 'good';
  return 'neutral';
}

function lerpColor(pct) {
  if (pct < 33) return '#e74c3c';
  if (pct < 66) return '#e67e22';
  return '#2ecc71';
}

function Bar({ label, val, max, urgent = false, invert = false }) {
  const pct = Math.round((val / max) * 100);
  const color = urgent ? '#e74c3c' : (invert ? lerpColor(pct) : lerpColor(100 - pct));
  return (
    <div className="bar-row">
      <span className="bar-label">{label}</span>
      <div className="bar-track">
        <div className="bar-fill" style={{ width: `${pct}%`, background: color }} />
      </div>
      <span className="bar-val">{val}</span>
    </div>
  );
}

function MiniBar({ val }) {
  const pct = Math.round(val);
  const color = pct > 66 ? '#2ecc71' : pct > 33 ? '#e67e22' : '#e74c3c';
  return (
    <div className="mini-bar-track">
      <div className="mini-bar-fill" style={{ width: `${pct}%`, background: color }} />
    </div>
  );
}

function Stat({ label, val }) {
  return (
    <div className="stat-cell">
      <span className="stat-label">{label}</span>
      <span className="stat-val">{Math.round(val)}</span>
    </div>
  );
}

function resolveLocName(locId, world) {
  if (!locId) return 'None';
  const loc = (world?.locations || []).find(l => l.id === locId);
  return loc ? loc.name : locId;
}

function NpcSpritePreview({ npcIndex, profession }) {
  const canvasRef = useRef(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    ctx.imageSmoothingEnabled = false;

    const sheet = generateNpcPreviewCanvas(npcIndex, profession);
    // Draw front-facing idle frame (dir=0, frame=0)
    ctx.drawImage(sheet, 0, 0, 32, 32, 0, 0, canvas.width, canvas.height);
  }, [npcIndex, profession]);

  return (
    <canvas ref={canvasRef} width={64} height={64} className="npc-preview-canvas" />
  );
}

function NpcInspector({ npc, world, selectNpc, myAgentNpcId, setMyAgent }) {
  const age = npc.age;
  const lifeStage = npc.lifeStage;
  const s = npc.stats;
  const n = npc.needs;
  const npcIndex = (world?.npcs || []).findIndex(np => np.id === npc.id);

  if (!npc.alive) {
    return (
      <>
        <div className="insp-header insp-dead">
          <h3>{npc.name} <span className="death-marker">DEAD</span></h3>
          <span className="insp-sub">{age}yo {npc.profession} ({lifeStage}) &mdash; Died of {npc.causeOfDeath}</span>
        </div>
        <div className="insp-section">
          <h4>Last Known Memories</h4>
          <div className="insp-mem">
            <Memories mems={npc.recentMemories} />
          </div>
        </div>
      </>
    );
  }

  return (
    <>
      <div className="insp-header insp-header-with-sprite">
        {npcIndex >= 0 && <NpcSpritePreview npcIndex={npcIndex} profession={npc.profession} />}
        <div className="insp-header-text">
          <h3><span className="insp-name-text">{npc.name}</span> <span className={`mood-badge mood-${moodClass(npc.mood)}`}>{npc.mood}</span></h3>
          <span className="insp-sub">{age}yo {npc.profession} ({lifeStage}) &mdash; {npc.personalitySummary}</span>
        </div>
      </div>

      {setMyAgent && (
        <div style={{ padding: '2px 6px 4px' }}>
          {npc.id === myAgentNpcId
            ? <span className="sdv-btn" style={{ fontSize: 7, padding: '3px 8px', opacity: 0.7, cursor: 'default' }} title="This is your prophet">{'\u2605'} My Prophet</span>
            : <button className="sdv-btn" style={{ fontSize: 7, padding: '3px 8px' }} onClick={() => setMyAgent(npc.id)}>Set as My Prophet</button>
          }
        </div>
      )}

      <ActivitySection npc={npc} world={world} />

      <div className="insp-section">
        <h4>Vitals</h4>
        <Bar label="HP" val={Math.round(npc.hp)} max={100} urgent={npc.hp < 40} />
        <Bar label="Hunger" val={Math.round(n.hunger)} max={100} urgent={n.hunger < 25} />
        <Bar label="Thirst" val={Math.round(n.thirst)} max={100} urgent={n.thirst < 15} />
        <Bar label="Fatigue" val={Math.round(n.fatigue)} max={100} urgent={n.fatigue > 85} invert />
        <Bar label="Social" val={Math.round(n.socialNeed)} max={100} urgent={n.socialNeed > 70} invert />
        <Bar label="Happiness" val={Math.round(npc.happiness)} max={100} />
        <Bar label="Stress" val={Math.round(npc.stress)} max={100} urgent={npc.stress > 75} invert />
        <Bar label="Hygiene" val={Math.round(npc.hygiene)} max={100} urgent={npc.hygiene < 20} />
        <Bar label="Sobriety" val={Math.round(npc.sobriety)} max={100} urgent={npc.sobriety < 30} />
      </div>

      <div className="insp-section">
        <h4>Skills &amp; Education</h4>
        <div className="insp-sub-row">
          <span className={`lit-badge lit-${npc.literacyLevel.replace(/\s/g, '-')}`}>{npc.literacyLevel}</span>
          <span style={{ color: '#888', fontSize: 11 }}>Literacy: {npc.literacy}/100</span>
          {npc.leadershipScore > 0 && <span style={{ color: '#f1c40f', fontSize: 11 }}>Leadership: {Math.round(npc.leadershipScore)}</span>}
        </div>
        <Skills skills={npc.skills} />
      </div>

      <FactionSection npc={npc} world={world} />
      <EmploymentSection npc={npc} world={world} />
      <MountSection npc={npc} world={world} />

      <div className="insp-section">
        <h4>Relationships</h4>
        <div className="insp-rel">
          <Relationships npc={npc} world={world} selectNpc={selectNpc} />
        </div>
      </div>

      <div className="insp-section">
        <h4>Equipment</h4>
        <div className="insp-equip-grid">
          <Equipment equipment={npc.equipment} />
        </div>
      </div>

      <div className="insp-section">
        <h4>Inventory <span style={{ fontSize: 9, color: '#888', textTransform: 'none', letterSpacing: 0 }}>({npc.usedSlots}/{npc.maxSlots} slots, {npc.usedWeight}/{Math.round(npc.maxWeight)} wt)</span></h4>
        <div className="insp-inv-grid">
          <Inventory inventory={npc.inventory} />
        </div>
      </div>

      <div className="insp-section">
        <h4>Recent Memories</h4>
        <div className="insp-mem">
          <Memories mems={npc.recentMemories} />
        </div>
      </div>

      <div className="insp-section">
        <h4>Location &amp; Home</h4>
        <div className="insp-kv">
          <span className="kv-label">At:</span> <span className="kv-val">{resolveLocName(npc.locationId, world)}</span>
        </div>
        <div className="insp-kv">
          <span className="kv-label">Home:</span> <span className="kv-val">{resolveLocName(npc.homeId, world)}</span>
        </div>
      </div>

      <div className="insp-section">
        <h4>Biology</h4>
        <div className="insp-kv"><span className="kv-label">Lifespan:</span> <span className="kv-val">~{npc.baseLifespan} years</span></div>
        <div className="insp-kv"><span className="kv-label">Fertility:</span> <span className="kv-val">{npc.fertility}/100</span></div>
        <div className="insp-kv"><span className="kv-label">Genetic Quality:</span> <span className="kv-val">{npc.geneticQuality}/100</span></div>
        <div className="insp-kv"><span className="kv-label">Addiction Risk:</span> <span className="kv-val">{s.addictionSusceptibility}/100</span></div>
      </div>

      <div className="insp-section">
        <h4>Physical Stats</h4>
        <div className="stat-grid">
          <Stat label="STR" val={s.strength} /><Stat label="AGI" val={s.agility} /><Stat label="END" val={s.endurance} />
          <Stat label="DEX" val={s.dexterity} /><Stat label="PTL" val={s.painTolerance} /><Stat label="DSR" val={s.diseaseResistance} />
        </div>
      </div>

      <div className="insp-section">
        <h4>Mental Stats</h4>
        <div className="stat-grid">
          <Stat label="INT" val={s.intelligence} /><Stat label="WIS" val={npc.effectiveWisdom} /><Stat label="CUR" val={s.curiosity} />
          <Stat label="CRT" val={s.creativity} /><Stat label="TAC" val={s.strategicThinking} /><Stat label="DEC" val={s.decisiveness} />
        </div>
      </div>

      <div className="insp-section">
        <h4>Social Stats</h4>
        <div className="stat-grid">
          <Stat label="CHA" val={s.charisma} /><Stat label="EMP" val={s.empathy} /><Stat label="PER" val={s.persuasion} />
          <Stat label="DOM" val={s.dominance} /><Stat label="EXT" val={s.extraversion} /><Stat label="SOC" val={s.socialAwareness} />
        </div>
      </div>

      <div className="insp-section">
        <h4>Social &amp; Reputation</h4>
        <div className="stat-grid">
          <Stat label="REP" val={s.reputation} /><Stat label="INF" val={s.infamy} /><Stat label="ATR" val={s.attractiveness} />
        </div>
      </div>

      <div className="insp-section">
        <h4>Mental State</h4>
        <div className="stat-grid">
          <Stat label="TRM" val={s.trauma} /><Stat label="DEP" val={s.depression} /><Stat label="SAN" val={s.sanity} />
          <Stat label="NEU" val={s.neuroticism} /><Stat label="CMP" val={s.composure} /><Stat label="FRG" val={s.mentalFragility} />
        </div>
      </div>

      <div className="insp-section">
        <h4>Personality</h4>
        <div className="stat-grid">
          <Stat label="AGG" val={s.aggression} /><Stat label="VIR" val={s.virtue} /><Stat label="GRD" val={s.greed} />
          <Stat label="AMB" val={s.ambition} /><Stat label="CRG" val={s.courage} /><Stat label="LOY" val={s.loyalty} />
          <Stat label="GEN" val={s.generosity} /><Stat label="RES" val={s.resilience} /><Stat label="CON" val={s.conscientiousness} />
          <Stat label="OPN" val={s.openness} /><Stat label="CNF" val={s.conformity} /><Stat label="JLY" val={s.jealousy} />
        </div>
      </div>

      <div className="insp-section">
        <h4>Faith &amp; Spiritual</h4>
        <div className="stat-grid">
          <Stat label="SPR" val={s.spiritualSensitivity} /><Stat label="FTH" val={s.faithCapacity} /><Stat label="PRY" val={s.prayerFrequency} />
          <Stat label="DBT" val={s.doubtResistance} /><Stat label="MYS" val={s.mysticalAptitude} /><Stat label="LCK" val={s.divineLuck} />
        </div>
      </div>
    </>
  );
}

function ActivitySection({ npc, world }) {
  const tickCount = useGameStore(s => s.tickCount);
  const parts = [];

  if (npc.goldCount != null) {
    parts.push(<div key="gold" className="insp-kv"><span className="kv-label">Gold:</span> <span className="kv-val kv-gold">{npc.goldCount}</span></div>);
  }
  if (npc.currentGoal) {
    parts.push(<div key="goal" className="insp-kv"><span className="kv-label">Goal:</span> <span className="kv-val">{npc.currentGoal}</span></div>);
  }

  const isTraveling = npc.targetLocationId && npc.travelArrivalTick > 0 && npc.travelArrivalTick > tickCount;

  if (isTraveling) {
    const destName = resolveLocName(npc.targetLocationId, world);
    const pendingLabel = npc.pendingActionId ? ` → then: ${npc.pendingActionId.replace(/_/g, ' ')}` : '';
    const isMounted = npc.mountId && (world?.mounts || []).some(m => m.id === npc.mountId && m.alive);
    const travelVerb = isMounted ? 'Riding to' : 'Traveling to';
    parts.push(<div key="doing" className="insp-kv"><span className="kv-label">Doing:</span> <span className="kv-val act-doing">{travelVerb} {destName}{pendingLabel}</span></div>);
  } else if (npc.pendingActionId) {
    const label = npc.pendingActionId.replace(/_/g, ' ');
    const target = npc.pendingTargetName ? ` → ${npc.pendingTargetName}` : '';
    parts.push(<div key="doing" className="insp-kv"><span className="kv-label">Doing:</span> <span className="kv-val act-doing">{label}{target}</span></div>);
  } else if (npc.lastAction) {
    parts.push(<div key="last" className="insp-kv"><span className="kv-label">Last:</span> <span className="kv-val">{npc.lastAction}</span></div>);
  }
  if (npc.pendingReason) {
    parts.push(<div key="reason" className="insp-kv"><span className="kv-label">Reason:</span> <span className="kv-val act-reason">{npc.pendingReason}</span></div>);
  }
  if (npc.resumeActionId) {
    const rlabel = npc.resumeActionId.replace(/_/g, ' ');
    parts.push(<div key="resume" className="insp-kv"><span className="kv-label">Interrupted:</span> <span className="kv-val act-interrupt">{rlabel} ({npc.resumeTicksLeft} ticks left)</span></div>);
  }
  if (npc.lastDialogue) {
    parts.push(<div key="dialogue" className="insp-dialogue">&ldquo;{npc.lastDialogue}&rdquo;</div>);
  }

  if (parts.length === 0) return null;
  return (
    <div className="insp-section insp-activity">
      <h4>Current Activity</h4>
      {parts}
    </div>
  );
}

function FactionSection({ npc, world }) {
  if (!npc.factionId) return null;
  const faction = (world?.factions || []).find(f => f.id === npc.factionId);
  if (!faction) return null;
  const isLeader = faction.leaderId === npc.id;
  return (
    <div className="insp-section">
      <h4>Faction</h4>
      <div className="faction-badge">
        <span className="faction-name">{faction.name}</span>
        <span className="faction-type">{faction.type}</span>
        {isLeader && <span className="faction-leader">LEADER</span>}
        <span className="faction-members">{faction.memberIds.length} members</span>
      </div>
    </div>
  );
}

function EmploymentSection({ npc, world }) {
  if (!npc.employerId && !npc.isBusinessOwner) return null;
  const parts = [];

  if (npc.isBusinessOwner) {
    parts.push(<div key="owns" className="insp-kv"><span className="kv-label">Owns:</span> <span className="kv-val">{resolveLocName(npc.workplaceId, world)}</span></div>);
  } else if (npc.employerId) {
    const employer = (world?.npcs || []).find(n => n.id === npc.employerId);
    const empName = employer ? employer.name : npc.employerId;
    parts.push(<div key="works" className="insp-kv"><span className="kv-label">Works at:</span> <span className="kv-val">{resolveLocName(npc.workplaceId, world)}</span></div>);
    parts.push(<div key="employer" className="insp-kv"><span className="kv-label">Employer:</span> <span className="kv-val">{empName}</span></div>);
  }
  if (npc.wage > 0) {
    parts.push(<div key="wage" className="insp-kv"><span className="kv-label">Wage:</span> <span className="kv-val">{npc.wage} gold/day</span></div>);
  }
  if (npc.unpaidDays > 0) {
    parts.push(<div key="unpaid" className="insp-kv"><span className="kv-label">Unpaid days:</span> <span className="kv-val act-interrupt">{npc.unpaidDays}</span></div>);
  }
  if ((npc.apprentices || []).length > 0) {
    const names = npc.apprentices.map(aId => {
      const a = (world?.npcs || []).find(n => n.id === aId);
      return a ? a.name : aId;
    });
    parts.push(<div key="apprentices" className="insp-kv"><span className="kv-label">Apprentices:</span> <span className="kv-val">{names.join(', ')}</span></div>);
  }
  if (npc.apprenticeTo) {
    const master = (world?.npcs || []).find(n => n.id === npc.apprenticeTo);
    parts.push(<div key="master" className="insp-kv"><span className="kv-label">Apprentice of:</span> <span className="kv-val">{master ? master.name : npc.apprenticeTo}</span></div>);
  }

  return (
    <div className="insp-section">
      <h4>Employment</h4>
      {parts}
    </div>
  );
}

function MountSection({ npc, world }) {
  const mount = npc.mountId ? (world?.mounts || []).find(m => m.id === npc.mountId) : null;
  const carriage = npc.carriageId ? (world?.carriages || []).find(c => c.id === npc.carriageId) : null;
  if (!mount && !carriage) return null;

  const isTraveling = npc.targetLocationId && npc.travelArrivalTick > 0;

  return (
    <div className="insp-section">
      <h4>Mount &amp; Transport</h4>
      {mount && (
        <>
          <div className="insp-kv">
            <span className="kv-label">Mount:</span>
            <span className="kv-val">{mount.name || mount.type} {isTraveling ? '(riding)' : '(idle)'}</span>
          </div>
          <Bar label="HP" val={mount.hp} max={mount.maxHp} urgent={mount.hp < mount.maxHp * 0.3} />
          <Bar label="Hunger" val={Math.round(mount.hunger)} max={100} urgent={mount.hunger < 25} />
          <Bar label="Grooming" val={Math.round(mount.grooming)} max={100} urgent={mount.grooming < 25} />
          <div className="insp-kv"><span className="kv-label">Speed:</span> <span className="kv-val">{mount.speed}x</span></div>
        </>
      )}
      {carriage && (
        <>
          <div className="insp-kv">
            <span className="kv-label">Carriage:</span>
            <span className="kv-val">{carriage.name || 'Cart'}</span>
          </div>
          <Bar label="Durability" val={Math.round(carriage.durability)} max={100} urgent={carriage.durability < 30} />
          <div className="insp-kv"><span className="kv-label">Cargo:</span> <span className="kv-val">{carriage.cargoSlots} slots / {carriage.cargoWeight} wt</span></div>
        </>
      )}
    </div>
  );
}

function Skills({ skills }) {
  if (!skills) return <div style={{ fontSize: 12, color: '#888' }}>No skills yet.</div>;
  const entries = Object.entries(skills).filter(([, v]) => v > 0).sort(([, a], [, b]) => b - a);
  if (entries.length === 0) return <div style={{ fontSize: 12, color: '#888' }}>No skills yet.</div>;
  return (
    <div className="skill-grid">
      {entries.map(([name, val]) => (
        <div key={name} className="skill-cell">
          <span className="skill-name">{name}</span>
          <MiniBar val={val} />
          <span className="skill-val">{Math.round(val)}</span>
        </div>
      ))}
    </div>
  );
}

function Relationships({ npc, world, selectNpc }) {
  const entries = Object.entries(npc.relationships || {}).filter(([, v]) => v !== 0).sort(([, a], [, b]) => b - a);
  if (entries.length === 0) return <em>No opinions formed yet.</em>;
  return entries.map(([id, val]) => {
    const other = (world?.npcs || []).find(n => n.id === id);
    const name = other ? other.name : 'Unknown';
    const cls = val > 20 ? 'rel-pos' : val < -20 ? 'rel-neg' : 'rel-neutral';
    const label = val > 50 ? 'close friend' : val > 20 ? 'friend' : val > 5 ? 'friendly' :
                  val < -50 ? 'nemesis' : val < -20 ? 'enemy' : val < -5 ? 'unfriendly' : 'neutral';
    return (
      <div key={id} className={`rel-entry ${cls}`} onClick={() => selectNpc(id)} style={{ cursor: 'pointer' }}>
        <span className="rel-name">{name}</span>
        <span className="rel-val">{val > 0 ? '+' : ''}{val}</span>
        <span className="rel-label">{label}</span>
      </div>
    );
  });
}

function Equipment({ equipment }) {
  if (!equipment) return null;
  const slots = [
    { key: 'weapon', label: 'Weapon' },
    { key: 'armor', label: 'Armor' },
    { key: 'bag1', label: 'Bag 1' },
    { key: 'bag2', label: 'Bag 2' },
  ];
  return slots.map(({ key, label }) => {
    const eq = equipment[key];
    if (!eq) return <div key={key} className="equip-slot equip-empty"><span className="equip-label">{label}</span><span className="equip-name">empty</span></div>;
    const dur = eq.durability != null ? eq.durability : 100;
    const durCls = dur < 30 ? 'dur-low' : dur < 60 ? 'dur-mid' : 'dur-ok';
    return (
      <div key={key} className="equip-slot">
        <span className="equip-label">{label}</span>
        <span className="equip-name">{eq.name}</span>
        <span className={`equip-dur ${durCls}`}>{Math.round(dur)}%</span>
      </div>
    );
  });
}

function Inventory({ inventory }) {
  if (!inventory || inventory.length === 0) return <div className="inv-empty">Empty</div>;
  return inventory.map((it, i) => {
    const dur = it.durability != null && it.durability < 100
      ? <span className="inv-dur">{Math.round(it.durability)}%</span> : null;
    return (
      <div key={i} className="inv-item">
        <span className="inv-name">{it.name}</span>
        <span className="inv-qty">x{it.qty}</span>
        {dur}
      </div>
    );
  });
}

function Memories({ mems }) {
  if (!mems || mems.length === 0) return <em>No memories yet.</em>;
  return mems.map((m, i) => (
    <div key={i} className="mem-entry">[{m.time || '?'}] {m.text}</div>
  ));
}

/* ── Location Inspector ────────────────────────────── */

function LocationSpritePreview({ locType, locIndex }) {
  const canvasRef = useRef(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    ctx.imageSmoothingEnabled = false;

    const src = generateBuildingPreviewCanvas(locType, locIndex);
    ctx.drawImage(src, 0, 0, 32, 32, 0, 0, canvas.width, canvas.height);
  }, [locType, locIndex]);

  return (
    <canvas ref={canvasRef} width={64} height={64} className="npc-preview-canvas" />
  );
}

function LocationInspector({ loc, world, selectNpc }) {
  const npcsHere = (world?.npcs || []).filter(n => n.alive && n.locationId === loc.id);
  const enemiesHere = (world?.aliveEnemies || []).filter(e => e.locationId === loc.id);
  const groundItems = (world?.groundItems || []).filter(g => g.locationId === loc.id);
  const constructions = (world?.constructions || []).filter(c => c.locationId === loc.id);
  const employees = (world?.npcs || []).filter(n => n.alive && n.workplaceId === loc.id && n.employerId);

  const owner = loc.ownerId ? (world?.npcs || []).find(n => n.id === loc.ownerId) : null;
  const bldgOwner = loc.buildingOwnerId ? (world?.npcs || []).find(n => n.id === loc.buildingOwnerId) : null;
  const locIndex = (world?.locations || []).findIndex(l => l.id === loc.id);

  return (
    <>
      <div className="insp-header insp-header-with-sprite">
        <LocationSpritePreview locType={loc.buildingType || loc.type} locIndex={Math.max(0, locIndex)} />
        <div className="insp-header-text">
          <h3><span className="insp-name-text">{loc.name}</span> <span className="loc-type-badge">{loc.type}</span></h3>
          {loc.description && <span className="insp-sub">{loc.description}</span>}
        </div>
      </div>

      <div className="insp-section">
        {owner && (
          <div className="insp-kv">
            <span className="kv-label">Owner:</span>
            <span className="kv-val loc-npc-link" onClick={() => selectNpc(owner.id)}>{owner.name}</span>
          </div>
        )}
        {loc.tier > 0 && (
          <div className="insp-kv"><span className="kv-label">Tier:</span> <span className="kv-val">{loc.tier}</span></div>
        )}
        <div className="insp-kv"><span className="kv-label">Position:</span> <span className="kv-val">({loc.x}, {loc.y}) {loc.w}x{loc.h}</span></div>
      </div>

      {loc.buildingType && (
        <div className="insp-section">
          <h4>Building</h4>
          <div className="insp-kv"><span className="kv-label">Type:</span> <span className="kv-val">{loc.buildingType}</span></div>
          {bldgOwner && (
            <div className="insp-kv">
              <span className="kv-label">Owner:</span>
              <span className="kv-val loc-npc-link" onClick={() => selectNpc(bldgOwner.id)}>{bldgOwner.name}</span>
            </div>
          )}
          {loc.buildingDurability != null && <Bar label="Durability" val={Math.round(loc.buildingDurability)} max={100} urgent={loc.buildingDurability < 30} />}
        </div>
      )}

      <LocResourceSection loc={loc} />
      <LocEntitySection title="NPCs Present" entities={npcsHere} selectNpc={selectNpc} />
      <LocEnemySection enemies={enemiesHere} />
      <LocEmployeeSection employees={employees} selectNpc={selectNpc} />
      <LocGroundItemSection items={groundItems} />
      <LocConstructionSection constructions={constructions} world={world} />
    </>
  );
}

function LocResourceSection({ loc }) {
  const res = loc.resources;
  if (!res || Object.keys(res).length === 0) return null;
  const maxRes = loc.maxResources || {};
  return (
    <div className="insp-section">
      <h4>Resources</h4>
      {Object.entries(res).map(([key, val]) => {
        const max = maxRes[key] || val;
        return <Bar key={key} label={key} val={val} max={max} urgent={val <= max * 0.2} />;
      })}
    </div>
  );
}

function LocEntitySection({ title, entities, selectNpc }) {
  if (entities.length === 0) return null;
  return (
    <div className="insp-section">
      <h4>{title} <span className="loc-count">({entities.length})</span></h4>
      {entities.map(n => (
        <div key={n.id} className="loc-entity-row loc-npc-link" onClick={() => selectNpc(n.id)}>
          <span className="loc-entity-dot" style={{ background: n.color }} />
          <span className="loc-entity-name">{n.name}</span>
          <span className="loc-entity-prof">{n.profession || ''}</span>
          <span className={`mood-badge mood-${moodClass(n.mood || '')}`} style={{ fontSize: 9, padding: '0 4px' }}>{n.mood || ''}</span>
          <span className="loc-entity-hp">HP {n.hp}</span>
        </div>
      ))}
    </div>
  );
}

function LocEnemySection({ enemies }) {
  if (enemies.length === 0) return null;
  return (
    <div className="insp-section">
      <h4>Enemies <span className="loc-count">({enemies.length})</span></h4>
      {enemies.map((e, i) => (
        <div key={i} className="loc-entity-row loc-enemy">
          <span className="loc-entity-dot" style={{ background: '#e74c3c' }} />
          <span className="loc-entity-name">{e.name}</span>
          <span className="loc-entity-prof">{e.category}</span>
          <span className="loc-entity-hp">HP {e.hp}/{e.maxHp}</span>
        </div>
      ))}
    </div>
  );
}

function LocEmployeeSection({ employees, selectNpc }) {
  if (employees.length === 0) return null;
  return (
    <div className="insp-section">
      <h4>Employees <span className="loc-count">({employees.length})</span></h4>
      {employees.map(n => (
        <div key={n.id} className="loc-entity-row loc-npc-link" onClick={() => selectNpc(n.id)}>
          <span className="loc-entity-dot" style={{ background: n.color }} />
          <span className="loc-entity-name">{n.name}</span>
          <span className="loc-entity-prof">{n.profession || ''}</span>
          <span className="loc-entity-hp">{n.wage || 0}g/day</span>
        </div>
      ))}
    </div>
  );
}

function LocGroundItemSection({ items }) {
  if (items.length === 0) return null;
  return (
    <div className="insp-section">
      <h4>Ground Items <span className="loc-count">({items.length})</span></h4>
      <div className="insp-inv-grid">
        {items.map((it, i) => {
          const dur = it.durability != null && it.durability < 100
            ? <span className="inv-dur">{Math.round(it.durability)}%</span> : null;
          return (
            <div key={i} className="inv-item">
              <span className="inv-name">{it.name}</span>
              <span className="inv-qty">x{it.qty}</span>
              {dur}
            </div>
          );
        })}
      </div>
    </div>
  );
}

function LocConstructionSection({ constructions, world }) {
  if (constructions.length === 0) return null;
  return (
    <div className="insp-section">
      <h4>Constructions <span className="loc-count">({constructions.length})</span></h4>
      {constructions.map((c, i) => {
        const pct = c.maxProgress > 0 ? Math.round((c.progress / c.maxProgress) * 100) : 0;
        const commissioner = c.commissionerId ? (world?.npcs || []).find(n => n.id === c.commissionerId) : null;
        return (
          <div key={i} className="loc-construction">
            <div className="loc-construction-name">{c.name || c.buildingType}</div>
            {commissioner && <div className="loc-construction-comm">By: {commissioner.name}</div>}
            <div className="bar-row">
              <span className="bar-label">Progress</span>
              <div className="bar-track"><div className="bar-fill" style={{ width: `${pct}%`, background: '#e67e22' }} /></div>
              <span className="bar-val">{pct}%</span>
            </div>
          </div>
        );
      })}
    </div>
  );
}

/* ── Main Inspector Component ────────────────── */

export default function Inspector() {
  const world = useGameStore(s => s.world);
  const selectedNpc = useGameStore(s => s.selectedNpc);
  const selectedLocation = useGameStore(s => s.selectedLocation);
  const selectNpc = useGameStore(s => s.selectNpc);
  const toggleInspector = useGameStore(s => s.toggleInspector);
  const myAgentNpcId = useGameStore(s => s.myAgentNpcId);
  const setMyAgent = useGameStore(s => s.setMyAgent);

  const npc = selectedNpc ? (world?.npcs || []).find(n => n.id === selectedNpc) : null;
  const loc = selectedLocation ? (world?.locations || []).find(l => l.id === selectedLocation) : null;

  return (
    <div className="right-panel sdv-frame" onPointerDown={e => e.stopPropagation()} onWheel={e => e.stopPropagation()}>
      <div className="panel-header">
        <h4 className="sdv-section-title">📖 Inspector</h4>
        <button className="panel-close sdv-btn" onClick={toggleInspector}>✕</button>
      </div>
      <div className="inspector">
        {npc ? (
          <NpcInspector npc={npc} world={world} selectNpc={selectNpc} myAgentNpcId={myAgentNpcId} setMyAgent={setMyAgent} />
        ) : loc ? (
          <LocationInspector loc={loc} world={world} selectNpc={selectNpc} />
        ) : (
          <p className="inspector-hint">Click an NPC or location on the map to inspect.</p>
        )}
      </div>
    </div>
  );
}
