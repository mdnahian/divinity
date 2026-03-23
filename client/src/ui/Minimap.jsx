import React, { useRef, useEffect, useCallback } from 'react';
import { useGameStore } from '../hooks/useGameStore';
import './Minimap.css';

const SIZE = 180;

const TILE_COLORS = {
  0:  [20, 50, 100],   // DEEP_WATER
  1:  [40, 90, 150],   // SHALLOW_WATER
  2:  [194, 178, 128],  // SAND
  3:  [80, 140, 60],   // GRASS
  4:  [50, 100, 40],   // DARK_GRASS
  5:  [40, 80, 35],    // FOREST_FLOOR
  6:  [120, 100, 70],  // DIRT
  7:  [130, 130, 130], // STONE
  8:  [230, 230, 240], // SNOW
  9:  [180, 210, 230], // ICE
  10: [60, 80, 50],    // SWAMP
  11: [60, 30, 20],    // VOLCANIC
  12: [200, 60, 20],   // LAVA
  13: [150, 140, 110], // ROAD
  14: [160, 170, 150], // TUNDRA
  15: [100, 160, 70],  // FLOWER
  16: [100, 100, 100], // MOUNTAIN
  17: [210, 195, 150], // COASTAL_SAND
  18: [50, 100, 160],  // RIVER
  19: [140, 120, 80],  // BRIDGE
};

