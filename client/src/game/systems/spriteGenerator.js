/**
 * Procedural pixel art sprite generator for buildings, trees, NPCs, and landmarks.
 * All sprites are drawn to off-screen canvases and loaded into Phaser as textures.
 */

const S = 32; // sprite size

function simpleRng(seed) {
  let s = seed | 0;
  return () => {
    s ^= s << 13; s ^= s >> 17; s ^= s << 5;
    return (s >>> 0) / 0xFFFFFFFF;
  };
}

function px(ctx, x, y, r, g, b, a = 255) {
  ctx.fillStyle = `rgba(${r},${g},${b},${a / 255})`;
  ctx.fillRect(x, y, 1, 1);
}

function rect(ctx, x, y, w, h, r, g, b) {
  ctx.fillStyle = `rgb(${r},${g},${b})`;
  ctx.fillRect(x, y, w, h);
}

function vary(base, range, rng) {
  return Math.max(0, Math.min(255, base + Math.floor((rng() - 0.5) * range)));
}

// ── TREES ──────────────────────────────────────────

function drawPineTree(ctx, ox, oy, rng) {
  rect(ctx, ox + 14, oy + 22, 4, 10, vary(90, 20, rng), vary(60, 15, rng), vary(30, 10, rng));
  for (let layer = 0; layer < 4; layer++) {
    const w = 8 + layer * 3;
    const x = ox + 16 - Math.floor(w / 2);
    const y = oy + 4 + layer * 5;
    const g = vary(30 + layer * 8, 15, rng);
    for (let py = 0; py < 7; py++) {
      const rowW = Math.max(1, Math.floor(w * (py + 1) / 7));
      const rx = x + Math.floor((w - rowW) / 2);
      for (let px = 0; px < rowW; px++) {
        const gr = vary(g, 12, rng);
        ctx.fillStyle = `rgb(${vary(20, 8, rng)},${gr},${vary(15, 8, rng)})`;
        ctx.fillRect(rx + px, y + py, 1, 1);
      }
    }
  }
}

function drawOakTree(ctx, ox, oy, rng) {
  rect(ctx, ox + 13, oy + 20, 6, 12, vary(85, 15, rng), vary(55, 12, rng), vary(25, 8, rng));
  for (let cy = -8; cy <= 4; cy++) {
    for (let cx = -10; cx <= 10; cx++) {
      if (cx * cx + cy * cy > 100) continue;
      const g = vary(55 + Math.floor(cy * 2), 18, rng);
      const x = ox + 16 + cx;
      const y = oy + 14 + cy;
      if (x >= ox && x < ox + S && y >= oy && y < oy + S) {
        ctx.fillStyle = `rgb(${vary(30, 10, rng)},${g},${vary(20, 10, rng)})`;
        ctx.fillRect(x, y, 1, 1);
      }
    }
  }
}

function drawPalmTree(ctx, ox, oy, rng) {
  for (let i = 0; i < 20; i++) {
    const x = ox + 15 + Math.floor(Math.sin(i * 0.3) * 2);
    rect(ctx, x, oy + 12 + i, 3, 1, vary(180, 15, rng), vary(150, 12, rng), vary(100, 10, rng));
  }
  const fronds = [[- 8, -4], [8, -4], [-6, -6], [6, -6], [0, -8]];
  fronds.forEach(([dx, dy]) => {
    for (let i = 0; i < 8; i++) {
      const x = ox + 16 + Math.floor(dx * i / 8);
      const y = oy + 12 + Math.floor(dy * i / 8) + Math.floor(i * i * 0.05);
      rect(ctx, x, y, 2, 1, vary(50, 12, rng), vary(120, 15, rng), vary(30, 10, rng));
    }
  });
}

function drawDeadTree(ctx, ox, oy, rng) {
  rect(ctx, ox + 14, oy + 14, 4, 18, vary(70, 12, rng), vary(50, 10, rng), vary(35, 8, rng));
  const branches = [[-6, -3], [5, -5], [-4, -7], [6, -2]];
  branches.forEach(([dx, dy]) => {
    for (let i = 0; i < 6; i++) {
      const x = ox + 16 + Math.floor(dx * i / 6);
      const y = oy + 14 + Math.floor(dy * i / 6);
      rect(ctx, x, y, 2, 1, vary(65, 10, rng), vary(45, 8, rng), vary(30, 8, rng));
    }
  });
}

// ── BUILDINGS ───────────────────────────────────────

// ── BUILDING VARIANTS ──────────────────────────────

function drawBuilding(ctx, ox, oy, type, rng, canvasW, canvasH, buildingData) {
  canvasW = canvasW || S;
  canvasH = canvasH || S;

  switch (type) {
    case 'well': return _drawWell(ctx, ox, oy, rng, canvasW, canvasH);
    case 'ruin': return _drawRuin(ctx, ox, oy, rng, canvasW, canvasH);
    case 'farm': return _drawFarm(ctx, ox, oy, rng, canvasW, canvasH);
    case 'mine': return _drawMine(ctx, ox, oy, rng, canvasW, canvasH);
    case 'forest': return _drawForestClearing(ctx, ox, oy, rng, canvasW, canvasH);
    case 'inn': case 'tavern': return _drawInn(ctx, ox, oy, rng, canvasW, canvasH);
    case 'forge': return _drawForge(ctx, ox, oy, rng, canvasW, canvasH);
    case 'market': return _drawMarket(ctx, ox, oy, rng, canvasW, canvasH);
    case 'shrine': return _drawShrine(ctx, ox, oy, rng, canvasW, canvasH);
    case 'mill': return _drawMill(ctx, ox, oy, rng, canvasW, canvasH);
    case 'library': case 'school': return _drawLibrary(ctx, ox, oy, rng, canvasW, canvasH);
    case 'workshop': return _drawWorkshop(ctx, ox, oy, rng, canvasW, canvasH);
    case 'dock': return _drawDock(ctx, ox, oy, rng, canvasW, canvasH, buildingData);
    case 'garden': return _drawGarden(ctx, ox, oy, rng, canvasW, canvasH);
    case 'warehouse': return _drawWarehouse(ctx, ox, oy, rng, canvasW, canvasH);
    case 'barracks': return _drawBarracks(ctx, ox, oy, rng, canvasW, canvasH);
    case 'palace': return _drawPalace(ctx, ox, oy, rng, canvasW, canvasH);
    case 'castle': return _drawCastle(ctx, ox, oy, rng, canvasW, canvasH);
    case 'manor': return _drawManor(ctx, ox, oy, rng, canvasW, canvasH);
    default: return _drawHouse(ctx, ox, oy, type, rng, canvasW, canvasH);
  }
}

// Helper to draw a basic rectangular building with wall/roof
function _drawBasicBuilding(ctx, ox, oy, wallC, roofC, rng, w, h, opts = {}) {
  const wallH = opts.wallH || Math.max(8, Math.floor(h * 0.55));
  const wallW = opts.wallW || Math.max(12, Math.floor(w * 0.7));
  const wallX = ox + Math.floor((w - wallW) / 2);
  const wallY = oy + h - wallH - Math.floor(h * 0.08);

  // Ground shadow (offset bottom-right, semi-transparent dark)
  const shOff = Math.max(1, Math.floor(Math.min(w, h) * 0.06));
  for (let y = 0; y < wallH + 2; y++) {
    for (let x = 0; x < wallW + 1; x++) {
      const sx = wallX + x + shOff;
      const sy = wallY + y + shOff;
      if (sx < ox + w && sy < oy + h) {
        rect(ctx, sx, sy, 1, 1, vary(20, 4, rng), vary(20, 4, rng), vary(15, 4, rng));
      }
    }
  }

  // Wall with texture
  for (let y = 0; y < wallH; y++) {
    for (let x = 0; x < wallW; x++) {
      const brick = (opts.bricks && ((y % 4 === 0) || (x % 6 === 0 && y % 4 === 2))) ? -15 : 0;
      const plank = (opts.planks && (x % 4 === 0)) ? -10 : 0;
      // Right edge darker for 3D depth (2px strip)
      const edgeDark = (x >= wallW - 2) ? -30 : 0;
      // Bottom edge slightly darker
      const bottomDark = (y >= wallH - 1) ? -15 : 0;
      rect(ctx, wallX + x, wallY + y, 1, 1,
        vary(wallC[0] + brick + plank + edgeDark + bottomDark, 8, rng),
        vary(wallC[1] + brick + plank + edgeDark + bottomDark, 6, rng),
        vary(wallC[2] + brick + plank + edgeDark + bottomDark, 6, rng));
    }
  }

  // Roof
  const roofH = opts.roofH || Math.max(4, Math.floor(wallH * 0.45));
  if (opts.flatRoof) {
    for (let y = 0; y < 3; y++) {
      for (let x = -1; x <= wallW; x++) {
        rect(ctx, wallX + x, wallY - y - 1, 1, 1,
          vary(roofC[0], 8, rng), vary(roofC[1], 6, rng), vary(roofC[2], 6, rng));
      }
    }
  } else {
    for (let y = 0; y < roofH; y++) {
      const rw = wallW + 2 - Math.floor(y * (wallW + 2) / roofH);
      if (rw < 2) break;
      const rx = wallX - 1 + Math.floor((wallW + 2 - rw) / 2);
      for (let x = 0; x < rw; x++) {
        const tile = (opts.tiledRoof && y % 2 === 0 && x % 3 === 0) ? 10 : 0;
        // Right half of roof slightly darker for 3D
        const roofEdge = (x > rw * 0.6) ? -12 : 0;
        rect(ctx, rx + x, wallY - y - 1, 1, 1,
          vary(roofC[0] + tile + roofEdge, 8, rng), vary(roofC[1] + tile + roofEdge, 6, rng), vary(roofC[2] + tile + roofEdge, 6, rng));
      }
    }
  }

  // Door
  const doorW = Math.max(3, Math.floor(wallW * 0.2));
  const doorH = Math.max(4, Math.floor(wallH * 0.4));
  const doorX = wallX + Math.floor((wallW - doorW) / 2);
  const doorY = wallY + wallH - doorH;
  const doorC = opts.doorColor || [70, 45, 20];
  rect(ctx, doorX, doorY, doorW, doorH, vary(doorC[0], 10, rng), vary(doorC[1], 8, rng), vary(doorC[2], 8, rng));

  // Windows
  if (wallW >= 14) {
    const winSize = Math.max(2, Math.floor(wallW * 0.12));
    const winY = wallY + Math.floor(wallH * 0.2);
    const winC = opts.windowGlow || [180, 200, 140];
    rect(ctx, wallX + 2, winY, winSize, winSize, vary(winC[0], 12, rng), vary(winC[1], 10, rng), vary(winC[2], 8, rng));
    rect(ctx, wallX + wallW - 2 - winSize, winY, winSize, winSize, vary(winC[0], 12, rng), vary(winC[1], 10, rng), vary(winC[2], 8, rng));
  }

  return { wallX, wallY, wallW, wallH };
}

// ── Helper: draw a thick stone wall segment with brick detail ──
function _drawStoneWall(ctx, x, y, w, h, stoneC, rng, darkSide) {
  for (let py = 0; py < h; py++) {
    for (let px = 0; px < w; px++) {
      const brick = ((py % 4 === 0) || (px % 6 === 0 && py % 4 === 2)) ? -12 : 0;
      const depth = darkSide && px >= w - 2 ? -20 : 0;
      rect(ctx, x + px, y + py, 1, 1,
        vary(stoneC[0] + brick + depth, 5, rng),
        vary(stoneC[1] + brick + depth, 4, rng),
        vary(stoneC[2] + brick + depth, 4, rng));
    }
  }
}

// ── Helper: draw a tower with pointed cap and crenellations ──
function _drawTower(ctx, tx, ty, tw, bodyH, capColor, stoneC, rng) {
  // Shadow
  for (let py = 0; py < bodyH; py++) {
    rect(ctx, tx + tw, ty + py, 2, 1, vary(20, 4, rng), vary(20, 4, rng), vary(15, 4, rng));
  }
  // Body
  for (let py = 0; py < bodyH; py++) {
    for (let px = 0; px < tw; px++) {
      const edgeDark = px >= tw - 2 ? -22 : 0;
      const brick = (py % 5 === 0 || px % 5 === 0) ? -10 : 0;
      rect(ctx, tx + px, ty + py, 1, 1,
        vary(stoneC[0] + edgeDark + brick, 5, rng),
        vary(stoneC[1] + edgeDark + brick, 4, rng),
        vary(stoneC[2] + edgeDark + brick, 4, rng));
    }
  }
  // Crenellations
  for (let px = 0; px < tw; px += 4) {
    rect(ctx, tx + px, ty - 3, 3, 3, vary(stoneC[0] + 5, 5, rng), vary(stoneC[1] + 5, 4, rng), vary(stoneC[2] + 5, 4, rng));
  }
  // Pointed cap
  const capH = Math.floor(tw * 0.7);
  for (let py = 0; py < capH; py++) {
    const rw = tw - Math.floor(py * tw / capH);
    if (rw < 1) break;
    const rx = tx + Math.floor((tw - rw) / 2);
    for (let px = 0; px < rw; px++) {
      const shade = px > rw * 0.6 ? -15 : 0;
      rect(ctx, rx + px, ty - 3 - py - 1, 1, 1,
        vary(capColor[0] + shade, 6, rng), vary(capColor[1] + shade, 5, rng), vary(capColor[2] + shade, 5, rng));
    }
  }
}

