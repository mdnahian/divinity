import { TS, TILE } from './tilesetGenerator.js';
import { generateNpcSpriteSheet } from './spriteGenerator.js';
import { useGameStore } from '../../hooks/useGameStore.js';
import { MOUNT_OFFSET_Y } from './mountManager.js';

const SEP_DIST = TS * 0.7;
const SEP_FORCE = 0.15;
const LERP_SPEED = 0.08;
const CULL_PADDING = TS * 8; // pixels of padding around camera viewport for NPC culling
const NPC_ZOOM_THRESHOLD = 0.45; // hide all NPC sprites below this zoom level

const ACTION_CATEGORY_MAP = {
  forage: 'gather', hunt: 'gather', farm: 'gather', fish: 'gather',
  chop_wood: 'gather', mine_stone: 'gather', mine_ore: 'gather',
  gather_thatch: 'gather', gather_clay: 'gather', scavenge: 'gather',
  cook: 'craft', brew_ale: 'craft', brew_potion: 'craft',
  smelt_ore: 'craft', forge_weapon: 'craft', forge_tool: 'craft',
  tan_hide: 'craft', tailor_craft: 'craft', mill_grain: 'craft',
  bake_bread_adv: 'craft', craft_pottery: 'craft', write_technique: 'craft',
  write_journal: 'craft', copy_text: 'craft',
  talk: 'social', gift: 'social', eat_together: 'social',
  drink_together: 'social', work_together: 'social', comfort: 'social',
  flirt: 'social', recruit_to_faction: 'social', share_journal: 'social',
  attack_enemy: 'combat', flee_area: 'combat', party_attack: 'combat',
  steal: 'combat', fight: 'combat',
  sleep: 'sleep', rest: 'sleep',
  eat: 'survival', drink: 'survival', bathe: 'survival', drink_ale: 'survival',
  heal: 'wellbeing', pray: 'wellbeing', read_book: 'wellbeing',
  trade: 'economy', buy_ale: 'economy', buy_food: 'economy',
  buy_supplies: 'economy', serve_customer: 'economy',
  heal_patient: 'economy', offer_counsel: 'economy',
  explore: 'movement', go_home: 'movement',
  begin_construction: 'construction', repair_building: 'construction',
  commission_construction: 'construction',
  teach: 'education', teach_literacy: 'education', teach_technique: 'education',
  hire_employee: 'economy', fire_employee: 'economy',
  quit_job: 'economy', seek_employment: 'economy', start_business: 'economy',
  equip_item: 'equipment', unequip_item: 'equipment',
  repair_metal: 'craft', repair_clothing: 'craft',
};

function _isWater(tile) {
  return tile === TILE.DEEP_WATER || tile === TILE.SHALLOW_WATER;
}

function _isBlocked(tile) {
  return tile === TILE.DEEP_WATER || tile === TILE.SHALLOW_WATER ||
         tile === TILE.MOUNTAIN || tile === TILE.SNOW ||
         tile === TILE.LAVA || tile === TILE.VOLCANIC;
}

// Location types where NPCs can walk freely inside (natural/open areas)
const WALKABLE_LOCATION_TYPES = new Set([
  'forest', 'farm', 'field', 'dock', 'garden', 'well', 'ruin',
]);

// Building types where NPCs should stand in front, not on top
const BUILDING_LOCATION_TYPES = new Set([
  'inn', 'market', 'shrine', 'forge', 'mill', 'home', 'library',
  'school', 'barracks', 'warehouse', 'workshop', 'tavern', 'mine',
]);

