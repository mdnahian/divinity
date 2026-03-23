import React, { useState, useRef, useEffect, useCallback, useMemo } from 'react';
import { useGameStore } from '../../hooks/useGameStore';
import PopupModal from '../PopupModal';

const MAX_GRAPH_NODES = 60;

function RelationshipGraph({ npcs }) {
  const canvasRef = useRef(null);
  const nodesRef = useRef([]);
  const animRef = useRef(null);
  const [hoveredNode, setHoveredNode] = useState(null);
  const [selectedNode, setSelectedNode] = useState(null);
  const dragRef = useRef(null);

  // Build nodes & edges from NPC data
  useEffect(() => {
    if (!npcs.length) return;

    const existing = nodesRef.current;
    const nodeIndex = new Map();
    for (const n of existing) nodeIndex.set(n.id, n);

    const nodes = npcs.map((npc, i) => {
      const prev = nodeIndex.get(npc.id);
      const angle = (2 * Math.PI * i) / npcs.length;
      const radius = Math.min(280, 140 + npcs.length * 4);
      return {
        id: npc.id,
        name: npc.name,
        profession: npc.profession || 'None',
        relationships: npc.relationships || {},
        x: prev?.x ?? 350 + Math.cos(angle) * radius,
        y: prev?.y ?? 280 + Math.sin(angle) * radius,
        vx: 0, vy: 0,
      };
    });
    nodesRef.current = nodes;
  }, [npcs]);

  // Force simulation + render loop
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    const W = canvas.width;
    const H = canvas.height;

    const tick = () => {
      const nodes = nodesRef.current;
      if (!nodes.length) { animRef.current = requestAnimationFrame(tick); return; }

      // Build index for fast lookups
      const idx = new Map();
      for (const n of nodes) idx.set(n.id, n);

      // Force simulation step
      const dt = 0.3;
      for (const n of nodes) { n.vx *= 0.85; n.vy *= 0.85; }

      // Repulsion between all nodes
      for (let i = 0; i < nodes.length; i++) {
        for (let j = i + 1; j < nodes.length; j++) {
          let dx = nodes[j].x - nodes[i].x;
          let dy = nodes[j].y - nodes[i].y;
          let dist = Math.sqrt(dx * dx + dy * dy) || 1;
          let force = 2000 / (dist * dist);
          let fx = (dx / dist) * force;
          let fy = (dy / dist) * force;
          nodes[i].vx -= fx * dt;
          nodes[i].vy -= fy * dt;
          nodes[j].vx += fx * dt;
          nodes[j].vy += fy * dt;
        }
      }

      // Attraction along edges (relationships)
      for (const n of nodes) {
        for (const [targetId, val] of Object.entries(n.relationships)) {
          const target = idx.get(targetId);
          if (!target) continue;
          let dx = target.x - n.x;
          let dy = target.y - n.y;
          let dist = Math.sqrt(dx * dx + dy * dy) || 1;
          const idealDist = Math.abs(val) > 30 ? 100 : 180;
          let force = (dist - idealDist) * 0.005;
          n.vx += (dx / dist) * force * dt;
          n.vy += (dy / dist) * force * dt;
        }
      }

      // Center gravity
      for (const n of nodes) {
        n.vx += (W / 2 - n.x) * 0.001;
        n.vy += (H / 2 - n.y) * 0.001;
      }

      // Integrate
      for (const n of nodes) {
        if (dragRef.current?.id === n.id) continue;
        n.x += n.vx;
        n.y += n.vy;
        n.x = Math.max(30, Math.min(W - 30, n.x));
        n.y = Math.max(30, Math.min(H - 30, n.y));
      }

      // Draw
      ctx.clearRect(0, 0, W, H);

      // Draw edges
      const drawnEdges = new Set();
      for (const n of nodes) {
        for (const [targetId, val] of Object.entries(n.relationships)) {
          const edgeKey = n.id < targetId ? `${n.id}-${targetId}` : `${targetId}-${n.id}`;
          if (drawnEdges.has(edgeKey)) continue;
          drawnEdges.add(edgeKey);
          const target = idx.get(targetId);
          if (!target) continue;

          const isHighlighted = selectedNode === n.id || selectedNode === targetId ||
            hoveredNode === n.id || hoveredNode === targetId;

          const alpha = selectedNode
            ? (isHighlighted ? 0.8 : 0.08)
            : (hoveredNode ? (isHighlighted ? 0.7 : 0.12) : 0.3);

          ctx.beginPath();
          ctx.moveTo(n.x, n.y);
          ctx.lineTo(target.x, target.y);
          const color = val > 20 ? `rgba(74, 154, 48, ${alpha})`
            : val < -20 ? `rgba(204, 68, 68, ${alpha})`
            : `rgba(139, 105, 20, ${alpha})`;
          ctx.strokeStyle = color;
          ctx.lineWidth = Math.min(4, Math.max(1, Math.abs(val) / 15));
          ctx.stroke();

          // Draw opinion value on highlighted edges
          if (isHighlighted && alpha > 0.3) {
            const mx = (n.x + target.x) / 2;
            const my = (n.y + target.y) / 2;
            ctx.font = '10px "Press Start 2P", monospace';
            ctx.fillStyle = val > 0 ? '#4A9A30' : val < 0 ? '#CC4444' : '#8B6914';
            ctx.textAlign = 'center';
            ctx.fillText((val > 0 ? '+' : '') + val, mx, my - 4);
          }
        }
      }

      // Draw nodes
      for (const n of nodes) {
        const isSelected = selectedNode === n.id;
        const isHovered = hoveredNode === n.id;
        const selectedRels = selectedNode ? idx.get(selectedNode)?.relationships : null;
        const isConnected = selectedRels ? selectedRels[n.id] !== undefined : false;
        const dimmed = (selectedNode && !isSelected && !isConnected) || (hoveredNode && !isHovered && !isConnected);

        const radius = isSelected ? 14 : isHovered ? 12 : 10;
        const alpha = dimmed ? 0.2 : 1;

        // Node circle
        ctx.beginPath();
        ctx.arc(n.x, n.y, radius, 0, Math.PI * 2);
        ctx.fillStyle = isSelected ? `rgba(212, 160, 32, ${alpha})` :
          isHovered ? `rgba(180, 138, 16, ${alpha})` :
          `rgba(139, 105, 20, ${alpha})`;
        ctx.fill();
        ctx.strokeStyle = `rgba(93, 65, 12, ${alpha})`;
        ctx.lineWidth = 2;
        ctx.stroke();

        // Name label
        ctx.font = '8px "Press Start 2P", monospace';
        ctx.textAlign = 'center';
        ctx.fillStyle = `rgba(60, 40, 10, ${alpha})`;
        ctx.fillText(n.name.split(' ')[0], n.x, n.y + radius + 12);

        if (isSelected || isHovered) {
          ctx.font = '7px "Press Start 2P", monospace';
          ctx.fillStyle = '#8B6914';
          ctx.fillText(n.profession, n.x, n.y + radius + 22);
        }
      }

      animRef.current = requestAnimationFrame(tick);
    };

    animRef.current = requestAnimationFrame(tick);
    return () => { if (animRef.current) cancelAnimationFrame(animRef.current); };
  }, [hoveredNode, selectedNode]);

  const getNodeAt = useCallback((x, y) => {
    for (const n of nodesRef.current) {
      const dx = n.x - x;
      const dy = n.y - y;
      if (dx * dx + dy * dy < 196) return n;
    }
    return null;
  }, []);

  const handleMouseMove = useCallback((e) => {
    const rect = canvasRef.current.getBoundingClientRect();
    const x = (e.clientX - rect.left) * (canvasRef.current.width / rect.width);
    const y = (e.clientY - rect.top) * (canvasRef.current.height / rect.height);

    if (dragRef.current) {
      const node = nodesRef.current.find(n => n.id === dragRef.current.id);
      if (node) { node.x = x; node.y = y; node.vx = 0; node.vy = 0; }
      return;
    }

    const node = getNodeAt(x, y);
    setHoveredNode(node?.id || null);
    canvasRef.current.style.cursor = node ? 'pointer' : 'default';
  }, [getNodeAt]);

  const handleMouseDown = useCallback((e) => {
    const rect = canvasRef.current.getBoundingClientRect();
    const x = (e.clientX - rect.left) * (canvasRef.current.width / rect.width);
    const y = (e.clientY - rect.top) * (canvasRef.current.height / rect.height);
    const node = getNodeAt(x, y);
    if (node) {
      dragRef.current = node;
      setSelectedNode(prev => prev === node.id ? null : node.id);
    } else {
      setSelectedNode(null);
    }
  }, [getNodeAt]);

  const handleMouseUp = useCallback(() => { dragRef.current = null; }, []);

  return (
    <canvas
      ref={canvasRef}
      width={700}
      height={500}
      style={{ width: '100%', height: 'auto', borderRadius: 4, background: 'rgba(245, 230, 200, 0.5)', border: '1px solid var(--sdv-parchment-deep)' }}
      onMouseMove={handleMouseMove}
      onMouseDown={handleMouseDown}
      onMouseUp={handleMouseUp}
      onMouseLeave={() => { setHoveredNode(null); dragRef.current = null; }}
    />
  );
}

