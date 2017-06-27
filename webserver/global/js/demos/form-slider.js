$(function () {

  /* ==================================================================
   Default Demo
   ================================================================== */

  $('#basic-slider').slider ()



  /* ==================================================================
   Range Demo
   ================================================================== */

  $('#slider-range').slider({
    range: true,
    min: 0,
    max: 500,
    animate: true,
    values: [ 50, 300 ]
  }).slider ('pips', {
    rest: 'label',
    step: 50
  }).slider ('float')



  /* ==================================================================
   Color Variation Demo
   ================================================================== */

  $('.demo-slider-colors').each (function (i) {
    var val = getRandomIntInclusive (10, 100)
    var step_val = (i % 2) ? 25 : 20

    $(this).empty ().slider ({
      value: val,
      range: 'min',
      animate: true
    })
  })



  /* ==================================================================
   Vertical Demo
   ================================================================== */

  $('#slider-vertical').slider({
    orientation: 'vertical',
    range: true,
    values: [ 25, 100 ],
    animate: true,
    slide: function( event, ui ) {
      $('#amount').val('$' + ui.values[ 0 ] + ' - $' + ui.values[ 1 ] )
    }
  }).slider ('pips', {
    rest: 'label',
    step: 25
  }).slider ('float')



  /* ==================================================================
   Default Vertical Demo
   ================================================================== */

  $('#eq-basic > span').each (function () {
    var value = parseInt( $(this).text(), 10 )

    $(this).empty ().slider ({
      value: value,
      range: 'min',
      animate: true,
      orientation: 'vertical'
    })
  })


  /* ==================================================================
   Pips Vertical Demo
   ================================================================== */

  $('#eq-pips > span').each (function () {
    var value = parseInt( $(this).text(), 10 )

    $(this).empty ().slider ({
      value: value,
      range: 'min',
      animate: true,
      orientation: 'vertical'
    }).slider ('pips', {
      rest: 'label',
      step: 25
    })
  })


  /* ==================================================================
   Select Demo
   ================================================================== */

  var select = $('#select-slider')
  var slider = $('<div id="select-slider-div" class="ui-slider-primary"></div>')

  slider.insertAfter( select ).slider ({
    min: 1,
    max: 6,
    range: 'min',
    animate: true,
    value: select[ 0 ].selectedIndex + 1,
    slide: function( event, ui ) {
      select[ 0 ].selectedIndex = ui.value - 1
    }
  }).slider ('pips', {
    rest: 'label'
  })

  $('#select-slider').change(function () {
    slider.slider('value', this.selectedIndex + 1 )
  })

})


function getRandomIntInclusive(min, max) {
  return Math.floor(Math.random() * (max - min + 1)) + min
}