function _computeLandPath(fromPx, toPx, tileData, gridW, gridH) {
  const sx = Math.floor(fromPx.x / TS), sy = Math.floor(fromPx.y / TS);
  const ex = Math.floor(toPx.x / TS), ey = Math.floor(toPx.y / TS);
  const clampX = v => Math.max(0, Math.min(gridW - 1, v));
  const clampY = v => Math.max(0, Math.min(gridH - 1, v));
  const startX = clampX(sx), startY = clampY(sy);
  const endX = clampX(ex), endY = clampY(ey);

  if (!tileData || startX === endX && startY === endY) {
    return [fromPx, toPx];
  }

  let crossesBlocked = false;
  const steps = Math.max(Math.abs(endX - startX), Math.abs(endY - startY));
  for (let i = 0; i <= steps; i++) {
    const t = steps > 0 ? i / steps : 0;
    const cx = Math.round(startX + (endX - startX) * t);
    const cy = Math.round(startY + (endY - startY) * t);
    if (tileData[cy] && _isBlocked(tileData[cy][cx])) {
      crossesBlocked = true;
      break;
    }
  }
  if (!crossesBlocked) return [fromPx, toPx];

  const key = (x, y) => y * gridW + x;
  const visited = new Uint8Array(gridW * gridH);
  const from = new Int32Array(gridW * gridH).fill(-1);
  const queue = [key(startX, startY)];
  visited[key(startX, startY)] = 1;
  let found = false;

  while (queue.length > 0) {
    const cur = queue.shift();
    const cx = cur % gridW, cy = (cur - cx) / gridW;
    if (cx === endX && cy === endY) { found = true; break; }
    for (const [dx, dy] of [[-1, 0], [1, 0], [0, -1], [0, 1], [-1, -1], [1, -1], [-1, 1], [1, 1]]) {
      const nx = cx + dx, ny = cy + dy;
      if (nx < 0 || nx >= gridW || ny < 0 || ny >= gridH) continue;
      const nk = key(nx, ny);
      if (visited[nk]) continue;
      if (tileData[ny] && _isBlocked(tileData[ny][nx])) continue;
      visited[nk] = 1;
      from[nk] = cur;
      queue.push(nk);
    }
  }

  if (!found) return [fromPx, toPx];

  const path = [];
  let c = key(endX, endY);
  while (c !== -1 && c !== key(startX, startY)) {
    const cx2 = c % gridW, cy2 = (c - cx2) / gridW;
    path.push({ x: (cx2 + 0.5) * TS, y: (cy2 + 0.5) * TS });
    c = from[c];
  }
  path.push(fromPx);
  path.reverse();
  path.push(toPx);

  if (path.length > 6) {
    const simplified = [path[0]];
    const step = Math.floor(path.length / 5);
    for (let i = step; i < path.length - 1; i += step) {
      simplified.push(path[i]);
    }
    simplified.push(path[path.length - 1]);
    return simplified;
  }
  return path;
}

export class NpcManager {
  constructor(scene) {
    this.scene = scene;
    this.sprites = new Map();
    this.npcMeta = new Map();
    this._npcIndexMap = new Map();
    this._nextIndex = 0;
    this._followIndicator = null;
    this._followTween = null;
    this._actionIcons = new Map();
    this._npcClickedId = null;
    this._travelLine = null;
    this._terrainData = null;
    this._pathCache = new Map();
  }

  setTerrainData(td) { this._terrainData = td; }

  _getNpcIndex(npcId) {
    if (!this._npcIndexMap.has(npcId)) {
      this._npcIndexMap.set(npcId, this._nextIndex++);
    }
    return this._npcIndexMap.get(npcId);
  }

  _ensureSprite(npc) {
    if (this.sprites.has(npc.id)) return this.sprites.get(npc.id);

    const idx = this._getNpcIndex(npc.id);
    const prof = npc.profession || 'default';
    const texKey = generateNpcSpriteSheet(this.scene, idx, npc.profession);
    const animPrefix = `npc_${idx}_${prof}`;

    const dirs = ['down', 'up', 'right', 'left'];
    dirs.forEach((dir, d) => {
      const walkKey = `${animPrefix}_walk_${dir}`;
      if (!this.scene.anims.exists(walkKey)) {
        this.scene.anims.create({
          key: walkKey,
          frames: this.scene.anims.generateFrameNumbers(texKey, { start: d * 4, end: d * 4 + 3 }),
          frameRate: 6,
          repeat: -1,
        });
      }
      const idleKey = `${animPrefix}_idle_${dir}`;
      if (!this.scene.anims.exists(idleKey)) {
        this.scene.anims.create({
          key: idleKey,
          frames: [{ key: texKey, frame: d * 4 }],
          frameRate: 1,
        });
      }
    });

    const sprite = this.scene.add.sprite(0, 0, texKey);
    sprite.setDepth(10);
    sprite.setInteractive({ useHandCursor: true });
    sprite.play(`${animPrefix}_idle_down`);

    sprite.on('pointerdown', (pointer) => {
      if (pointer.event && pointer.event.target !== this.scene.game.canvas) return;
      useGameStore.getState().selectNpc(npc.id);
      this._npcClickedId = npc.id;
    });

    this.sprites.set(npc.id, sprite);
    this.npcMeta.set(npc.id, { prevX: null, prevY: null, dir: 'down', idx, animPrefix, traveling: false });
    return sprite;
  }

