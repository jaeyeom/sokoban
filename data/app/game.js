const config = {
  type: Phaser.AUTO,
  width: 608,
  height: 512,
  scale: {
    // Fit to window
    mode: Phaser.Scale.FIT,
    // Center vertically and horizontally
    autoCenter: Phaser.Scale.CENTER_BOTH
  },
  scene: {
    preload: preload,
    create: create,
  }
};

let tileMap, layer, asciiMap, level, levelText;

const game = new Phaser.Game(config);

const tileMapping = {
  '.': 0,  // Empty
  'O': 1,  // Player
  'X': 2,  // Box
  '#': 3,  // Wall
  '+': 4,  // Target
  '@': 5,  // PlayerOnTarget
  '*': 6   // BoxOnTarget
};

const tileSize = 32;

const mapWidth = 19;
const mapHeight = 16;

function preload() {
  this.load.image('tiles', '/assets/tileset.png');
}

function create() {
  tileMap = this.make.tilemap({
    tileWidth: tileSize,
    tileHeight: tileSize,
    width: mapWidth,
    height: mapHeight
  });

  const tileset = tileMap.addTilesetImage('tiles', null, tileSize, tileSize);

  layer = tileMap.createBlankLayer('layer1', tileset, 0, 0);
  levelText = this.add.text(0, 0);
  let undoButton = this.add.text(150, 0, '[Undo]').setInteractive();
  let resetButton = this.add.text(400, 0, '[Reset]').setInteractive();
  undoButton.on('pointerdown', () => {
    sendAction('undo');
  });
  resetButton.on('pointerdown', () => {
    sendAction('reset');
  });

  this.input.on('pointerdown', (pointer, currentlyOver) => {
    // Skip if pointer is over an interactive object (like buttons)
    if (currentlyOver.length > 0) {
      return;
    }
    // Calculate direction from pointer.x and pointer.y
    const halfWidth = this.cameras.main.width / 2;
    const halfHeight = this.cameras.main.height / 2;
    const direction = Math.abs(pointer.x - halfWidth) > Math.abs(pointer.y - halfHeight) ?
      (pointer.x < halfWidth ? 'left' : 'right') :
      (pointer.y < halfHeight ? 'up' : 'down');
    sendAction(direction);
  });
  this.input.keyboard.on('keydown', handleKeyPress);

  handleKeyPress({ key: '.' });
}

function renderMap(asciiMap) {
  layer.forEachTile(tile => tile.index = -1); // Clear existing tiles
  const rows = asciiMap.length;
  const cols = asciiMap[0].length;
  // Center the map
	const offsetX = Math.floor((mapWidth - cols) / 2);
	const offsetY = Math.floor((mapHeight - rows) / 2);

  asciiMap.forEach((row, y) => {
    Array.from(row).forEach((char, x) => {
      const tileIndex = tileMapping[char];
		  if (tileIndex !== undefined) {
		    layer.putTileAt(tileIndex, x + offsetX, y + offsetY);
		  }
    });
  });
  levelText.setText('Level: ' + level);
}

function handleKeyPress(event) {
  // Ctrl, Alt, Meta, Shift keys
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
    sendAction(action);
  }
}

function sendAction(action) {
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
        level = data.current;
        renderMap(data.grid.trimEnd().split('\n'));
      }
    })
    .catch(error => console.error('Error:', error));
}
