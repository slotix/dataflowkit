function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

async function scroll2bottom() {
  let docHeight = 0;
  let delay = 500;
  while (docHeight < window.document.body.scrollHeight) {
    docHeight = window.document.body.scrollHeight;
    window.scrollTo(0, docHeight);
    await sleep(delay);
    if (docHeight == window.document.body.scrollHeight) {
      delay += 500;
      await sleep(3000);
    }
  }
}

scroll2bottom();