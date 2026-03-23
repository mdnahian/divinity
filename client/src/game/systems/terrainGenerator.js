import { fbm, noise2 } from '../../utils/noise.js';
import { clamp, smoothstep } from '../../utils/math.js';
import { TILE } from './tilesetGenerator.js';

const OFFSET_X = 47.3;
const OFFSET_Y = 13.7;
const MOISTURE_OX = 123.4;
const MOISTURE_OY = 87.6;

function heightAt(x, y, gridW, gridH) {
  const nx = x / gridW;
  const ny = y / gridH;
  const raw = fbm(nx * 6 + OFFSET_X, ny * 6 + OFFSET_Y, 5) * 0.5 + 0.5;
  const edgeX = Math.min(x, gridW - 1 - x) / (gridW * 0.15);
  const edgeY = Math.min(y, gridH - 1 - y) / (gridH * 0.15);
  const edgeFade = smoothstep(clamp(Math.min(edgeX, edgeY), 0, 1));
  return clamp(raw * edgeFade, 0, 1);
}

function moistureAt(x, y, gridW, gridH) {
  const nx = x / gridW;
  const ny = y / gridH;
  return fbm(nx * 4 + MOISTURE_OX, ny * 4 + MOISTURE_OY, 4) * 0.5 + 0.5;
}

// biomeShifts maps biome hint strings to height/moisture adjustments.
// These shift the noise values to favor terrain types appropriate for each biome.
const biomeShifts = {
  desert:   { heightShift: -0.10, moistureShift: -0.35, heat: 0.4 },
  swamp:    { heightShift: -0.15, moistureShift:  0.30, heat: 0.0 },
  forest:   { heightShift:  0.00, moistureShift:  0.20, heat: 0.0 },
  tundra:   { heightShift:  0.05, moistureShift: -0.20, heat: -0.3 },
  plains:   { heightShift: -0.05, moistureShift:  0.00, heat: 0.0 },
  mountain: { heightShift:  0.20, moistureShift: -0.10, heat: -0.1 },
  lake:     { heightShift: -0.25, moistureShift:  0.25, heat: 0.0 },
};

function tileFromHeightMoisture(h, m, heat) {
  // Lake biome override
  if (h < 0.15 && heat === undefined) return TILE.DEEP_WATER;

  if (h < 0.22) return TILE.DEEP_WATER;
  if (h < 0.30) return heat > 0.2 ? TILE.LAKE : TILE.SHALLOW_WATER;
  if (h < 0.34) return heat > 0.2 ? TILE.DESERT_SAND : TILE.SAND;
  if (h < 0.37) return TILE.COASTAL_SAND;

  if (h < 0.65) {
    if (heat > 0.3) {
      // Hot desert zone
      if (m > 0.6) return TILE.OASIS;
      return TILE.DESERT_SAND;
    }
    if (m > 0.75) return TILE.SWAMP;
    if (m > 0.65) return TILE.DEEP_FOREST;
    if (m > 0.55) return TILE.FOREST_FLOOR;
    if (m > 0.45) return TILE.DARK_GRASS;
    if (heat < -0.15 || m < 0.2) return TILE.TUNDRA;
    if (m < 0.25 && heat < 0) return TILE.ICE;
    const detail = noise2(h * 50, m * 50);
    if (detail > 0.4) return TILE.FLOWER;
    return TILE.GRASS;
  }

  if (h < 0.75) {
    if (m > 0.6) return TILE.DIRT;
    return heat < -0.1 ? TILE.CAVE_FLOOR : TILE.STONE;
  }

  if (h < 0.85) return TILE.MOUNTAIN;
  return heat < 0 ? TILE.ICE : TILE.SNOW;
}

// computeBiomeInfluence returns {heightShift, moistureShift, heat} for a given tile
// based on nearby territories and biome overrides.
function computeBiomeInfluence(x, y, territories, biomeOverrides) {
  let hShift = 0, mShift = 0, heat = 0;
  let totalWeight = 0;

  // Territory influence (broad, gradual)
  if (territories && territories.length > 0) {
    for (const t of territories) {
      if (!t.biomeHint || !biomeShifts[t.biomeHint]) continue;
      const dx = x - t.centerX;
      const dy = y - t.centerY;
      const dist = Math.sqrt(dx * dx + dy * dy);
      const radius = t.radius || 50;
      if (dist > radius * 1.5) continue;
      const weight = Math.max(0, 1 - dist / (radius * 1.5));
      const shifts = biomeShifts[t.biomeHint];
      hShift += shifts.heightShift * weight;
      mShift += shifts.moistureShift * weight;
      heat += (shifts.heat || 0) * weight;
      totalWeight += weight;
    }
  }

  // Biome overrides (localized, stronger)
  if (biomeOverrides && biomeOverrides.length > 0) {
    for (const bo of biomeOverrides) {
      if (!biomeShifts[bo.biomeType]) continue;
      const dx = x - bo.x;
      const dy = y - bo.y;
      const dist = Math.sqrt(dx * dx + dy * dy);
      const radius = bo.radius || 10;
      if (dist > radius) continue;
      const weight = (1 - dist / radius) * 1.5; // Stronger than territory
      const shifts = biomeShifts[bo.biomeType];
      hShift += shifts.heightShift * weight;
      mShift += shifts.moistureShift * weight;
      heat += (shifts.heat || 0) * weight;
      totalWeight += weight;
    }
  }

  if (totalWeight > 0) {
    hShift /= totalWeight;
    mShift /= totalWeight;
    heat /= totalWeight;
  }

  return { heightShift: hShift, moistureShift: mShift, heat };
}

