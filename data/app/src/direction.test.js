const test = require("node:test");
const assert = require("node:assert/strict");
const { directionFromPoint } = require("./direction");

const width = 608;
const height = 512;

test("screen quadrants", () => {
  assert.equal(directionFromPoint(500, 256, width, height), "right");
  assert.equal(directionFromPoint(80, 256, width, height), "left");
  assert.equal(directionFromPoint(304, 40, width, height), "up");
  assert.equal(directionFromPoint(304, 400, width, height), "down");
});

test("ties go to the vertical axis", () => {
  assert.equal(directionFromPoint(404, 356, width, height), "down");
  assert.equal(directionFromPoint(204, 156, width, height), "up");
  assert.equal(directionFromPoint(width / 2, height / 2, width, height), "down");
});
