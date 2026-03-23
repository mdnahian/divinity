import Phaser from 'phaser';
import { generateTileset, TS } from '../systems/tilesetGenerator.js';

export class BootScene extends Phaser.Scene {
  constructor() {
    super('BootScene');
  }

  preload() {
    const dataUrl = generateTileset();
    this.load.spritesheet('terrain', dataUrl, {
      frameWidth: TS,
      frameHeight: TS,
    });
  }

  create() {
    this.scene.start('WorldScene');
  }
}