// ── PALACE — Massive royal compound: 16x14 tiles (512x448px) ──
function _drawPalace(ctx, ox, oy, rng, w, h) {
  const stoneC = [185, 180, 170];
  const darkStone = [160, 155, 148];
  const goldRoof = [175, 145, 40];
  const courtC = [170, 160, 125];

  const wallThick = Math.max(6, Math.floor(w * 0.04));
  const margin = Math.floor(w * 0.02);
  const wL = ox + margin;
  const wR = ox + w - margin;
  const wT = oy + Math.floor(h * 0.18);
  const wB = oy + h - margin;

  // Ground shadow for entire compound
  for (let py = wT + 2; py <= wB + 3; py++) {
    for (let px = wL + 2; px <= wR + 3; px++) {
      if (py > wB || px > wR) {
        rect(ctx, px, py, 1, 1, vary(18, 3, rng), vary(18, 3, rng), vary(12, 3, rng));
      }
    }
  }

  // Courtyard fill
  for (let py = wT + wallThick; py < wB - wallThick; py++) {
    for (let px = wL + wallThick; px < wR - wallThick; px++) {
      const tile = ((px + py) % 8 < 4) ? 5 : -5;
      rect(ctx, px, py, 1, 1, vary(courtC[0] + tile, 4, rng), vary(courtC[1] + tile, 3, rng), vary(courtC[2] + tile, 3, rng));
    }
  }

  // Curtain walls — 4 sides
  _drawStoneWall(ctx, wL, wT, wR - wL, wallThick, stoneC, rng, false);           // top
  _drawStoneWall(ctx, wL, wB - wallThick, wR - wL, wallThick, stoneC, rng, false); // bottom
  _drawStoneWall(ctx, wL, wT, wallThick, wB - wT, stoneC, rng, false);            // left
  _drawStoneWall(ctx, wR - wallThick, wT, wallThick, wB - wT, darkStone, rng, true); // right (darker)

  // Crenellations on all 4 walls
  for (let px = wL; px < wR; px += 5) {
    rect(ctx, px, wT - 4, 3, 4, vary(stoneC[0] + 5, 4, rng), vary(stoneC[1] + 5, 3, rng), vary(stoneC[2] + 5, 3, rng));
    rect(ctx, px, wB, 3, 4, vary(stoneC[0] - 5, 4, rng), vary(stoneC[1] - 5, 3, rng), vary(stoneC[2] - 5, 3, rng));
  }
  for (let py = wT; py < wB; py += 5) {
    rect(ctx, wL - 3, py, 3, 3, vary(stoneC[0] + 5, 4, rng), vary(stoneC[1] + 5, 3, rng), vary(stoneC[2] + 5, 3, rng));
    rect(ctx, wR, py, 3, 3, vary(darkStone[0] + 5, 4, rng), vary(darkStone[1] + 5, 3, rng), vary(darkStone[2] + 5, 3, rng));
  }

  // Corner towers (4 corners)
  const tw = Math.max(14, Math.floor(w * 0.06));
  const th = Math.floor(h * 0.35);
  const corners = [[wL - 2, wT - th], [wR - tw + 2, wT - th], [wL - 2, wB - tw], [wR - tw + 2, wB - tw]];
  for (const [tx, ty] of corners) {
    _drawTower(ctx, tx, ty, tw, th + tw, goldRoof, [170, 165, 158], rng);
  }

  // Mid-wall towers (one on each side)
  const midX = ox + Math.floor(w / 2) - Math.floor(tw / 2);
  const midY = oy + Math.floor(h / 2) - Math.floor(tw / 2);
  _drawTower(ctx, midX, wT - Math.floor(th * 0.7), tw, Math.floor(th * 0.7) + wallThick, goldRoof, [170, 165, 158], rng);
  _drawTower(ctx, wL - 2, midY - Math.floor(th * 0.4), tw, Math.floor(th * 0.6) + tw, goldRoof, [170, 165, 158], rng);
  _drawTower(ctx, wR - tw + 2, midY - Math.floor(th * 0.4), tw, Math.floor(th * 0.6) + tw, goldRoof, [170, 165, 158], rng);

  // Main keep — large central building
  const keepW = Math.floor(w * 0.4);
  const keepH = Math.floor(h * 0.4);
  const keepX = ox + Math.floor((w - keepW) / 2);
  const keepY = wT + Math.floor((wB - wT - keepH) / 2) - Math.floor(keepH * 0.05);
  _drawBasicBuilding(ctx, keepX, keepY, [195, 190, 180], goldRoof, rng, keepW, keepH, {
    bricks: true, tiledRoof: true,
    wallH: Math.floor(keepH * 0.6),
    windowGlow: [230, 210, 100],
  });

  // Throne room wing (wider, behind keep)
  const throneW = Math.floor(keepW * 0.6);
  const throneH = Math.floor(keepH * 0.5);
  _drawBasicBuilding(ctx, keepX + Math.floor((keepW - throneW) / 2), keepY - Math.floor(throneH * 0.3), [190, 185, 175], goldRoof, rng, throneW, throneH, {
    bricks: true, tiledRoof: true, wallH: Math.floor(throneH * 0.55),
  });

  // Side wings
  const wingW = Math.floor(keepW * 0.35);
  const wingH = Math.floor(keepH * 0.6);
  _drawBasicBuilding(ctx, keepX - wingW - 4, keepY + Math.floor(keepH * 0.2), [180, 175, 165], [140, 115, 35], rng, wingW, wingH, {
    bricks: true, wallH: Math.floor(wingH * 0.55),
  });
  _drawBasicBuilding(ctx, keepX + keepW + 4, keepY + Math.floor(keepH * 0.2), [180, 175, 165], [140, 115, 35], rng, wingW, wingH, {
    bricks: true, wallH: Math.floor(wingH * 0.55),
  });

  // Grand gatehouse (center bottom)
  const ghW = Math.max(20, Math.floor(w * 0.1));
  const ghH = Math.max(16, wallThick + 10);
  const ghX = ox + Math.floor((w - ghW) / 2);
  const ghY = wB - ghH;
  _drawStoneWall(ctx, ghX, ghY, ghW, ghH, [175, 170, 162], rng, true);
  // Gate arch
  const archW = Math.floor(ghW * 0.5);
  const archH = Math.floor(ghH * 0.7);
  const archX = ghX + Math.floor((ghW - archW) / 2);
  rect(ctx, archX, ghY + ghH - archH, archW, archH, vary(35, 4, rng), vary(25, 3, rng), vary(15, 3, rng));
  // Portcullis
  for (let px = archX + 1; px < archX + archW - 1; px += 3) {
    rect(ctx, px, ghY + ghH - archH + 1, 1, archH - 2, vary(80, 5, rng), vary(75, 4, rng), vary(70, 4, rng));
  }
  // Gatehouse towers
  const gtw = Math.floor(ghW * 0.35);
  _drawTower(ctx, ghX - gtw + 2, ghY - Math.floor(th * 0.5), gtw, Math.floor(th * 0.5) + ghH, goldRoof, [175, 170, 162], rng);
  _drawTower(ctx, ghX + ghW - 2, ghY - Math.floor(th * 0.5), gtw, Math.floor(th * 0.5) + ghH, goldRoof, [175, 170, 162], rng);

  // Royal banners on main keep
  const flagPole = Math.floor(h * 0.06);
  for (const fx of [keepX + Math.floor(keepW * 0.3), keepX + Math.floor(keepW * 0.7)]) {
    rect(ctx, fx, keepY - flagPole - 5, 1, flagPole + 5, vary(95, 5, rng), vary(75, 4, rng), vary(45, 4, rng));
    rect(ctx, fx + 1, keepY - flagPole - 5, 7, 4, vary(195, 8, rng), vary(40, 8, rng), vary(40, 8, rng));
    rect(ctx, fx + 2, keepY - flagPole - 3, 5, 1, vary(220, 6, rng), vary(200, 6, rng), vary(50, 6, rng));
  }

  // Garden courtyard patches
  const gardenY = keepY + keepH + 8;
  for (let gx = wL + wallThick + 10; gx < wR - wallThick - 15; gx += Math.floor(w * 0.12)) {
    for (let gy = gardenY; gy < Math.min(gardenY + 12, wB - wallThick - 2); gy++) {
      for (let px = 0; px < 10; px++) {
        rect(ctx, gx + px, gy, 1, 1, vary(60, 12, rng), vary(120 + Math.floor(rng() * 30), 10, rng), vary(40, 8, rng));
      }
    }
  }
}

// ── CASTLE — Fortified stronghold: 10x9 tiles (320x288px) ──
function _drawCastle(ctx, ox, oy, rng, w, h) {
  const stoneC = [145, 140, 132];
  const darkStone = [120, 115, 108];
  const roofC = [88, 82, 76];
  const courtC = [155, 145, 112];

  const wallThick = Math.max(5, Math.floor(w * 0.04));
  const margin = Math.floor(w * 0.02);
  const wL = ox + margin;
  const wR = ox + w - margin;
  const wT = oy + Math.floor(h * 0.2);
  const wB = oy + h - margin;

  // Shadow
  for (let py = wT + 2; py <= wB + 3; py++) {
    for (let px = wL + 2; px <= wR + 3; px++) {
      if (py > wB || px > wR) {
        rect(ctx, px, py, 1, 1, vary(18, 3, rng), vary(18, 3, rng), vary(12, 3, rng));
      }
    }
  }

  // Courtyard
  for (let py = wT + wallThick; py < wB - wallThick; py++) {
    for (let px = wL + wallThick; px < wR - wallThick; px++) {
      rect(ctx, px, py, 1, 1, vary(courtC[0], 4, rng), vary(courtC[1], 3, rng), vary(courtC[2], 3, rng));
    }
  }

  // Curtain walls
  _drawStoneWall(ctx, wL, wT, wR - wL, wallThick, stoneC, rng, false);
  _drawStoneWall(ctx, wL, wB - wallThick, wR - wL, wallThick, stoneC, rng, false);
  _drawStoneWall(ctx, wL, wT, wallThick, wB - wT, stoneC, rng, false);
  _drawStoneWall(ctx, wR - wallThick, wT, wallThick, wB - wT, darkStone, rng, true);

  // Crenellations
  for (let px = wL; px < wR; px += 5) {
    rect(ctx, px, wT - 3, 3, 3, vary(stoneC[0] + 5, 4, rng), vary(stoneC[1] + 5, 3, rng), vary(stoneC[2] + 5, 3, rng));
    rect(ctx, px, wB, 3, 3, vary(stoneC[0] - 5, 4, rng), vary(stoneC[1] - 5, 3, rng), vary(stoneC[2] - 5, 3, rng));
  }
  for (let py = wT; py < wB; py += 5) {
    rect(ctx, wL - 2, py, 2, 3, vary(stoneC[0] + 5, 4, rng), vary(stoneC[1] + 5, 3, rng), vary(stoneC[2] + 5, 3, rng));
    rect(ctx, wR, py, 2, 3, vary(darkStone[0] + 5, 4, rng), vary(darkStone[1] + 5, 3, rng), vary(darkStone[2] + 5, 3, rng));
  }

  // Corner towers
  const tw = Math.max(12, Math.floor(w * 0.07));
  const th = Math.floor(h * 0.3);
  const corners = [[wL - 2, wT - th], [wR - tw + 2, wT - th], [wL - 2, wB - tw], [wR - tw + 2, wB - tw]];
  for (const [tx, ty] of corners) {
    _drawTower(ctx, tx, ty, tw, th + tw, roofC, [155, 150, 142], rng);
  }

  // Central keep
  const keepW = Math.floor(w * 0.38);
  const keepH = Math.floor(h * 0.42);
  const keepX = ox + Math.floor((w - keepW) / 2);
  const keepY = wT + Math.floor((wB - wT - keepH) / 2) - Math.floor(keepH * 0.05);
  _drawBasicBuilding(ctx, keepX, keepY, darkStone, roofC, rng, keepW, keepH, {
    bricks: true,
    wallH: Math.floor(keepH * 0.62),
  });

  // Gatehouse
  const ghW = Math.max(16, Math.floor(w * 0.1));
  const ghH = Math.max(12, wallThick + 6);
  const ghX = ox + Math.floor((w - ghW) / 2);
  const ghY = wB - ghH;
  _drawStoneWall(ctx, ghX, ghY, ghW, ghH, [150, 145, 138], rng, true);
  const archW = Math.floor(ghW * 0.45);
  const archH = Math.floor(ghH * 0.65);
  const archX = ghX + Math.floor((ghW - archW) / 2);
  rect(ctx, archX, ghY + ghH - archH, archW, archH, vary(30, 4, rng), vary(22, 3, rng), vary(14, 3, rng));
  for (let px = archX + 1; px < archX + archW - 1; px += 3) {
    rect(ctx, px, ghY + ghH - archH + 1, 1, archH - 2, vary(75, 4, rng), vary(70, 3, rng), vary(65, 3, rng));
  }

  // Banner
  const flagX = keepX + Math.floor(keepW / 2);
  const flagY = keepY - Math.floor(h * 0.06);
  rect(ctx, flagX, flagY, 1, Math.floor(h * 0.06) + 4, vary(90, 4, rng), vary(70, 3, rng), vary(45, 3, rng));
  rect(ctx, flagX + 1, flagY, 6, 4, vary(50, 8, rng), vary(50, 8, rng), vary(180, 8, rng));
}

// ── MANOR — Noble estate with surrounding wall and garden ──
function _drawManor(ctx, ox, oy, rng, w, h) {
  const wallC = [175, 165, 145];
  const roofC = [125, 50, 35];

  // Stone wall around property
  const wallThick = 3;
  const mL = ox + 2, mR = ox + w - 2, mT = oy + Math.floor(h * 0.25), mB = oy + h - 2;
  for (let px = mL; px <= mR; px++) {
    for (let y = 0; y < wallThick; y++) {
      rect(ctx, px, mT + y, 1, 1, vary(150, 5, rng), vary(140, 4, rng), vary(120, 4, rng));
      rect(ctx, px, mB - y, 1, 1, vary(150, 5, rng), vary(140, 4, rng), vary(120, 4, rng));
    }
  }
  for (let py = mT; py <= mB; py++) {
    for (let x = 0; x < wallThick; x++) {
      rect(ctx, mL + x, py, 1, 1, vary(150, 5, rng), vary(140, 4, rng), vary(120, 4, rng));
      rect(ctx, mR - x, py, 1, 1, vary(140, 5, rng), vary(130, 4, rng), vary(110, 4, rng));
    }
  }

  // Garden inside wall
  for (let py = mT + wallThick; py < mB - wallThick; py++) {
    for (let px = mL + wallThick; px < mR - wallThick; px++) {
      rect(ctx, px, py, 1, 1, vary(75, 8, rng), vary(130, 8, rng), vary(55, 6, rng));
    }
  }

  // Main building
  const bldgW = Math.floor(w * 0.6);
  const bldgH = Math.floor(h * 0.5);
  const bldgX = ox + Math.floor((w - bldgW) / 2);
  const bldgY = mT + Math.floor((mB - mT - bldgH) / 2);
  const dims = _drawBasicBuilding(ctx, bldgX, bldgY, wallC, roofC, rng, bldgW, bldgH, {
    bricks: true, tiledRoof: true,
    wallH: Math.floor(bldgH * 0.55),
    windowGlow: [200, 190, 130],
  });

  // Wing
  const wingW = Math.floor(bldgW * 0.4);
  const wingH = Math.floor(bldgH * 0.7);
  _drawBasicBuilding(ctx, bldgX + bldgW + 2, bldgY + Math.floor(bldgH * 0.15), [165, 155, 135], roofC, rng, wingW, wingH, {
    bricks: true, tiledRoof: true, wallH: Math.floor(wingH * 0.5),
  });

  // Chimney
  const { wallX, wallY } = dims;
  rect(ctx, wallX + 3, wallY - Math.floor(h * 0.12), 3, Math.floor(h * 0.08), vary(120, 5, rng), vary(115, 4, rng), vary(110, 4, rng));
}

function _drawHouse(ctx, ox, oy, type, rng, w, h) {
  const variant = Math.floor(rng() * 4);
  const walls = [
    [180, 160, 130], // thatched cottage
    [160, 155, 150], // stone house
    [170, 140, 90],  // timber frame
    [190, 160, 120], // clay hut
  ];
  const roofs = [
    [140, 120, 60],  // thatch
    [100, 95, 90],   // slate
    [130, 55, 40],   // wood shingle
    [160, 130, 80],  // clay tile
  ];
  const wallC = walls[variant];
  const roofC = roofs[variant];
  const opts = {
    bricks: variant === 1,
    planks: variant === 2,
    tiledRoof: variant === 3,
  };

  const dims = _drawBasicBuilding(ctx, ox, oy, wallC, roofC, rng, w, h, opts);

  // Variant details
  if (variant === 2) {
    // timber frame cross beams
    const { wallX, wallY, wallW, wallH } = dims;
    for (let x = 0; x < wallW; x += Math.floor(wallW / 3)) {
      for (let y = 0; y < wallH; y++) {
        rect(ctx, wallX + x, wallY + y, 1, 1, vary(90, 8, rng), vary(60, 6, rng), vary(30, 6, rng));
      }
    }
  }
  if (variant === 0) {
    // flower box under windows
    if (w >= 20) {
      rect(ctx, dims.wallX + 1, dims.wallY + Math.floor(dims.wallH * 0.35), 3, 1,
        vary(60, 10, rng), vary(140, 12, rng), vary(40, 8, rng));
    }
  }
}

function _drawInn(ctx, ox, oy, rng, w, h) {
  const variant = Math.floor(rng() * 3);
  const walls = [[160, 130, 90], [140, 110, 75], [170, 140, 100]];
  const roofs = [[130, 55, 40], [110, 80, 50], [140, 60, 45]];
  const wallC = walls[variant];
  const roofC = roofs[variant];

  const dims = _drawBasicBuilding(ctx, ox, oy, wallC, roofC, rng, w, h, {
    bricks: variant === 0,
    planks: variant === 1,
    wallH: Math.max(10, Math.floor(h * 0.6)),
    windowGlow: [200, 180, 100],
  });

  // Chimney
  const { wallX, wallY, wallW } = dims;
  const chimX = wallX + wallW - 4;
  const chimY = wallY - Math.floor(h * 0.2);
  rect(ctx, chimX, chimY, 3, Math.floor(h * 0.15),
    vary(80, 8, rng), vary(70, 8, rng), vary(65, 8, rng));

  // Hanging sign
  if (w >= 24) {
    rect(ctx, wallX - 2, wallY + 2, 1, 5, vary(80, 6, rng), vary(60, 6, rng), vary(40, 6, rng));
    rect(ctx, wallX - 4, wallY + 6, 5, 3, vary(160, 10, rng), vary(130, 8, rng), vary(60, 8, rng));
  }

  // Second floor windows for variant 0
  if (variant === 0 && dims.wallH >= 14) {
    const winY2 = wallY + 2;
    rect(ctx, wallX + 3, winY2, 2, 2, vary(200, 10, rng), vary(180, 10, rng), vary(100, 8, rng));
    rect(ctx, wallX + wallW - 5, winY2, 2, 2, vary(200, 10, rng), vary(180, 10, rng), vary(100, 8, rng));
  }
}

