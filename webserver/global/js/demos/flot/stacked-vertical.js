$(function () {

  var d1, d2, d3, data, chartOptions

  d1 = [
    [1325376000000, 1200], [1328054400000, 700], [1330560000000, 1000], [1333238400000, 600],
    [1335830400000, 350]
  ]

  d2 = [
    [1325376000000, 800], [1328054400000, 600], [1330560000000, 300], [1333238400000, 350],
    [1335830400000, 300]
  ]

  d3 = [
    [1325376000000, 650], [1328054400000, 450], [1330560000000, 150], [1333238400000, 200],
    [1335830400000, 150]
  ]

  data = [{
    label: 'Product 1',
    data: d1
  }, {
    label: 'Product 2',
    data: d2
  }, {
    label: 'Product 3',
    data: d3
  }]

  chartOptions = {
    xaxis: {
      min: (new Date(2011, 11, 15)).getTime(),
      max: (new Date(2012, 04, 18)).getTime(),
      mode: "time",
      tickSize: [2, "month"],
      monthNames: ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"],
      tickLength: 0
    },
    grid: {
      hoverable: true,
      clickable: false,
      borderWidth: 0
    },
    series: {
      stack: true
    },
    bars: {
      show: true,
      barWidth: 36*24*60*60*300,
      fill: true,
      align: 'center',
      lineWidth: 1,
      lineWidth: 0,
      fillColor: { colors: [ { opacity: 1 }, { opacity: 1 } ] }
    },
    tooltip: true,
    tooltipOpts: {
      content: '%s: %y'
    },
    colors: mvpready_core.layoutColors
  }

  var holder = $('#stacked-vertical-chart')

  if (holder.length) {
    $.plot(holder, data, chartOptions )
  }



})