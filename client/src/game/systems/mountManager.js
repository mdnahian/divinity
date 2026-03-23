import { TS } from './tilesetGenerator.js';
import { generateMountSpriteSheet, generateCarriageSpriteSheet } from './spriteGenerator.js';
import { useGameStore } from '../../hooks/useGameStore.js';

const MOUNT_OFFSET_Y = 3; // NPC shifts up this many pixels when mounted

export { MOUNT_OFFSET_Y };

export class MountManager {
  constructor(scene) {
    this.scene = scene;
    this.sprites = new Map();          // mountId -> Phaser.Sprite
    this.carriageSprites = new Map();  // carriageId -> Phaser.Sprite
    this._mountIndexMap = new Map();
    this._carriageIndexMap = new Map();
    this._nextMountIndex = 0;
    this._nextCarriageIndex = 0;
    this._mountMeta = new Map();       // mountId -> { dir, animPrefix }
    this._carriageMeta = new Map();
  }

  _getMountIndex(mountId) {
    if (!this._mountIndexMap.has(mountId)) {
      this._mountIndexMap.set(mountId, this._nextMountIndex++);
    }
    return this._mountIndexMap.get(mountId);
  }

  _getCarriageIndex(carriageId) {
    if (!this._carriageIndexMap.has(carriageId)) {
      this._carriageIndexMap.set(carriageId, this._nextCarriageIndex++);
    }
    return this._carriageIndexMap.get(carriageId);
  }

  _ensureMountSprite(mount) {
    if (this.sprites.has(mount.id)) return this.sprites.get(mount.id);

    const idx = this._getMountIndex(mount.id);
    const mountType = mount.type || 'horse';
    const texKey = generateMountSpriteSheet(this.scene, idx, mountType);
    const animPrefix = `mount_${idx}_${mountType}`;

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
    sprite.setDepth(9); // Below NPC (10) so rider appears on top
    sprite.setVisible(false); // Hidden until positioned in update()
    sprite.play(`${animPrefix}_idle_down`);

    this.sprites.set(mount.id, sprite);
    this._mountMeta.set(mount.id, { dir: 'down', animPrefix });
    return sprite;
  }

  _ensureCarriageSprite(carriage) {
    if (this.carriageSprites.has(carriage.id)) return this.carriageSprites.get(carriage.id);

    const idx = this._getCarriageIndex(carriage.id);
    const cType = carriage.cargoSlots > 20 ? 'wagon' : 'cart';
    const texKey = generateCarriageSpriteSheet(this.scene, idx, cType);

    // Carriages have 1 frame per direction (no walk anim), 4 directions
    const sprite = this.scene.add.sprite(0, 0, texKey);
    sprite.setDepth(8); // Behind mount
    sprite.setVisible(false); // Hidden until positioned in update()
    sprite.setFrame(0);

    this.carriageSprites.set(carriage.id, sprite);
    this._carriageMeta.set(carriage.id, { dir: 'down', texKey });
    return sprite;
  }

  _locCenter(loc) {
    return {
      x: (loc.x + (loc.w || 1) * 0.5) * TS,
      y: (loc.y + (loc.h || 1) * 0.5) * TS,
    };
  }