function _drawForge(ctx, ox, oy, rng, w, h) {
  const variant = Math.floor(rng() * 2);
  const wallC = variant === 0 ? [100, 80, 65] : [120, 90, 60];
  const roofC = variant === 0 ? [70, 55, 50] : [80, 65, 45];

  const dims = _drawBasicBuilding(ctx, ox, oy, wallC, roofC, rng, w, h, {
    bricks: true,
    windowGlow: [240, 160, 60],
  });

  // Large chimney
  const { wallX, wallY, wallW } = dims;
  const chimW = Math.max(3, Math.floor(wallW * 0.2));
  rect(ctx, wallX + wallW - chimW - 1, wallY - Math.floor(h * 0.3), chimW, Math.floor(h * 0.25),
    vary(80, 8, rng), vary(70, 8, rng), vary(65, 8, rng));

  // Anvil
  if (variant === 0) {
    rect(ctx, ox + Math.floor(w * 0.15), oy + h - 4, 4, 2, vary(60, 6, rng), vary(60, 6, rng), vary(65, 6, rng));
    rect(ctx, ox + Math.floor(w * 0.15) + 1, oy + h - 6, 2, 2, vary(70, 6, rng), vary(70, 6, rng), vary(75, 6, rng));
  }

  // Orange glow from door
  const doorX = dims.wallX + Math.floor(dims.wallW / 2) - 1;
  px(ctx, doorX, oy + h - 4, 240, 140, 40, 180);
  px(ctx, doorX + 1, oy + h - 4, 240, 140, 40, 180);
}

function _drawMarket(ctx, ox, oy, rng, w, h) {
  // Draw multiple small stall-buildings using the same style as other buildings
  const stallPalettes = [
    { wall: [180, 155, 115], roof: [140, 100, 55] },  // warm wood
    { wall: [165, 150, 130], roof: [120, 85, 50] },   // weathered timber
    { wall: [175, 145, 100], roof: [130, 70, 40] },   // rustic brown
    { wall: [170, 160, 140], roof: [110, 95, 70] },   // stone grey
  ];

  // Determine stall grid
  const stallW = Math.min(28, Math.max(20, Math.floor(w / 3)));
  const stallH = Math.min(26, Math.max(18, Math.floor(h / 2)));
  const cols = Math.max(1, Math.floor((w - 2) / (stallW + 1)));
  const rows = Math.max(1, Math.floor((h - 2) / (stallH + 1)));
  const gapX = Math.floor((w - cols * stallW) / (cols + 1));
  const gapY = Math.floor((h - rows * stallH) / (rows + 1));

  // Draw each stall as a small building
  for (let r = 0; r < rows; r++) {
    for (let c = 0; c < cols; c++) {
      const sx = ox + gapX + c * (stallW + gapX);
      const sy = oy + gapY + r * (stallH + gapY);
      const pal = stallPalettes[Math.floor(rng() * stallPalettes.length)];

      // Use _drawBasicBuilding for each stall — same style as houses
      _drawBasicBuilding(ctx, sx, sy, pal.wall, pal.roof, rng, stallW, stallH, {
        planks: rng() > 0.5,
        doorColor: [90, 60, 30],
      });

      // Small crate or barrel next to some stalls
      if (rng() > 0.4) {
        const crateX = sx + stallW - 4;
        const crateY = sy + stallH - 5;
        rect(ctx, crateX, crateY, 3, 4,
          vary(120, 12, rng), vary(85, 10, rng), vary(45, 8, rng));
        // lid line
        rect(ctx, crateX, crateY, 3, 1,
          vary(100, 10, rng), vary(70, 8, rng), vary(35, 6, rng));
      }
    }
  }
}

function _drawShrine(ctx, ox, oy, rng, w, h) {
  const variant = Math.floor(rng() * 3);

  if (variant === 0) {
    // Stone altar
    const altW = Math.floor(w * 0.5);
    const altH = Math.floor(h * 0.35);
    const ax = ox + Math.floor((w - altW) / 2);
    const ay = oy + h - altH - 2;
    for (let y = 0; y < altH; y++) {
      for (let x = 0; x < altW; x++) {
        rect(ctx, ax + x, ay + y, 1, 1, vary(170, 10, rng), vary(165, 8, rng), vary(160, 8, rng));
      }
    }
    // Cross/symbol on top
    rect(ctx, ox + Math.floor(w / 2), ay - 5, 1, 5, vary(200, 8, rng), vary(180, 8, rng), vary(100, 8, rng));
    rect(ctx, ox + Math.floor(w / 2) - 2, ay - 3, 5, 1, vary(200, 8, rng), vary(180, 8, rng), vary(100, 8, rng));
  } else if (variant === 1) {
    // Small wooden chapel
    _drawBasicBuilding(ctx, ox, oy, [190, 180, 165], [150, 130, 200], rng, w, h, {});
    // Steeple
    rect(ctx, ox + Math.floor(w / 2) - 1, oy + Math.floor(h * 0.1), 2, Math.floor(h * 0.2),
      vary(160, 8, rng), vary(155, 8, rng), vary(150, 8, rng));
  } else {
    // Pillar circle
    const radius = Math.floor(Math.min(w, h) * 0.35);
    const cx = ox + Math.floor(w / 2);
    const cy = oy + Math.floor(h * 0.55);
    for (let a = 0; a < 360; a += 60) {
      const px2 = cx + Math.floor(Math.cos(a * Math.PI / 180) * radius);
      const py2 = cy + Math.floor(Math.sin(a * Math.PI / 180) * (radius * 0.7));
      rect(ctx, px2, py2 - 3, 2, 6, vary(160, 10, rng), vary(155, 8, rng), vary(150, 8, rng));
    }
    // Center glow
    px(ctx, cx, cy, vary(180, 10, rng), vary(160, 10, rng), vary(220, 10, rng), 200);
  }
}

function _drawMill(ctx, ox, oy, rng, w, h) {
  const variant = Math.floor(rng() * 2);
  const wallC = [175, 155, 120];
  const roofC = [130, 100, 60];

  if (variant === 0) {
    // Windmill — taller narrower tower with blades
    const towerW = Math.max(6, Math.floor(w * 0.4));
    const towerH = Math.floor(h * 0.75);
    const towerX = ox + Math.floor((w - towerW) / 2);
    const towerY = oy + h - towerH;

    // Tower body (tapers slightly toward top)
    for (let y = 0; y < towerH; y++) {
      const taper = Math.floor(y * 0.5 / towerH);
      const rowW = towerW - taper * 2;
      const rowX = towerX + taper;
      for (let x = 0; x < rowW; x++) {
        rect(ctx, rowX + x, towerY + y, 1, 1,
          vary(wallC[0], 6, rng), vary(wallC[1], 6, rng), vary(wallC[2], 6, rng));
      }
    }

    // Conical roof cap
    const capH = Math.floor(h * 0.15);
    for (let y = 0; y < capH; y++) {
      const capW = Math.max(1, Math.floor(towerW * (1 - y / capH)));
      const capX = towerX + Math.floor((towerW - capW) / 2);
      for (let x = 0; x < capW; x++) {
        rect(ctx, capX + x, towerY - capH + y, 1, 1,
          vary(roofC[0], 6, rng), vary(roofC[1], 6, rng), vary(roofC[2], 6, rng));
      }
    }

    // Door
    rect(ctx, towerX + Math.floor(towerW / 2) - 1, towerY + towerH - 4, 2, 4,
      vary(90, 8, rng), vary(60, 6, rng), vary(30, 6, rng));

    // Window
    rect(ctx, towerX + Math.floor(towerW / 2) - 1, towerY + Math.floor(towerH * 0.3), 2, 2,
      vary(160, 10, rng), vary(190, 8, rng), vary(220, 8, rng));

    // Blades — 4 rectangular blades from hub at roof center
    const hubX = towerX + Math.floor(towerW / 2);
    const hubY = towerY - Math.floor(capH * 0.5);
    const bladeLen = Math.max(4, Math.floor(Math.min(w, h) * 0.35));
    // Draw 4 blades as thick lines
    const bladeOffsets = [[-1, -1], [1, -1], [1, 1], [-1, 1]];
    for (const [dx, dy] of bladeOffsets) {
      for (let i = 1; i <= bladeLen; i++) {
        rect(ctx, hubX + dx * i, hubY + dy * i, 1, 1,
          vary(170, 8, rng), vary(155, 6, rng), vary(130, 6, rng));
        // Blade width (paddle part on outer half)
        if (i > bladeLen * 0.4) {
          rect(ctx, hubX + dx * i + dy, hubY + dy * i - dx, 1, 1,
            vary(180, 8, rng), vary(165, 6, rng), vary(140, 6, rng));
        }
      }
    }
    // Hub dot
    rect(ctx, hubX - 1, hubY - 1, 2, 2, vary(80, 6, rng), vary(70, 6, rng), vary(60, 6, rng));

  } else {
    // Watermill — building with large water wheel on the side
    const bldgW = Math.max(8, Math.floor(w * 0.55));
    const bldgH = Math.floor(h * 0.6);
    const bldgX = ox + 1;
    const bldgY = oy + h - bldgH;

    _drawBasicBuilding(ctx, bldgX, bldgY, wallC, roofC, rng, bldgW, bldgH, {
      bricks: true,
      wallH: Math.floor(bldgH * 0.65),
    });

    // Water channel underneath right side
    const chanY = oy + h - 3;
    for (let x = bldgX + bldgW - 2; x < ox + w; x++) {
      for (let y = chanY; y < oy + h; y++) {
        rect(ctx, x, y, 1, 1, vary(40, 8, rng), vary(80, 10, rng), vary(140, 10, rng));
      }
    }

    // Water wheel — drawn as a circle with spokes and paddles
    const wheelR = Math.max(4, Math.floor(Math.min(w, h) * 0.25));
    const wheelCx = bldgX + bldgW + 1;
    const wheelCy = bldgY + Math.floor(bldgH * 0.5);

    // Outer rim (filled circle outline)
    for (let a = 0; a < 360; a += 8) {
      const rad = a * Math.PI / 180;
      const wx = Math.round(Math.cos(rad) * wheelR);
      const wy = Math.round(Math.sin(rad) * wheelR);
      rect(ctx, wheelCx + wx, wheelCy + wy, 1, 1,
        vary(100, 6, rng), vary(75, 6, rng), vary(45, 6, rng));
    }

    // Spokes — 8 lines from center to rim
    for (let s = 0; s < 8; s++) {
      const rad = s * Math.PI / 4;
      for (let i = 1; i < wheelR; i++) {
        const sx = Math.round(Math.cos(rad) * i);
        const sy = Math.round(Math.sin(rad) * i);
        px(ctx, wheelCx + sx, wheelCy + sy,
          vary(110, 8, rng), vary(80, 6, rng), vary(50, 6, rng));
      }
    }

    // Paddles at rim — small rectangles at each spoke end
    for (let s = 0; s < 8; s++) {
      const rad = s * Math.PI / 4;
      const px2 = Math.round(Math.cos(rad) * wheelR);
      const py2 = Math.round(Math.sin(rad) * wheelR);
      rect(ctx, wheelCx + px2 - 1, wheelCy + py2, 2, 1,
        vary(90, 6, rng), vary(65, 6, rng), vary(35, 6, rng));
    }

    // Hub
    rect(ctx, wheelCx - 1, wheelCy - 1, 2, 2,
      vary(70, 6, rng), vary(55, 6, rng), vary(40, 6, rng));
  }
}

function _drawLibrary(ctx, ox, oy, rng, w, h) {
  const variant = Math.floor(rng() * 2);
  const wallC = variant === 0 ? [160, 150, 140] : [150, 140, 130];
  const roofC = variant === 0 ? [60, 80, 130] : [70, 70, 120];

  const dims = _drawBasicBuilding(ctx, ox, oy, wallC, roofC, rng, w, h, {
    bricks: true,
    tiledRoof: true,
    wallH: variant === 0 ? Math.floor(h * 0.65) : Math.floor(h * 0.5),
  });

  // Window grid
  const { wallX, wallY, wallW, wallH } = dims;
  const winCols = Math.max(2, Math.floor(wallW / 6));
  for (let i = 0; i < winCols; i++) {
    const wx = wallX + 2 + Math.floor(i * (wallW - 4) / winCols);
    rect(ctx, wx, wallY + 2, 2, 3, vary(180, 10, rng), vary(200, 8, rng), vary(220, 8, rng));
  }
}

function _drawWorkshop(ctx, ox, oy, rng, w, h) {
  const wallC = [165, 140, 100];
  const roofC = [120, 90, 55];

  _drawBasicBuilding(ctx, ox, oy, wallC, roofC, rng, w, h, {
    planks: true,
  });

  // Wood pile beside
  for (let i = 0; i < 3; i++) {
    rect(ctx, ox + 2, oy + h - 3 - i * 2, 3, 2, vary(120, 10, rng), vary(80, 8, rng), vary(40, 8, rng));
  }
}

function _drawWell(ctx, ox, oy, rng, w, h) {
  const variant = Math.floor(rng() * 2);
  const cx = ox + Math.floor(w / 2);
  const cy = oy + Math.floor(h * 0.55);

  if (variant === 0) {
    // Stone circle well
    const radius = Math.floor(Math.min(w, h) * 0.2);
    for (let a = 0; a < 360; a += 15) {
      const wx = cx + Math.floor(Math.cos(a * Math.PI / 180) * radius);
      const wy = cy + Math.floor(Math.sin(a * Math.PI / 180) * (radius * 0.7));
      rect(ctx, wx, wy, 2, 2, vary(140, 10, rng), vary(140, 8, rng), vary(145, 8, rng));
    }
    // Crossbar
    rect(ctx, cx - radius, cy - radius - 1, radius * 2, 2, vary(100, 8, rng), vary(100, 8, rng), vary(110, 8, rng));
    // Water
    px(ctx, cx - 1, cy, 40, 90, 150);
    px(ctx, cx, cy, 40, 90, 150);
    px(ctx, cx + 1, cy + 1, 35, 80, 140);
  } else {
    // Wooden bucket well
    const radius = Math.floor(Math.min(w, h) * 0.18);
    for (let a = 0; a < 360; a += 20) {
      const wx = cx + Math.floor(Math.cos(a * Math.PI / 180) * radius);
      const wy = cy + Math.floor(Math.sin(a * Math.PI / 180) * (radius * 0.7));
      rect(ctx, wx, wy, 2, 2, vary(120, 10, rng), vary(80, 8, rng), vary(40, 8, rng));
    }
    // Posts
    rect(ctx, cx - radius - 1, cy - radius - 3, 2, radius + 3, vary(100, 8, rng), vary(70, 6, rng), vary(40, 6, rng));
    rect(ctx, cx + radius, cy - radius - 3, 2, radius + 3, vary(100, 8, rng), vary(70, 6, rng), vary(40, 6, rng));
    // Rope
    rect(ctx, cx, cy - radius - 2, 1, radius, vary(160, 8, rng), vary(140, 8, rng), vary(100, 8, rng));
    // Water
    px(ctx, cx, cy, 40, 90, 150);
  }
}

function _drawRuin(ctx, ox, oy, rng, w, h) {
  // Scattered stone blocks
  const numBlocks = Math.max(4, Math.floor((w * h) / 80));
  for (let i = 0; i < numBlocks; i++) {
    const bx = ox + Math.floor(rng() * (w - 6)) + 2;
    const by = oy + Math.floor(h * 0.3) + Math.floor(rng() * (h * 0.6));
    const bw = 3 + Math.floor(rng() * Math.min(6, w * 0.2));
    const bh = 3 + Math.floor(rng() * Math.min(8, h * 0.25));
    rect(ctx, bx, by, bw, bh, vary(110, 20, rng), vary(105, 18, rng), vary(100, 15, rng));
  }
  // Vine/moss details
  for (let i = 0; i < 3; i++) {
    const vx = ox + Math.floor(rng() * w);
    const vy = oy + Math.floor(rng() * h);
    px(ctx, vx, vy, vary(40, 10, rng), vary(100, 15, rng), vary(30, 10, rng));
  }
}

