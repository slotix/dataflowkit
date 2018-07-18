function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

async function scroll2bottom() {
  let docHeight = 0;
  while (docHeight < window.document.body.scrollHeight) {
    docHeight = window.document.body.scrollHeight;
    window.scrollTo(0, docHeight);
    await sleep(200);
  }
}



scroll2bottom();