// directionFromPoint maps a point in game coordinates to a move.
// The vertical axis wins ties, including the exact center, which is down.
function directionFromPoint(x, y, width, height) {
  const halfWidth = width / 2;
  const halfHeight = height / 2;
  if (Math.abs(x - halfWidth) > Math.abs(y - halfHeight)) {
    return x < halfWidth ? "left" : "right";
  }
  return y < halfHeight ? "up" : "down";
}

module.exports = { directionFromPoint };