function _drawFarm(ctx, ox, oy, rng, w, h) {
  const variant = Math.floor(rng() * 3);

  // Dirt ground base
  for (let y = 0; y < h; y++) {
    for (let x = 0; x < w; x++) {
      rect(ctx, ox + x, oy + y, 1, 1,
        vary(140, 8, rng), vary(115, 8, rng), vary(70, 6, rng));
    }
  }

  // Fence on all 4 sides
  const fenceC = () => [vary(120, 8, rng), vary(90, 6, rng), vary(50, 6, rng)];
  // Top & bottom horizontal rails
  for (let x = 0; x < w; x += 3) {
    // Vertical post
    const c1 = fenceC();
    rect(ctx, ox + x, oy, 1, 2, c1[0], c1[1], c1[2]);
    const c2 = fenceC();
    rect(ctx, ox + x, oy + h - 2, 1, 2, c2[0], c2[1], c2[2]);
    // Horizontal rail
    if (x + 2 < w) {
      const c3 = fenceC();
      rect(ctx, ox + x, oy + 1, 3, 1, c3[0], c3[1], c3[2]);
      const c4 = fenceC();
      rect(ctx, ox + x, oy + h - 1, 3, 1, c4[0], c4[1], c4[2]);
    }
  }
  // Left & right vertical rails
  for (let y = 0; y < h; y += 3) {
    const c1 = fenceC();
    rect(ctx, ox, oy + y, 2, 1, c1[0], c1[1], c1[2]);
    const c2 = fenceC();
    rect(ctx, ox + w - 2, oy + y, 2, 1, c2[0], c2[1], c2[2]);
    if (y + 2 < h) {
      const c3 = fenceC();
      rect(ctx, ox, oy + y, 1, 3, c3[0], c3[1], c3[2]);
      const c4 = fenceC();
      rect(ctx, ox + w - 1, oy + y, 1, 3, c4[0], c4[1], c4[2]);
    }
  }

  // Small barn in top-left corner (inside fence)
  const barnW = Math.min(Math.floor(w * 0.25), 14);
  const barnH = Math.min(Math.floor(h * 0.35), 12);
  _drawBasicBuilding(ctx, ox + 3, oy + 3, [155, 80, 60], [120, 95, 55], rng, barnW, barnH, {
    planks: true,
  });

  // Field area (to the right of barn)
  const fieldX = ox + barnW + 5;
  const fieldY = oy + 3;
  const fieldW = w - barnW - 8;
  const fieldH = h - 6;

  if (variant === 0) {
    // Crop rows (green) with dirt furrows
    for (let row = 0; row < fieldH; row += 3) {
      for (let col = 0; col < fieldW; col++) {
        // Dirt furrow
        rect(ctx, fieldX + col, fieldY + row, 1, 1,
          vary(110, 8, rng), vary(80, 6, rng), vary(45, 6, rng));
        // Green crop tops
        if (row + 1 < fieldH) {
          rect(ctx, fieldX + col, fieldY + row + 1, 1, 1,
            vary(50, 12, rng), vary(130, 15, rng), vary(40, 8, rng));
        }
        // Second green row
        if (row + 2 < fieldH) {
          rect(ctx, fieldX + col, fieldY + row + 2, 1, 1,
            vary(40, 10, rng), vary(110, 12, rng), vary(35, 8, rng));
        }
      }
    }
  } else if (variant === 1) {
    // Wheat field (golden) with wave pattern
    for (let y = 0; y < fieldH; y++) {
      for (let x = 0; x < fieldW; x++) {
        const wave = Math.sin((x + y * 0.3) * 0.8) * 10;
        rect(ctx, fieldX + x, fieldY + y, 1, 1,
          vary(200 + wave, 12, rng), vary(170 + wave, 10, rng), vary(60, 8, rng));
      }
    }
  } else {
    // Orchard (grid of small trees)
    for (let ty = 2; ty < fieldH - 2; ty += 5) {
      for (let tx = 2; tx < fieldW - 2; tx += 5) {
        // Tree trunk
        rect(ctx, fieldX + tx + 1, fieldY + ty + 3, 1, 2,
          vary(100, 8, rng), vary(70, 6, rng), vary(40, 6, rng));
        // Tree canopy
        rect(ctx, fieldX + tx, fieldY + ty, 3, 3,
          vary(40, 10, rng), vary(100, 15, rng), vary(30, 8, rng));
      }
    }
  }

  // Scarecrow (in the field)
  const scX = fieldX + Math.floor(fieldW * 0.6);
  const scY = fieldY + Math.floor(fieldH * 0.3);
  // Pole
  rect(ctx, scX, scY, 1, 6, vary(100, 8, rng), vary(75, 6, rng), vary(40, 6, rng));
  // Crossbar (arms)
  rect(ctx, scX - 2, scY + 1, 5, 1, vary(100, 8, rng), vary(75, 6, rng), vary(40, 6, rng));
  // Head
  rect(ctx, scX - 1, scY - 1, 2, 2, vary(190, 10, rng), vary(160, 8, rng), vary(100, 8, rng));
  // Hat
  rect(ctx, scX - 2, scY - 2, 4, 1, vary(80, 8, rng), vary(55, 6, rng), vary(30, 6, rng));

  // Hay bales near barn
  const hayX = ox + 3;
  const hayY = oy + barnH + 5;
  for (let i = 0; i < 2; i++) {
    rect(ctx, hayX + i * 5, hayY, 4, 3,
      vary(190, 10, rng), vary(170, 8, rng), vary(80, 8, rng));
    // Hay rope/band
    rect(ctx, hayX + i * 5, hayY + 1, 4, 1,
      vary(160, 8, rng), vary(140, 6, rng), vary(60, 6, rng));
  }
}

function _drawMine(ctx, ox, oy, rng, w, h) {
  const variant = Math.floor(rng() * 2);

  if (variant === 0) {
    // Cave entrance in rock face
    // Rock wall
    const rockH = Math.floor(h * 0.7);
    const rockY = oy + Math.floor(h * 0.15);
    for (let y = 0; y < rockH; y++) {
      const rowW = w - Math.floor(y * 0.5);
      for (let x = 0; x < rowW; x++) {
        rect(ctx, ox + x, rockY + y, 1, 1, vary(90, 15, rng), vary(85, 12, rng), vary(80, 10, rng));
      }
    }
    // Cave opening
    const caveW = Math.floor(w * 0.3);
    const caveH = Math.floor(h * 0.35);
    const caveX = ox + Math.floor((w - caveW) / 2);
    const caveY = rockY + rockH - caveH;
    for (let y = 0; y < caveH; y++) {
      const rowW = caveW - Math.floor(y * caveW / caveH / 2);
      const rx = caveX + Math.floor((caveW - rowW) / 2);
      for (let x = 0; x < rowW; x++) {
        rect(ctx, rx + x, caveY + y, 1, 1, vary(30, 8, rng), vary(25, 6, rng), vary(20, 6, rng));
      }
    }
    // Support beams
    rect(ctx, caveX, caveY, 2, caveH, vary(100, 8, rng), vary(70, 6, rng), vary(40, 6, rng));
    rect(ctx, caveX + caveW - 2, caveY, 2, caveH, vary(100, 8, rng), vary(70, 6, rng), vary(40, 6, rng));
    rect(ctx, caveX, caveY, caveW, 2, vary(100, 8, rng), vary(70, 6, rng), vary(40, 6, rng));
  } else {
    // Open pit with scaffolding
    // Pit
    const pitW = Math.floor(w * 0.6);
    const pitH = Math.floor(h * 0.4);
    const pitX = ox + Math.floor((w - pitW) / 2);
    const pitY = oy + Math.floor(h * 0.45);
    for (let y = 0; y < pitH; y++) {
      for (let x = 0; x < pitW; x++) {
        const depth = y / pitH;
        rect(ctx, pitX + x, pitY + y, 1, 1,
          vary(Math.floor(70 - depth * 30), 10, rng),
          vary(Math.floor(65 - depth * 25), 8, rng),
          vary(Math.floor(60 - depth * 20), 8, rng));
      }
    }
    // Scaffolding
    rect(ctx, pitX - 1, pitY - 6, 2, 8, vary(100, 8, rng), vary(70, 6, rng), vary(40, 6, rng));
    rect(ctx, pitX + pitW - 1, pitY - 6, 2, 8, vary(100, 8, rng), vary(70, 6, rng), vary(40, 6, rng));
    rect(ctx, pitX - 1, pitY - 6, pitW + 2, 1, vary(100, 8, rng), vary(70, 6, rng), vary(40, 6, rng));
    // Cart
    rect(ctx, ox + 2, oy + h - 5, 5, 3, vary(120, 8, rng), vary(90, 6, rng), vary(50, 6, rng));
    px(ctx, ox + 2, oy + h - 2, vary(60, 6, rng), vary(60, 6, rng), vary(65, 6, rng));
    px(ctx, ox + 6, oy + h - 2, vary(60, 6, rng), vary(60, 6, rng), vary(65, 6, rng));
  }
}

function _drawForestClearing(ctx, ox, oy, rng, w, h) {
  // Forest floor base — dark mossy ground
  for (let y = 0; y < h; y++) {
    for (let x = 0; x < w; x++) {
      rect(ctx, ox + x, oy + y, 1, 1,
        vary(45, 8, rng), vary(75, 10, rng), vary(30, 8, rng));
    }
  }

  // Small pond/stream (30% chance)
  const hasPond = rng() < 0.3;
  if (hasPond) {
    const pondCx = ox + Math.floor(w * 0.3 + rng() * w * 0.4);
    const pondCy = oy + Math.floor(h * 0.4 + rng() * h * 0.3);
    const pondRx = Math.max(3, Math.floor(w * 0.1 + rng() * w * 0.08));
    const pondRy = Math.max(2, Math.floor(h * 0.08 + rng() * h * 0.06));
    for (let py = -pondRy - 1; py <= pondRy + 1; py++) {
      for (let px2 = -pondRx - 1; px2 <= pondRx + 1; px2++) {
        const dist = (px2 * px2) / ((pondRx + 1) * (pondRx + 1)) + (py * py) / ((pondRy + 1) * (pondRy + 1));
        const gx = pondCx + px2, gy = pondCy + py;
        if (dist <= 1 && gx >= ox && gx < ox + w && gy >= oy && gy < oy + h) {
          if (dist > 0.7) {
            // Muddy shore
            rect(ctx, gx, gy, 1, 1, vary(80, 8, rng), vary(70, 8, rng), vary(40, 6, rng));
          } else {
            // Water
            const shimmer = Math.sin((px2 + py * 0.5) * 1.5) * 8;
            rect(ctx, gx, gy, 1, 1,
              vary(35 + shimmer, 6, rng), vary(70 + shimmer, 8, rng), vary(120 + shimmer, 10, rng));
          }
        }
      }
    }
  }

  // Undergrowth — scattered small bushes and ferns
  const bushCount = Math.max(3, Math.floor(w * h / 200));
  for (let i = 0; i < bushCount; i++) {
    const bx = ox + 2 + Math.floor(rng() * (w - 4));
    const by = oy + 2 + Math.floor(rng() * (h - 4));
    const bw = 2 + Math.floor(rng() * 2);
    const bh = 1 + Math.floor(rng() * 2);
    for (let y = 0; y < bh; y++) {
      for (let x = 0; x < bw; x++) {
        const gx = bx + x, gy = by + y;
        if (gx < ox + w && gy < oy + h) {
          rect(ctx, gx, gy, 1, 1, vary(35, 10, rng), vary(95, 15, rng), vary(25, 8, rng));
        }
      }
    }
  }

  // Dense trees — sorted back-to-front for overlap
  const treeCount = Math.max(6, Math.floor(w * h / 80));
  const trees = [];
  for (let i = 0; i < treeCount; i++) {
    trees.push({
      x: ox + 3 + Math.floor(rng() * (w - 6)),
      y: oy + 3 + Math.floor(rng() * (h - 6)),
      r: 3 + Math.floor(rng() * 3),
      shade: Math.floor(rng() * 30),
    });
  }
  trees.sort((a, b) => a.y - b.y);

  for (const t of trees) {
    // Trunk
    const trunkH = t.r + 1;
    if (t.y + t.r + trunkH < oy + h) {
      rect(ctx, t.x, t.y + t.r, 2, trunkH,
        vary(85, 10, rng), vary(55, 8, rng), vary(25, 6, rng));
    }
    // Canopy — filled circle with shading (top lighter, bottom darker)
    for (let cy = -t.r; cy <= t.r; cy++) {
      for (let cx = -t.r; cx <= t.r; cx++) {
        if (cx * cx + cy * cy > t.r * t.r) continue;
        const gx = t.x + cx, gy = t.y + cy;
        if (gx >= ox && gx < ox + w && gy >= oy && gy < oy + h) {
          // Top of canopy lighter, bottom darker
          const shade = cy * 4;
          rect(ctx, gx, gy, 1, 1,
            vary(30 - t.shade * 0.3, 10, rng),
            vary(80 + shade - t.shade * 0.5, 12, rng),
            vary(20, 8, rng));
        }
      }
    }
  }

  // Fallen logs (1-2)
  const logCount = 1 + Math.floor(rng() * 2);
  for (let i = 0; i < logCount; i++) {
    const lx = ox + 3 + Math.floor(rng() * (w - 10));
    const ly = oy + Math.floor(h * 0.6 + rng() * h * 0.3);
    const ll = 3 + Math.floor(rng() * 4);
    if (ly < oy + h - 1) {
      rect(ctx, lx, ly, ll, 1, vary(80, 8, rng), vary(55, 6, rng), vary(30, 6, rng));
      rect(ctx, lx, ly + 1, ll, 1, vary(65, 8, rng), vary(45, 6, rng), vary(25, 6, rng));
    }
  }

  // Mushrooms (tiny colored dots)
  const mushCount = 2 + Math.floor(rng() * 3);
  for (let i = 0; i < mushCount; i++) {
    const mx = ox + 2 + Math.floor(rng() * (w - 4));
    const my = oy + 2 + Math.floor(rng() * (h - 4));
    if (mx < ox + w && my < oy + h - 1) {
      px(ctx, mx, my, vary(180, 20, rng), vary(50, 20, rng), vary(40, 15, rng));
      px(ctx, mx, my + 1, vary(200, 10, rng), vary(190, 10, rng), vary(170, 10, rng));
    }
  }
}

