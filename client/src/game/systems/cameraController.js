import { TS } from './tilesetGenerator.js';
import { useGameStore } from '../../hooks/useGameStore.js';

const MAX_ZOOM = 3.0;
const ZOOM_STEP = 0.08;
const PAN_DECEL = 0.92;
const FOLLOW_LERP = 0.06;

/** Snap zoom to nearest value where zoom*TS is integer — eliminates sub-pixel tile gaps */
function snapZoom(z) {
  return Math.max(1 / TS, Math.round(z * TS) / TS);
}

export class CameraController {
  constructor(scene) {
    this.scene = scene;
    this.cam = scene.cameras.main;

    this.cam.setZoom(0.18);
    this.cam.centerOn(0, 0);

    this.dragging = false;
    this.dragStartX = 0;
    this.dragStartY = 0;
    this.dragMoved = false;
    this.velocityX = 0;
    this.velocityY = 0;
    this._boundsSet = false;
    this._minZoom = 0.15;

    // Pinch-to-zoom state
    this.pinching = false;
    this._pinchDist = 0;

    this._onFitWorld = () => this.fitWorld();
    window.addEventListener('divinity-fit-world', this._onFitWorld);

    this._bindInput();
  }

  _computeMinZoom(worldPxW, worldPxH) {
    const zoomX = this.cam.width / worldPxW;
    const zoomY = this.cam.height / worldPxH;
    return Math.max(zoomX, zoomY);
  }

  _bindInput() {
    const scene = this.scene;

    scene.input.on('pointerdown', (pointer) => {
      if (pointer.rightButtonDown() || pointer.middleButtonDown()) return;
      // Ignore events that didn't originate from the game canvas
      if (pointer.event && pointer.event.target !== scene.game.canvas) return;
      this.dragging = true;
      this.dragStartX = pointer.x;
      this.dragStartY = pointer.y;
      this.dragMoved = false;
      this.velocityX = 0;
      this.velocityY = 0;
    });

    scene.input.on('pointermove', (pointer) => {
      // Pinch-to-zoom: detect two active pointers
      const p1 = scene.input.pointer1;
      const p2 = scene.input.pointer2;
      if (p1.isDown && p2.isDown) {
        const dist = Math.hypot(p2.x - p1.x, p2.y - p1.y);
        if (!this.pinching) {
          this.pinching = true;
          this._pinchDist = dist;
          this.dragging = false;
          this.dragMoved = false;
          return;
        }
        if (this._pinchDist > 0 && dist > 0) {
          const ratio = dist / this._pinchDist;
          const oldZoom = this.cam.zoom;
          const newZoom = snapZoom(Math.max(this._minZoom, Math.min(MAX_ZOOM, oldZoom * ratio)));

          // Zoom centered on midpoint of the two fingers
          const midX = (p1.x + p2.x) / 2;
          const midY = (p1.y + p2.y) / 2;
          const worldX = this.cam.scrollX + midX / oldZoom;
          const worldY = this.cam.scrollY + midY / oldZoom;
          this.cam.zoom = newZoom;
          this.cam.scrollX = worldX - midX / newZoom;
          this.cam.scrollY = worldY - midY / newZoom;
        }
        this._pinchDist = dist;
        return;
      }

      // Single-pointer drag (suppress during/after pinch)
      if (this.pinching) return;
      if (!this.dragging) return;
      const dx = pointer.x - this.dragStartX;
      const dy = pointer.y - this.dragStartY;
      if (!this.dragMoved) {
        if (Math.abs(dx) > 5 || Math.abs(dy) > 5) {
          this.dragMoved = true;
          useGameStore.getState().setFollowNpc(null);
          this.dragStartX = pointer.x;
          this.dragStartY = pointer.y;
        }
        return;
      }
      const zoom = this.cam.zoom;
      this.cam.scrollX -= dx / zoom;
      this.cam.scrollY -= dy / zoom;
      this.velocityX = -dx / zoom;
      this.velocityY = -dy / zoom;
      this.dragStartX = pointer.x;
      this.dragStartY = pointer.y;
    });

    scene.input.on('pointerup', () => {
      this.dragging = false;
      // Clear pinch state when fewer than 2 pointers remain
      const p1 = scene.input.pointer1;
      const p2 = scene.input.pointer2;
      if (!p1.isDown || !p2.isDown) {
        this.pinching = false;
        this._pinchDist = 0;
      }
    });

    scene.input.on('wheel', (pointer, gameObjects, deltaX, deltaY) => {
      const oldZoom = this.cam.zoom;
      const direction = deltaY < 0 ? 1 : -1;
      // Guarantee at least 1 snap step in the zoom direction
      const snapped = snapZoom(oldZoom * (1 + ZOOM_STEP * direction));
      const newZoom = Math.max(this._minZoom, Math.min(MAX_ZOOM,
        snapped === oldZoom ? oldZoom + direction / TS : snapped));

      const worldX = pointer.worldX;
      const worldY = pointer.worldY;
      this.cam.zoom = newZoom;
      const newWorldX = pointer.worldX;
      const newWorldY = pointer.worldY;
      this.cam.scrollX += worldX - newWorldX;
      this.cam.scrollY += worldY - newWorldY;
    });

    scene.input.keyboard.on('keydown-ESC', () => {
      useGameStore.getState().setFollowNpc(null);
      useGameStore.getState().clearSelection();
    });
  }

