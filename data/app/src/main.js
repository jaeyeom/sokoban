import Phaser from 'phaser';
import Game from './game';

const config = {
  type: Phaser.AUTO,
  width: 608,
  height: 512,
  scale: {
    mode: Phaser.Scale.FIT,
    autoCenter: Phaser.Scale.CENTER_BOTH
  },
  scene: {
    preload: function () { game.preload(this); },
    create: function () { game.create(this); }
  },
  tileMapping: {
    '.': 0,
    'O': 1,
    'X': 2,
    '#': 3,
    '+': 4,
    '@': 5,
    '*': 6
  },
  tileSize: 32,
  mapWidth: 19,
  mapHeight: 16,
  level: 1
};

const game = new Game(config);