function _drawDock(ctx, ox, oy, rng, w, h, buildingData) {
  // Determine which edge the dock is nearest to — pier extends toward that edge (toward water)
  // Edges: 'right', 'left', 'bottom', 'top'
  let waterDir = 'bottom'; // default
  if (buildingData && buildingData._worldW && buildingData._worldH) {
    const bx = buildingData.tx || 0;
    const by = buildingData.tz || 0;
    const gw = buildingData.gridW || 1;
    const gh = buildingData.gridH || 1;
    const worldW = buildingData._worldW;
    const worldH = buildingData._worldH;
    const distLeft = bx;
    const distRight = worldW - (bx + gw);
    const distTop = by;
    const distBottom = worldH - (by + gh);
    const minDist = Math.min(distLeft, distRight, distTop, distBottom);
    if (minDist === distBottom) waterDir = 'bottom';
    else if (minDist === distTop) waterDir = 'top';
    else if (minDist === distRight) waterDir = 'right';
    else waterDir = 'left';
  }

  const isHoriz = (waterDir === 'left' || waterDir === 'right');

  // Land/sand half and water half
  const halfW = isHoriz ? Math.floor(w / 2) : w;
  const halfH = isHoriz ? h : Math.floor(h / 2);

  // Draw sandy ground on the land side
  for (let y = 0; y < h; y++) {
    for (let x = 0; x < w; x++) {
      let isLand = false;
      if (waterDir === 'bottom') isLand = y < halfH;
      else if (waterDir === 'top') isLand = y >= h - halfH;
      else if (waterDir === 'right') isLand = x < halfW;
      else isLand = x >= w - halfW;

      if (isLand) {
        // Sandy beach ground
        rect(ctx, ox + x, oy + y, 1, 1,
          vary(194, 8, rng), vary(178, 8, rng), vary(128, 8, rng));
      } else {
        // Water with wave pattern
        const wave = Math.sin((x + y * 0.5) * 0.8) * 8;
        rect(ctx, ox + x, oy + y, 1, 1,
          vary(35 + wave, 6, rng), vary(75 + wave, 8, rng), vary(135 + wave, 8, rng));
      }
    }
  }

  // Draw the pier extending from land into water
  let pierX, pierY, pierW, pierH;
  if (waterDir === 'bottom') {
    pierW = Math.max(4, Math.floor(w * 0.25));
    pierH = Math.max(8, Math.floor(h * 0.7));
    pierX = ox + Math.floor((w - pierW) / 2);
    pierY = oy + Math.floor(h * 0.15);
  } else if (waterDir === 'top') {
    pierW = Math.max(4, Math.floor(w * 0.25));
    pierH = Math.max(8, Math.floor(h * 0.7));
    pierX = ox + Math.floor((w - pierW) / 2);
    pierY = oy + Math.floor(h * 0.15);
  } else if (waterDir === 'right') {
    pierW = Math.max(8, Math.floor(w * 0.7));
    pierH = Math.max(4, Math.floor(h * 0.25));
    pierX = ox + Math.floor(w * 0.15);
    pierY = oy + Math.floor((h - pierH) / 2);
  } else {
    pierW = Math.max(8, Math.floor(w * 0.7));
    pierH = Math.max(4, Math.floor(h * 0.25));
    pierX = ox + Math.floor(w * 0.15);
    pierY = oy + Math.floor((h - pierH) / 2);
  }

  // Pier planks — gaps perpendicular to pier direction
  for (let py = 0; py < pierH; py++) {
    for (let px2 = 0; px2 < pierW; px2++) {
      const alongPier = isHoriz ? px2 : py;
      const acrossPier = isHoriz ? py : px2;
      const gap = (alongPier % 5 === 0) ? -15 : 0;
      const edgeGap = (acrossPier === 0 || acrossPier === (isHoriz ? pierH : pierW) - 1) ? -10 : 0;
      rect(ctx, pierX + px2, pierY + py, 1, 1,
        vary(150 + gap + edgeGap, 6, rng), vary(120 + gap + edgeGap, 6, rng), vary(75 + gap + edgeGap, 6, rng));
    }
  }

  // Support posts along the pier (on the water side)
  if (isHoriz) {
    const postSpacing = Math.max(3, Math.floor(pierW / 4));
    for (let i = 0; i < 4; i++) {
      const postX = pierX + 1 + i * postSpacing;
      if (postX >= pierX + pierW) break;
      rect(ctx, postX, pierY + pierH, 1, 3,
        vary(80, 6, rng), vary(55, 6, rng), vary(30, 6, rng));
      rect(ctx, postX, pierY - 1, 1, 1,
        vary(90, 6, rng), vary(65, 6, rng), vary(35, 6, rng));
    }
  } else {
    const postSpacing = Math.max(3, Math.floor(pierH / 4));
    for (let i = 0; i < 4; i++) {
      const postY = pierY + 1 + i * postSpacing;
      if (postY >= pierY + pierH) break;
      rect(ctx, pierX + pierW, postY, 3, 1,
        vary(80, 6, rng), vary(55, 6, rng), vary(30, 6, rng));
      rect(ctx, pierX - 1, postY, 1, 1,
        vary(90, 6, rng), vary(65, 6, rng), vary(35, 6, rng));
    }
  }

  // Mooring posts at pier ends
  if (isHoriz) {
    const moor1X = pierX + 2;
    const moor2X = pierX + pierW - 3;
    for (const mx of [moor1X, moor2X]) {
      rect(ctx, mx, pierY - 3, 2, 3, vary(85, 6, rng), vary(60, 6, rng), vary(35, 6, rng));
    }
  } else {
    const moor1Y = pierY + 2;
    const moor2Y = pierY + pierH - 3;
    for (const my of [moor1Y, moor2Y]) {
      rect(ctx, pierX - 3, my, 3, 2, vary(85, 6, rng), vary(60, 6, rng), vary(35, 6, rng));
    }
  }

  // Small boat beside pier in the water
  if (isHoriz) {
    const boatX = pierX + Math.floor(pierW * 0.5);
    const boatY = pierY + pierH + 3;
    const boatW = Math.min(8, Math.floor(w * 0.2));
    const boatH = 3;
    for (let by = 0; by < boatH; by++) {
      const taper = by === 0 ? 1 : 0;
      for (let bx = taper; bx < boatW - taper; bx++) {
        rect(ctx, boatX + bx, boatY + by, 1, 1,
          vary(110, 8, rng), vary(75, 6, rng), vary(40, 6, rng));
      }
    }
  } else {
    const boatX = pierX + pierW + 3;
    const boatY = pierY + Math.floor(pierH * 0.5);
    const boatW = 3;
    const boatH = Math.min(8, Math.floor(h * 0.2));
    for (let bx = 0; bx < boatW; bx++) {
      const taper = bx === 0 ? 1 : 0;
      for (let by = taper; by < boatH - taper; by++) {
        rect(ctx, boatX + bx, boatY + by, 1, 1,
          vary(110, 8, rng), vary(75, 6, rng), vary(40, 6, rng));
      }
    }
  }

  // Crates and barrels on the land side of the pier
  const crateR = vary(130, 8, rng), crateG = vary(95, 6, rng), crateB = vary(55, 6, rng);
  if (waterDir === 'bottom') {
    rect(ctx, pierX - 4, pierY + 1, 3, 3, crateR, crateG, crateB);
    rect(ctx, pierX + pierW + 1, pierY + 2, 2, 2, vary(100, 8, rng), vary(70, 6, rng), vary(40, 6, rng));
  } else if (waterDir === 'right') {
    rect(ctx, pierX + 1, pierY - 4, 3, 3, crateR, crateG, crateB);
    rect(ctx, pierX + 2, pierY + pierH + 1, 2, 2, vary(100, 8, rng), vary(70, 6, rng), vary(40, 6, rng));
  } else if (waterDir === 'top') {
    rect(ctx, pierX - 4, pierY + pierH - 4, 3, 3, crateR, crateG, crateB);
    rect(ctx, pierX + pierW + 1, pierY + pierH - 3, 2, 2, vary(100, 8, rng), vary(70, 6, rng), vary(40, 6, rng));
  } else {
    rect(ctx, pierX + pierW - 4, pierY - 4, 3, 3, crateR, crateG, crateB);
    rect(ctx, pierX + pierW - 3, pierY + pierH + 1, 2, 2, vary(100, 8, rng), vary(70, 6, rng), vary(40, 6, rng));
  }
}

function _drawGarden(ctx, ox, oy, rng, w, h) {
  const variant = Math.floor(rng() * 3);

  // Ground base
  for (let y = 0; y < h; y++) {
    for (let x = 0; x < w; x++) {
      rect(ctx, ox + x, oy + y, 1, 1, vary(80, 8, rng), vary(120, 10, rng), vary(50, 8, rng));
    }
  }

  if (variant === 0) {
    // Flower garden with rows of different colored flowers
    const flowerColors = [[200, 60, 80], [220, 180, 50], [160, 80, 200], [220, 120, 60]];
    for (let row = 2; row < h - 2; row += 3) {
      const fc = flowerColors[row % flowerColors.length];
      for (let col = 2; col < w - 2; col += 2) {
        px(ctx, ox + col, oy + row, vary(fc[0], 15, rng), vary(fc[1], 12, rng), vary(fc[2], 10, rng));
        px(ctx, ox + col, oy + row + 1, vary(40, 8, rng), vary(100, 10, rng), vary(30, 8, rng));
      }
    }
  } else if (variant === 1) {
    // Herb garden with small bushy patches
    for (let py = 3; py < h - 3; py += 4) {
      for (let px2 = 3; px2 < w - 3; px2 += 4) {
        const bushG = vary(90, 20, rng);
        rect(ctx, ox + px2, oy + py, 3, 2, vary(30, 8, rng), bushG, vary(25, 8, rng));
        rect(ctx, ox + px2 + 1, oy + py - 1, 1, 1, vary(35, 8, rng), vary(bushG + 10, 8, rng), vary(20, 8, rng));
      }
    }
  } else {
    // Zen/stone garden with gravel paths and shrubs
    for (let y = 0; y < h; y++) {
      for (let x = 0; x < w; x++) {
        if ((x + y) % 3 === 0) {
          rect(ctx, ox + x, oy + y, 1, 1, vary(160, 10, rng), vary(155, 8, rng), vary(140, 8, rng));
        }
      }
    }
    // Shrubs at corners
    rect(ctx, ox + 2, oy + 2, 3, 3, vary(40, 10, rng), vary(110, 12, rng), vary(35, 8, rng));
    rect(ctx, ox + w - 5, oy + 2, 3, 3, vary(40, 10, rng), vary(110, 12, rng), vary(35, 8, rng));
    rect(ctx, ox + Math.floor(w / 2) - 1, oy + h - 5, 3, 3, vary(40, 10, rng), vary(110, 12, rng), vary(35, 8, rng));
  }

  // Stone border
  for (let x = 0; x < w; x++) {
    rect(ctx, ox + x, oy, 1, 1, vary(140, 8, rng), vary(135, 8, rng), vary(130, 8, rng));
    rect(ctx, ox + x, oy + h - 1, 1, 1, vary(140, 8, rng), vary(135, 8, rng), vary(130, 8, rng));
  }
  for (let y = 0; y < h; y++) {
    rect(ctx, ox, oy + y, 1, 1, vary(140, 8, rng), vary(135, 8, rng), vary(130, 8, rng));
    rect(ctx, ox + w - 1, oy + y, 1, 1, vary(140, 8, rng), vary(135, 8, rng), vary(130, 8, rng));
  }
}

function _drawWarehouse(ctx, ox, oy, rng, w, h) {
  const variant = Math.floor(rng() * 2);

  if (variant === 0) {
    // Long barn style
    const wallC = [150, 120, 80];
    const roofC = [110, 85, 50];
    _drawBasicBuilding(ctx, ox, oy, wallC, roofC, rng, w, h, {
      planks: true,
      wallH: Math.floor(h * 0.6),
      wallW: Math.floor(w * 0.85),
    });
    // Double doors
    const doorW = Math.max(5, Math.floor(w * 0.25));
    const doorH = Math.max(6, Math.floor(h * 0.35));
    const doorX = ox + Math.floor((w - doorW) / 2);
    const doorY = oy + h - doorH - Math.floor(h * 0.08);
    rect(ctx, doorX, doorY, doorW, doorH, vary(80, 10, rng), vary(55, 8, rng), vary(30, 8, rng));
    rect(ctx, doorX + Math.floor(doorW / 2), doorY, 1, doorH, vary(60, 6, rng), vary(40, 6, rng), vary(20, 6, rng));
  } else {
    // Stacked crate depot
    const baseH = Math.floor(h * 0.3);
    // Platform/floor
    rect(ctx, ox + 2, oy + h - baseH, w - 4, baseH, vary(130, 8, rng), vary(100, 8, rng), vary(60, 8, rng));
    // Stacked crates
    const crateColors = [[140, 100, 50], [120, 80, 40], [160, 120, 60]];
    for (let row = 0; row < 3; row++) {
      for (let col = 0; col < Math.floor(w / 8); col++) {
        const cc = crateColors[(row + col) % crateColors.length];
        const cx = ox + 3 + col * 7;
        const cy = oy + h - baseH - 5 - row * 5;
        if (cy < oy + 2) continue;
        rect(ctx, cx, cy, 5, 5, vary(cc[0], 10, rng), vary(cc[1], 8, rng), vary(cc[2], 8, rng));
        // Cross marking on crate
        rect(ctx, cx + 2, cy + 1, 1, 3, vary(cc[0] - 20, 6, rng), vary(cc[1] - 15, 6, rng), vary(cc[2] - 10, 6, rng));
        rect(ctx, cx + 1, cy + 2, 3, 1, vary(cc[0] - 20, 6, rng), vary(cc[1] - 15, 6, rng), vary(cc[2] - 10, 6, rng));
      }
    }
    // Simple roof overhang
    for (let x = 0; x < w; x++) {
      rect(ctx, ox + x, oy + 2, 1, 2, vary(100, 8, rng), vary(70, 6, rng), vary(40, 6, rng));
    }
  }
}

function _drawBarracks(ctx, ox, oy, rng, w, h) {
  const variant = Math.floor(rng() * 2);

  if (variant === 0) {
    // Stone fort
    const wallC = [130, 125, 120];
    const roofC = [90, 85, 80];
    const dims = _drawBasicBuilding(ctx, ox, oy, wallC, roofC, rng, w, h, {
      bricks: true,
      flatRoof: true,
      wallH: Math.floor(h * 0.65),
    });
    // Battlements (crenellations)
    const { wallX, wallY, wallW } = dims;
    for (let x = 0; x < wallW; x += 3) {
      rect(ctx, wallX + x, wallY - 3, 2, 3, vary(130, 8, rng), vary(125, 8, rng), vary(120, 8, rng));
    }
    // Banner/flag
    rect(ctx, wallX + Math.floor(wallW / 2), wallY - 8, 1, 5, vary(100, 8, rng), vary(70, 6, rng), vary(40, 6, rng));
    rect(ctx, wallX + Math.floor(wallW / 2) + 1, wallY - 8, 3, 2, vary(180, 10, rng), vary(40, 10, rng), vary(40, 10, rng));
  } else {
    // Wooden palisade
    const palisadeH = Math.floor(h * 0.55);
    const palisadeY = oy + h - palisadeH - 2;
    // Vertical logs forming wall
    for (let x = 0; x < w - 2; x += 2) {
      const logH = palisadeH + Math.floor(rng() * 3) - 1;
      rect(ctx, ox + x + 1, palisadeY + palisadeH - logH, 2, logH,
        vary(100, 10, rng), vary(70, 8, rng), vary(40, 8, rng));
      // Pointed top
      px(ctx, ox + x + 1, palisadeY + palisadeH - logH - 1, vary(90, 8, rng), vary(65, 6, rng), vary(35, 6, rng));
      px(ctx, ox + x + 2, palisadeY + palisadeH - logH - 1, vary(90, 8, rng), vary(65, 6, rng), vary(35, 6, rng));
    }
    // Gate opening
    const gateW = Math.max(4, Math.floor(w * 0.2));
    const gateX = ox + Math.floor((w - gateW) / 2);
    const gateH = Math.floor(palisadeH * 0.6);
    rect(ctx, gateX, palisadeY + palisadeH - gateH, gateW, gateH,
      vary(50, 8, rng), vary(35, 6, rng), vary(20, 6, rng));
    // Watchtower post
    rect(ctx, ox + 1, palisadeY - 4, 3, 4, vary(95, 8, rng), vary(65, 6, rng), vary(35, 6, rng));
  }
}

// ── NPC ────────────────────────────────────────────

const NPC_COLORS = [
  [180, 80, 80], [80, 130, 180], [180, 160, 80], [80, 160, 80],
  [160, 100, 160], [200, 130, 80], [100, 180, 160], [180, 120, 120],
];

// Profession-specific visual styles
const PROFESSION_STYLES = {
  farmer:     { body: [140, 120, 70],  accent: [180, 160, 100], hair: 'short',  accessory: 'hat' },
  hunter:     { body: [80, 110, 60],   accent: [100, 80, 50],   hair: 'short',  accessory: 'bow' },
  merchant:   { body: [130, 70, 140],  accent: [200, 170, 50],  hair: 'long',   accessory: 'pouch' },
  herbalist:  { body: [90, 130, 80],   accent: [140, 100, 160], hair: 'long',   accessory: 'leaf' },
  healer:     { body: [210, 210, 220], accent: [120, 160, 200], hair: 'long',   accessory: 'staff' },
  blacksmith: { body: [90, 70, 55],    accent: [200, 130, 50],  hair: 'bald',   accessory: 'hammer' },
  carpenter:  { body: [160, 130, 80],  accent: [100, 70, 40],   hair: 'short',  accessory: 'hammer' },
  scribe:     { body: [60, 60, 110],   accent: [200, 190, 160], hair: 'hooded', accessory: 'book' },
  miner:      { body: [100, 95, 85],   accent: [190, 170, 50],  hair: 'bald',   accessory: 'pickaxe' },
  fisher:     { body: [70, 110, 150],  accent: [80, 150, 140],  hair: 'short',  accessory: 'rod' },
  tailor:     { body: [160, 120, 170], accent: [200, 140, 160], hair: 'long',   accessory: 'scissors' },
  potter:     { body: [170, 120, 80],  accent: [200, 180, 150], hair: 'short',  accessory: null },
  barmaid:    { body: [160, 90, 70],   accent: [180, 150, 100], hair: 'long',   accessory: 'mug' },
  elder:      { body: [140, 135, 130], accent: [180, 160, 80],  hair: 'long',   accessory: null },
  guard:      { body: [120, 120, 130], accent: [90, 90, 100],   hair: 'short',  accessory: 'sword' },
  baker:      { body: [210, 200, 180], accent: [180, 60, 50],   hair: 'short',  accessory: 'apron' },
  cook:       { body: [200, 190, 170], accent: [170, 50, 40],   hair: 'short',  accessory: 'apron' },
  innkeeper:  { body: [140, 90, 60],   accent: [180, 150, 100], hair: 'short',  accessory: 'mug' },
  priest:     { body: [200, 200, 210], accent: [180, 160, 60],  hair: 'hooded', accessory: 'staff' },
};

