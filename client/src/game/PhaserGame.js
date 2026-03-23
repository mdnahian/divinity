import Phaser from 'phaser';
import { BootScene } from './scenes/BootScene.js';
import { WorldScene } from './scenes/WorldScene.js';

let gameInstance = null;

export function createGame(parentElement) {
  if (gameInstance) {
    gameInstance.destroy(true);
    gameInstance = null;
  }

  const config = {
    type: Phaser.AUTO,
    parent: parentElement,
    width: parentElement.clientWidth || 800,
    height: parentElement.clientHeight || 600,
    backgroundColor: '#0e0e18',
    pixelArt: true,
    roundPixels: true,
    antialias: false,
    scale: {
      mode: Phaser.Scale.RESIZE,
      autoCenter: Phaser.Scale.CENTER_BOTH,
    },
    scene: [BootScene, WorldScene],
    input: {
      mouse: { preventDefaultWheel: true },
      touch: { capture: true },
      activePointers: 2,
    },
  };

  gameInstance = new Phaser.Game(config);
  return gameInstance;
}

export function destroyGame() {
  if (gameInstance) {
    gameInstance.destroy(true);
    gameInstance = null;
  }
}
