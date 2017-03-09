(function ($) {

  $.dataTableHelper = function (el, options) {
    var base = this;

    base.$el = $(el);
    base.el = el;

    base.init = function () {
      base.metadata = base.$el.data ()

      base.options = $.extend({},$.dataTableHelper.defaultOptions, options, base.metadata);

      base.options.ordering = true    
      base.options.searching = true
    };

    base.init();

    var sortableArray = []

    base.$el.find ('thead th').each (function (i) {
      var bool = ($(this).data ('sortable')) ? true : false
      sortableArray.push ({ orderable: bool })
    })

    base.options.columns = sortableArray

      base.options.initComplete = function () {

        var api = this.api ()
            , row = $('<tr>')
            , orderArray = []
            , filterCount = 0

        api.columns ().indexes ().flatten ().each (function (colIndex) {
          var column = api.column (colIndex)
              , $col = $(column.header ())
              , th = $('<th>&nbsp;</th>')

          if ($col.data ('filterable') == 'text') {
            var input = $('<input type="text" class="form-control input-sm" />')
                .appendTo(th.empty ())
                .on ('keyup change', function () {
                  api
                    .column (colIndex)
                    .search (this.value)
                    .draw ()

                var nodes = $(column.nodes ())
                    , index = $col.index () + 1


                highlightColumn (column, index, this.value)                  
            })                  

            filterCount++
          }

          if ($col.data ('filterable') == 'select') {
            var select = $('<select class="form-control input-sm"><option value=""></option></select>')
                .appendTo(th.empty ())
                .on( 'change', function () {
                  var val = $(this).val()
                      , index = $col.index () + 1

                  column
                    .search( val ? '^'+val+'$' : '', true, false )
                    .draw()

                  highlightColumn (column, index, val)
                })

            column.data ().unique ().sort ().each (function (d, j) {
              select.append('<option value="' + d + '">' + d + '</option>')
            })              

            filterCount++
          }

          base.$el.find ('thead th').each (function (i) {              
            if ($(this).data ('sort-order') === undefined) { $(this).data ('sort-order', 1000) }
          }).sort (sortLI).each (function () {
            orderArray.push ([$(this).index (), $(this).data ('direction') ])
          })

          row.append (th)                                                                                                                                       
        })

        if (orderArray.length > 0) {
          api.order (orderArray).draw ()
        }

        if (filterCount) {
          base.$el.find ('thead').append (row)
        }

        api.draw ()
      }
      
      var $thisTable = base.$el.dataTable (base.options).addClass ('dataTable-helper')

      if (!base.options.globalSearch) {
        $thisTable.parents ('.dataTables_wrapper').find ('.dataTables_filter').remove ()
      } 


  };

  function sortLI (a, b) {
      return ($(b).data ('sort-order')) < ($(a).data ('sort-order')) ? 1 : -1;    
    }

    function highlightColumn (column, index, val) {
      console.log (column)
      if (val == '') {
        $(column.header ()).removeClass ('highlight')
        $(column.nodes ()).removeClass ('highlight')
        $(column.footer ()).removeClass ('highlight')
      } else {
        $(column.header ()).addClass ('highlight')
        $(column.nodes ()).addClass ('highlight')
        $(column.footer ()).addClass ('highlight')        
      }
    }

  $.dataTableHelper.defaultOptions = {
    paging: false
    , globalSearch: false
    , info: false
    , lengthChange: false
    , pageLength: 10
  };

  $.fn.dataTableHelper = function (options) {
    return this.each (function () {
      (new $.dataTableHelper(this, options));
    });
  };

})(jQuery);