function _drawHair(ctx, ox, oy, bob, style, bodyColor, rng) {
  const hr = vary(60, 20, rng), hg = vary(40, 15, rng), hb = vary(20, 10, rng);
  // Head top is at oy+10+bob (oval centered at oy+14, ry=4)
  const headTop = oy + 10 + bob;
  switch (style) {
    case 'bald':
      break;
    case 'long':
      // cap on top of head
      for (let x = -3; x <= 3; x++) {
        rect(ctx, ox + 16 + x, headTop, 1, 2, hr, hg, hb);
      }
      // side strands hanging down past ears
      rect(ctx, ox + 13, headTop + 2, 1, 5, hr, hg, hb);
      rect(ctx, ox + 19, headTop + 2, 1, 5, hr, hg, hb);
      break;
    case 'hooded': {
      const bc = [vary(bodyColor[0] - 20, 8, rng), vary(bodyColor[1] - 20, 8, rng), vary(bodyColor[2] - 20, 8, rng)];
      // hood over top
      for (let x = -3; x <= 3; x++) {
        rect(ctx, ox + 16 + x, headTop, 1, 2, bc[0], bc[1], bc[2]);
      }
      // hood sides
      rect(ctx, ox + 12, headTop + 1, 1, 5, bc[0], bc[1], bc[2]);
      rect(ctx, ox + 20, headTop + 1, 1, 5, bc[0], bc[1], bc[2]);
      break;
    }
    default: // 'short'
      for (let x = -2; x <= 2; x++) {
        rect(ctx, ox + 16 + x, headTop, 1, 2, hr, hg, hb);
      }
      break;
  }
}

function _drawAccessory(ctx, ox, oy, bob, accessory, accent, dir, rng) {
  if (!accessory || dir === 1) return; // skip when facing away
  const ar = vary(accent[0], 8, rng), ag = vary(accent[1], 8, rng), ab = vary(accent[2], 8, rng);
  switch (accessory) {
    case 'hat':
      // wide brim on top of head
      rect(ctx, ox + 12, oy + 8 + bob, 9, 1, ar, ag, ab);
      rect(ctx, ox + 14, oy + 7 + bob, 5, 1, ar, ag, ab);
      break;
    case 'sword':
      if (dir === 2) {
        rect(ctx, ox + 20, oy + 20 + bob, 1, 7, vary(180, 10, rng), vary(180, 10, rng), vary(190, 10, rng));
        px(ctx, ox + 20, oy + 19 + bob, ar, ag, ab);
      } else if (dir === 3) {
        rect(ctx, ox + 12, oy + 20 + bob, 1, 7, vary(180, 10, rng), vary(180, 10, rng), vary(190, 10, rng));
        px(ctx, ox + 12, oy + 19 + bob, ar, ag, ab);
      } else {
        rect(ctx, ox + 20, oy + 21 + bob, 1, 6, vary(180, 10, rng), vary(180, 10, rng), vary(190, 10, rng));
        px(ctx, ox + 20, oy + 20 + bob, ar, ag, ab);
      }
      break;
    case 'bow':
      if (dir === 2 || dir === 0) {
        rect(ctx, ox + 20, oy + 17 + bob, 1, 8, vary(120, 10, rng), vary(80, 8, rng), vary(40, 8, rng));
        px(ctx, ox + 21, oy + 20 + bob, vary(120, 10, rng), vary(80, 8, rng), vary(40, 8, rng));
      } else {
        rect(ctx, ox + 12, oy + 17 + bob, 1, 8, vary(120, 10, rng), vary(80, 8, rng), vary(40, 8, rng));
        px(ctx, ox + 11, oy + 20 + bob, vary(120, 10, rng), vary(80, 8, rng), vary(40, 8, rng));
      }
      break;
    case 'staff':
      if (dir === 2 || dir === 0) {
        rect(ctx, ox + 21, oy + 13 + bob, 1, 16, vary(140, 10, rng), vary(100, 8, rng), vary(50, 8, rng));
        px(ctx, ox + 21, oy + 12 + bob, ar, ag, ab);
      } else {
        rect(ctx, ox + 11, oy + 13 + bob, 1, 16, vary(140, 10, rng), vary(100, 8, rng), vary(50, 8, rng));
        px(ctx, ox + 11, oy + 12 + bob, ar, ag, ab);
      }
      break;
    case 'hammer':
      if (dir === 2 || dir === 0) {
        rect(ctx, ox + 20, oy + 19 + bob, 1, 6, vary(100, 8, rng), vary(70, 8, rng), vary(40, 8, rng));
        rect(ctx, ox + 19, oy + 18 + bob, 3, 2, vary(140, 10, rng), vary(140, 10, rng), vary(150, 10, rng));
      } else {
        rect(ctx, ox + 12, oy + 19 + bob, 1, 6, vary(100, 8, rng), vary(70, 8, rng), vary(40, 8, rng));
        rect(ctx, ox + 11, oy + 18 + bob, 3, 2, vary(140, 10, rng), vary(140, 10, rng), vary(150, 10, rng));
      }
      break;
    case 'pickaxe':
      if (dir === 2 || dir === 0) {
        rect(ctx, ox + 20, oy + 18 + bob, 1, 7, vary(100, 8, rng), vary(70, 8, rng), vary(40, 8, rng));
        px(ctx, ox + 21, oy + 18 + bob, vary(140, 10, rng), vary(140, 10, rng), vary(150, 10, rng));
        px(ctx, ox + 21, oy + 19 + bob, vary(140, 10, rng), vary(140, 10, rng), vary(150, 10, rng));
      } else {
        rect(ctx, ox + 12, oy + 18 + bob, 1, 7, vary(100, 8, rng), vary(70, 8, rng), vary(40, 8, rng));
        px(ctx, ox + 11, oy + 18 + bob, vary(140, 10, rng), vary(140, 10, rng), vary(150, 10, rng));
        px(ctx, ox + 11, oy + 19 + bob, vary(140, 10, rng), vary(140, 10, rng), vary(150, 10, rng));
      }
      break;
    case 'book':
      if (dir === 0) {
        rect(ctx, ox + 14, oy + 22 + bob, 4, 3, ar, ag, ab);
      } else if (dir === 2) {
        rect(ctx, ox + 19, oy + 22 + bob, 3, 4, ar, ag, ab);
      } else {
        rect(ctx, ox + 11, oy + 22 + bob, 3, 4, ar, ag, ab);
      }
      break;
    case 'pouch':
      if (dir === 2 || dir === 0) {
        rect(ctx, ox + 19, oy + 25 + bob, 2, 2, ar, ag, ab);
      } else {
        rect(ctx, ox + 12, oy + 25 + bob, 2, 2, ar, ag, ab);
      }
      break;
    case 'mug':
      if (dir === 2 || dir === 0) {
        rect(ctx, ox + 20, oy + 22 + bob, 2, 3, ar, ag, ab);
        px(ctx, ox + 22, oy + 23 + bob, ar, ag, ab);
      } else {
        rect(ctx, ox + 11, oy + 22 + bob, 2, 3, ar, ag, ab);
        px(ctx, ox + 10, oy + 23 + bob, ar, ag, ab);
      }
      break;
    case 'rod':
      if (dir === 2 || dir === 0) {
        rect(ctx, ox + 21, oy + 14 + bob, 1, 15, vary(140, 10, rng), vary(120, 8, rng), vary(80, 8, rng));
      } else {
        rect(ctx, ox + 11, oy + 14 + bob, 1, 15, vary(140, 10, rng), vary(120, 8, rng), vary(80, 8, rng));
      }
      break;
    case 'leaf':
      if (dir === 0) {
        px(ctx, ox + 15, oy + 8 + bob, vary(50, 10, rng), vary(140, 10, rng), vary(40, 10, rng));
        px(ctx, ox + 16, oy + 7 + bob, vary(50, 10, rng), vary(140, 10, rng), vary(40, 10, rng));
      }
      break;
    case 'scissors':
      if (dir === 2 || dir === 0) {
        px(ctx, ox + 20, oy + 23 + bob, vary(180, 10, rng), vary(180, 10, rng), vary(190, 10, rng));
        px(ctx, ox + 21, oy + 24 + bob, vary(180, 10, rng), vary(180, 10, rng), vary(190, 10, rng));
        px(ctx, ox + 20, oy + 25 + bob, vary(180, 10, rng), vary(180, 10, rng), vary(190, 10, rng));
      }
      break;
    case 'apron':
      // white apron strip over front of body
      if (dir === 0) {
        rect(ctx, ox + 14, oy + 22 + bob, 5, 5, vary(230, 8, rng), vary(230, 8, rng), vary(225, 8, rng));
      }
      break;
  }
}

function drawNpcFrame(ctx, ox, oy, colorIdx, dir, frame, rng, profession) {
  const style = PROFESSION_STYLES[profession];
  const c = style ? style.body : NPC_COLORS[colorIdx % NPC_COLORS.length];
  const accent = style ? style.accent : c;
  const bob = (frame === 1) ? -1 : (frame === 3) ? 1 : 0;

  // Shadow
  for (let x = -3; x <= 3; x++) {
    px(ctx, ox + 16 + x, oy + 30, 0, 0, 0, 50);
  }

  // Body
  for (let y = 0; y < 10; y++) {
    for (let x = -3; x <= 3; x++) {
      rect(ctx, ox + 16 + x, oy + 18 + y + bob, 1, 1,
        vary(c[0], 10, rng), vary(c[1], 8, rng), vary(c[2], 8, rng));
    }
  }

  // Accent belt/trim at mid-body
  if (style) {
    for (let x = -3; x <= 3; x++) {
      rect(ctx, ox + 16 + x, oy + 23 + bob, 1, 1,
        vary(accent[0], 8, rng), vary(accent[1], 8, rng), vary(accent[2], 8, rng));
    }
  }

  // Head (skin) — oval: rx=3, ry=4, centered at (16, 14)
  for (let y = -4; y <= 3; y++) {
    for (let x = -3; x <= 3; x++) {
      if ((x * x) / 9 + (y * y) / 16 > 1) continue;
      rect(ctx, ox + 16 + x, oy + 14 + y + bob, 1, 1,
        vary(210, 12, rng), vary(180, 10, rng), vary(150, 10, rng));
    }
  }

  // Hair
  _drawHair(ctx, ox, oy, bob, style ? style.hair : 'short', c, rng);

  // Legs
  const legOff = (frame % 2 === 1) ? 1 : 0;
  rect(ctx, ox + 14, oy + 28 + bob - legOff, 2, 3, vary(60, 10, rng), vary(50, 10, rng), vary(40, 8, rng));
  rect(ctx, ox + 17, oy + 28 + bob + legOff, 2, 3, vary(60, 10, rng), vary(50, 10, rng), vary(40, 8, rng));

  // Eyes (head centered at oy+14)
  if (dir === 0) { // down
    px(ctx, ox + 14, oy + 15 + bob, 30, 30, 40);
    px(ctx, ox + 18, oy + 15 + bob, 30, 30, 40);
  } else if (dir === 1) { // up
    // no eyes visible
  } else if (dir === 2) { // right
    px(ctx, ox + 18, oy + 15 + bob, 30, 30, 40);
  } else { // left
    px(ctx, ox + 14, oy + 15 + bob, 30, 30, 40);
  }

  // Accessory
  if (style) {
    _drawAccessory(ctx, ox, oy, bob, style.accessory, accent, dir, rng);
  }
}

// ── LANDMARKS ──────────────────────────────────────

function drawLandmark(ctx, ox, oy, type, rng) {
  switch (type) {
    case 'stone_circle':
      for (let a = 0; a < 360; a += 45) {
        const x = ox + 16 + Math.floor(Math.cos(a * Math.PI / 180) * 10);
        const y = oy + 16 + Math.floor(Math.sin(a * Math.PI / 180) * 8);
        rect(ctx, x, y, 3, 5, vary(130, 12, rng), vary(128, 10, rng), vary(125, 10, rng));
      }
      break;
    case 'ancient_tree':
      rect(ctx, ox + 10, oy + 14, 12, 18, vary(80, 12, rng), vary(55, 10, rng), vary(25, 8, rng));
      for (let cy = -12; cy <= 2; cy++) {
        for (let cx = -14; cx <= 14; cx++) {
          if (cx * cx + cy * cy > 200) continue;
          const x = ox + 16 + cx;
          const y = oy + 10 + cy;
          if (x >= ox && x < ox + S && y >= oy) {
            ctx.fillStyle = `rgb(${vary(25, 12, rng)},${vary(70, 18, rng)},${vary(18, 10, rng)})`;
            ctx.fillRect(x, y, 1, 1);
          }
        }
      }
      break;
    case 'crater':
      for (let cy = -10; cy <= 10; cy++) {
        for (let cx = -12; cx <= 12; cx++) {
          const d = cx * cx + cy * cy;
          if (d > 144) continue;
          const x = ox + 16 + cx;
          const y = oy + 16 + cy;
          if (d > 80) {
            rect(ctx, x, y, 1, 1, vary(90, 10, rng), vary(75, 8, rng), vary(60, 8, rng));
          } else {
            rect(ctx, x, y, 1, 1, vary(40, 8, rng), vary(35, 6, rng), vary(30, 6, rng));
          }
        }
      }
      break;
    case 'oasis':
      for (let cy = -6; cy <= 6; cy++) {
        for (let cx = -8; cx <= 8; cx++) {
          if (cx * cx + cy * cy > 64) continue;
          const x = ox + 16 + cx;
          const y = oy + 18 + cy;
          rect(ctx, x, y, 1, 1, vary(40, 8, rng), vary(100, 10, rng), vary(150, 10, rng));
        }
      }
      drawPalmTree(ctx, ox, oy, rng);
      break;
    default:
      rect(ctx, ox + 10, oy + 10, 12, 16, vary(150, 15, rng), vary(140, 12, rng), vary(130, 12, rng));
      break;
  }
}

// ── PUBLIC API ─────────────────────────────────────

export function generateTreeTextures(scene) {
  const types = ['pine', 'oak', 'palm', 'dead'];
  types.forEach((type, ti) => {
    for (let v = 0; v < 4; v++) {
      const key = `tree_${type}_${v}`;
      if (scene.textures.exists(key)) return;
      const canvas = document.createElement('canvas');
      canvas.width = S;
      canvas.height = S;
      const ctx = canvas.getContext('2d');
      const rng = simpleRng(ti * 1000 + v * 100 + 42);
      if (type === 'pine') drawPineTree(ctx, 0, 0, rng);
      else if (type === 'oak') drawOakTree(ctx, 0, 0, rng);
      else if (type === 'palm') drawPalmTree(ctx, 0, 0, rng);
      else drawDeadTree(ctx, 0, 0, rng);
      scene.textures.addCanvas(key, canvas);
    }
  });
}

export function generateBuildingTextures(scene, buildings) {
  const generated = new Set();
  buildings.forEach((b, i) => {
    const key = `bld_${b.type}_${i}`;
    if (generated.has(key) || scene.textures.exists(key)) return;
    const canvasW = (b.gridW || 1) * S;
    const canvasH = (b.gridH || 1) * S;
    const canvas = document.createElement('canvas');
    canvas.width = canvasW;
    canvas.height = canvasH;
    const ctx = canvas.getContext('2d');
    const rng = simpleRng(i * 137 + 99);
    drawBuilding(ctx, 0, 0, b.type, rng, canvasW, canvasH, b);
    scene.textures.addCanvas(key, canvas);
    generated.add(key);
  });
}