  _ensureFollowIndicator() {
    if (this._followIndicator) return this._followIndicator;

    const g = this.scene.add.graphics();
    g.setDepth(15);
    g.lineStyle(1, 0xff3333, 0.85);
    g.strokeCircle(0, 0, 4);
    g.fillStyle(0xff3333, 0.7);
    g.fillTriangle(-1.5, -6, 1.5, -6, 0, -8);
    g.setVisible(false);
    this._followIndicator = g;
    return g;
  }

  _locCenter(loc) {
    return {
      x: (loc.x + (loc.w || 1) * 0.5) * TS,
      y: (loc.y + (loc.h || 1) * 0.5) * TS,
    };
  }

  _drawActionIcon(category) {
    const g = this.scene.add.graphics();
    g.setDepth(16);
    const s = 4;
    switch (category) {
      case 'gather':
        g.lineStyle(1, 0xccaa44, 0.9);
        g.lineBetween(-s, s, 0, -s);
        g.lineBetween(0, -s, s * 0.3, -s * 0.6);
        g.lineBetween(0, -s, -s * 0.3, -s * 0.6);
        break;
      case 'craft':
        g.fillStyle(0xbb8833, 0.9);
        g.fillRect(-1, -s, 2, s * 2);
        g.fillRect(-s, -s, s * 2, 2);
        break;
      case 'social':
        g.fillStyle(0x55aaff, 0.9);
        g.fillRoundedRect(-s, -s, s * 2, s * 1.4, 2);
        g.fillTriangle(-s + 1, s * 0.4, -s + 1, s, -s + 3, s * 0.4);
        break;
      case 'combat':
        g.lineStyle(1.5, 0xff4444, 0.9);
        g.lineBetween(-s, -s, s, s);
        g.lineBetween(-s, s, s, -s);
        break;
      case 'sleep':
        g.fillStyle(0x9999dd, 0.85);
        const txt = this.scene.add.text(0, 0, 'z', {
          fontSize: '7px', fontFamily: 'monospace',
          color: '#aaaaee', stroke: '#222244', strokeThickness: 1,
        }).setOrigin(0.5, 0.5).setDepth(16);
        g._zzText = txt;
        break;
      case 'economy':
        g.fillStyle(0xffcc33, 0.9);
        g.fillCircle(0, 0, s * 0.7);
        g.fillStyle(0xaa8800, 0.9);
        g.fillCircle(0, 0, s * 0.3);
        break;
      case 'movement':
        g.fillStyle(0x88cc88, 0.9);
        g.fillTriangle(0, -s, -s * 0.6, s * 0.5, s * 0.6, s * 0.5);
        break;
      case 'construction':
        g.fillStyle(0xcc8844, 0.9);
        g.fillRect(-1, -s, 2, s * 1.5);
        g.fillRect(-s * 0.8, -s, s * 1.6, 2);
        break;
      case 'education':
        g.fillStyle(0x44aacc, 0.9);
        g.fillRect(-s * 0.7, -s, s * 1.4, s * 1.8);
        g.lineStyle(0.5, 0x226688, 0.9);
        g.lineBetween(-s * 0.3, -s * 0.5, s * 0.3, -s * 0.5);
        g.lineBetween(-s * 0.3, 0, s * 0.3, 0);
        break;
      default:
        g.fillStyle(0xcccccc, 0.7);
        g.fillCircle(0, 0, 2);
        break;
    }
    return g;
  }