  /** Zoom into and center the camera on world coordinates (e.g. NPC position). */
  zoomTo(x, y) {
    const targetZoom = snapZoom(Math.max(this._minZoom, Math.min(MAX_ZOOM, 1.8)));
    this.cam.setZoom(targetZoom);
    this.cam.centerOn(x, y);
  }

  /** Fully zoom in and center on a location's world coordinates. */
  zoomToLocation(x, y) {
    const targetZoom = snapZoom(Math.max(this._minZoom, MAX_ZOOM));
    this.cam.setZoom(targetZoom);
    this.cam.centerOn(x, y);
  }

  fitWorld() {
    const world = useGameStore.getState().world;
    if (!world) return;

    const gridW = world.gridW || world.minGridW || 20;
    const gridH = world.gridH || world.minGridH || 20;
    const worldPxW = gridW * TS;
    const worldPxH = gridH * TS;

    this._minZoom = this._computeMinZoom(worldPxW, worldPxH);
    const fitZoom = Math.max(this.cam.width / worldPxW, this.cam.height / worldPxH);
    // Fit the world into view with slight breathing room
    const zoom = snapZoom(Math.max(this._minZoom, Math.min(MAX_ZOOM, fitZoom)));

    this.cam.setZoom(zoom);
    this.cam.centerOn(worldPxW / 2, worldPxH / 2);
  }

  update(time, delta) {
    const world = useGameStore.getState().world;

    if (world && !this._boundsSet) {
      const gridW = world.gridW || world.minGridW || 20;
      const gridH = world.gridH || world.minGridH || 20;
      const worldPxW = gridW * TS;
      const worldPxH = gridH * TS;

      this._minZoom = this._computeMinZoom(worldPxW, worldPxH);
      this.cam.setBounds(0, 0, worldPxW, worldPxH);
      this._boundsSet = true;
      // Skip fitWorld if we're about to follow/zoom to an NPC
      if (!useGameStore.getState().followNpc) {
        this.fitWorld();
      }
    }

    const followId = useGameStore.getState().followNpc;
    if (followId != null) {
      const npcManager = this.scene.npcManager;
      if (npcManager) {
        const sprite = npcManager.getNpcSprite(followId);
        if (sprite) {
          // On first follow, zoom in so the NPC is visible
          if (this._followTarget !== followId) {
            this._followTarget = followId;
            const targetZoom = snapZoom(Math.max(this._minZoom, Math.min(MAX_ZOOM, 1.8)));
            this.cam.setZoom(targetZoom);
            this.cam.centerOn(sprite.x, sprite.y);
          }
          const midX = this.cam.midPoint.x;
          const midY = this.cam.midPoint.y;
          const dx = (sprite.x - midX) * FOLLOW_LERP;
          const dy = (sprite.y - midY) * FOLLOW_LERP;
          this.cam.scrollX += dx;
          this.cam.scrollY += dy;
          return;
        }
      }
    } else {
      this._followTarget = null;
    }

    if (!this.dragging && !this.pinching && (Math.abs(this.velocityX) > 0.5 || Math.abs(this.velocityY) > 0.5)) {
      this.cam.scrollX += this.velocityX * 0.3;
      this.cam.scrollY += this.velocityY * 0.3;
      this.velocityX *= PAN_DECEL;
      this.velocityY *= PAN_DECEL;
    }

    // Round scroll to prevent sub-pixel gaps between tiles
    this.cam.scrollX = Math.round(this.cam.scrollX);
    this.cam.scrollY = Math.round(this.cam.scrollY);
  }

  destroy() {
    window.removeEventListener('divinity-fit-world', this._onFitWorld);
  }
}
