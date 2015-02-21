'use strict';

describe('Controller: ProblemCtrl', function () {

  // load the controller's module
  beforeEach(module('onlinejudgeApp'));

  var ProblemCtrl,
    scope;

  // Initialize the controller and a mock scope
  beforeEach(inject(function ($controller, $rootScope) {
    scope = $rootScope.$new();
    ProblemCtrl = $controller('ProblemCtrl', {
      $scope: scope
    });
  }));

  it('should attach a list of awesomeThings to the scope', function () {
    expect(scope.awesomeThings.length).toBe(3);
  });
});
