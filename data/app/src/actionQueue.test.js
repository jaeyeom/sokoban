const test = require("node:test");
const assert = require("node:assert/strict");
const { createActionQueue } = require("./actionQueue");

test("runs one action at a time in order", async () => {
  const enqueue = createActionQueue();
  const events = [];
  let releaseFirst;
  const gate = new Promise((resolve) => {
    releaseFirst = resolve;
  });
  const first = enqueue(async () => {
    events.push("start1");
    await gate;
    events.push("end1");
  });
  const second = enqueue(async () => {
    events.push("start2");
  });
  await Promise.resolve();
  assert.deepEqual(events, ["start1"]);
  releaseFirst();
  await first;
  await second;
  assert.deepEqual(events, ["start1", "end1", "start2"]);
});

test("a failed action does not block the next one", async () => {
  const enqueue = createActionQueue();
  const events = [];
  const first = enqueue(async () => {
    events.push("fail");
    throw new Error("network");
  });
  const second = enqueue(async () => {
    events.push("next");
  });
  await assert.rejects(first, /network/);
  await second;
  assert.deepEqual(events, ["fail", "next"]);
});