function traceRivers(heightMap, gridW, gridH, tileData) {
  const sources = [];
  for (let y = 0; y < gridH; y++) {
    for (let x = 0; x < gridW; x++) {
      const h = heightMap[y * gridW + x];
      if (h > 0.72 && h < 0.82) {
        const detail = noise2(x * 0.5, y * 0.5);
        if (detail > 0.35) sources.push({ x, y });
      }
    }
  }

  const riverSet = new Set();
  const maxRivers = Math.max(2, Math.floor(Math.sqrt(gridW * gridH) / 6));
  const picked = [];
  for (let i = 0; i < sources.length && picked.length < maxRivers; i++) {
    const s = sources[i];
    let tooClose = false;
    for (const p of picked) {
      const dx = s.x - p.x, dy = s.y - p.y;
      if (dx * dx + dy * dy < 100) { tooClose = true; break; }
    }
    if (!tooClose) picked.push(s);
  }

  for (const src of picked) {
    let cx = src.x, cy = src.y;
    const visited = new Set();
    for (let step = 0; step < gridW + gridH; step++) {
      const key = `${cx},${cy}`;
      if (visited.has(key)) break;
      visited.add(key);

      const h = heightMap[cy * gridW + cx];
      if (h < 0.25) break;

      if (h < 0.65 && h > 0.34) {
        riverSet.add(key);
        tileData[cy][cx] = TILE.RIVER;
      }

      let bestH = h;
      let bx = cx, by = cy;
      const dirs = [[-1, 0], [1, 0], [0, -1], [0, 1], [-1, -1], [1, -1], [-1, 1], [1, 1]];
      for (const [dx, dy] of dirs) {
        const nx = cx + dx, ny = cy + dy;
        if (nx < 0 || nx >= gridW || ny < 0 || ny >= gridH) continue;
        const nh = heightMap[ny * gridW + nx];
        if (nh < bestH) { bestH = nh; bx = nx; by = ny; }
      }
      if (bx === cx && by === cy) break;
      cx = bx;
      cy = by;
    }
  }

  return riverSet;
}

function connectLocationsWithRoads(tileData, locations, gridW, gridH, riverSet, backendRoads) {
  const roadSet = new Set();

  const drawRoad = (x0, y0, x1, y1) => {
    let x = x0, y = y0;
    const maxSteps = Math.abs(x1 - x0) + Math.abs(y1 - y0) + 4;
    let steps = 0;
    while ((x !== x1 || y !== y1) && steps++ < maxSteps) {
      if (x >= 0 && x < gridW && y >= 0 && y < gridH) {
        const key = `${x},${y}`;
        const cur = tileData[y][x];
        if (riverSet.has(key)) {
          tileData[y][x] = TILE.BRIDGE;
        } else if (cur !== TILE.DEEP_WATER && cur !== TILE.SHALLOW_WATER && cur !== TILE.MOUNTAIN && cur !== TILE.SNOW) {
          tileData[y][x] = TILE.ROAD;
        }
        roadSet.add(key);
      }

      const dx = x1 - x, dy = y1 - y;
      const jitter = noise2(x * 0.8, y * 0.8);
      if (Math.abs(dx) > Math.abs(dy) + jitter * 2) {
        x += dx > 0 ? 1 : -1;
      } else {
        y += dy > 0 ? 1 : -1;
      }
    }
  };

  // Draw backend roads first (town road grid from city centers)
  if (backendRoads && backendRoads.length > 0) {
    for (const road of backendRoads) {
      drawRoad(road.x1, road.y1, road.x2, road.y2);
    }
  }

  // Then connect locations to nearest neighbors
  for (let i = 0; i < locations.length; i++) {
    const a = locations[i];
    const acx = a.x + Math.floor((a.w || 1) / 2);
    const acy = a.y + Math.floor((a.h || 1) / 2);

    let nearest = null;
    let nearestDist = Infinity;
    for (let j = 0; j < locations.length; j++) {
      if (i === j) continue;
      const b = locations[j];
      const bcx = b.x + Math.floor((b.w || 1) / 2);
      const bcy = b.y + Math.floor((b.h || 1) / 2);
      const d = Math.abs(acx - bcx) + Math.abs(acy - bcy);
      if (d < nearestDist) { nearestDist = d; nearest = { cx: bcx, cy: bcy }; }
    }
    if (nearest && nearestDist < gridW * 0.6) {
      drawRoad(acx, acy, nearest.cx, nearest.cy);
    }
  }

  return roadSet;
}

