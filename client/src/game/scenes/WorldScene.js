import Phaser from 'phaser';
import { TS, TILE } from '../systems/tilesetGenerator.js';
import { generateTreeTextures, generateBuildingTextures } from '../systems/spriteGenerator.js';
import { generateTerrain } from '../systems/terrainGenerator.js';
import { NpcManager } from '../systems/npcManager.js';
import { MountManager } from '../systems/mountManager.js';
import { EnemyManager } from '../systems/enemyManager.js';
import { CameraController } from '../systems/cameraController.js';
import { WeatherSystem } from '../systems/weatherSystem.js';
import { ChunkManager } from '../systems/chunkManager.js';
import { useGameStore } from '../../hooks/useGameStore.js';
import { fbm } from '../../utils/noise.js';

const BUILDING_TYPES = new Set([
  'inn', 'market', 'shrine', 'forge', 'mill', 'home', 'library',
  'school', 'barracks', 'warehouse', 'workshop', 'tavern', 'mine',
  'farm', 'forest', 'dock', 'garden', 'well', 'ruin',
  'palace', 'castle', 'manor', 'stable', 'arena', 'cave', 'dungeon_entrance',
]);

const SECONDARY_BUILDINGS = ['house', 'house', 'stall', 'house', 'stable', 'house'];

function simpleRng(seed) {
  let s = seed | 0;
  return () => { s ^= s << 13; s ^= s >> 17; s ^= s << 5; return (s >>> 0) / 0xFFFFFFFF; };
}

export class WorldScene extends Phaser.Scene {
  constructor() {
    super('WorldScene');
    this.npcManager = null;
    this.mountManager = null;
    this.enemyManager = null;
    this.cameraController = null;
    this.weatherSystem = null;
    this._lastWorldRef = null;
    this._builtTilemap = false;
    this._locationOverlay = null;
    this._locationRects = [];
    this._buildingSprites = [];
    this._treeSprites = [];
    this._labels = [];
    this._labelLocIds = [];
    this._terrainData = null;
    this._lastZoomedNpcId = null;
    this._lastZoomedLocId = null;
    this.chunkManager = null;
  }

  create() {
    useGameStore.getState().setPhaserScene(this);

    this._locationOverlay = this.add.graphics();
    this._locationOverlay.setDepth(2);

    this.npcManager = new NpcManager(this);
    this.mountManager = new MountManager(this);
    this.enemyManager = new EnemyManager(this);
    this.cameraController = new CameraController(this);
    this.weatherSystem = new WeatherSystem(this);
    this.chunkManager = new ChunkManager(this);

    this.input.on('pointerup', (pointer) => {
      // Ignore events that didn't originate from the game canvas
      if (pointer.event && pointer.event.target !== this.game.canvas) return;
      if (this.npcManager?._npcClickedId) {
        this.npcManager._npcClickedId = null;
        return;
      }
      if (this.cameraController.dragMoved) return;
      this._handleMapClick(pointer);
    });

    this.input.on('pointermove', (pointer) => {
      this._handleHover(pointer);
    });
  }

  _handleMapClick(pointer) {
    const worldX = pointer.worldX;
    const worldY = pointer.worldY;

    const clickedNpc = this.npcManager?.sprites
      ? [...this.npcManager.sprites.entries()].find(([, sprite]) => {
          const halfW = (sprite.width || 16) / 2;
          const halfH = (sprite.height || 16) / 2;
          return worldX >= sprite.x - halfW && worldX <= sprite.x + halfW &&
                 worldY >= sprite.y - halfH && worldY <= sprite.y + halfH;
        })
      : null;

    if (clickedNpc) return;

    for (let i = this._locationRects.length - 1; i >= 0; i--) {
      const r = this._locationRects[i];
      if (worldX >= r.px && worldX < r.px + r.pw && worldY >= r.py && worldY < r.py + r.ph) {
        useGameStore.getState().selectLocation(r.locId);
        return;
      }
    }

    useGameStore.getState().clearSelection();
  }

  _handleHover(pointer) {
    const worldX = pointer.worldX;
    const worldY = pointer.worldY;
    const selectedLocId = useGameStore.getState().selectedLocation;

    let hoveredLocId = null;
    for (let i = this._locationRects.length - 1; i >= 0; i--) {
      const r = this._locationRects[i];
      if (worldX >= r.px && worldX < r.px + r.pw && worldY >= r.py && worldY < r.py + r.ph) {
        hoveredLocId = r.locId;
        break;
      }
    }

    for (let i = 0; i < this._labels.length; i++) {
      const label = this._labels[i];
      const locId = this._labelLocIds[i];
      label.setVisible(locId === hoveredLocId || locId === selectedLocId);
    }
  }