export default function Minimap() {
  const canvasRef = useRef(null);
  const terrainDrawn = useRef(false);
  const terrainImageRef = useRef(null);
  const lastWorldKey = useRef(null);
  const visible = useGameStore(s => s.minimapVisible);

  const drawTerrain = useCallback((ctx, world) => {
    if (!world || !world.gridW) return;

    // Get tile data from the scene's terrain generator
    const scene = useGameStore.getState().phaserScene;
    const tileData = scene?._terrainData?.tileData;

    const gridW = world.gridW;
    const gridH = world.gridH || gridW;
    const scaleX = SIZE / gridW;
    const scaleY = SIZE / gridH;

    const imageData = ctx.createImageData(SIZE, SIZE);
    const data = imageData.data;

    for (let py = 0; py < SIZE; py++) {
      for (let px = 0; px < SIZE; px++) {
        const tx = Math.floor(px / scaleX);
        const ty = Math.floor(py / scaleY);
        const idx = (py * SIZE + px) * 4;

        // Look up actual tile type from terrain data
        let tileType = 3; // default to GRASS
        if (tileData && tileData[ty] && tileData[ty][tx] != null) {
          tileType = tileData[ty][tx];
        }

        const color = TILE_COLORS[tileType] || TILE_COLORS[3];
        data[idx] = color[0];
        data[idx + 1] = color[1];
        data[idx + 2] = color[2];
        data[idx + 3] = 255;
      }
    }

    ctx.putImageData(imageData, 0, 0);

    // Draw location overlays
    const locs = world.locations || [];
    for (const loc of locs) {
      const lx = Math.floor(loc.x * scaleX);
      const ly = Math.floor(loc.y * scaleY);
      const lw = Math.max(2, Math.floor((loc.w || 1) * scaleX));
      const lh = Math.max(2, Math.floor((loc.h || 1) * scaleY));
      ctx.fillStyle = loc.color || '#6a5a30';
      ctx.globalAlpha = 0.5;
      ctx.fillRect(lx, ly, lw, lh);
      ctx.globalAlpha = 1;
    }

    // Cache the terrain as an image for fast redraw
    terrainImageRef.current = ctx.getImageData(0, 0, SIZE, SIZE);
  }, []);

  const drawNpcs = useCallback((ctx, world, selectedNpc) => {
    if (!world || !world.gridW) return;
    const gridW = world.gridW;
    const gridH = world.gridH || gridW;
    const scaleX = SIZE / gridW;
    const scaleY = SIZE / gridH;

    // Build location lookup for position resolution
    const locMap = {};
    for (const loc of (world.locations || [])) locMap[loc.id] = loc;

    const npcs = world.npcs || [];
    for (const npc of npcs) {
      if (!npc.alive) continue;
      const loc = locMap[npc.locationId];
      if (!loc) continue;
      const nx = loc.x + (loc.w || 1) * 0.5;
      const ny = loc.y + (loc.h || 1) * 0.5;
      const px = Math.floor(nx * scaleX);
      const py = Math.floor(ny * scaleY);
      const isSelected = npc.id === selectedNpc;

      ctx.fillStyle = isSelected ? '#f0d060' : (npc.color || '#4a9a30');
      const dotSize = isSelected ? 3 : 2;
      ctx.fillRect(px - Math.floor(dotSize / 2), py - Math.floor(dotSize / 2), dotSize, dotSize);
    }

    const enemies = world.aliveEnemies || [];
    for (const e of enemies) {
      const loc = locMap[e.locationId];
      if (!loc) continue;
      const ex = loc.x + (loc.w || 1) * 0.5;
      const ey = loc.y + (loc.h || 1) * 0.5;
      ctx.fillStyle = '#cc4444';
      ctx.fillRect(Math.floor(ex * scaleX) - 1, Math.floor(ey * scaleY) - 1, 3, 3);
    }
  }, []);

  const drawViewport = useCallback((ctx) => {
    const scene = useGameStore.getState().phaserScene;
    if (!scene || !scene.cameras?.main) return;

    const world = useGameStore.getState().world;
    if (!world || !world.gridW) return;

    const cam = scene.cameras.main;
    const gridW = world.gridW;
    const gridH = world.gridH || gridW;
    const worldPxW = gridW * 32;
    const worldPxH = gridH * 32;

    // Use worldView for accurate visible rectangle
    const wv = cam.worldView;
    const vx = Math.floor((wv.x / worldPxW) * SIZE);
    const vy = Math.floor((wv.y / worldPxH) * SIZE);
    const vw = Math.floor((wv.width / worldPxW) * SIZE);
    const vh = Math.floor((wv.height / worldPxH) * SIZE);

    ctx.strokeStyle = '#c8a84e';
    ctx.lineWidth = 1.5;
    ctx.strokeRect(
      Math.max(0, vx) + 0.5,
      Math.max(0, vy) + 0.5,
      Math.min(vw, SIZE - Math.max(0, vx)),
      Math.min(vh, SIZE - Math.max(0, vy))
    );
  }, []);

  useEffect(() => {
    if (!visible) return;

    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');

    let rafId;
    const render = () => {
      const world = useGameStore.getState().world;
      const selectedNpc = useGameStore.getState().selectedNpc;

      if (!world || !world.gridW) {
        rafId = requestAnimationFrame(render);
        return;
      }

      // Check if terrain needs redraw (new world or terrain data appeared)
      const scene = useGameStore.getState().phaserScene;
      const hasTerrain = !!scene?._terrainData?.tileData;
      const worldKey = `${world.gridW}_${(world.locations || []).length}_${hasTerrain}`;

      if (worldKey !== lastWorldKey.current) {
        drawTerrain(ctx, world);
        lastWorldKey.current = worldKey;
        terrainDrawn.current = true;
      }

      if (terrainDrawn.current && terrainImageRef.current) {
        // Restore cached terrain, then draw dynamic elements on top
        ctx.putImageData(terrainImageRef.current, 0, 0);
        drawNpcs(ctx, world, selectedNpc);
        drawViewport(ctx);
      }

      rafId = requestAnimationFrame(render);
    };

    rafId = requestAnimationFrame(render);
    return () => cancelAnimationFrame(rafId);
  }, [visible, drawTerrain, drawNpcs, drawViewport]);

  const handleClick = useCallback((e) => {
    const world = useGameStore.getState().world;
    if (!world || !world.gridW) return;

    const canvas = canvasRef.current;
    const rect = canvas.getBoundingClientRect();
    const mx = e.clientX - rect.left;
    const my = e.clientY - rect.top;

    const worldPxW = world.gridW * 32;
    const worldPxH = (world.gridH || world.gridW) * 32;

    const worldX = (mx / SIZE) * worldPxW;
    const worldY = (my / SIZE) * worldPxH;

    const scene = useGameStore.getState().phaserScene;
    if (scene && scene.cameras?.main) {
      const cam = scene.cameras.main;
      cam.scrollX = worldX - (cam.width / cam.zoom) / 2;
      cam.scrollY = worldY - (cam.height / cam.zoom) / 2;
    }
  }, []);

  if (!visible) return null;

  return (
    <div className="minimap-wrap sdv-frame" onPointerDown={e => e.stopPropagation()} onWheel={e => e.stopPropagation()}>
      <canvas
        ref={canvasRef}
        width={SIZE}
        height={SIZE}
        className="minimap-canvas"
        onClick={handleClick}
      />
      <div className="minimap-label">Map</div>
    </div>
  );
}