function shapeTerrainAroundLocations(heightMap, moistureMap, gridW, gridH, locations) {
  const radius = 4;

  for (const loc of (locations || [])) {
    const cx = loc.x + (loc.w || 1) / 2;
    const cy = loc.y + (loc.h || 1) / 2;

    for (let dy = -radius; dy <= (loc.h || 1) + radius; dy++) {
      for (let dx = -radius; dx <= (loc.w || 1) + radius; dx++) {
        const ty = loc.y + dy;
        const tx = loc.x + dx;
        if (ty < 0 || ty >= gridH || tx < 0 || tx >= gridW) continue;
        const idx = ty * gridW + tx;

        const distX = Math.max(0, Math.max(loc.x - tx, tx - (loc.x + (loc.w || 1) - 1)));
        const distY = Math.max(0, Math.max(loc.y - ty, ty - (loc.y + (loc.h || 1) - 1)));
        const dist = Math.sqrt(distX * distX + distY * distY);
        const influence = 1 - clamp(dist / radius, 0, 1);

        switch (loc.type) {
          case 'dock':
            if (dist <= 1) {
              heightMap[idx] = Math.max(heightMap[idx], 0.35);
            } else if (dist <= 3) {
              heightMap[idx] = heightMap[idx] * (1 - influence * 0.5) + 0.18 * influence * 0.5;
              moistureMap[idx] = Math.min(1, moistureMap[idx] + influence * 0.3);
            }
            break;
          case 'mine':
            heightMap[idx] = Math.min(0.74, heightMap[idx] + influence * 0.3);
            moistureMap[idx] = Math.max(0, moistureMap[idx] - influence * 0.2);
            break;
          case 'farm':
            if (dist <= 1) {
              heightMap[idx] = clamp(heightMap[idx], 0.38, 0.55);
            }
            heightMap[idx] = heightMap[idx] * (1 - influence * 0.4) + 0.45 * influence * 0.4;
            moistureMap[idx] = moistureMap[idx] * (1 - influence * 0.3) + 0.4 * influence * 0.3;
            break;
          case 'forest':
            if (dist <= 1) {
              heightMap[idx] = clamp(heightMap[idx], 0.38, 0.60);
            }
            moistureMap[idx] = Math.min(1, moistureMap[idx] + influence * 0.25);
            break;
          default:
            if (dist <= 1) {
              heightMap[idx] = Math.max(heightMap[idx], 0.5);
            }
            break;
        }
      }
    }
  }
}

function isWaterTile(tile) {
  return tile === TILE.DEEP_WATER || tile === TILE.SHALLOW_WATER;
}

function floodFillEliminateIslands(tileData, gridW, gridH) {
  const visited = new Uint8Array(gridW * gridH);
  const cx = Math.floor(gridW / 2);
  const cy = Math.floor(gridH / 2);

  let startX = cx, startY = cy;
  if (isWaterTile(tileData[startY][startX])) {
    let found = false;
    for (let r = 1; r < Math.max(gridW, gridH) && !found; r++) {
      for (let dy = -r; dy <= r && !found; dy++) {
        for (let dx = -r; dx <= r && !found; dx++) {
          const sx = cx + dx, sy = cy + dy;
          if (sx >= 0 && sx < gridW && sy >= 0 && sy < gridH && !isWaterTile(tileData[sy][sx])) {
            startX = sx; startY = sy; found = true;
          }
        }
      }
    }
    if (!found) return;
  }

  const stack = [[startX, startY]];
  visited[startY * gridW + startX] = 1;
  while (stack.length > 0) {
    const [x, y] = stack.pop();
    for (const [dx, dy] of [[-1, 0], [1, 0], [0, -1], [0, 1]]) {
      const nx = x + dx, ny = y + dy;
      if (nx < 0 || nx >= gridW || ny < 0 || ny >= gridH) continue;
      const idx = ny * gridW + nx;
      if (visited[idx]) continue;
      if (isWaterTile(tileData[ny][nx])) continue;
      visited[idx] = 1;
      stack.push([nx, ny]);
    }
  }

  for (let y = 0; y < gridH; y++) {
    for (let x = 0; x < gridW; x++) {
      if (!visited[y * gridW + x] && !isWaterTile(tileData[y][x])) {
        tileData[y][x] = TILE.SHALLOW_WATER;
      }
    }
  }
}