  update(time, delta) {
    const world = useGameStore.getState().world;

    if (world && world !== this._lastWorldRef) {
      this._lastWorldRef = world;

      if (!this._builtTilemap && world.gridW > 0) {
        this._buildTerrainFromWorld(world);
        this._builtTilemap = true;
        this.chunkManager.init(world.gridW, world.gridH);
        this.chunkManager.drawTerritoryBorders(world.territories);
        useGameStore.getState().setSceneReady(true);
      }

      this.chunkManager.clear();
      this._drawLocationOverlays(world);
      this._placeBuildings(world);
      this._placeTrees(world);
      this.chunkManager.updateMinimap(world);
    }

    this.chunkManager?.updateVisibility();
    this.npcManager?.update(time, delta);
    this.mountManager?.update(time, delta);
    this.enemyManager?.update(time, delta);

    const store = useGameStore.getState();
    const selectedNpc = store.selectedNpc;
    const selectedLoc = store.selectedLocation;

    if (selectedNpc && selectedNpc !== this._lastZoomedNpcId) {
      const npc = world?.npcs?.find(n => n.id === selectedNpc);
      if (npc?.alive) {
        const sprite = this.npcManager?.getNpcSprite(selectedNpc);
        if (sprite) {
          this.cameraController?.zoomTo(sprite.x, sprite.y);
          this._lastZoomedNpcId = selectedNpc;
        } else {
          // Sprite not rendered yet — zoom to NPC's location
          const loc = world?.locations?.find(l => l.id === npc.locationId);
          if (loc) {
            const cx = (loc.x + (loc.w || 1) / 2) * TS;
            const cy = (loc.y + (loc.h || 1) / 2) * TS;
            this.cameraController?.zoomTo(cx, cy);
            this._lastZoomedNpcId = selectedNpc;
          }
        }
      } else if (npc && !npc.alive) {
        this._lastZoomedNpcId = selectedNpc;
      }
    } else if (!selectedNpc) {
      this._lastZoomedNpcId = null;
    }

    if (selectedLoc && selectedLoc !== this._lastZoomedLocId) {
      const loc = world?.locations?.find(l => l.id === selectedLoc);
      if (loc) {
        const cx = (loc.x + (loc.w || 1) / 2) * TS;
        const cy = (loc.y + (loc.h || 1) / 2) * TS;
        this.cameraController?.zoomToLocation(cx, cy);
        this._lastZoomedLocId = selectedLoc;
      }
    } else if (!selectedLoc) {
      this._lastZoomedLocId = null;
    }

    if (selectedNpc) {
      const npc = world?.npcs?.find(n => n.id === selectedNpc);
      if (npc?.alive && store.followNpc !== selectedNpc) {
        store.setFollowNpc(selectedNpc);
      }
    } else if (store.followNpc) {
      store.setFollowNpc(null);
    }

    this.cameraController?.update(time, delta);
    this.weatherSystem?.update(time, delta);
  }

  _buildTerrainFromWorld(world) {
    const gridW = world.gridW || 30;
    const gridH = world.gridH || 30;

    const result = generateTerrain(gridW, gridH, world.locations || [], world.roads || null, world.territories || [], world.biomeOverrides || []);
    this._terrainData = result;
    this.npcManager?.setTerrainData(result);

    const map = this.make.tilemap({ data: result.tileData, tileWidth: TS, tileHeight: TS });
    const tileset = map.addTilesetImage('terrain', 'terrain', TS, TS, 0, 0);
    const layer = map.createLayer(0, tileset, 0, 0);
    layer.setDepth(0);

    this.tilemap = map;
    this.terrainLayer = layer;
  }

