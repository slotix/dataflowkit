/* ========================================================
*
* MVP Ready - Lightweight & Responsive Admin Template
*
* ========================================================
*
* File: mvpready-helpers.js
* Theme Version: 3.0.0
* Bootstrap Version: 3.3.6
* Author: Jumpstart Themes
* Website: http://mvpready.com
*
* ======================================================== */

var mvpready_helpers = function () {

  "use strict"

  var initFormValidation = function( ) {
    if ($.fn.parsley) {
      $('.parsley-form').each (function () {
        $(this).parsley ({
          trigger: 'change'
          , errorsContainer: function (el) {
            if (el.$element.parents ('form').is ('.form-horizontal')) {
              return el.$element.parents ("*[class^='col-']")
            }

            return el.$element.parents ('.form-group')
          }
          , errorsWrapper: '<ul class="parsley-error-list"></ul>'
        })
      })
    }
  }

  var initAccordions = function () {
    $('.accordion-simple, .accordion-panel').each (function (i) {
      var accordion = $(this)
          , toggle = accordion.find ('.accordion-toggle')
          , activePanel = accordion.find ('.panel-collapse.in').parent ()

      activePanel.addClass ('is-open')

      toggle.prepend('<i class="fa accordion-caret"></i>')

      toggle.on ('click', function (e) {
        var panel = $(this).parents ('.panel')

        panel.toggleClass ('is-open')
        panel.siblings ().removeClass ('is-open')
      })
    })
  }

  var initTooltips = function () {
    if ($.fn.tooltip) { $('.ui-tooltip').tooltip ({ container: 'body' }) }
    if ($.fn.popover) { $('.ui-popover').popover ({ container: 'body' }) }
  }

  var initLightbox = function () {
    if ($.fn.magnificPopup) {
      $('.ui-lightbox').magnificPopup ({
        type: 'image'
        , closeOnContentClick: false
        , closeBtnInside: true
        , fixedContentPos: true
        , mainClass: 'mfp-no-margins mfp-with-zoom'
        , image: {
          verticalFit: true
          , tError: '<a href="%url%">The image #%curr%</a> could not be loaded.'
        }
      })

      $('.ui-lightbox-video, .ui-lightbox-iframe').magnificPopup ({
        disableOn: 700
        , type: 'iframe'
        , mainClass: 'mfp-fade'
        , removalDelay: 160
        , preloader: false
        , fixedContentPos: false
      })

      $('.ui-lightbox-gallery').magnificPopup ({
        delegate: 'a'
        , type: 'image'
        , tLoading: 'Loading image #%curr%...'
        , mainClass: 'mfp-img-mobile'
        , gallery: {
          enabled: true
          , navigateByImgClick: true
          , preload: [0,1]
        },
        image: {
          tError: '<a href="%url%">The image #%curr%</a> could not be loaded.'
          , titleSrc: function(item) {
            return item.el.attr('title')
          }
        }
      })
    }
  }

  var initSelect = function () {
    if ($.fn.select2) {
      $('.ui-select').select2({
        allowClear: true
        , placeholder: "Select..." })
    }
  }

  var initIcheck = function () {
    if ($.fn.iCheck) {
      $('.ui-check').iCheck ({
        checkboxClass: 'ui-icheck icheckbox_minimal-grey'
        , radioClass: 'ui-icheck iradio_minimal-grey'
        , inheritClass: true
      }).on ('ifChanged', function (e) {
        $(e.currentTarget).trigger ('change')
        console.log ($(e.currentTarget))
        console.log ('changed')
      })
    }
  }

  var initDataTableHelper = function () {
    if ($.fn.dataTableHelper) {
      $('.ui-datatable').dataTableHelper ()
    }
  }

  var initiTimePicker = function () {
    if ($.fn.timepicker) {
      $('.ui-timepicker').timepicker ()
      $('.ui-timepicker-modal').timepicker ({ template: 'modal' })
    }
  }

  var initDatePicker = function () {
    if ($.fn.datepicker) {
      $('.ui-datepicker').datepicker({
        autoclose: true
        , todayHighlight: true
      })
    }
  }

  var initColorPicker = function () {
    if ($.fn.simplecolorpicker) {
      $('.ui-colorpicker').each (function (i) {
        var picker = $(this).data ('picker')

        $(this).simplecolorpicker({
          picker: picker
        })
      })
    }
  }

  return {
    initAccordions: initAccordions
    , initFormValidation: initFormValidation
    , initTooltips: initTooltips
    , initLightbox: initLightbox
    , initSelect: initSelect
    , initIcheck: initIcheck
    , initDataTableHelper: initDataTableHelper
    , initiTimePicker: initiTimePicker
    , initDatePicker: initDatePicker
    , initColorPicker: initColorPicker
  }

}()