export function generateNpcSpriteSheet(scene, npcIndex, profession) {
  const key = `npc_${npcIndex}_${profession || 'default'}`;
  if (scene.textures.exists(key)) return key;

  const frames = 4;
  const dirs = 4;
  const canvas = document.createElement('canvas');
  canvas.width = S * frames;
  canvas.height = S * dirs;
  const ctx = canvas.getContext('2d');
  const rng = simpleRng(npcIndex * 73 + 17);
  const colorIdx = npcIndex;

  for (let dir = 0; dir < dirs; dir++) {
    for (let frame = 0; frame < frames; frame++) {
      const fRng = simpleRng(npcIndex * 73 + dir * 10 + frame + 17);
      drawNpcFrame(ctx, frame * S, dir * S, colorIdx, dir, frame, fRng, profession);
    }
  }

  scene.textures.addSpriteSheet(key, canvas, { frameWidth: S, frameHeight: S });
  return key;
}

/**
 * Generate a standalone preview canvas for an NPC (no Phaser dependency).
 * Returns a canvas with 4 columns (frames) x 4 rows (directions), each 32x32.
 * Directions: 0=down, 1=up, 2=right, 3=left
 */
export function generateNpcPreviewCanvas(npcIndex, profession) {
  const frames = 4;
  const dirs = 4;
  const canvas = document.createElement('canvas');
  canvas.width = S * frames;
  canvas.height = S * dirs;
  const ctx = canvas.getContext('2d');

  for (let dir = 0; dir < dirs; dir++) {
    for (let frame = 0; frame < frames; frame++) {
      const fRng = simpleRng(npcIndex * 73 + dir * 10 + frame + 17);
      drawNpcFrame(ctx, frame * S, dir * S, npcIndex, dir, frame, fRng, profession);
    }
  }

  return canvas;
}

export function generateBuildingPreviewCanvas(locType, locIndex) {
  const btype = locType || 'house';
  const canvasW = S;
  const canvasH = S;
  const canvas = document.createElement('canvas');
  canvas.width = canvasW;
  canvas.height = canvasH;
  const ctx = canvas.getContext('2d');
  const rng = simpleRng((locIndex || 0) * 137 + 99);
  drawBuilding(ctx, 0, 0, btype, rng, canvasW, canvasH, {});
  return canvas;
}

export function generateLandmarkTextures(scene, landmarks) {
  landmarks.forEach((l, i) => {
    const key = `lmk_${i}`;
    if (scene.textures.exists(key)) return;
    const canvas = document.createElement('canvas');
    canvas.width = S;
    canvas.height = S;
    const ctx = canvas.getContext('2d');
    const rng = simpleRng(i * 211 + 55);
    drawLandmark(ctx, 0, 0, l.type, rng);
    scene.textures.addCanvas(key, canvas);
  });
}

// ── ENEMY SPRITES ─────────────────────────────────

const ENEMY_STYLES = {
  wolf:         { type: 'beast', body: [100, 90, 80],  accent: [80, 70, 60],  shape: 'canine' },
  bear:         { type: 'beast', body: [100, 70, 40],  accent: [80, 55, 30],  shape: 'bulky' },
  wild_boar:    { type: 'beast', body: [120, 90, 60],  accent: [90, 70, 40],  shape: 'canine' },
  giant_spider: { type: 'beast', body: [40, 40, 50],   accent: [80, 30, 30],  shape: 'spider' },
  cave_bat:     { type: 'beast', body: [50, 45, 55],   accent: [70, 60, 70],  shape: 'bat' },
  serpent:      { type: 'beast', body: [60, 100, 50],  accent: [80, 120, 60], shape: 'serpent' },
  bandit:       { type: 'human', body: [90, 60, 50],   accent: [60, 40, 30],  hair: 'hooded', accessory: 'sword' },
  robber:       { type: 'human', body: [70, 55, 45],   accent: [50, 35, 25],  hair: 'hooded', accessory: null },
  pirate:       { type: 'human', body: [80, 50, 50],   accent: [60, 40, 30],  hair: 'short',  accessory: 'sword' },
  brigand:      { type: 'human', body: [60, 55, 45],   accent: [80, 50, 40],  hair: 'bald',   accessory: 'sword' },
  cultist:      { type: 'human', body: [50, 30, 60],   accent: [80, 50, 90],  hair: 'hooded', accessory: 'staff' },
};

function drawEnemyBeastFrame(ctx, ox, oy, style, dir, frame, rng) {
  const c = style.body;
  const ac = style.accent;
  const bob = (frame === 1) ? -1 : (frame === 3) ? 1 : 0;
  const shape = style.shape || 'canine';

  // Shadow
  for (let x = -4; x <= 4; x++) {
    px(ctx, ox + 16 + x, oy + 30, 0, 0, 0, 50);
  }

  if (shape === 'spider') {
    // Spider: round body + 8 legs
    for (let cy = -3; cy <= 3; cy++) {
      for (let cx = -4; cx <= 4; cx++) {
        if (cx * cx + cy * cy > 20) continue;
        rect(ctx, ox + 16 + cx, oy + 20 + cy + bob, 1, 1,
          vary(c[0], 10, rng), vary(c[1], 8, rng), vary(c[2], 8, rng));
      }
    }
    // Eyes
    if (dir !== 1) {
      px(ctx, ox + 14, oy + 18 + bob, 200, 30, 30);
      px(ctx, ox + 18, oy + 18 + bob, 200, 30, 30);
      px(ctx, ox + 13, oy + 19 + bob, 200, 30, 30);
      px(ctx, ox + 19, oy + 19 + bob, 200, 30, 30);
    }
    // Legs (4 pairs)
    const legOff = (frame % 2 === 1) ? 1 : 0;
    for (let i = -3; i <= 3; i += 2) {
      const ly = oy + 22 + bob + ((i + legOff) % 2);
      rect(ctx, ox + 10 + i, ly, 1, 3, vary(ac[0], 8, rng), vary(ac[1], 6, rng), vary(ac[2], 6, rng));
      rect(ctx, ox + 22 - i, ly, 1, 3, vary(ac[0], 8, rng), vary(ac[1], 6, rng), vary(ac[2], 6, rng));
    }
  } else if (shape === 'bat') {
    // Bat: small body + wings
    for (let cy = -2; cy <= 2; cy++) {
      for (let cx = -2; cx <= 2; cx++) {
        rect(ctx, ox + 16 + cx, oy + 18 + cy + bob, 1, 1,
          vary(c[0], 8, rng), vary(c[1], 6, rng), vary(c[2], 6, rng));
      }
    }
    // Wings
    const wingSpread = (frame % 2 === 0) ? 6 : 4;
    for (let wy = 0; wy < 3; wy++) {
      rect(ctx, ox + 16 - wingSpread - wy, oy + 17 + wy + bob, wingSpread, 1,
        vary(ac[0], 10, rng), vary(ac[1], 8, rng), vary(ac[2], 8, rng));
      rect(ctx, ox + 18 + wy, oy + 17 + wy + bob, wingSpread, 1,
        vary(ac[0], 10, rng), vary(ac[1], 8, rng), vary(ac[2], 8, rng));
    }
    // Eyes
    if (dir !== 1) {
      px(ctx, ox + 15, oy + 17 + bob, 200, 50, 50);
      px(ctx, ox + 17, oy + 17 + bob, 200, 50, 50);
    }
  } else if (shape === 'serpent') {
    // Serpent: wavy body
    for (let i = 0; i < 10; i++) {
      const sx = ox + 10 + i * 1.2;
      const wave = Math.sin((i + frame) * 0.8) * 2;
      rect(ctx, Math.floor(sx), oy + 20 + Math.floor(wave) + bob, 2, 3,
        vary(c[0], 10, rng), vary(c[1], 8, rng), vary(c[2], 8, rng));
    }
    // Head
    rect(ctx, ox + 10, oy + 18 + bob, 4, 4,
      vary(ac[0], 8, rng), vary(ac[1], 6, rng), vary(ac[2], 6, rng));
    if (dir !== 1) {
      px(ctx, ox + 11, oy + 18 + bob, 200, 50, 30);
    }
  } else {
    // Canine/bulky (wolf, bear, boar): 4-legged body
    const bodyW = shape === 'bulky' ? 10 : 8;
    const bodyH = shape === 'bulky' ? 7 : 5;
    const bodyY = shape === 'bulky' ? 16 : 18;

    // Body
    for (let cy = 0; cy < bodyH; cy++) {
      for (let cx = 0; cx < bodyW; cx++) {
        rect(ctx, ox + 16 - bodyW / 2 + cx, oy + bodyY + cy + bob, 1, 1,
          vary(c[0], 10, rng), vary(c[1], 8, rng), vary(c[2], 8, rng));
      }
    }

    // Head
    const headSize = shape === 'bulky' ? 5 : 4;
    const headX = dir === 2 ? ox + 16 + bodyW / 2 - 1 : dir === 3 ? ox + 16 - bodyW / 2 - headSize + 1 : ox + 16 - headSize / 2;
    const headY = oy + bodyY - 2 + bob;
    for (let cy = 0; cy < headSize; cy++) {
      for (let cx = 0; cx < headSize; cx++) {
        rect(ctx, headX + cx, headY + cy, 1, 1,
          vary(ac[0], 8, rng), vary(ac[1], 6, rng), vary(ac[2], 6, rng));
      }
    }

    // Eyes
    if (dir === 0) {
      px(ctx, headX + 1, headY + 1, 30, 30, 30);
      px(ctx, headX + headSize - 2, headY + 1, 30, 30, 30);
    } else if (dir === 2) {
      px(ctx, headX + headSize - 2, headY + 1, 30, 30, 30);
    } else if (dir === 3) {
      px(ctx, headX + 1, headY + 1, 30, 30, 30);
    }

    // Legs
    const legOff = (frame % 2 === 1) ? 1 : 0;
    const legY = oy + bodyY + bodyH + bob;
    rect(ctx, ox + 13, legY - legOff, 2, 4, vary(c[0] - 20, 8, rng), vary(c[1] - 15, 6, rng), vary(c[2] - 10, 6, rng));
    rect(ctx, ox + 17, legY + legOff, 2, 4, vary(c[0] - 20, 8, rng), vary(c[1] - 15, 6, rng), vary(c[2] - 10, 6, rng));
    if (shape === 'bulky') {
      rect(ctx, ox + 11, legY + legOff, 2, 4, vary(c[0] - 20, 8, rng), vary(c[1] - 15, 6, rng), vary(c[2] - 10, 6, rng));
      rect(ctx, ox + 19, legY - legOff, 2, 4, vary(c[0] - 20, 8, rng), vary(c[1] - 15, 6, rng), vary(c[2] - 10, 6, rng));
    }
  }
}

function drawEnemyHumanFrame(ctx, ox, oy, style, dir, frame, rng) {
  // Reuse NPC drawing with enemy-specific colors
  const pseudoStyle = {
    body: style.body,
    accent: style.accent,
    hair: style.hair || 'hooded',
    accessory: style.accessory || null,
  };
  const c = pseudoStyle.body;
  const accent = pseudoStyle.accent;
  const bob = (frame === 1) ? -1 : (frame === 3) ? 1 : 0;

  // Shadow
  for (let x = -3; x <= 3; x++) {
    px(ctx, ox + 16 + x, oy + 30, 0, 0, 0, 50);
  }

  // Body
  for (let y = 0; y < 10; y++) {
    for (let x = -3; x <= 3; x++) {
      rect(ctx, ox + 16 + x, oy + 18 + y + bob, 1, 1,
        vary(c[0], 10, rng), vary(c[1], 8, rng), vary(c[2], 8, rng));
    }
  }

  // Belt
  for (let x = -3; x <= 3; x++) {
    rect(ctx, ox + 16 + x, oy + 23 + bob, 1, 1,
      vary(accent[0], 8, rng), vary(accent[1], 8, rng), vary(accent[2], 8, rng));
  }

  // Head (darker skin tone for enemies)
  for (let y = -4; y <= 3; y++) {
    for (let x = -3; x <= 3; x++) {
      if ((x * x) / 9 + (y * y) / 16 > 1) continue;
      rect(ctx, ox + 16 + x, oy + 14 + y + bob, 1, 1,
        vary(180, 12, rng), vary(150, 10, rng), vary(120, 10, rng));
    }
  }

  // Hair
  _drawHair(ctx, ox, oy, bob, pseudoStyle.hair, c, rng);

  // Legs
  const legOff = (frame % 2 === 1) ? 1 : 0;
  rect(ctx, ox + 14, oy + 28 + bob - legOff, 2, 3, vary(40, 10, rng), vary(35, 10, rng), vary(30, 8, rng));
  rect(ctx, ox + 17, oy + 28 + bob + legOff, 2, 3, vary(40, 10, rng), vary(35, 10, rng), vary(30, 8, rng));

  // Eyes (red-ish for enemies)
  if (dir === 0) {
    px(ctx, ox + 14, oy + 15 + bob, 180, 40, 40);
    px(ctx, ox + 18, oy + 15 + bob, 180, 40, 40);
  } else if (dir === 2) {
    px(ctx, ox + 18, oy + 15 + bob, 180, 40, 40);
  } else if (dir === 3) {
    px(ctx, ox + 14, oy + 15 + bob, 180, 40, 40);
  }

  // Accessory
  if (pseudoStyle.accessory) {
    _drawAccessory(ctx, ox, oy, bob, pseudoStyle.accessory, accent, dir, rng);
  }
}

export function generateEnemySpriteSheet(scene, enemyType, enemyIndex) {
  const key = `enemy_${enemyType}_${enemyIndex}`;
  if (scene.textures.exists(key)) return key;

  const style = ENEMY_STYLES[enemyType] || ENEMY_STYLES.wolf;
  const isHuman = style.type === 'human';

  const frames = 4;
  const dirs = 4;
  const canvas = document.createElement('canvas');
  canvas.width = S * frames;
  canvas.height = S * dirs;
  const ctx = canvas.getContext('2d');

  for (let dir = 0; dir < dirs; dir++) {
    for (let frame = 0; frame < frames; frame++) {
      const fRng = simpleRng(enemyIndex * 97 + dir * 13 + frame + 31);
      if (isHuman) {
        drawEnemyHumanFrame(ctx, frame * S, dir * S, style, dir, frame, fRng);
      } else {
        drawEnemyBeastFrame(ctx, frame * S, dir * S, style, dir, frame, fRng);
      }
    }
  }

  scene.textures.addSpriteSheet(key, canvas, { frameWidth: S, frameHeight: S });
  return key;
}

// ── MOUNT & CARRIAGE SPRITES ──────────────────────

const MOUNT_STYLES = {
  horse:     { body: [140, 100, 60],  mane: [80, 55, 30],   scale: 1.0 },
  war_horse: { body: [70, 60, 55],    mane: [40, 35, 30],   scale: 1.15, armor: [150, 140, 120] },
  pony:      { body: [180, 150, 100], mane: [130, 100, 60], scale: 0.8 },
};

