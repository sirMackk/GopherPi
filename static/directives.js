app = angular.module('gopi_media.directives', [])

app.directive('embedTarget', function(){
  return {
    restrict: 'A',
    link: function(scope, element, attrs) {
      var current = element;
      scope.$watch(function() { return attrs.embedTarget }, function(newVal, oldVal) {
        var clone = element.clone().attr('target', attrs.embedTarget);
        current.replaceWith(clone);
        current = clone;
      });
    }
  }
});

