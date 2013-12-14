app = angular.module('gopi_media.controllers', [])

app.controller('HomeCtrl', ['$scope', 'Media', function(sc, Media) {
  Media.index().success(function(data, status) {
    sc.MediaListing = data
  });
}]);

app.controller('ShowMediaCtrl', ['$scope', 'Media', '$route', '$routeParams', function(sc, Media, rt, rtParams) {
  var mediaId = rtParams.id;
  Media.show(mediaId).success(function(data, status) {
    sc.Media = data
  });
}]);