export function generateTerrain(gridW, gridH, locations, backendRoads, territories, biomeOverrides) {
  const heightMap = new Float32Array(gridW * gridH);
  const moistureMap = new Float32Array(gridW * gridH);
  const heatMap = new Float32Array(gridW * gridH);

  const hasBiomes = (territories && territories.length > 0) || (biomeOverrides && biomeOverrides.length > 0);

  for (let y = 0; y < gridH; y++) {
    for (let x = 0; x < gridW; x++) {
      const idx = y * gridW + x;
      let h = heightAt(x, y, gridW, gridH);
      let m = moistureAt(x, y, gridW, gridH);
      let heat = 0;

      if (hasBiomes) {
        const influence = computeBiomeInfluence(x, y, territories, biomeOverrides);
        h = clamp(h + influence.heightShift, 0, 1);
        m = clamp(m + influence.moistureShift, 0, 1);
        heat = influence.heat;
      }

      heightMap[idx] = h;
      moistureMap[idx] = m;
      heatMap[idx] = heat;
    }
  }

  shapeTerrainAroundLocations(heightMap, moistureMap, gridW, gridH, locations);

  for (const loc of (locations || [])) {
    for (let dy = -1; dy <= (loc.h || 1); dy++) {
      for (let dx = -1; dx <= (loc.w || 1); dx++) {
        const ty = loc.y + dy;
        const tx = loc.x + dx;
        if (ty >= 0 && ty < gridH && tx >= 0 && tx < gridW) {
          heightMap[ty * gridW + tx] = clamp(Math.max(heightMap[ty * gridW + tx], 0.5), 0, 0.74);
        }
      }
    }
  }

  const tileData = [];
  for (let y = 0; y < gridH; y++) {
    const row = [];
    for (let x = 0; x < gridW; x++) {
      const idx = y * gridW + x;
      row.push(tileFromHeightMoisture(heightMap[idx], moistureMap[idx], heatMap[idx]));
    }
    tileData.push(row);
  }

  floodFillEliminateIslands(tileData, gridW, gridH);

  const riverSet = traceRivers(heightMap, gridW, gridH, tileData);

  const LOC_TYPE_TILE = {
    forest: TILE.FOREST_FLOOR,
    farm:   TILE.FLOWER,
    mine:   TILE.STONE,
    dock:   TILE.COASTAL_SAND,
    well:   TILE.DIRT,
    inn:    TILE.ROAD,
    market: TILE.ROAD,
    shrine: TILE.STONE,
    forge:  TILE.DIRT,
    mill:   TILE.DIRT,
    home:   TILE.GRASS,
    library: TILE.STONE,
    school: TILE.ROAD,
    barracks: TILE.DIRT,
    warehouse: TILE.ROAD,
    workshop: TILE.DIRT,
    tavern: TILE.ROAD,
    garden: TILE.FLOWER,
    palace: TILE.ROAD,
    castle: TILE.STONE,
    manor: TILE.ROAD,
    stable: TILE.DIRT,
    arena: TILE.SAND,
    cave: TILE.CAVE_FLOOR,
    dungeon_entrance: TILE.CAVE_FLOOR,
    desert: TILE.DESERT_SAND,
    swamp: TILE.SWAMP,
    tundra: TILE.TUNDRA,
  };

  for (const loc of (locations || [])) {
    const tile = LOC_TYPE_TILE[loc.type] ?? TILE.GRASS;
    for (let dy = -1; dy <= (loc.h || 1); dy++) {
      for (let dx = -1; dx <= (loc.w || 1); dx++) {
        const ty = loc.y + dy;
        const tx = loc.x + dx;
        if (ty < 0 || ty >= gridH || tx < 0 || tx >= gridW) continue;
        const isEdge = dy === -1 || dx === -1 || dy === (loc.h || 1) || dx === (loc.w || 1);
        if (isEdge) {
          const cur = tileData[ty][tx];
          if (cur === TILE.DEEP_WATER || cur === TILE.SHALLOW_WATER) {
            tileData[ty][tx] = TILE.SAND;
          }
        } else {
          tileData[ty][tx] = tile;
        }
      }
    }
  }

  const roadSet = connectLocationsWithRoads(tileData, locations || [], gridW, gridH, riverSet, backendRoads || null);

  return { tileData, heightMap, moistureMap, riverSet, roadSet };
}
