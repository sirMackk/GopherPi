app = angular.module('gopi_media.controllers', [])

app.controller('HomeCtrl', ['$scope', 'Media', function(sc, Media) {
  Media.index().success(function(data, status) {
    sc.MediaListing = data
  });
}]);

app.controller('ShowMediaCtrl', ['$scope', 'Media', '$location', '$route', '$routeParams', function(sc, Media, location, rt, rtParams) {
  sc.mediaId = rtParams.id;
  Media.show(sc.mediaId).success(function(data, status) {
    sc.Media = data
  });

  sc.deleteMedia = function(id) {
      Media.remove(id).success(function(data, status) {
        location.path('/');
      });
  }
}]);
