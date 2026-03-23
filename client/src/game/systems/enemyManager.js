import { TS } from './tilesetGenerator.js';
import { generateEnemySpriteSheet } from './spriteGenerator.js';
import { useGameStore } from '../../hooks/useGameStore.js';

const LERP_SPEED = 0.06;
const SEP_DIST = TS * 0.6;
const SEP_FORCE = 0.12;

const BUILDING_LOCATION_TYPES = new Set([
  'inn', 'market', 'shrine', 'forge', 'mill', 'home', 'library',
  'school', 'barracks', 'warehouse', 'workshop', 'tavern', 'mine',
]);

export class EnemyManager {
  constructor(scene) {
    this.scene = scene;
    this.sprites = new Map();
    this.enemyMeta = new Map();
    this._enemyIndexMap = new Map();
    this._nextIndex = 0;
    this._healthBars = new Map();
  }

  _getEnemyIndex(enemyId) {
    if (!this._enemyIndexMap.has(enemyId)) {
      this._enemyIndexMap.set(enemyId, this._nextIndex++);
    }
    return this._enemyIndexMap.get(enemyId);
  }

  _locCenter(loc) {
    return {
      x: (loc.x + (loc.w || 1) * 0.5) * TS,
      y: (loc.y + (loc.h || 1) * 0.5) * TS,
    };
  }

