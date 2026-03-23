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

export class WeatherSystem {
  constructor(scene) {
    this.scene = scene;
    this.particles = [];
    this.currentWeather = '';
    this._lastZoom = null;
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

    if (wData.particles > 0 && this.particles.length === 0) {
      this._spawnParticles(wData);
    }

    if (this._lastZoom !== null && Math.abs(cam.zoom - this._lastZoom) > 0.001 && this.particles.length > 0) {
      this._redistributeParticles(cam);
    }
    this._lastZoom = cam.zoom;

    this._updateParticles(wData, delta, cam);
  }

  _spawnParticles(wData) {
    const cam = this.scene.cameras.main;
    const v = cam.worldView;
    const count = Math.min(wData.particles, 800);
    for (let i = 0; i < count; i++) {
      const color = wData.snow ? 0xeef4ff : 0x8899bb;
      const pw = wData.snow ? 3 : 2;
      const ph = wData.snow ? 3 : 8;
      const alpha = wData.snow ? 0.85 : 0.6;
      const p = this.scene.add.rectangle(
        v.x + Math.random() * v.width,
        v.y + Math.random() * v.height,
        pw, ph, color, alpha
      );
      p.setScrollFactor(1);
      p.setDepth(101);
      this.particles.push({
        obj: p,
        vx: (Math.random() - 0.3) * (wData.snow ? 18 : 50),
        vy: wData.snow ? 25 + Math.random() * 30 : 250 + Math.random() * 200,
      });
    }
  }

  _updateParticles(wData, delta, cam) {
    const dt = delta / 1000;
    const v = cam.worldView;
    const left = v.x;
    const top = v.y;
    const right = v.x + v.width;
    const bottom = v.y + v.height;
    const invZoom = 1 / cam.zoom;

    this.particles.forEach(p => {
      p.obj.x += p.vx * invZoom * dt;
      p.obj.y += p.vy * invZoom * dt;
      p.obj.setScale(invZoom);

      if (p.obj.x < left || p.obj.x > right ||
          p.obj.y < top - 10 || p.obj.y > bottom) {
        p.obj.x = left + Math.random() * v.width;
        p.obj.y = top + Math.random() * v.height;
      }
    });
  }

  _redistributeParticles(cam) {
    const v = cam.worldView;
    this.particles.forEach(p => {
      p.obj.x = v.x + Math.random() * v.width;
      p.obj.y = v.y + Math.random() * v.height;
    });
  }

  _clearParticles() {
    this.particles.forEach(p => p.obj.destroy());
    this.particles = [];
  }

  destroy() {
    this._clearParticles();
  }
}
