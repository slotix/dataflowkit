/* ========================================================
*
* MVP Ready - Lightweight & Responsive Admin Template
*
* ========================================================
*
* File: mvpready-core.js
* Theme Version: 3.0.0
* Bootstrap Version: 3.3.6
* Author: Jumpstart Themes
* Website: http://mvpready.com
*
* ======================================================== */

var mvpready_core = function () {

  "use strict"

  var getLayoutColors = function (theme) {

    var theme_choice = (theme === undefined || theme === '') ? 'slate' : theme

    var color_object = {
      slate: ['#D74B4B', '#475F77', '#BCBCBC', '#777777', '#6685a4', '#E68E8E']
      , belize: ['#2980b9', '#7CB268', '#A9A9A9', '#888888', '#74B5E0', '#B3D1A7']
      , square: ['#6B5C93', '#444444', '#569BAA', '#AFB7C2', '#A89EC2', '#A9CCD3']
      , pom: ['#e74c3c', '#444444', '#569BAA', '#AFB7C2', '#F2A299', '#A9CBD3']
      , royal: ['#3498DB', '#2c3e50', '#569BAA', '#AFB7C2', '#ACCDD5', '#6487AA']
      , carrot: ['#E5723F', '#67B0DE', '#373737', '#BCBCBC', '#F2BAA2', '#267BAE']
    }

    return color_object[theme_choice]
  }

  var scrollConfig = {
    height: '100%',
    railVisible: true,
    railColor: '#999',
    size: '5px',
    color: '#888',
    touchScrollStep: 60
  }

  var isLayoutCollapsed = function () {
    return $('.navbar-toggle').is (':visible')
  }

  var isSidebarCollapsed = function () {
    return $('.sidebar-toggle').is (':visible')
  }

  var initLayoutToggles = function () {
    $('.navbar-toggle, .mainnav-toggle').click (function (e) {
      $(this).toggleClass ('is-open')
    })
  }

  var initBackToTop = function () {
    var backToTop = $('<a>', { id: 'back-to-top', href: '#top' })
        , icon = $('<i>', { 'class': 'fa fa-chevron-up' })

    backToTop.appendTo ('body')
    icon.appendTo (backToTop)

    backToTop.hide ()

    $(window).scroll (function () {
      if ($(this).scrollTop () > 150) {
        backToTop.fadeIn ()
      } else {
        backToTop.fadeOut ()
      }
    })

    backToTop.click (function (e) {
      e.preventDefault ()

      $('body, html').animate({
        scrollTop: 0
      }, 600)
    })
  }

  var initSidebarNav = function () {

    var resizeTimer

    $('.sidebar .dropdown > a').click (function(e) {

      if($(this).parent ().hasClass ('has_sub')) {
        e.preventDefault();
      }

      var $this = $(this),
          $li = $this.parents ('li')

      if(!$li.hasClass ("open")) {

        $li.siblings ().find ('ul').slideUp (250, function () {
          $(this).parent ().removeClass ('open')
        })

        $li.find ('ul').slideDown (250, function () {
          $(this).parent ().addClass ('open')
        })

      } else {
        $li.find ('ul').slideUp (250, function () {
          $(this).parent ().removeClass ('open')
        })
      }

    })

    initSidebarScroll ()

    $(window).on('resize', function(e) {
      clearTimeout(resizeTimer)
      resizeTimer = setTimeout(function() {
        initSidebarScroll ()
      }, 250)
    })
  }

  var initSidebarScroll = function () {
    if (!isSidebarCollapsed ()) {
      $('.sidebar-inner').slimScroll (scrollConfig)
    }
  }

  var initNavEnhanced = function () {
    $('.mainnav-menu .dropdown-toggle').click (function (e) {
      e.preventDefault ()
      e.stopPropagation ()

      var $toggle = $(this)

      $toggle.parent ().addClass ('open').trigger ('show.bs.dropdown')
      $toggle.parent ().siblings ('.dropdown').removeClass ('open').trigger ('hide.bs.dropdown')
    })
  }

  var initNavHover = function (config) {
    $('[data-hover="dropdown"]').each (function (e) {

      var $this = $(this)
          , defaults = { delay: { show: 1000, hide: 1000 } }
          , $parent = $this.parent ()
          , settings = $.extend (defaults, config)
          , timeout

      if (!('ontouchstart' in document.documentElement)) {
        $parent.find ('.dropdown-toggle').click (function (e) {
            if (!isLayoutCollapsed ()) {
              e.preventDefault ()
              e.stopPropagation ()
            }
        })
      }

      $parent.mouseenter(function () {
        if (isLayoutCollapsed ()) { return false }

        timeout = setTimeout (function () {
          $parent.addClass ('open')
          $parent.trigger ('show.bs.dropdown')
        }, settings.delay.show)
      })

      $parent.mouseleave(function () {
        if (isLayoutCollapsed ()) { return false }

        clearTimeout (timeout)

        timeout = setTimeout (function () {
          $parent.removeClass ('open keep-open')
          $parent.trigger ('hide.bs.dropdown')
        }, settings.delay.hide)
      })
    })
  }

  var initNavbarNotifications = function () {
    var notifications = $('.navbar-notification'),
        notificationsScrollConfig = {}

    notifications.find ('> .dropdown-toggle').click (function (e) {
      if (mvpready_core.isLayoutCollapsed ()) {
        window.location = $(this).prop ('href')
      }
    })

    notificationsScrollConfig = $.extend (notificationsScrollConfig, scrollConfig, { height: 225 })

    notifications.find ('.notification-list').slimScroll (notificationsScrollConfig)
  }

  return {
    initNavEnhanced: initNavEnhanced
    , initNavHover: initNavHover
    , initSidebarNav: initSidebarNav
    , initNavbarNotifications: initNavbarNotifications

    , initBackToTop: initBackToTop
    , isLayoutCollapsed: isLayoutCollapsed
    , initLayoutToggles: initLayoutToggles

    , layoutColors: getLayoutColors ('slate')
  }

}()