  update(time, delta) {
    const state = useGameStore.getState();
    const world = state.world;
    if (!world || !world.npcs) return;

    const selectedNpc = state.selectedNpc;
    const followNpc = state.followNpc;
    const tickCount = state.tickCount || 0;
    const locations = world.locations || [];
    const locMap = {};
    for (const loc of locations) locMap[loc.id] = loc;

    // Camera viewport for culling off-screen NPCs
    const cam = this.scene.cameras.main;
    const zoom = cam.zoom;
    const view = cam.worldView;
    const cullLeft = view.x - CULL_PADDING;
    const cullRight = view.x + view.width + CULL_PADDING;
    const cullTop = view.y - CULL_PADDING;
    const cullBottom = view.y + view.height + CULL_PADDING;

    // When zoomed out far, NPCs are sub-pixel — hide them all and skip processing
    if (zoom < NPC_ZOOM_THRESHOLD) {
      for (const [, sprite] of this.sprites) sprite.setVisible(false);
      for (const [, icon] of this._actionIcons) {
        icon.setVisible(false);
        if (icon._zzText) icon._zzText.setVisible(false);
      }
      this._ensureFollowIndicator().setVisible(false);
      if (this._travelLine) this._travelLine.clear();
      return;
    }

    const locNpcCounts = {};
    for (const npc of world.npcs) {
      const effectiveLocId = npc.targetLocationId && npc.travelArrivalTick > tickCount
        ? npc.previousLocationId || npc.locationId
        : npc.locationId;
      locNpcCounts[effectiveLocId] = (locNpcCounts[effectiveLocId] || 0) + 1;
    }

    const locPosIndex = {};
    const seenIds = new Set();
    const visibleSprites = [];

    for (const npc of world.npcs) {
      if (!npc.alive) continue;

      seenIds.add(npc.id);

      const sprite = this._ensureSprite(npc);
      const meta = this.npcMeta.get(npc.id);
      const idx = meta.idx;

      sprite.setAlpha(1);
      sprite.clearTint();

      let targetX, targetY;
      const isTraveling = npc.targetLocationId &&
        npc.travelArrivalTick > 0 &&
        npc.locationId !== npc.targetLocationId;

      if (isTraveling) {
        const fromLoc = locMap[npc.previousLocationId || npc.locationId];
        const toLoc = locMap[npc.targetLocationId];

        if (fromLoc && toLoc) {
          const from = this._locCenter(fromLoc);
          const to = this._locCenter(toLoc);
          const tickMs = world.tickIntervalMs || 6000;

          if (!meta.traveling) {
            meta.traveling = true;
            const td = this._terrainData;
            if (td?.tileData) {
              const gridW = world.gridW || td.tileData[0]?.length || 30;
              const gridH = world.gridH || td.tileData.length || 30;
              meta.travelPath = _computeLandPath(from, to, td.tileData, gridW, gridH);
            } else {
              meta.travelPath = [from, to];
            }
          }

          // Tick-based progress (survives page refresh)
          const startTick = npc.travelStartTick || 0;
          const totalTicks = Math.max(1, npc.travelArrivalTick - startTick);
          const elapsedTicks = Math.max(0, tickCount - startTick);
          const tickProgress = Math.min(1, elapsedTicks / totalTicks);

          // Sub-tick smoothing for fluid animation between ticks
          if (tickCount !== meta._lastTickCount) {
            meta._lastTickTime = Date.now();
            meta._lastTickCount = tickCount;
          }
          const msSinceLastTick = Date.now() - (meta._lastTickTime || Date.now());
          const subTickFraction = Math.min(1, msSinceLastTick / tickMs) / totalTicks;
          const progress = Math.min(1, tickProgress + subTickFraction);

          const waypoints = meta.travelPath || [from, to];
          const totalLen = this._waypointLength(waypoints);
          const targetDist = progress * totalLen;
          let accumulated = 0;
          targetX = waypoints[waypoints.length - 1].x;
          targetY = waypoints[waypoints.length - 1].y;
          for (let wi = 0; wi < waypoints.length - 1; wi++) {
            const segLen = Math.hypot(waypoints[wi + 1].x - waypoints[wi].x, waypoints[wi + 1].y - waypoints[wi].y);
            if (accumulated + segLen >= targetDist) {
              const t = segLen > 0 ? (targetDist - accumulated) / segLen : 0;
              targetX = waypoints[wi].x + (waypoints[wi + 1].x - waypoints[wi].x) * t;
              targetY = waypoints[wi].y + (waypoints[wi + 1].y - waypoints[wi].y) * t;
              break;
            }
            accumulated += segLen;
          }
        } else {
          const loc = locMap[npc.locationId] || locations[0];
          if (!loc) continue;
          const c = this._locCenter(loc);
          targetX = c.x;
          targetY = c.y;
        }
      } else {
        meta.traveling = false;
        meta.travelPath = null;
        meta.travelStartTime = null;
        meta.travelDurationMs = null;
        const loc = locMap[npc.locationId] || locations[0];
        if (!loc) continue;

        locPosIndex[loc.id] = (locPosIndex[loc.id] || 0);
        const posIdx = locPosIndex[loc.id]++;

        const locW = loc.w || 1;
        const locH = loc.h || 1;
        const isBuilding = BUILDING_LOCATION_TYPES.has(loc.type);

        if (isBuilding) {
          // Stand in front of the building (below it), spread horizontally
          const frontY = (loc.y + locH) * TS - TS * 0.2;
          const centerX = (loc.x + locW * 0.5) * TS;
          const spread = Math.min(locW * TS * 0.8, TS * 3);
          const npcCount = locNpcCounts[loc.id] || 1;
          const offset = npcCount > 1 ? (posIdx / (npcCount - 1) - 0.5) * spread : 0;
          targetX = centerX + offset;
          targetY = frontY + Math.floor(posIdx / 4) * TS * 0.2;
        } else {
          // Natural/open areas: spread inside the location area
          targetX = (loc.x + locW * 0.5) * TS + (posIdx % 3 - 1) * TS * 0.28;
          targetY = (loc.y + locH * 0.5) * TS + Math.floor(posIdx / 3) * TS * 0.3 + TS * 0.2;
        }
      }

      let cx, cy;
      if (isTraveling) {
        cx = targetX;
        cy = targetY;
      } else if (meta.prevX != null) {
        cx = meta.prevX + (targetX - meta.prevX) * LERP_SPEED;
        cy = meta.prevY + (targetY - meta.prevY) * LERP_SPEED;
        if (Math.abs(targetX - cx) < 0.5 && Math.abs(targetY - cy) < 0.5) {
          cx = targetX;
          cy = targetY;
        }
      } else {
        cx = targetX;
        cy = targetY;
      }

      // Cull off-screen NPCs: snap position but skip animation/icon/separation
      const onScreen = cx >= cullLeft && cx <= cullRight && cy >= cullTop && cy <= cullBottom;
      if (!onScreen) {
        sprite.setVisible(false);
        sprite.setPosition(cx, cy);
        meta.prevX = cx;
        meta.prevY = cy;
        // Hide action icon if it exists
        const offIcon = this._actionIcons.get(npc.id);
        if (offIcon) {
          offIcon.setVisible(false);
          if (offIcon._zzText) offIcon._zzText.setVisible(false);
        }
        continue;
      }

      sprite.setVisible(true);

      const dx = cx - (meta.prevX ?? cx);
      const dy = cy - (meta.prevY ?? cy);
      const moving = Math.abs(dx) > 0.3 || Math.abs(dy) > 0.3;

      if (moving && npc.alive) {
        let newDir;
        if (Math.abs(dx) > Math.abs(dy)) {
          newDir = dx > 0 ? 'right' : 'left';
        } else {
          newDir = dy > 0 ? 'down' : 'up';
        }
        if (newDir !== meta.dir || !sprite.anims.currentAnim?.key?.includes('walk')) {
          meta.dir = newDir;
          sprite.play(`${meta.animPrefix}_walk_${newDir}`, true);
        }
      } else if (npc.alive && sprite.anims.currentAnim?.key?.includes('walk')) {
        sprite.play(`${meta.animPrefix}_idle_${meta.dir}`, true);
      }

      // Shift NPC up when mounted and traveling so they appear seated on the horse
      const isTravelingMounted = npc.mountId && npc.targetLocationId && npc.travelArrivalTick > 0 &&
        (world.mounts || []).some(m => m.id === npc.mountId && m.alive);
      const displayY = isTravelingMounted ? cy - MOUNT_OFFSET_Y : cy;

      sprite.setPosition(cx, displayY);
      meta.prevX = cx;
      meta.prevY = cy;

      sprite.setDepth(npc.id === selectedNpc ? 12 : 10);

      const isBusy = npc.pendingActionId && npc.busyUntilTick > tickCount && !isTraveling;
      const actionCategory = isBusy ? (ACTION_CATEGORY_MAP[npc.pendingActionId] || 'default') : null;
      let icon = this._actionIcons.get(npc.id);

      if (isBusy) {
        if (!icon || icon._category !== actionCategory) {
          if (icon) {
            icon._zzText?.destroy();
            icon.destroy();
          }
          icon = this._drawActionIcon(actionCategory);
          icon._category = actionCategory;
          this._actionIcons.set(npc.id, icon);
        }
        const bob = Math.sin(time * 0.005 + cx) * 1.5;
        icon.setPosition(cx, cy - 12 + bob);
        icon.setVisible(true);
        if (icon._zzText) {
          icon._zzText.setPosition(cx, cy - 12 + bob);
          icon._zzText.setVisible(true);
        }
      } else if (icon) {
        icon.setVisible(false);
        if (icon._zzText) icon._zzText.setVisible(false);
      }

      visibleSprites.push(sprite);
    }

    this._applySeparation(visibleSprites);

    if (!this._travelLine) {
      this._travelLine = this.scene.add.graphics();
      this._travelLine.setDepth(9);
    }
    this._travelLine.clear();
    if (selectedNpc) {
      const selNpc = (world.npcs || []).find(n => n.id === selectedNpc);
      const selSprite = this.sprites.get(selectedNpc);
      if (selNpc && selSprite &&
          selNpc.targetLocationId &&
          selNpc.travelArrivalTick > 0 &&
          selNpc.travelArrivalTick > tickCount) {
        const toLoc = locMap[selNpc.targetLocationId];
        if (toLoc) {
          const to = this._locCenter(toLoc);
          this._travelLine.lineStyle(2, 0xff0000, 0.9);
          this._travelLine.lineBetween(selSprite.x, selSprite.y, to.x, to.y);
        }
      }
    }

    const indicator = this._ensureFollowIndicator();
    const indicatorTarget = selectedNpc || followNpc;
    if (indicatorTarget != null) {
      const targetSprite = this.sprites.get(indicatorTarget);
      if (targetSprite) {
        const bob = Math.sin(time * 0.004) * 1.5;
        indicator.setPosition(targetSprite.x, targetSprite.y - 10 + bob);
        indicator.setVisible(true);
      } else {
        indicator.setVisible(false);
      }
    } else {
      indicator.setVisible(false);
    }

    for (const [id, sprite] of this.sprites) {
      if (!seenIds.has(id)) {
        sprite.destroy();
        this.sprites.delete(id);
        this.npcMeta.delete(id);
        const icon = this._actionIcons.get(id);
        if (icon) {
          icon._zzText?.destroy();
          icon.destroy();
          this._actionIcons.delete(id);
        }
      }
    }
  }

