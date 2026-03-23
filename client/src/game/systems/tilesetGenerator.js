/**
 * Procedural pixel art tileset generator.
 * Draws all terrain tiles onto a canvas at runtime -- no external asset files needed.
 * Each tile is 32x32 pixels. Tiles are laid out in a single row for Phaser.
 */

const TS = 32;

const TILE = {
  DEEP_WATER:    0,
  SHALLOW_WATER: 1,
  SAND:          2,
  GRASS:         3,
  DARK_GRASS:    4,
  FOREST_FLOOR:  5,
  DIRT:          6,
  STONE:         7,
  SNOW:          8,
  ICE:           9,
  SWAMP:        10,
  VOLCANIC:     11,
  LAVA:         12,
  ROAD:         13,
  TUNDRA:       14,
  FLOWER:       15,
  MOUNTAIN:     16,
  COASTAL_SAND: 17,
  RIVER:        18,
  BRIDGE:       19,
  DESERT_SAND:  20,
  OASIS:        21,
  CAVE_FLOOR:   22,
  DEEP_FOREST:  23,
  LAKE:         24,
};

const TILE_COUNT = Object.keys(TILE).length;

function hexToRgb(hex) {
  const n = parseInt(hex.replace('#', ''), 16);
  return [(n >> 16) & 0xff, (n >> 8) & 0xff, n & 0xff];
}

function drawPixel(ctx, x, y, r, g, b) {
  ctx.fillStyle = `rgb(${r},${g},${b})`;
  ctx.fillRect(x, y, 1, 1);
}

function vary(base, range, rng) {
  return Math.max(0, Math.min(255, base + Math.floor((rng() - 0.5) * range)));
}

function simpleRng(seed) {
  let s = seed | 0;
  return () => {
    s ^= s << 13;
    s ^= s >> 17;
    s ^= s << 5;
    return (s >>> 0) / 0xFFFFFFFF;
  };
}

function fillTile(ctx, ox, oy, baseR, baseG, baseB, variance, seed) {
  const rng = simpleRng(seed);
  for (let y = 0; y < TS; y++) {
    for (let x = 0; x < TS; x++) {
      drawPixel(ctx, ox + x, oy + y,
        vary(baseR, variance, rng),
        vary(baseG, variance, rng),
        vary(baseB, variance, rng));
    }
  }
}

function drawWaterTile(ctx, ox, oy, deep) {
  const rng = simpleRng(deep ? 1001 : 2001);
  const baseR = deep ? 20 : 40;
  const baseG = deep ? 50 : 90;
  const baseB = deep ? 100 : 150;
  for (let y = 0; y < TS; y++) {
    for (let x = 0; x < TS; x++) {
      const wave = Math.sin((x + y * 0.7) * 0.4) * 8;
      drawPixel(ctx, ox + x, oy + y,
        vary(baseR, 10, rng) + wave,
        vary(baseG, 12, rng) + wave,
        vary(baseB, 15, rng) + wave * 0.5);
    }
  }
}

function drawGrassTile(ctx, ox, oy, dark, seed) {
  const rng = simpleRng(seed);
  const baseR = dark ? 50 : 80;
  const baseG = dark ? 100 : 140;
  const baseB = dark ? 35 : 50;
  for (let y = 0; y < TS; y++) {
    for (let x = 0; x < TS; x++) {
      const blade = (rng() > 0.92) ? 15 : 0;
      drawPixel(ctx, ox + x, oy + y,
        vary(baseR, 18, rng) - blade,
        vary(baseG, 22, rng) + blade,
        vary(baseB, 12, rng));
    }
  }
}

function drawSwampTile(ctx, ox, oy) {
  const rng = simpleRng(5001);
  for (let y = 0; y < TS; y++) {
    for (let x = 0; x < TS; x++) {
      const pool = rng() > 0.85;
      if (pool) {
        drawPixel(ctx, ox + x, oy + y, vary(40, 8, rng), vary(70, 10, rng), vary(50, 8, rng));
      } else {
        drawPixel(ctx, ox + x, oy + y, vary(55, 14, rng), vary(85, 16, rng), vary(40, 10, rng));
      }
    }
  }
}

function drawVolcanicTile(ctx, ox, oy) {
  const rng = simpleRng(6001);
  for (let y = 0; y < TS; y++) {
    for (let x = 0; x < TS; x++) {
      const crack = rng() > 0.95;
      if (crack) {
        drawPixel(ctx, ox + x, oy + y, vary(180, 30, rng), vary(60, 20, rng), vary(10, 8, rng));
      } else {
        drawPixel(ctx, ox + x, oy + y, vary(40, 12, rng), vary(30, 10, rng), vary(25, 8, rng));
      }
    }
  }
}

function drawLavaTile(ctx, ox, oy) {
  const rng = simpleRng(6501);
  for (let y = 0; y < TS; y++) {
    for (let x = 0; x < TS; x++) {
      const flow = Math.sin((x * 0.3 + y * 0.5)) * 0.5 + 0.5;
      drawPixel(ctx, ox + x, oy + y,
        vary(200 + flow * 55, 15, rng),
        vary(80 + flow * 40, 12, rng),
        vary(10, 8, rng));
    }
  }
}

function drawFlowerTile(ctx, ox, oy) {
  const rng = simpleRng(8001);
  const flowers = [[200, 100, 120], [220, 180, 80], [180, 120, 200], [240, 200, 100]];
  for (let y = 0; y < TS; y++) {
    for (let x = 0; x < TS; x++) {
      if (rng() > 0.88) {
        const f = flowers[Math.floor(rng() * flowers.length)];
        drawPixel(ctx, ox + x, oy + y, vary(f[0], 15, rng), vary(f[1], 15, rng), vary(f[2], 15, rng));
      } else {
        drawPixel(ctx, ox + x, oy + y, vary(80, 16, rng), vary(145, 20, rng), vary(55, 12, rng));
      }
    }
  }
}

