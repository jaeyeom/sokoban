import Phaser from 'phaser';
import { fetch } from 'whatwg-fetch'; // Use whatwg-fetch
import direction from './direction';

class Game {
  constructor(config) {
    this.config = config;
    this.game = new Phaser.Game(config);
  }

  preload(scene) {
    scene.load.image('tiles', '/assets/tileset.png');
  }

  create(scene) {
    this.tileMap = scene.make.tilemap({
      tileWidth: this.config.tileSize,
      tileHeight: this.config.tileSize,
      width: this.config.mapWidth,
      height: this.config.mapHeight
    });

    const tileset = this.tileMap.addTilesetImage('tiles', null, this.config.tileSize, this.config.tileSize);
    this.layer = this.tileMap.createBlankLayer('layer1', tileset, 0, 0);
    this.levelText = scene.add.text(0, 0);
    let undoButton = scene.add.text(150, 0, '[Undo]').setInteractive();
    let resetButton = scene.add.text(400, 0, '[Reset]').setInteractive();
    undoButton.on('pointerdown', () => {
      this.sendAction('undo');
    });
    resetButton.on('pointerdown', () => {
      this.sendAction('reset');
    });

    const handlePointer = (pointer) => {
      // Letterboxing can emit pointerdownoutside while the target is still the canvas.
      if (pointer.downElement !== scene.game.canvas) {
        return;
      }
      if (undoButton.getBounds().contains(pointer.x, pointer.y) ||
          resetButton.getBounds().contains(pointer.x, pointer.y)) {
        return;
      }
      const action = direction.directionFromPoint(
        pointer.x,
        pointer.y,
        scene.cameras.main.width,
        scene.cameras.main.height
      );
      this.sendAction(action);
    };
    scene.input.on('pointerdown', handlePointer);
    scene.input.on('pointerdownoutside', handlePointer);

    scene.input.keyboard.on('keydown', (event) => this.handleKeyPress(event));

    this.handleKeyPress({ key: '.' });
  }

  renderMap(asciiMap) {
    this.layer.forEachTile(tile => tile.index = -1); // Clear existing tiles
    const rows = asciiMap.length;
    const cols = asciiMap[0].length;
    const offsetX = Math.floor((this.config.mapWidth - cols) / 2);
    const offsetY = Math.floor((this.config.mapHeight - rows) / 2);

    asciiMap.forEach((row, y) => {
      Array.from(row).forEach((char, x) => {
        const tileIndex = this.config.tileMapping[char];
        if (tileIndex !== undefined) {
          this.layer.putTileAt(tileIndex, x + offsetX, y + offsetY);
        }
      });
    });

    this.levelText.setText('Level: ' + this.config.level);
  }

  handleKeyPress(event) {
    if (event.ctrlKey || event.altKey || event.metaKey || event.shiftKey) {
      return;
    }

    const actionMap = {
      'ArrowUp': 'up',
      'ArrowDown': 'down',
      'ArrowLeft': 'left',
      'ArrowRight': 'right',
      'w': 'up',
      's': 'down',
      'a': 'left',
      'd': 'right',
      'k': 'up',
      'j': 'down',
      'h': 'left',
      'l': 'right',
      'u': 'undo',
      'r': 'reset',
      '.': '.'
    };

    const action = actionMap[event.key];
    if (action) {
      this.sendAction(action);
    }
  }

  sendAction(action) {
    fetch(window.location.href, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ action: action })
    })
      .then(response => response.json())
      .then(data => {
        if (data.grid) {
          this.config.level = data.current;
          this.renderMap(data.grid.trimEnd().split('\n'));
        }
      })
      .catch(error => console.error('Error:', error));
  }
}

export default Game;