  _waypointLength(waypoints) {
    let len = 0;
    for (let i = 0; i < waypoints.length - 1; i++) {
      len += Math.hypot(waypoints[i + 1].x - waypoints[i].x, waypoints[i + 1].y - waypoints[i].y);
    }
    return len;
  }


  _applySeparation(sprites) {
    const sepDist2 = SEP_DIST * SEP_DIST;
    for (let i = 0; i < sprites.length; i++) {
      const a = sprites[i];
      for (let j = i + 1; j < sprites.length; j++) {
        const b = sprites[j];
        const dx = a.x - b.x;
        const dy = a.y - b.y;
        const d2 = dx * dx + dy * dy;
        if (d2 < sepDist2 && d2 > 0.01) {
          const d = Math.sqrt(d2);
          const push = (SEP_DIST - d) * SEP_FORCE;
          const nx = (dx / d) * push;
          const ny = (dy / d) * push;
          a.x += nx;
          a.y += ny;
          b.x -= nx;
          b.y -= ny;
        }
      }
    }
  }

  getNpcSprite(npcId) {
    return this.sprites.get(npcId) || null;
  }

  destroy() {
    for (const s of this.sprites.values()) s.destroy();
    this.sprites.clear();
    this.npcMeta.clear();
    this._followIndicator?.destroy();
    this._followIndicator = null;
    this._travelLine?.destroy();
    this._travelLine = null;
    for (const icon of this._actionIcons.values()) {
      icon._zzText?.destroy();
      icon.destroy();
    }
    this._actionIcons.clear();
  }
}