function drawMountainTile(ctx, ox, oy) {
  const rng = simpleRng(9001);
  for (let y = 0; y < TS; y++) {
    for (let x = 0; x < TS; x++) {
      const height = (1 - y / TS) * 30;
      drawPixel(ctx, ox + x, oy + y,
        vary(110 + height, 18, rng),
        vary(100 + height, 16, rng),
        vary(90 + height, 14, rng));
    }
  }
}

function drawRoadTile(ctx, ox, oy) {
  const rng = simpleRng(7001);
  for (let y = 0; y < TS; y++) {
    for (let x = 0; x < TS; x++) {
      const pebble = rng() > 0.9 ? 12 : 0;
      drawPixel(ctx, ox + x, oy + y,
        vary(150, 14, rng) + pebble,
        vary(130, 12, rng) + pebble,
        vary(100, 10, rng));
    }
  }
}

function drawBridgeTile(ctx, ox, oy) {
  const rng = simpleRng(7501);
  for (let y = 0; y < TS; y++) {
    for (let x = 0; x < TS; x++) {
      const plank = Math.floor(x / 8) % 2 === 0;
      const base = plank ? 120 : 105;
      drawPixel(ctx, ox + x, oy + y,
        vary(base + 20, 10, rng),
        vary(base, 8, rng),
        vary(base - 30, 8, rng));
    }
  }
}

function drawOasisTile(ctx, ox, oy) {
  const rng = simpleRng(7101);
  for (let y = 0; y < TS; y++) {
    for (let x = 0; x < TS; x++) {
      const distCenter = Math.sqrt((x - 16) * (x - 16) + (y - 16) * (y - 16));
      if (distCenter < 10) {
        // Water center
        drawPixel(ctx, ox + x, oy + y,
          vary(40, 10, rng), vary(120, 15, rng), vary(180, 12, rng));
      } else if (distCenter < 14) {
        // Green vegetation ring
        drawPixel(ctx, ox + x, oy + y,
          vary(50, 12, rng), vary(140, 15, rng), vary(40, 10, rng));
      } else {
        // Sand border
        drawPixel(ctx, ox + x, oy + y,
          vary(220, 12, rng), vary(200, 10, rng), vary(150, 10, rng));
      }
    }
  }
}

function drawLakeTile(ctx, ox, oy) {
  const rng = simpleRng(7401);
  for (let y = 0; y < TS; y++) {
    for (let x = 0; x < TS; x++) {
      const wave = Math.sin(x * 0.3 + y * 0.2) * 8;
      drawPixel(ctx, ox + x, oy + y,
        vary(50, 8, rng), vary(100 + wave, 12, rng), vary(170 + wave, 10, rng));
    }
  }
}

export function generateTileset() {
  const canvas = document.createElement('canvas');
  canvas.width = TILE_COUNT * TS;
  canvas.height = TS;
  const ctx = canvas.getContext('2d');

  drawWaterTile(ctx, TILE.DEEP_WATER * TS, 0, true);
  drawWaterTile(ctx, TILE.SHALLOW_WATER * TS, 0, false);
  fillTile(ctx, TILE.SAND * TS, 0, 210, 190, 140, 16, 3001);
  drawGrassTile(ctx, TILE.GRASS * TS, 0, false, 3501);
  drawGrassTile(ctx, TILE.DARK_GRASS * TS, 0, true, 3601);
  fillTile(ctx, TILE.FOREST_FLOOR * TS, 0, 60, 90, 40, 20, 4001);
  fillTile(ctx, TILE.DIRT * TS, 0, 130, 100, 60, 18, 4501);
  fillTile(ctx, TILE.STONE * TS, 0, 128, 128, 128, 14, 4801);
  fillTile(ctx, TILE.SNOW * TS, 0, 235, 238, 242, 8, 5501);
  fillTile(ctx, TILE.ICE * TS, 0, 200, 220, 240, 10, 5601);
  drawSwampTile(ctx, TILE.SWAMP * TS, 0);
  drawVolcanicTile(ctx, TILE.VOLCANIC * TS, 0);
  drawLavaTile(ctx, TILE.LAVA * TS, 0);
  drawRoadTile(ctx, TILE.ROAD * TS, 0);
  fillTile(ctx, TILE.TUNDRA * TS, 0, 155, 170, 155, 14, 5801);
  drawFlowerTile(ctx, TILE.FLOWER * TS, 0);
  drawMountainTile(ctx, TILE.MOUNTAIN * TS, 0);
  fillTile(ctx, TILE.COASTAL_SAND * TS, 0, 200, 185, 140, 14, 6101);
  drawWaterTile(ctx, TILE.RIVER * TS, 0, false);
  drawBridgeTile(ctx, TILE.BRIDGE * TS, 0);

  // New biome tiles
  fillTile(ctx, TILE.DESERT_SAND * TS, 0, 230, 210, 160, 12, 7001);  // hot desert sand
  drawOasisTile(ctx, TILE.OASIS * TS, 0);
  fillTile(ctx, TILE.CAVE_FLOOR * TS, 0, 80, 75, 70, 16, 7201);      // dark cave stone
  fillTile(ctx, TILE.DEEP_FOREST * TS, 0, 30, 65, 25, 22, 7301);     // very dark forest
  drawLakeTile(ctx, TILE.LAKE * TS, 0);

  return canvas.toDataURL();
}

export { TILE, TILE_COUNT, TS };
