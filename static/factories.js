app = angular.module('gopi_media.factories', [])

app.factory('Media', ['$http', function(http) {
  return {
    index: function() {
      return http.get('/home')
    },
    show: function(id) {
      return http.get('/media/' + id)
    },
    remove: function(id) {
      return http.delete('/media/' + id)
    }
  }
}]);

