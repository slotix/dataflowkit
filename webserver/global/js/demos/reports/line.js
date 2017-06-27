$(function () {
    
  var d1, d2, data, chartOptions;

  d1 = [
    [1262304000000, 5], [1264982400000, 200], [1267401600000, 1605], [1270080000000, 1129], 
    [1272672000000, 1163], [1275350400000, 1905], [1277942400000, 2002], [1280620800000, 2917], 
    [1283299200000, 2700], [1285891200000, 2700], [1288569600000, 3500], [1291161600000, 4157]
  ];

  data = [{ 
    label: "Revenue", 
    data: d1
  }];
 
  chartOptions = {
    xaxis: {
      min: (new Date(2009, 12, 1)).getTime(),
      max: (new Date(2010, 11, 2)).getTime(),
      mode: "time",
      tickSize: [1, "month"],
      monthNames: ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"]
    },

    yaxis: {

    },

    series: {
      lines: {
        show: true, 
        fill: true,
        lineWidth: 2,
        fillColor: {
          colors: [{
            opacity: 0.2
          }, {
            opacity: 0.01
          }]
        } 
      },

      points: {
        show: true,
        radius: 3,
        fill: true,
        fillColor: "#ffffff",
        lineWidth: 2
      }
    },

    grid: {
      hoverable: true, 
      clickable: false, 
      borderWidth: 0,
      tickColor: 'rgba(255,255,255,0.22)'
    },

    legend: {
      show: false
    },

    tooltip: true,
    tooltipOpts: {
      content: '%s: %y'
    },

    colors: ['#ffffff']
  };

    

  var holder = $('#reports-line-chart');

  if (holder.length) {
    $.plot(holder, data, chartOptions );
  }


});