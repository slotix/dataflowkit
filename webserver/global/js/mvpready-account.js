/* ========================================================
*
* MVP Ready - Lightweight & Responsive Admin Template
*
* ========================================================
*
* File: mvpready-account.js
* Theme Version: 3.0.0
* Bootstrap Version: 3.3.6
* Author: Jumpstart Themes
* Website: http://mvpready.com
*
* ======================================================== */

var mvpready_account = function () {

  "use strict"

  var initPlaceholder = function () {
    $.support.placeholder = false
    var test = document.createElement('input')
    if('placeholder' in test) $.support.placeholder = true

    if (!$.support.placeholder) {
      $('.placeholder-hidden').show ()
    }
  }

  return {
    init: function() {
      initPlaceholder ()
    }
  }

}()

$(function () {
  mvpready_account.init ()
})