  _ensureSprite(enemy) {
    if (this.sprites.has(enemy.id)) return this.sprites.get(enemy.id);

    const idx = this._getEnemyIndex(enemy.id);
    const enemyType = enemy.type || 'wolf';
    const texKey = generateEnemySpriteSheet(this.scene, enemyType, idx);
    const animPrefix = `enemy_${enemyType}_${idx}`;

    const dirs = ['down', 'up', 'right', 'left'];
    dirs.forEach((dir, d) => {
      const walkKey = `${animPrefix}_walk_${dir}`;
      if (!this.scene.anims.exists(walkKey)) {
        this.scene.anims.create({
          key: walkKey,
          frames: this.scene.anims.generateFrameNumbers(texKey, { start: d * 4, end: d * 4 + 3 }),
          frameRate: 5,
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
      // Select the location this enemy is at
      useGameStore.getState().selectLocation(enemy.locationId);
    });

    this.sprites.set(enemy.id, sprite);
    this.enemyMeta.set(enemy.id, { prevX: null, prevY: null, dir: 'down', animPrefix });
    return sprite;
  }

  _ensureHealthBar(enemyId) {
    if (this._healthBars.has(enemyId)) return this._healthBars.get(enemyId);
    const g = this.scene.add.graphics();
    g.setDepth(14);
    this._healthBars.set(enemyId, g);
    return g;
  }

  _drawHealthBar(enemy, sprite) {
    const bar = this._ensureHealthBar(enemy.id);
    bar.clear();

    const barW = 16;
    const barH = 2;
    const x = sprite.x - barW / 2;
    const y = sprite.y - 12;

    // Background
    bar.fillStyle(0x333333, 0.8);
    bar.fillRect(x, y, barW, barH);

    // Health fill
    const ratio = Math.max(0, enemy.hp / enemy.maxHp);
    const color = ratio > 0.5 ? 0xcc4444 : (ratio > 0.25 ? 0xcc8844 : 0xcc2222);
    bar.fillStyle(color, 0.9);
    bar.fillRect(x, y, barW * ratio, barH);
  }

  update(time, delta) {
    const state = useGameStore.getState();
    const world = state.world;
    if (!world) return;

    const enemies = world.aliveEnemies || [];
    const locations = world.locations || [];
    const locMap = {};
    for (const loc of locations) locMap[loc.id] = loc;

    // Count enemies per location for spread positioning
    const locEnemyCounts = {};
    for (const e of enemies) {
      locEnemyCounts[e.locationId] = (locEnemyCounts[e.locationId] || 0) + 1;
    }

    const locPosIndex = {};
    const seenIds = new Set();

    for (const enemy of enemies) {
      seenIds.add(enemy.id);

      const sprite = this._ensureSprite(enemy);
      const meta = this.enemyMeta.get(enemy.id);

      const loc = locMap[enemy.locationId];
      if (!loc) continue;

      // Calculate position within location
      locPosIndex[loc.id] = (locPosIndex[loc.id] || 0);
      const posIdx = locPosIndex[loc.id]++;

      const locW = loc.w || 1;
      const locH = loc.h || 1;
      const isBuilding = BUILDING_LOCATION_TYPES.has(loc.type);

      let targetX, targetY;
      if (isBuilding) {
        // Enemies lurk further below and to the side of buildings
        const frontY = (loc.y + locH) * TS + TS * 0.8;
        const centerX = (loc.x + locW * 0.5) * TS;
        const count = locEnemyCounts[loc.id] || 1;
        const spread = Math.min(locW * TS * 0.8, TS * 3);
        const offset = count > 1 ? (posIdx / (count - 1) - 0.5) * spread : 0;
        targetX = centerX + offset;
        targetY = frontY + Math.floor(posIdx / 3) * TS * 0.3;
      } else {
        // Position enemies at the edges of the location, away from NPC center
        const edgeAngle = (posIdx / Math.max(1, locEnemyCounts[loc.id] || 1)) * Math.PI * 2 + Math.PI * 0.75;
        const radiusX = locW * TS * 0.4;
        const radiusY = locH * TS * 0.4;
        targetX = (loc.x + locW * 0.5) * TS + Math.cos(edgeAngle) * radiusX;
        targetY = (loc.y + locH * 0.5) * TS + Math.sin(edgeAngle) * radiusY;
      }

      // LERP to target
      let cx, cy;
      if (meta.prevX == null) {
        cx = targetX;
        cy = targetY;
      } else {
        cx = meta.prevX + (targetX - meta.prevX) * LERP_SPEED;
        cy = meta.prevY + (targetY - meta.prevY) * LERP_SPEED;
        if (Math.abs(cx - targetX) < 0.5) cx = targetX;
        if (Math.abs(cy - targetY) < 0.5) cy = targetY;
      }

      // Direction-based idle animation
      const dx = cx - (meta.prevX ?? cx);
      const dy = cy - (meta.prevY ?? cy);
      if (Math.abs(dx) > 0.3 || Math.abs(dy) > 0.3) {
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
      } else if (sprite.anims.currentAnim?.key?.includes('walk')) {
        sprite.play(`${meta.animPrefix}_idle_${meta.dir}`, true);
      }

      sprite.setPosition(cx, cy);
      meta.prevX = cx;
      meta.prevY = cy;

      // Health bar
      this._drawHealthBar(enemy, sprite);
    }

    // Separation between enemy sprites
    this._applySeparation();

    // Clean up dead/removed enemies
    for (const [id, sprite] of this.sprites) {
      if (!seenIds.has(id)) {
        sprite.destroy();
        this.sprites.delete(id);
        this.enemyMeta.delete(id);
        const bar = this._healthBars.get(id);
        if (bar) {
          bar.destroy();
          this._healthBars.delete(id);
        }
      }
    }
  }

  _applySeparation() {
    const entries = [...this.sprites.entries()];
    for (let i = 0; i < entries.length; i++) {
      for (let j = i + 1; j < entries.length; j++) {
        const a = entries[i][1];
        const b = entries[j][1];
        const dx = a.x - b.x;
        const dy = a.y - b.y;
        const d2 = dx * dx + dy * dy;
        if (d2 < SEP_DIST * SEP_DIST && d2 > 0.01) {
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

  destroy() {
    for (const s of this.sprites.values()) s.destroy();
    this.sprites.clear();
    this.enemyMeta.clear();
    for (const bar of this._healthBars.values()) bar.destroy();
    this._healthBars.clear();
  }
}
