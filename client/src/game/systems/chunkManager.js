/**
 * ChunkManager provides viewport-based sprite culling for large worlds.
 * Instead of replacing Phaser's tilemap (which already culls tiles), this manages
 * visibility of tree/building sprites based on camera position.
 *
 * It also renders a minimap showing territory outlines and NPC positions.
 */
import { TS } from './tilesetGenerator.js';

const CHUNK_SIZE = 32; // tiles per chunk
const BUFFER_CHUNKS = 1; // extra chunks around viewport

export class ChunkManager {
  constructor(scene) {
    this.scene = scene;
    this.gridW = 0;
    this.gridH = 0;
    this.chunksW = 0;
    this.chunksH = 0;

    // Sprites organized by chunk key "cx,cy"
    this._chunkSprites = new Map();
    this._visibleChunks = new Set();

    // Territory borders drawn on the main scene
    this._borderGraphics = null;
  }

  init(gridW, gridH) {
    this.gridW = gridW;
    this.gridH = gridH;
    this.chunksW = Math.ceil(gridW / CHUNK_SIZE);
    this.chunksH = Math.ceil(gridH / CHUNK_SIZE);
  }

  /**
   * Register a sprite so it can be culled by chunk.
   * @param {Phaser.GameObjects.Sprite} sprite
   * @param {number} tileX - tile coordinate
   * @param {number} tileY - tile coordinate
   */
  addSprite(sprite, tileX, tileY) {
    const cx = Math.floor(tileX / CHUNK_SIZE);
    const cy = Math.floor(tileY / CHUNK_SIZE);
    const key = `${cx},${cy}`;
    if (!this._chunkSprites.has(key)) {
      this._chunkSprites.set(key, []);
    }
    this._chunkSprites.get(key).push(sprite);
  }

  /**
   * Clear all tracked sprites (call before rebuilding).
   */
  clear() {
    this._chunkSprites.clear();
    this._visibleChunks.clear();
  }

  /**
   * Update visibility based on camera viewport.
   * Call this each frame in the scene update.
   */
  updateVisibility() {
    const cam = this.scene.cameras.main;
    if (!cam) return;

    const view = cam.worldView;
    const minCX = Math.max(0, Math.floor(view.x / (CHUNK_SIZE * TS)) - BUFFER_CHUNKS);
    const minCY = Math.max(0, Math.floor(view.y / (CHUNK_SIZE * TS)) - BUFFER_CHUNKS);
    const maxCX = Math.min(this.chunksW - 1, Math.floor((view.x + view.width) / (CHUNK_SIZE * TS)) + BUFFER_CHUNKS);
    const maxCY = Math.min(this.chunksH - 1, Math.floor((view.y + view.height) / (CHUNK_SIZE * TS)) + BUFFER_CHUNKS);

    const newVisible = new Set();
    for (let cy = minCY; cy <= maxCY; cy++) {
      for (let cx = minCX; cx <= maxCX; cx++) {
        newVisible.add(`${cx},${cy}`);
      }
    }

    // Hide sprites in chunks that left the viewport
    for (const key of this._visibleChunks) {
      if (!newVisible.has(key)) {
        const sprites = this._chunkSprites.get(key);
        if (sprites) {
          for (const s of sprites) s.setVisible(false);
        }
      }
    }

    // Show sprites in chunks that entered the viewport
    for (const key of newVisible) {
      if (!this._visibleChunks.has(key)) {
        const sprites = this._chunkSprites.get(key);
        if (sprites) {
          for (const s of sprites) s.setVisible(true);
        }
      }
    }

    this._visibleChunks = newVisible;
  }

  /**
   * Draw territory border outlines on the main scene.
   */
  drawTerritoryBorders(territories) {
    if (!territories || territories.length === 0) return;

    if (!this._borderGraphics) {
      this._borderGraphics = this.scene.add.graphics();
      this._borderGraphics.setDepth(1);
    }
    this._borderGraphics.clear();

    const colors = [0xff4444, 0x44ff44, 0x4488ff, 0xffaa00, 0xff44ff, 0x44ffff];

    for (let i = 0; i < territories.length; i++) {
      const t = territories[i];
      const color = colors[i % colors.length];
      const cx = (t.centerX || 0) * TS;
      const cy = (t.centerY || 0) * TS;
      const r = (t.radius || 50) * TS;

      // Draw territory boundary circle
      this._borderGraphics.lineStyle(2, color, 0.3);
      this._borderGraphics.strokeCircle(cx, cy, r);

      // Draw territory name at center
      if (!this._territoryLabels) this._territoryLabels = [];
    }

    // Add territory labels (only once)
    if (!this._territoryLabelsBuilt) {
      this._territoryLabels = [];
      for (let i = 0; i < territories.length; i++) {
        const t = territories[i];
        const color = colors[i % colors.length];
        const hexColor = '#' + color.toString(16).padStart(6, '0');
        const label = this.scene.add.text(
          (t.centerX || 0) * TS,
          (t.centerY || 0) * TS - (t.radius || 50) * TS - 10,
          t.name || t.id,
          {
            fontSize: '14px', fontFamily: 'monospace',
            color: hexColor, stroke: '#000000', strokeThickness: 3,
            align: 'center',
          }
        );
        label.setOrigin(0.5, 1);
        label.setDepth(21);
        this._territoryLabels.push(label);
      }
      this._territoryLabelsBuilt = true;
    }
  }

  /** Minimap is handled by the React Minimap component — no-op. */
  updateMinimap() {}

  /**
   * Destroy all managed resources.
   */
  destroy() {
    this._chunkSprites.clear();
    this._borderGraphics?.destroy();
    if (this._territoryLabels) {
      this._territoryLabels.forEach(l => l.destroy());
    }
  }
}

export { CHUNK_SIZE };
