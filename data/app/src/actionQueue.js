function createActionQueue() {
  let tail = Promise.resolve();
  return function enqueue(run) {
    const result = tail.then(run, run);
    tail = result.then(
      () => {},
      () => {},
    );
    return result;
  };
}

module.exports = { createActionQueue };
