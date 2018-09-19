function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

async function ScrollDown(pageCount, buttonCSSSelector = '') {
  let docHeight = 0;
  let delay = 500;
  currentPageNum = 1;
  while (docHeight < window.document.body.scrollHeight && currentPageNum < pageCount) {
    // We have to reGet more button element reference every scroll 'cause
    // element changes its location and old reference is not valid any more
    moreButton = (buttonCSSSelector == '')? null: document.querySelector(buttonCSSSelector);
    docHeight = window.document.body.scrollHeight;
    if (moreButton == null) {
      window.scrollTo(0, docHeight);
    } else {
      moreButton.click();
    }
    await sleep(delay);
    if (docHeight == window.document.body.scrollHeight) {
      delay += 500;
      await sleep(3000);
    } else {
      currentPageNum++;
    }
  }
}