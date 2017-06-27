/* Lines with autodrowing */

$(function () {
  // we use an inline data source in the example, usually data would
  // be fetched from a server
  var data = [], totalPoints = 100

  function getRandomData() {
    if (data.length > 0)
    data = data.slice(1)

    // do a random walk
    while (data.length < totalPoints) {
      var prev = data.length > 0 ? data[data.length - 1] : 50,
          y = prev + Math.random() * 10 - 5

      if (y < 0) { y = 0 }

      if (y > 100) { y = 100 }
        
      data.push(y)
    }


    // zip the generated y values with the x values
    var res = []
    for (var i = 0; i < data.length; ++i) {
      res.push([i, data[i]])
    }
    return res
  }

  // setup control widget
  var updateInterval = 225

  $("#updateInterval").val(updateInterval).change(function () {
    var v = $(this).val()

    if (v && !isNaN(+v)) {
      updateInterval = +v

      if (updateInterval < 1) { updateInterval = 1 }

      if (updateInterval > 2000) { updateInterval = 2000 }

      $(this).val("" + updateInterval)
    }
  })

  // setup plot
  var options = {
    yaxis: { min: 0, max: 125 },
    xaxis: { show: false, min: 0, max: 100 },
    colors: [mvpready_core.layoutColors[0]],

    grid: { 
      hoverable: true, 
      clickable: false, 
      borderWidth: 0 
    },
    series: {
      lines: { 
        lineWidth: 2, 
        fill: true,
        fillColor: { colors: [ { opacity: 0.6 }, { opacity: 0.2 } ] },
        steps: false
      }
    }
  }

  var plot = $.plot($("#auto-chart"), [ getRandomData() ], options)

  function update() {
    plot.setData([ getRandomData() ])
    // since the axes don't change, we don't need to call plot.setupGrid()
    plot.draw()

    setTimeout(update, updateInterval)
  }

  update()
})