function drawMountFrame(ctx, ox, oy, mountType, dir, frame, rng) {
  const style = MOUNT_STYLES[mountType] || MOUNT_STYLES.horse;
  const c = style.body;
  const mc = style.mane;
  const sc = style.scale;
  const bob = (frame === 1) ? -1 : (frame === 3) ? 1 : 0;
  const legOff = (frame % 2 === 1) ? 1 : 0;
  const legC = [Math.max(0, c[0] - 25), Math.max(0, c[1] - 20), Math.max(0, c[2] - 15)];
  const hc = [c[0] + 10, c[1] + 8, c[2] + 5]; // head color (slightly lighter)

  // Center x/y in tile
  const cx = ox + 16;
  const baseY = oy + 22;

  if (dir === 2 || dir === 3) {
    // ── SIDE VIEW (right=2, left=3) ──
    const flip = dir === 3 ? -1 : 1;
    const bodyLen = Math.round(12 * sc);
    const bodyH = Math.round(7 * sc);
    const bx = cx - Math.floor(bodyLen / 2);
    const by = baseY - bodyH + bob;
    const legH = Math.round(4 * sc);

    // Shadow
    for (let x = -Math.floor(bodyLen / 2) - 2; x <= Math.floor(bodyLen / 2) + 3; x++) {
      px(ctx, cx + x, oy + 30, 0, 0, 0, 45);
    }

    // Body (rounded rect — slightly tapered at rear)
    for (let y = 0; y < bodyH; y++) {
      const taper = y <= 1 || y >= bodyH - 1 ? 1 : 0;
      for (let x = taper; x < bodyLen - taper; x++) {
        rect(ctx, bx + x, by + y, 1, 1,
          vary(c[0], 10, rng), vary(c[1], 8, rng), vary(c[2], 8, rng));
      }
    }

    // Belly highlight
    for (let x = 2; x < bodyLen - 2; x++) {
      rect(ctx, bx + x, by + bodyH - 2, 1, 1,
        vary(c[0] + 15, 6, rng), vary(c[1] + 12, 6, rng), vary(c[2] + 8, 6, rng));
    }

    // Armor plating for war_horse (saddle blanket)
    if (style.armor) {
      const ac = style.armor;
      for (let x = 3; x < bodyLen - 3; x++) {
        for (let y = 0; y < 2; y++) {
          rect(ctx, bx + x, by + y + 1, 1, 1,
            vary(ac[0], 8, rng), vary(ac[1], 6, rng), vary(ac[2], 6, rng));
        }
      }
    }

    // Neck (extends forward-up from body)
    const neckX = flip === 1 ? bx + bodyLen - 1 : bx;
    const neckLen = Math.round(3 * sc);
    for (let i = 0; i < neckLen; i++) {
      const nx = neckX + flip * i;
      rect(ctx, nx, by - i - 1, 2, 2,
        vary(c[0] + 5, 8, rng), vary(c[1] + 3, 6, rng), vary(c[2] + 2, 6, rng));
    }

    // Head (elongated snout extending forward)
    const headX = neckX + flip * neckLen;
    const headY = by - neckLen - 1;
    const headLen = Math.round(4 * sc);
    const headH = Math.round(3 * sc);
    for (let y = 0; y < headH; y++) {
      for (let x = 0; x < headLen; x++) {
        rect(ctx, headX + flip * x, headY + y, 1, 1,
          vary(hc[0], 8, rng), vary(hc[1], 6, rng), vary(hc[2], 6, rng));
      }
    }

    // Nostril
    px(ctx, headX + flip * (headLen - 1), headY + headH - 1, 50, 40, 35);

    // Eye
    px(ctx, headX + flip * 1, headY + 1, 20, 20, 25);

    // Ears (two small triangles on top of head)
    px(ctx, headX + flip * 1, headY - 1, vary(hc[0], 6, rng), vary(hc[1], 6, rng), vary(hc[2], 6, rng));
    px(ctx, headX + flip * 2, headY - 1, vary(hc[0], 6, rng), vary(hc[1], 6, rng), vary(hc[2], 6, rng));

    // Mane (runs along neck and top of body)
    for (let i = 0; i < neckLen + 2; i++) {
      const mx = neckX + flip * (i - 1);
      const my = by - Math.min(i, neckLen) - 2;
      rect(ctx, mx, my, 1, 2,
        vary(mc[0], 10, rng), vary(mc[1], 8, rng), vary(mc[2], 8, rng));
    }

    // Tail (flowing from rear)
    const tailX = flip === 1 ? bx : bx + bodyLen - 1;
    for (let i = 0; i < 4; i++) {
      const sway = (frame % 2 === 0) ? i % 2 : (i + 1) % 2;
      rect(ctx, tailX - flip * (i + 1), by + 2 + i + sway, 1, 2,
        vary(mc[0], 10, rng), vary(mc[1], 8, rng), vary(mc[2], 8, rng));
    }

    // Legs (4 legs — front pair and rear pair, with walk cycle)
    const frontLegX = flip === 1 ? bx + bodyLen - 3 : bx + 1;
    const rearLegX = flip === 1 ? bx + 1 : bx + bodyLen - 3;
    // Front legs
    rect(ctx, frontLegX, by + bodyH - legOff, 2, legH, vary(legC[0], 8, rng), vary(legC[1], 6, rng), vary(legC[2], 6, rng));
    rect(ctx, frontLegX + 1, by + bodyH + legOff, 1, legH, vary(legC[0] - 8, 6, rng), vary(legC[1] - 6, 6, rng), vary(legC[2] - 4, 6, rng));
    // Rear legs
    rect(ctx, rearLegX, by + bodyH + legOff, 2, legH, vary(legC[0], 8, rng), vary(legC[1], 6, rng), vary(legC[2], 6, rng));
    rect(ctx, rearLegX + 1, by + bodyH - legOff, 1, legH, vary(legC[0] - 8, 6, rng), vary(legC[1] - 6, 6, rng), vary(legC[2] - 4, 6, rng));

    // Hooves
    const hoofC = [40, 35, 30];
    px(ctx, frontLegX, by + bodyH + legH - legOff, hoofC[0], hoofC[1], hoofC[2]);
    px(ctx, frontLegX + 1, by + bodyH + legH + legOff, hoofC[0], hoofC[1], hoofC[2]);
    px(ctx, rearLegX, by + bodyH + legH + legOff, hoofC[0], hoofC[1], hoofC[2]);
    px(ctx, rearLegX + 1, by + bodyH + legH - legOff, hoofC[0], hoofC[1], hoofC[2]);

  } else {
    // ── FRONT/BACK VIEW (down=0, up=1) ──
    const bodyW = Math.round(8 * sc);
    const bodyH = Math.round(8 * sc);
    const bx = cx - Math.floor(bodyW / 2);
    const by = baseY - bodyH + bob;
    const legH = Math.round(3 * sc);

    // Shadow
    for (let x = -Math.floor(bodyW / 2) - 1; x <= Math.floor(bodyW / 2) + 1; x++) {
      px(ctx, cx + x, oy + 30, 0, 0, 0, 45);
    }

    // Body (rounded)
    for (let y = 0; y < bodyH; y++) {
      const taper = y <= 0 || y >= bodyH - 1 ? 1 : 0;
      for (let x = taper; x < bodyW - taper; x++) {
        rect(ctx, bx + x, by + y, 1, 1,
          vary(c[0], 10, rng), vary(c[1], 8, rng), vary(c[2], 8, rng));
      }
    }

    // Armor for war_horse
    if (style.armor) {
      const ac = style.armor;
      for (let x = 1; x < bodyW - 1; x++) {
        rect(ctx, bx + x, by + 1, 1, 2,
          vary(ac[0], 8, rng), vary(ac[1], 6, rng), vary(ac[2], 6, rng));
      }
    }

    if (dir === 0) {
      // FACING DOWN — we see the front: head below body, rump above

      // Neck
      rect(ctx, cx - 1, by + bodyH - 1, 3, 3,
        vary(c[0] + 5, 8, rng), vary(c[1] + 3, 6, rng), vary(c[2] + 2, 6, rng));

      // Head (below body, facing toward camera)
      const headY = by + bodyH + 2;
      const headW = Math.round(4 * sc);
      const hx = cx - Math.floor(headW / 2);
      for (let y = 0; y < 4; y++) {
        const taper = y >= 3 ? 1 : 0;
        for (let x = taper; x < headW - taper; x++) {
          rect(ctx, hx + x, headY + y, 1, 1,
            vary(hc[0], 8, rng), vary(hc[1], 6, rng), vary(hc[2], 6, rng));
        }
      }

      // Ears (pointing up/out from top of head)
      px(ctx, hx, headY - 1, vary(hc[0], 6, rng), vary(hc[1], 6, rng), vary(hc[2], 6, rng));
      px(ctx, hx + headW - 1, headY - 1, vary(hc[0], 6, rng), vary(hc[1], 6, rng), vary(hc[2], 6, rng));

      // Eyes
      px(ctx, hx + 1, headY + 1, 20, 20, 25);
      px(ctx, hx + headW - 2, headY + 1, 20, 20, 25);

      // Nostrils
      px(ctx, hx + 1, headY + 3, 50, 40, 35);
      px(ctx, hx + headW - 2, headY + 3, 50, 40, 35);

      // Mane (between ears, across top of head)
      for (let x = 1; x < headW - 1; x++) {
        px(ctx, hx + x, headY - 1, vary(mc[0], 10, rng), vary(mc[1], 8, rng), vary(mc[2], 8, rng));
      }

      // Tail (behind, sticking up from top of body)
      for (let i = 0; i < 3; i++) {
        const sway = (frame % 2 === 0) ? 0 : (i % 2 === 0 ? -1 : 1);
        rect(ctx, cx + sway, by - i - 1, 1, 1,
          vary(mc[0], 10, rng), vary(mc[1], 8, rng), vary(mc[2], 8, rng));
      }

      // Legs (front pair visible below head)
      rect(ctx, cx - 3, by + bodyH + 5 - legOff, 2, legH, vary(legC[0], 8, rng), vary(legC[1], 6, rng), vary(legC[2], 6, rng));
      rect(ctx, cx + 2, by + bodyH + 5 + legOff, 2, legH, vary(legC[0], 8, rng), vary(legC[1], 6, rng), vary(legC[2], 6, rng));

    } else {
      // FACING UP — we see the rear: tail below, head hidden above

      // Rump highlight
      for (let x = 1; x < bodyW - 1; x++) {
        rect(ctx, bx + x, by + bodyH - 2, 1, 1,
          vary(c[0] + 12, 6, rng), vary(c[1] + 10, 6, rng), vary(c[2] + 6, 6, rng));
      }

      // Tail (hanging down from bottom of body)
      for (let i = 0; i < 5; i++) {
        const sway = (frame % 2 === 0) ? (i % 2 === 0 ? -1 : 0) : (i % 2 === 0 ? 0 : 1);
        rect(ctx, cx + sway, by + bodyH + i, 1, 1,
          vary(mc[0], 10, rng), vary(mc[1], 8, rng), vary(mc[2], 8, rng));
      }

      // Ears (visible at top from behind)
      px(ctx, cx - 2, by - 1, vary(hc[0], 6, rng), vary(hc[1], 6, rng), vary(hc[2], 6, rng));
      px(ctx, cx + 2, by - 1, vary(hc[0], 6, rng), vary(hc[1], 6, rng), vary(hc[2], 6, rng));

      // Mane (top of head, between ears)
      for (let x = -1; x <= 1; x++) {
        px(ctx, cx + x, by - 1, vary(mc[0], 10, rng), vary(mc[1], 8, rng), vary(mc[2], 8, rng));
      }

      // Legs (rear pair visible below body)
      rect(ctx, cx - 3, by + bodyH - legOff, 2, legH, vary(legC[0], 8, rng), vary(legC[1], 6, rng), vary(legC[2], 6, rng));
      rect(ctx, cx + 2, by + bodyH + legOff, 2, legH, vary(legC[0], 8, rng), vary(legC[1], 6, rng), vary(legC[2], 6, rng));
    }
  }
}

export function generateMountSpriteSheet(scene, mountIndex, mountType) {
  const key = `mount_${mountIndex}_${mountType || 'horse'}`;
  if (scene.textures.exists(key)) return key;

  const frames = 4;
  const dirs = 4;
  const canvas = document.createElement('canvas');
  canvas.width = S * frames;
  canvas.height = S * dirs;
  const ctx = canvas.getContext('2d');

  for (let dir = 0; dir < dirs; dir++) {
    for (let frame = 0; frame < frames; frame++) {
      const fRng = simpleRng(mountIndex * 89 + dir * 11 + frame + 43);
      drawMountFrame(ctx, frame * S, dir * S, mountType, dir, frame, fRng);
    }
  }

  scene.textures.addSpriteSheet(key, canvas, { frameWidth: S, frameHeight: S });
  return key;
}

function drawCarriageFrame(ctx, ox, oy, carriageType, dir, rng) {
  const isWagon = carriageType === 'wagon';
  const bodyW = isWagon ? 16 : 12;
  const bodyH = isWagon ? 8 : 6;
  const woodC = isWagon ? [100, 70, 40] : [130, 90, 50];

  // Shadow
  for (let x = -Math.floor(bodyW / 2); x <= Math.floor(bodyW / 2); x++) {
    px(ctx, ox + 16 + x, oy + 30, 0, 0, 0, 40);
  }

  // Body/bed
  const bodyX = ox + 16 - Math.floor(bodyW / 2);
  const bodyY = oy + 18;
  for (let cy = 0; cy < bodyH; cy++) {
    for (let cx = 0; cx < bodyW; cx++) {
      rect(ctx, bodyX + cx, bodyY + cy, 1, 1,
        vary(woodC[0], 10, rng), vary(woodC[1], 8, rng), vary(woodC[2], 8, rng));
    }
  }

  // Side rails
  rect(ctx, bodyX, bodyY, bodyW, 1, vary(woodC[0] - 20, 8, rng), vary(woodC[1] - 15, 6, rng), vary(woodC[2] - 10, 6, rng));
  rect(ctx, bodyX, bodyY + bodyH - 1, bodyW, 1, vary(woodC[0] - 20, 8, rng), vary(woodC[1] - 15, 6, rng), vary(woodC[2] - 10, 6, rng));

  // Canvas cover for wagon
  if (isWagon) {
    const coverC = [200, 190, 170];
    for (let cx = 0; cx < bodyW; cx++) {
      for (let cy = 0; cy < 4; cy++) {
        rect(ctx, bodyX + cx, bodyY - 4 + cy, 1, 1,
          vary(coverC[0], 8, rng), vary(coverC[1], 6, rng), vary(coverC[2], 6, rng));
      }
    }
    rect(ctx, bodyX + 2, bodyY - 5, 1, 6, vary(woodC[0] - 10, 6, rng), vary(woodC[1] - 8, 5, rng), vary(woodC[2] - 5, 5, rng));
    rect(ctx, bodyX + bodyW - 3, bodyY - 5, 1, 6, vary(woodC[0] - 10, 6, rng), vary(woodC[1] - 8, 5, rng), vary(woodC[2] - 5, 5, rng));
  }

  // Wheels
  const wheelC = [60, 50, 40];
  if (isWagon) {
    for (const wx of [bodyX + 1, bodyX + 4, bodyX + bodyW - 5, bodyX + bodyW - 2]) {
      for (let cy = 0; cy < 3; cy++) {
        rect(ctx, wx, bodyY + bodyH + cy, 2, 1, vary(wheelC[0], 8, rng), vary(wheelC[1], 6, rng), vary(wheelC[2], 6, rng));
      }
    }
  } else {
    for (const wx of [bodyX + 2, bodyX + bodyW - 4]) {
      for (let cy = 0; cy < 3; cy++) {
        rect(ctx, wx, bodyY + bodyH + cy, 2, 1, vary(wheelC[0], 8, rng), vary(wheelC[1], 6, rng), vary(wheelC[2], 6, rng));
      }
    }
  }

  // Hitch pole
  if (dir === 2) {
    rect(ctx, bodyX + bodyW, bodyY + Math.floor(bodyH / 2), 4, 1, vary(woodC[0] - 10, 6, rng), vary(woodC[1] - 8, 5, rng), vary(woodC[2], 5, rng));
  } else if (dir === 3) {
    rect(ctx, bodyX - 4, bodyY + Math.floor(bodyH / 2), 4, 1, vary(woodC[0] - 10, 6, rng), vary(woodC[1] - 8, 5, rng), vary(woodC[2], 5, rng));
  } else if (dir === 0) {
    rect(ctx, ox + 15, bodyY + bodyH, 2, 3, vary(woodC[0] - 10, 6, rng), vary(woodC[1] - 8, 5, rng), vary(woodC[2], 5, rng));
  } else {
    rect(ctx, ox + 15, bodyY - 3, 2, 3, vary(woodC[0] - 10, 6, rng), vary(woodC[1] - 8, 5, rng), vary(woodC[2], 5, rng));
  }
}

export function generateCarriageSpriteSheet(scene, carriageIndex, carriageType) {
  const key = `carriage_${carriageIndex}_${carriageType || 'cart'}`;
  if (scene.textures.exists(key)) return key;

  const dirs = 4;
  const canvas = document.createElement('canvas');
  canvas.width = S;
  canvas.height = S * dirs;
  const ctx = canvas.getContext('2d');

  for (let dir = 0; dir < dirs; dir++) {
    const fRng = simpleRng(carriageIndex * 113 + dir * 17 + 59);
    drawCarriageFrame(ctx, 0, dir * S, carriageType, dir, fRng);
  }

  scene.textures.addSpriteSheet(key, canvas, { frameWidth: S, frameHeight: S });
  return key;
}
