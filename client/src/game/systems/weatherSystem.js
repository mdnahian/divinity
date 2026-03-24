import Phaser from 'phaser';
import { useGameStore } from '../../hooks/useGameStore.js';

const WEATHERS = {
  sunny:   { particles: 0,    snow: false },
  clear:   { particles: 0,    snow: false },
  cloudy:  { particles: 0,    snow: false },
  showers: { particles: 500,  snow: false },
  rain:    { particles: 1200, snow: false },
  storm:   { particles: 2000, snow: false },
  snow:    { particles: 900,  snow: true  },
};

const DEFAULT_WEATHER = { particles: 0, snow: false };
const MAX_PARTICLES = 800;

export class WeatherSystem {
  constructor(scene) {
    this.scene = scene;
    this.currentWeather = '';
    this._emitter = null;
    this._lastZoom = null;
    this._ensureTextures();
  }

  _ensureTextures() {
    if (!this.scene.textures.exists('rain_particle')) {
      const g = this.scene.add.graphics();
      g.fillStyle(0x8899bb, 0.6);
      g.fillRect(0, 0, 2, 8);
      g.generateTexture('rain_particle', 2, 8);
      g.destroy();
    }
    if (!this.scene.textures.exists('snow_particle')) {
      const g = this.scene.add.graphics();
      g.fillStyle(0xeef4ff, 0.85);
      g.fillRect(0, 0, 3, 3);
      g.generateTexture('snow_particle', 3, 3);
      g.destroy();
    }
  }

  update(time, delta) {
    const world = useGameStore.getState().world;
    const weather = world?.weather || 'clear';

    const cam = this.scene.cameras.main;
    const wData = WEATHERS[weather] || DEFAULT_WEATHER;

    if (weather !== this.currentWeather) {
      this._clearParticles();
      this.currentWeather = weather;
    }

    if (wData.particles > 0 && !this._emitter) {
      this._spawnEmitter(wData, cam);
    }

    if (this._emitter) {
      // Update emitter zone to match camera viewport
      const v = cam.worldView;
      const invZoom = 1 / cam.zoom;

      this._emitter.setPosition(v.x, v.y);
      this._emitter.addEmitZone({
        type: 'random',
        source: new Phaser.Geom.Rectangle(0, -20, v.width, 10),
        quantity: 1,
      });
      // Clear old zones, keep only the latest
      if (this._emitter.emitZones && this._emitter.emitZones.length > 1) {
        this._emitter.emitZones.splice(0, this._emitter.emitZones.length - 1);
      }

      this._emitter.setParticleScale(invZoom);
    }

    this._lastZoom = cam.zoom;
  }

  _spawnEmitter(wData, cam) {
    const v = cam.worldView;
    const isSnow = wData.snow;
    const texKey = isSnow ? 'snow_particle' : 'rain_particle';
    const count = Math.min(wData.particles, MAX_PARTICLES);

    this._emitter = this.scene.add.particles(v.x, v.y, texKey, {
      emitZone: {
        type: 'random',
        source: new Phaser.Geom.Rectangle(0, -20, v.width, 10),
        quantity: 1,
      },
      lifespan: isSnow ? 5000 : 1800,
      speedX: isSnow ? { min: -18, max: 18 } : { min: -30, max: 20 },
      speedY: isSnow ? { min: 25, max: 55 } : { min: 250, max: 450 },
      quantity: Math.max(1, Math.ceil(count / 50)),
      frequency: 16,
      maxParticles: count,
      gravityY: isSnow ? 5 : 20,
    });
    this._emitter.setDepth(101);
  }

  _clearParticles() {
    if (this._emitter) {
      this._emitter.destroy();
      this._emitter = null;
    }
  }

  destroy() {
    this._clearParticles();
  }
}
