app = angular.module("gopi_media.router", ['ngRoute'])

app.config(['$routeProvider', function($routeProvider) {
  //move views to angular-views
  $routeProvider
    .when('/', {
      templateUrl: 'static/angular-views/home.html',
      controller: 'HomeCtrl'
    })
    .when('/media/:id', {
      templateUrl: 'static/angular-views/showMedia.html',
      controller: 'ShowMediaCtrl'
    })
}]);
