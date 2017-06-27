$(function () {

  /* ==================================================================
   Select2 Examples 
   ================================================================== */

  $(".js-example-basic-single").select2({ placeholder: 'Select a State', allowClear: true, theme: 'bootstrap' })


  $(".js-example-basic-multiple").select2({ theme: 'bootstrap' })

  $(".js-example-tokenizer").select2({
    tags: true,
    tokenSeparators: [',', ' '],
    theme: 'bootstrap'
  })

  var data = [{ id: 0, text: 'enhancement' }, { id: 1, text: 'bug' }, { id: 2, text: 'duplicate' }, { id: 3, text: 'invalid' }, { id: 4, text: 'wontfix' }]

  $(".js-example-data-array").select2({
    data: data,
    theme: 'bootstrap'
  })

  $(".js-data-example-ajax").select2({
    theme: 'bootstrap',
    ajax: {
      url: "https://api.github.com/search/repositories",
      dataType: 'json',
      delay: 250,
      data: function (params) {
        return {
          q: params.term, // search term
          page: params.page
        }
      },
      processResults: function (data, page) {
        // parse the results into the format expected by Select2.
        // since we are using custom formatting functions we do not need to
        // alter the remote JSON data
        return {
          results: data.items
        }
      },
      cache: true
    },
    escapeMarkup: function (markup) { return markup }, // let our custom formatter work
    minimumInputLength: 1,
    templateResult: formatRepo, // omitted for brevity, see the source of this page
    templateSelection: formatRepoSelection // omitted for brevity, see the source of this page
  })



  /* ==================================================================
   Simple Color Picker Examples
   ================================================================== */

  $('#cp-ex-1').simplecolorpicker ()

  $('#cp-ex-2').simplecolorpicker({ 
    picker: true
  })

  $('#cp-ex-3').simplecolorpicker({ 
    picker: true
  })



  /* ==================================================================
   Time Picker Examples 
   ================================================================== */

  $('#timepicker1').timepicker ()

  $('#timepicker2').timepicker ()

  $('#timepicker3').timepicker ({ template: 'modal' })

  $('#timepicker4').timepicker ({ template: false })

  $('#timepicker5').timepicker ({ showMeridian: false })

  $('#timepicker6').timepicker ({ showSeconds: true })



  /* ==================================================================
   Date Picker Examples 
   ================================================================== */

  $('#datepicker1').datepicker ({
    autoclose: true,
    todayHighlight: true
  })

  $('#datepicker2').datepicker ({
    autoclose: true,
    todayHighlight: true
  })

  $('#datepicker3').datepicker ({
    autoclose: true,
    todayHighlight: true
  })



  /* ==================================================================
   iCheck Examples
   ================================================================== */

  $('.demo-icheck input').iCheck ({
    checkboxClass: 'ui-icheck icheckbox_minimal-grey',
    radioClass: 'ui-icheck iradio_minimal-grey'
  }).on ('ifChanged', function (e) {
    $(e.currentTarget).trigger ('change')
  })  

})


/* ==================================================================
   Select2 Helper Functions
   ================================================================== */

function formatRepo (repo) {
  if (repo.loading) return repo.text

  var markup = '<div class="clearfix">' +
  '<div class="col-sm-1">' +
  '<img src="' + repo.owner.avatar_url + '" style="max-width: 100%" />' +
  '</div>' +
  '<div clas="col-sm-10">' +
  '<div class="clearfix">' +
  '<div class="col-sm-6">' + repo.full_name + '</div>' +
  '<div class="col-sm-3"><i class="fa fa-code-fork"></i> ' + repo.forks_count + '</div>' +
  '<div class="col-sm-2"><i class="fa fa-star"></i> ' + repo.stargazers_count + '</div>' +
  '</div>'

  if (repo.description) {
    markup += '<div>' + repo.description + '</div>'
  }

  markup += '</div></div>'

  return markup
}

function formatRepoSelection (repo) {
  return repo.full_name || repo.text
}