  _drawLocationOverlays(world) {
    const g = this._locationOverlay;
    g.clear();
    this._locationRects = [];

    this._labels.forEach(l => l.destroy());
    this._labels = [];
    this._labelLocIds = [];

    const selectedLocId = useGameStore.getState().selectedLocation;
    const locations = world.locations || [];
    const gridW = world.gridW || 30;
    const gridH = world.gridH || 30;
    const worldPxW = gridW * TS;
    const worldPxH = gridH * TS;

    // Track placed label bounding boxes for collision avoidance
    const placedLabels = [];

    for (const loc of locations) {
      const px = loc.x * TS;
      const py = loc.y * TS;
      const pw = (loc.w || 1) * TS;
      const ph = (loc.h || 1) * TS;

      const isSelected = loc.id === selectedLocId;

      if (isSelected) {
        const colorInt = loc.color ? parseInt(loc.color.replace('#', ''), 16) : 0x555577;
        g.fillStyle(colorInt, 0.15);
        g.fillRect(px, py, pw, ph);
        g.lineStyle(2, 0xffffff, 0.9);
        g.strokeRect(px, py, pw, ph);
      }

      this._locationRects.push({ locId: loc.id, px, py, pw, ph });

      const labelCenterX = px + pw / 2;
      let labelY = py + 3;
      let labelOriginY = 0;

      if (py < TS) {
        labelY = py + ph - 3;
        labelOriginY = 1;
      }

      let labelX = Math.max(4, Math.min(labelCenterX, worldPxW - 4));

      const label = this.add.text(labelX, labelY, loc.name, {
        fontSize: '10px',
        fontFamily: 'monospace',
        color: '#dddddd',
        stroke: '#000000',
        strokeThickness: 2,
        align: 'center',
      });
      label.setOrigin(0.5, labelOriginY);
      label.setDepth(20);

      const halfW = label.width / 2;
      const lh = label.height || 12;
      if (labelX - halfW < 2) label.setX(halfW + 2);
      if (labelX + halfW > worldPxW - 2) label.setX(worldPxW - halfW - 2);

      // Nudge label if it overlaps a previously placed label
      let finalX = label.x;
      let finalY = labelY;
      for (let attempt = 0; attempt < 6; attempt++) {
        let overlaps = false;
        for (const prev of placedLabels) {
          const ox = Math.abs(finalX - prev.x);
          const oy = Math.abs(finalY - prev.y);
          if (ox < (halfW + prev.hw + 4) && oy < (lh / 2 + prev.hh + 2)) {
            overlaps = true;
            break;
          }
        }
        if (!overlaps) break;
        finalY += lh + 2; // shift down
        if (finalY > worldPxH - 4) { finalY = labelY - lh - 2; break; }
      }
      label.setY(finalY);
      placedLabels.push({ x: finalX, y: finalY, hw: halfW, hh: lh / 2 });

      label.setVisible(isSelected);
      this._labels.push(label);
      this._labelLocIds.push(loc.id);
    }
  }

  _placeBuildings(world) {
    this._buildingSprites.forEach(s => s.destroy());
    this._buildingSprites = [];

    const locations = world.locations || [];
    const rng = simpleRng(55555);
    const buildings = [];

    // Multi-tile building types that fill their grid allocation
    const MULTI_TILE_TYPES = new Set(['farm', 'forest', 'mine', 'market', 'inn', 'dock', 'tavern', 'library', 'workshop', 'garden', 'warehouse', 'barracks', 'palace', 'castle', 'manor', 'stable', 'arena']);

    for (const loc of locations) {
      const hasBuildingType = loc.buildingType || BUILDING_TYPES.has(loc.type);
      if (!hasBuildingType) continue;

      const locW = loc.w || 1;
      const locH = loc.h || 1;
      const btype = loc.buildingType || loc.type;
      const isMultiTile = MULTI_TILE_TYPES.has(btype) && locW * locH > 1;

      buildings.push({
        type: btype,
        tx: loc.x, tz: loc.y,
        gridW: isMultiTile ? locW : 1,
        gridH: isMultiTile ? locH : 1,
        name: loc.name, locId: loc.id,
        _index: buildings.length,
        _worldW: world.gridW || 30,
        _worldH: world.gridH || 30,
      });

      // Only add secondary filler buildings for large locations that don't fill their grid
      const area = locW * locH;
      if (area >= 4 && !isMultiTile) {
        const cx = loc.x + Math.floor(locW / 2);
        const cy = loc.y + Math.floor(locH / 2);
        const extra = Math.min(4, Math.floor(area / 3));
        for (let e = 0; e < extra; e++) {
          const stype = SECONDARY_BUILDINGS[e % SECONDARY_BUILDINGS.length];
          const ox = loc.x + Math.floor(rng() * locW);
          const oy = loc.y + Math.floor(rng() * locH);
          if (ox === cx && oy === cy) continue;
          buildings.push({
            type: stype, tx: ox, tz: oy,
            gridW: 1, gridH: 1,
            name: stype, locId: loc.id,
            _index: buildings.length,
          });
        }
      }
    }

    if (buildings.length === 0) return;

    generateBuildingTextures(this, buildings);

    this._buildingSprites = buildings.map((b, i) => {
      const key = `bld_${b.type}_${i}`;
      if (!this.textures.exists(key)) return null;
      const gw = b.gridW || 1;
      const gh = b.gridH || 1;
      const sprite = this.add.sprite(
        b.tx * TS + (gw * TS) / 2,
        b.tz * TS + (gh * TS) / 2,
        key
      );
      sprite.setDepth(5);
      sprite.setVisible(false); // Start hidden; chunk manager will show visible ones
      sprite.setInteractive({ useHandCursor: true });

      sprite.on('pointerdown', (pointer) => {
        if (pointer.event && pointer.event.target !== this.game.canvas) return;
        useGameStore.getState().selectLocation(b.locId);
      });

      this.chunkManager?.addSprite(sprite, b.tx, b.tz);
      return sprite;
    }).filter(Boolean);
  }

