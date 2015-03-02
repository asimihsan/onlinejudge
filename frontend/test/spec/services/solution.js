'use strict';

describe('Service: solution', function () {

  // load the service's module
  beforeEach(module('onlinejudgeApp'));

  // instantiate service
  var solution;
  beforeEach(inject(function (_solution_) {
    solution = _solution_;
  }));

  it('should do something', function () {
    expect(!!solution).toBe(true);
  });

});
