app = angular.module('gopi_media.factories', [])

app.factory('Media', ['$http', function(http) {
  return {
    index: function() {
      return http.get('/')
    }
  }
}]);