  _placeTrees(world) {
    this._treeSprites.forEach(s => s.destroy());
    this._treeSprites = [];

    generateTreeTextures(this);

    const locations = world.locations || [];
    const gridW = world.gridW || 30;
    const gridH = world.gridH || 30;
    const rng = simpleRng(7777);

    // Forest locations already draw their own trees in the sprite — skip placing
    // additional tree sprites on them to avoid visual overlap.

    const locSet = new Set();
    for (const loc of locations) {
      // Forest locations draw their own trees — use a wider exclusion buffer
      // to prevent terrain trees from overlapping with the forest sprite edges
      const buf = (loc.type === 'forest') ? 2 : 1;
      for (let dy = -buf; dy <= (loc.h || 1) + (buf - 1); dy++) {
        for (let dx = -buf; dx <= (loc.w || 1) + (buf - 1); dx++) {
          locSet.add(`${loc.x + dx},${loc.y + dy}`);
        }
      }
    }

    const tileData2 = this._terrainData?.tileData;
    const roadSet = this._terrainData?.roadSet;
    const treeTiles = new Set([TILE.FOREST_FLOOR, TILE.DARK_GRASS, TILE.GRASS, TILE.SAND, TILE.COASTAL_SAND, TILE.DEEP_FOREST, TILE.DESERT_SAND, TILE.OASIS]);

    for (let y = 0; y < gridH; y++) {
      for (let x = 0; x < gridW; x++) {
        if (locSet.has(`${x},${y}`)) continue;
        if (roadSet?.has(`${x},${y}`)) continue;

        const tile = tileData2 ? tileData2[y][x] : TILE.GRASS;
        if (!treeTiles.has(tile)) continue;

        const density = fbm(x * 0.3, y * 0.3, 3) * 0.5 + 0.5;
        let threshold = 0.72;
        if (tile === TILE.FOREST_FLOOR) threshold = 0.35;
        else if (tile === TILE.DEEP_FOREST) threshold = 0.25;
        else if (tile === TILE.DARK_GRASS) threshold = 0.55;
        else if (tile === TILE.SAND || tile === TILE.COASTAL_SAND) threshold = 0.85;
        else if (tile === TILE.DESERT_SAND) threshold = 0.92;
        else if (tile === TILE.OASIS) threshold = 0.30;

        if (density < threshold) continue;
        if (rng() > 0.6) continue;

        let sp;
        if (tile === TILE.SAND || tile === TILE.COASTAL_SAND || tile === TILE.DESERT_SAND) {
          sp = rng() < 0.8 ? 'palm' : 'oak';
        } else if (tile === TILE.OASIS) {
          sp = 'palm';
        } else if (tile === TILE.FOREST_FLOOR || tile === TILE.DEEP_FOREST) {
          sp = rng() < 0.6 ? 'pine' : 'oak';
        } else {
          sp = rng() < 0.7 ? 'oak' : 'pine';
        }
        const variant = Math.floor(rng() * 4);
        const key = `tree_${sp}_${variant}`;

        const sprite = this.add.sprite(
          x * TS + TS * rng(),
          y * TS + TS * rng(),
          key
        );
        sprite.setDepth(4);
        sprite.setVisible(false); // Start hidden; chunk manager will show visible ones
        this._treeSprites.push(sprite);
        this.chunkManager?.addSprite(sprite, x, y);
      }
    }
  }
}
