app = angular.module('gopi_media.controllers', [])

app.controller('HomeCtrl', ['$scope', 'Media', function(sc, media) {
  sc.Test = "Test"
  console.log("ctrl init")
}]);