  update(time, delta) {
    const state = useGameStore.getState();
    const world = state.world;
    if (!world) return;

    const mounts = world.mounts || [];
    const carriages = world.carriages || [];
    const npcs = world.npcs || [];
    const locations = world.locations || [];
    const locMap = {};
    for (const loc of locations) locMap[loc.id] = loc;

    // Build NPC lookup and get NPC sprite positions from npcManager
    const npcMap = {};
    for (const npc of npcs) npcMap[npc.id] = npc;

    const npcManager = this.scene.npcManager;
    const activeMountIds = new Set();
    const activeCarriageIds = new Set();

    // Camera viewport for culling
    const cam = this.scene.cameras.main;
    const view = cam.worldView;
    const pad = TS * 8;
    const cullLeft = view.x - pad;
    const cullRight = view.x + view.width + pad;
    const cullTop = view.y - pad;
    const cullBottom = view.y + view.height + pad;

    for (const mount of mounts) {
      if (!mount.alive) continue;
      activeMountIds.add(mount.id);

      const sprite = this._ensureMountSprite(mount);
      const meta = this._mountMeta.get(mount.id);
      const owner = mount.ownerId ? npcMap[mount.ownerId] : null;

      let mx, my, ownerDir = null;
      const ownerTraveling = owner && owner.targetLocationId && owner.travelArrivalTick > 0;

      if (owner && ownerTraveling && npcManager) {
        // Only show mount with owner when traveling
        const npcSprite = npcManager.sprites.get(owner.id);
        const npcMeta = npcManager.npcMeta.get(owner.id);
        if (npcSprite) {
          mx = npcSprite.x;
          my = npcSprite.y + 2; // Slightly below NPC center
          ownerDir = npcMeta?.dir || 'down';
        }
      }

      if (mx == null) {
        // Not traveling or no owner — hide mount
        sprite.setVisible(false);
        continue;
      }

      // Cull off-screen
      if (mx < cullLeft || mx > cullRight || my < cullTop || my > cullBottom) {
        sprite.setVisible(false);
        continue;
      }

      sprite.setVisible(true);
      sprite.setPosition(mx, my);

      // Match owner's direction for animation
      const dir = ownerDir || meta.dir;
      const isMoving = owner && owner.targetLocationId && owner.travelArrivalTick > 0;

      if (isMoving) {
        const walkKey = `${meta.animPrefix}_walk_${dir}`;
        if (!sprite.anims.currentAnim?.key?.includes(walkKey)) {
          sprite.play(walkKey, true);
        }
      } else {
        const idleKey = `${meta.animPrefix}_idle_${dir}`;
        if (sprite.anims.currentAnim?.key !== idleKey) {
          sprite.play(idleKey, true);
        }
      }
      meta.dir = dir;
    }

    // Handle carriages
    const mountMap = {};
    for (const m of mounts) mountMap[m.id] = m;

    for (const carriage of carriages) {
      activeCarriageIds.add(carriage.id);

      const sprite = this._ensureCarriageSprite(carriage);
      const meta = this._carriageMeta.get(carriage.id);
      const horse = carriage.horseId ? mountMap[carriage.horseId] : null;

      let cx, cy, dir = 'down';

      if (horse && horse.alive) {
        const mountSprite = this.sprites.get(horse.id);
        if (mountSprite && mountSprite.visible) {
          const mountMeta = this._mountMeta.get(horse.id);
          dir = mountMeta?.dir || 'down';
          // Position behind the mount based on direction
          const offset = 10;
          cx = mountSprite.x;
          cy = mountSprite.y;
          if (dir === 'down') cy -= offset;
          else if (dir === 'up') cy += offset;
          else if (dir === 'right') cx -= offset;
          else if (dir === 'left') cx += offset;
        }
      }

      if (cx == null) {
        // Not traveling — hide carriage
        sprite.setVisible(false);
        continue;
      }

      // Cull off-screen
      if (cx < cullLeft || cx > cullRight || cy < cullTop || cy > cullBottom) {
        sprite.setVisible(false);
        continue;
      }

      sprite.setVisible(true);
      sprite.setPosition(cx, cy);

      // Set frame based on direction (0=down, 1=up, 2=right, 3=left)
      const dirIdx = dir === 'down' ? 0 : dir === 'up' ? 1 : dir === 'right' ? 2 : 3;
      sprite.setFrame(dirIdx);
      meta.dir = dir;
    }

    // Clean up sprites for mounts/carriages no longer in world
    for (const [id, sprite] of this.sprites) {
      if (!activeMountIds.has(id)) {
        sprite.destroy();
        this.sprites.delete(id);
        this._mountMeta.delete(id);
      }
    }
    for (const [id, sprite] of this.carriageSprites) {
      if (!activeCarriageIds.has(id)) {
        sprite.destroy();
        this.carriageSprites.delete(id);
        this._carriageMeta.delete(id);
      }
    }
  }

  destroy() {
    for (const s of this.sprites.values()) s.destroy();
    this.sprites.clear();
    for (const s of this.carriageSprites.values()) s.destroy();
    this.carriageSprites.clear();
  }
}
