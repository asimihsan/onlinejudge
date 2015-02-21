'use strict';

/**
 * @ngdoc service
 * @name onlinejudgeApp.evaluate
 * @description
 * # evaluate
 * Factory in the onlinejudgeApp.
 */
angular.module('onlinejudgeApp')
  .factory('evaluateService', function($http, $q) {
    var evaluateAttempt = function(problem, language, code) {
      var deferred = $q.defer();
      var data = {
        'code': code,
      };
      var url = '/evaluator/evaluate/' + problem + '/' + language;
      $http({
        url: url,
        method: 'POST',
        data: JSON.stringify(data),
        headers: {'Content-Type': 'application/json'}
      }).success(function(response) {
        deferred.resolve(response);
      }).error(function(msg, code) {
        deferred.reject(msg);
        console.log(msg, code);
      });
      return deferred.promise;
    };

    return {
      evaluateAttempt: evaluateAttempt,
    };

  });
