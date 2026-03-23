/**
 * Perlin noise implementation for terrain generation.
 * Provides noise2 (2D Perlin) and fbm (fractal Brownian motion).
 */
import { smoothstep, lerp } from './math';

const _P = new Uint8Array(512);

// Initialize permutation table
(() => {
  const p = new Uint8Array(256);
  for (let i = 0; i < 256; i++) p[i] = i;
  for (let i = 255; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [p[i], p[j]] = [p[j], p[i]];
  }
  for (let i = 0; i < 512; i++) _P[i] = p[i & 255];
})();

function _g(h, x, y) {
  h &= 7;
  return (
    ((h & 1) ? -(h < 4 ? x : y) : (h < 4 ? x : y)) +
    ((h & 2) ? -(h < 4 ? y : x) : (h < 4 ? y : x))
  );
}

/** 2D Perlin noise, returns roughly -1..1 */
export function noise2(x, y) {
  const X = Math.floor(x) & 255;
  const Y = Math.floor(y) & 255;
  const xf = x - Math.floor(x);
  const yf = y - Math.floor(y);
  const u = smoothstep(xf);
  const v = smoothstep(yf);
  const a = _P[X] + Y;
  const b = _P[X + 1] + Y;
  return lerp(
    lerp(_g(_P[a], xf, yf), _g(_P[b], xf - 1, yf), u),
    lerp(_g(_P[a + 1], xf, yf - 1), _g(_P[b + 1], xf - 1, yf - 1), u),
    v,
  );
}

/** Fractal Brownian motion — layered Perlin for natural-looking terrain. */
export function fbm(x, y, octaves = 5) {
  let v = 0;
  let a = 0.5;
  let f = 1;
  let m = 0;
  for (let i = 0; i < octaves; i++) {
    v += noise2(x * f, y * f) * a;
    m += a;
    a *= 0.5;
    f *= 2.07;
  }
  return v / m;
}
