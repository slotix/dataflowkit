$(function () {

  /* ==================================================================
   Advanced Form (Horizontal) Wizard Demo
   ================================================================== */
	
  var form = $('#example-advanced-form').show()
  
  form.steps({    
    stepsOrientation: "horizontal",
    headerTag: 'h3',
    bodyTag: 'fieldset',
    transitionEffect: 'slideLeft',
    onInit: function (event, currentIndex) {
      var el = $(event.currentTarget)

      el.find ('.steps li').each (function (i) {
        $(this).find (".number").text (i+1)
      })

      el.find ('fieldset').each(function(index, section) {
        $(section).find(':input').attr('data-parsley-group', 'block-' + index);  
      })

    },
    onStepChanging: function (event, currentIndex, newIndex) { 
      var el = $(event.currentTarget)   

      if (currentIndex > newIndex) {
        if (el.find ('fieldset').eq (currentIndex).find ('.parsley-error').length > 0) {
          setTimeout (function () {
            form.find ('.steps ul li').eq (currentIndex).addClass ('error')
          }, 1)
         }

        return true;
      }

      return form.parsley().validate('block-' + currentIndex)
    },
    onFinishing: function (event, currentIndex) {
      return form.parsley ().validate ()
    },
    onFinished: function (event, currentIndex) {
      alert('Submitted!')
    }
  })

  form.parsley ().subscribe ('parsley:field:validate', function (formInstance) {
    var currentForm = $(formInstance.$element).parents ('form'),
        currentFieldset = $(formInstance.$element).parents ('fieldset'),
        currentIndex = currentForm.find ('fieldset').index(currentForm.find ('fieldset').filter('.current'))

    if (currentFieldset.find ('.parsley-error').length < 1) {
      currentForm.find ('.steps ul li').eq (currentIndex).removeClass ('error')
      return
    }

    if (form.parsley().isValid('block-' + currentIndex)) {
      currentForm.find ('.steps ul li').eq (currentIndex).removeClass ('error')
    } else {
      currentForm.find ('.steps ul li').eq (currentIndex).addClass ('error')
    }
  })



  /* ==================================================================
   Vertical Tab Demo 
   ================================================================== */

  $("#example-vertical").steps ({
    headerTag: "h3",
    bodyTag: "section",
    transitionEffect: "slideLeft",
    stepsOrientation: "vertical",
    onInit: function (event, currentIndex) {
      var el = $(event.currentTarget)

      el.find ('.steps li').each (function (i) {
        $(this).find (".number").text (i+1)
      })
    }
  })
	
})