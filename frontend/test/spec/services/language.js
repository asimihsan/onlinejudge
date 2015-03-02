'use strict';

describe('Service: language', function () {

  // load the service's module
  beforeEach(module('onlinejudgeApp'));

  // instantiate service
  var language;
  beforeEach(inject(function (_language_) {
    language = _language_;
  }));

  it('should do something', function () {
    expect(!!language).toBe(true);
  });

});
