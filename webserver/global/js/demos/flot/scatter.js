$(function () {

  var d1, d2, data, chartOptions;

  d1 = [[0,1],[1,2],[3,8],[5,4],[2,10],[1.2,9],[9,2],[46,41],[22,14],[20,12],[20,25],[7,5],[18,11],[31,20]];
  d2 = [[50,40],[24,36],[20,42],[33,41],[51,39],[11,28],[32,16],[38,40],[35,22],[41,30],[21,18]];

  data = [{ 
    label: "Total visitors", 
    data: d1
  }, {
    label: "Total Sales",
    data: d2
  }];

  chartOptions = {
    xaxis: {

    },
    yaxis: {

    },
    series: {
      lines: {
        show: false, 
        fill: false,
        lineWidth: 3
      },
      points: {
        show: true,
        radius: 4,
        fill: true,
        lineWidth: 3
      }
    },
    grid: { 
      hoverable: true, 
      clickable: false, 
      borderWidth: 0 
    },
    legend: {
      show: true
    },
    tooltip: true,
    tooltipOpts: {
      content: '%s: %y'
    },
    colors: mvpready_core.layoutColors
  };

  var holder = $('#scatter-chart');

  if (holder.length) {
    $.plot(holder, data, chartOptions );
  }


});