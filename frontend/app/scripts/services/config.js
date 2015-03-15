'use strict';

/**
 * @ngdoc service
 * @name onlinejudgeApp.config
 * @description
 * # config
 * Factory in the onlinejudgeApp.
 */
angular.module('onlinejudgeApp')
  .factory('configService', function() {
    var options = {
      backendBaseUrl: '//www.runsomecode.com'
    };
    return {
      backendBaseUrl: function() {
        return options.backendBaseUrl;
      }
    };
  });
