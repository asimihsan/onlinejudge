'use strict';

describe('Service: persona', function () {

  // load the service's module
  beforeEach(module('onlinejudgeApp'));

  // instantiate service
  var persona;
  beforeEach(inject(function (_persona_) {
    persona = _persona_;
  }));

  it('should do something', function () {
    expect(!!persona).toBe(true);
  });

});