export default function RelationshipsPopup() {
  const world = useGameStore(s => s.world);
  const [view, setView] = useState('graph');
  const [search, setSearch] = useState('');
  const [territory, setTerritory] = useState('all');
  const [minStrength, setMinStrength] = useState(10);

  const allNpcs = useMemo(() => (world?.npcs || []).filter(n => n.alive), [world]);
  const territories = useMemo(() => world?.territories || [], [world]);

  // Build location → territory lookup
  const locTerritoryMap = useMemo(() => {
    const m = {};
    for (const loc of (world?.locations || [])) {
      if (loc.territoryId) m[loc.id] = loc.territoryId;
    }
    return m;
  }, [world]);

  // Filter NPCs: only those with at least one meaningful relationship
  const filteredNpcs = useMemo(() => {
    const npcIdSet = new Set(allNpcs.map(n => n.id));

    return allNpcs.filter(npc => {
      // Territory filter
      if (territory !== 'all') {
        const npcTerritory = locTerritoryMap[npc.locationId];
        if (npcTerritory !== territory) return false;
      }

      // Search filter
      if (search) {
        const s = search.toLowerCase();
        if (!npc.name.toLowerCase().includes(s) && !(npc.profession || '').toLowerCase().includes(s)) return false;
      }

      // Must have at least one relationship meeting the strength threshold
      const rels = npc.relationships || {};
      return Object.entries(rels).some(([targetId, val]) =>
        Math.abs(val) >= minStrength && npcIdSet.has(targetId)
      );
    });
  }, [allNpcs, territory, search, minStrength, locTerritoryMap]);

  // For the graph, cap to MAX_GRAPH_NODES — prioritize most connected
  const graphNpcs = useMemo(() => {
    if (filteredNpcs.length <= MAX_GRAPH_NODES) return filteredNpcs;

    // Score by total relationship weight
    const scored = filteredNpcs.map(npc => {
      const rels = Object.values(npc.relationships || {});
      const score = rels.reduce((s, v) => s + Math.abs(v), 0);
      return { npc, score };
    });
    scored.sort((a, b) => b.score - a.score);

    const topNpcs = scored.slice(0, MAX_GRAPH_NODES).map(s => s.npc);
    // Filter relationships to only include nodes in the graph
    const idSet = new Set(topNpcs.map(n => n.id));
    return topNpcs.map(npc => ({
      ...npc,
      relationships: Object.fromEntries(
        Object.entries(npc.relationships || {}).filter(([id]) => idSet.has(id))
      ),
    }));
  }, [filteredNpcs]);

  // For the matrix, limit to 40 NPCs
  const matrixNpcs = useMemo(() => {
    if (filteredNpcs.length <= 40) return filteredNpcs;
    const scored = filteredNpcs.map(npc => {
      const rels = Object.values(npc.relationships || {});
      return { npc, score: rels.reduce((s, v) => s + Math.abs(v), 0) };
    });
    scored.sort((a, b) => b.score - a.score);
    return scored.slice(0, 40).map(s => s.npc);
  }, [filteredNpcs]);

  // Compute global stats from all alive NPCs
  const totalRelationships = allNpcs.reduce((sum, n) => sum + Object.keys(n.relationships || {}).length, 0);
  const avgOpinion = allNpcs.reduce((sum, n) => {
    const vals = Object.values(n.relationships || {});
    return sum + vals.reduce((s, v) => s + v, 0);
  }, 0) / Math.max(totalRelationships, 1);

  const friendships = allNpcs.reduce((sum, n) => {
    return sum + Object.values(n.relationships || {}).filter(v => v > 20).length;
  }, 0);
  const rivalries = allNpcs.reduce((sum, n) => {
    return sum + Object.values(n.relationships || {}).filter(v => v < -20).length;
  }, 0);

  return (
    <PopupModal icon="❤️" title="Relationships">
      <div className="popup-section">
        <div className="popup-grid" style={{ gridTemplateColumns: 'repeat(4, 1fr)' }}>
          <div className="popup-card" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)' }}>Total Bonds</div>
            <div className="popup-stat-value" style={{ fontSize: 18 }}>{totalRelationships}</div>
          </div>
          <div className="popup-card" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)' }}>Avg Opinion</div>
            <div className="popup-stat-value" style={{ fontSize: 18, color: avgOpinion >= 0 ? 'var(--sdv-green)' : 'var(--sdv-red)' }}>
              {avgOpinion >= 0 ? '+' : ''}{avgOpinion.toFixed(1)}
            </div>
          </div>
          <div className="popup-card" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)' }}>Friendships</div>
            <div className="popup-stat-value" style={{ fontSize: 18, color: 'var(--sdv-green)' }}>{friendships}</div>
          </div>
          <div className="popup-card" style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)' }}>Rivalries</div>
            <div className="popup-stat-value" style={{ fontSize: 18, color: 'var(--sdv-red)' }}>{rivalries}</div>
          </div>
        </div>
      </div>

      {/* Filters */}
      <div className="popup-section" style={{ display: 'flex', gap: 8, flexWrap: 'wrap', alignItems: 'center' }}>
        <input
          type="text"
          placeholder="Search NPC..."
          value={search}
          onChange={e => setSearch(e.target.value)}
          className="sdv-input"
          style={{ fontSize: 9, padding: '3px 8px', width: 140 }}
        />
        <select
          value={territory}
          onChange={e => setTerritory(e.target.value)}
          className="sdv-input"
          style={{ fontSize: 9, padding: '3px 6px' }}
        >
          <option value="all">All Territories</option>
          {territories.map(t => (
            <option key={t.id} value={t.id}>{t.name || t.id}</option>
          ))}
        </select>
        <label style={{ fontSize: 9, color: 'var(--sdv-text-dim)', display: 'flex', alignItems: 'center', gap: 4 }}>
          Min |opinion|:
          <input
            type="range" min="0" max="50" value={minStrength}
            onChange={e => setMinStrength(Number(e.target.value))}
            style={{ width: 80 }}
          />
          <span style={{ width: 20, textAlign: 'right' }}>{minStrength}</span>
        </label>
        <span style={{ fontSize: 9, color: 'var(--sdv-text-dim)' }}>
          Showing {filteredNpcs.length} / {allNpcs.length} NPCs
        </span>
      </div>

      <div className="popup-section">
        <div style={{ display: 'flex', gap: 4, marginBottom: 8 }}>
          <button className={`sdv-btn ${view === 'graph' ? 'active' : ''}`} style={{ fontSize: 7, padding: '4px 10px' }} onClick={() => setView('graph')}>
            Network Graph
          </button>
          <button className={`sdv-btn ${view === 'matrix' ? 'active' : ''}`} style={{ fontSize: 7, padding: '4px 10px' }} onClick={() => setView('matrix')}>
            Opinion Matrix
          </button>
        </div>

        {view === 'graph' && (
          <>
            <div style={{ fontSize: 10, color: 'var(--sdv-text-dim)', marginBottom: 6 }}>
              Click a node to highlight connections. Drag nodes to rearrange. Green = friends, Red = rivals.
              {graphNpcs.length < filteredNpcs.length && (
                <span> Showing top {graphNpcs.length} most connected.</span>
              )}
            </div>
            <RelationshipGraph npcs={graphNpcs} />
          </>
        )}

        {view === 'matrix' && (
          <div style={{ overflowX: 'auto' }}>
            {matrixNpcs.length < filteredNpcs.length && (
              <div style={{ fontSize: 9, color: 'var(--sdv-text-dim)', marginBottom: 4 }}>
                Showing top {matrixNpcs.length} most connected NPCs.
              </div>
            )}
            <table className="popup-table" style={{ fontSize: 9 }}>
              <thead>
                <tr>
                  <th style={{ position: 'sticky', left: 0, background: 'var(--sdv-parchment)', zIndex: 1 }}></th>
                  {matrixNpcs.map(n => (
                    <th key={n.id} style={{ writingMode: 'vertical-rl', textOrientation: 'mixed', maxWidth: 20, padding: '4px 2px', fontSize: 6 }}>
                      {n.name.split(' ')[0]}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {matrixNpcs.map(n => (
                  <tr key={n.id}>
                    <td style={{ fontWeight: 600, whiteSpace: 'nowrap', position: 'sticky', left: 0, background: 'var(--sdv-parchment)', zIndex: 1, fontSize: 9 }}>
                      {n.name.split(' ')[0]}
                    </td>
                    {matrixNpcs.map(other => {
                      if (n.id === other.id) return <td key={other.id} style={{ background: 'var(--sdv-parchment-deep)', textAlign: 'center' }}>-</td>;
                      const val = (n.relationships || {})[other.id];
                      const bg = val === undefined ? 'transparent' :
                        val > 20 ? `rgba(74, 154, 48, ${Math.min(1, Math.abs(val) / 80)})` :
                        val < -20 ? `rgba(204, 68, 68, ${Math.min(1, Math.abs(val) / 80)})` :
                        'rgba(139, 105, 20, 0.1)';
                      return (
                        <td key={other.id} style={{ background: bg, textAlign: 'center', fontSize: 8, padding: '2px 3px', color: val === undefined ? 'var(--sdv-text-dim)' : 'var(--sdv-text-bright)' }}>
                          {val !== undefined ? val : ''}
                        </td>
                      );
                    })}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </PopupModal>
  );
}
