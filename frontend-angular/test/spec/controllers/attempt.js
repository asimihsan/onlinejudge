'use strict';

describe('Controller: AttemptCtrl', function () {

  // load the controller's module
  beforeEach(module('onlinejudgeApp'));

  var AttemptCtrl,
    scope;

  // Initialize the controller and a mock scope
  beforeEach(inject(function ($controller, $rootScope) {
    scope = $rootScope.$new();
    AttemptCtrl = $controller('AttemptCtrl', {
      $scope: scope
    });
  }));

  it('should attach a list of awesomeThings to the scope', function () {
    expect(scope.awesomeThings.length).toBe(3);
  });
});
