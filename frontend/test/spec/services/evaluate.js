'use strict';

describe('Service: evaluate', function () {

  // load the service's module
  beforeEach(module('onlinejudgeApp'));

  // instantiate service
  var evaluate;
  beforeEach(inject(function (_evaluate_) {
    evaluate = _evaluate_;
  }));

  it('should do something', function () {
    expect(!!evaluate).toBe(true);
  });

});
