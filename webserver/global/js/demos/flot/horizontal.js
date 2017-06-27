$(function () {

  var ds=[], data, chartOptions

  ds.push ([[2500, 1],[3400, 2],[3700, 3],[4500, 4]])
  ds.push ([[1300, 1],[2900, 2],[2500, 3],[2300, 4]])
  ds.push ([[800, 1],[1300, 2],[1900, 3],[1500, 4]])

  data = [{
    label: 'Product 1',
    data: ds[1],
    bars: {
      order: 0
    }
  }, {
    label: 'Product 2',
    data: ds[0],
    bars: {
      order: 1
    }
  }, {
    label: 'Product 3',
    data: ds[2],
    bars: {
      order: 2
    }
  }]

  chartOptions = {
    xaxis: {

    },
    grid: {
      hoverable: true,
      clickable: false,
      borderWidth: 0
    },
    bars: {
      horizontal: true,
      show: true,
      barWidth: 8*24*60*60*300,
      barWidth: .2,
      fill: true,
      lineWidth: 1,
      order: true,
      lineWidth: 0,
      fillColor: { colors: [ { opacity: 1 }, { opacity: 1 } ] }
    },
    tooltip: true,
    tooltipOpts: {
      content: '%s: %x'
    },
    colors: mvpready_core.layoutColors
  }

  var holder = $('#horizontal-chart')

  if (holder.length) {
    $.plot(holder, data, chartOptions )
  }

})