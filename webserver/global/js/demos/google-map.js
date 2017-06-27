$(function () {
  var mapMarkers = [{
        address: "New York, NY 10017",
        html: "<strong>New York Office</strong><br>New York, NY 10017",
        icon: {
          image: "",
          iconsize: [26, 46],
          iconanchor: [12, 46]
        },
        popup: true
      }];

      // Map Initial Location
      var initLatitude = 40.75198;
      var initLongitude = -73.96978;

      // Map Extended Settings
      var mapSettings = {
        controls: {
          draggable: true,
          panControl: false,
          zoomControl: false,
          mapTypeControl: false,
          scaleControl: false,
          streetViewControl: false,
          overviewMapControl: false
        },
        scrollwheel: false,
        markers: mapMarkers,
        latitude: initLatitude,
        longitude: initLongitude,
        zoom: 16
      };

      var map = $("#googlemaps").gMap(mapSettings),
        mapRef = $("#googlemaps").data('gMap.reference');

      // Create an array of styles.
      var mapColor = "#0088cc";

      var styles = [{
        stylers: [{
          hue: mapColor
        }]
      }, {
        featureType: "road",
        elementType: "geometry",
        stylers: [{
          lightness: 0
        }, {
          visibility: "simplified"
        }]
      }, {
        featureType: "road",
        elementType: "labels",
        stylers: [{
          visibility: "off"
        }]
      }];

      var styledMap = new google.maps.StyledMapType(styles, {
        name: "Styled Map"
      });

      mapRef.mapTypes.set('map_style', styledMap);
      mapRef.setMapTypeId('map_style');
})