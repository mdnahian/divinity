/** Core math utilities used throughout Divinity. */

export function lerp(a, b, t) {
  return a + (b - a) * t;
}

export function clamp(v, lo, hi) {
  return Math.max(lo, Math.min(hi, v));
}

export function smoothstep(t) {
  t = clamp(t, 0, 1);
  return t * t * (3 - 2 * t);
}

export function dist2(ax, az, bx, bz) {
  const dx = ax - bx;
  const dz = az - bz;
  return Math.sqrt(dx * dx + dz * dz);
}

/**
 * Deterministic seeded RNG — same seed always produces the same sequence.
 * Critical for multiplayer consistency and reproducible world generation.
 */
export function seededRng(seed) {
  let s = 0;
  const str = String(seed);
  for (let i = 0; i < str.length; i++) {
    s = Math.imul(31, s) + str.charCodeAt(i) | 0;
  }
  return () => {
    s ^= s << 13;
    s ^= s >> 17;
    s ^= s << 5;
    return (s >>> 0) / 0xFFFFFFFF;
  };
}
