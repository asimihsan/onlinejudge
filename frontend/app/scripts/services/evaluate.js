'use strict';

/**
 * @ngdoc service
 * @name onlinejudgeApp.evaluate
 * @description
 * # evaluate
 * Factory in the onlinejudgeApp.
 */
angular.module('onlinejudgeApp')
  .factory('evaluateService', function($http, $q, configService) {
    var evaluateAttempt = function(problem, language, code) {
      var deferred = $q.defer();
      var data = {
        'code': code,
      };
      var url = configService.backendBaseUrl() + '/evaluator/evaluate/' + problem + '/' + language;
      $http({
        url: url,
        method: 'POST',
        data: JSON.stringify(data),
        headers: {'Content-Type': 'application/json'}
      }).success(function(response) {
        console.log('evaluateService.evaluateAttempt success.');
        console.log(response);
        deferred.resolve(response);
      }).error(function(msg, code) {
        console.log('evaluateService.evaluateAttempt error.');
        console.log(msg, code);
        deferred.reject(msg);
      });
      return deferred.promise;
    };

    var submitAttempt = function(problemId, language, code) {
      var deferred = $q.defer();
      var data = {
        'problem_id': problemId,
        'language': language,
        'code': code,
      };
      var url = configService.backendBaseUrl() + '/user_data/solution/submit';
      $http({
        url: url,
        method: 'POST',
        data: JSON.stringify(data),
        headers: {'Content-Type': 'application/json'}
      }).success(function(response) {
        console.log('evaluateService.submitAttempt success.');
        console.log(response);
        deferred.resolve(response);
      }).error(function(msg, code) {
        console.log('evaluateService.submitAttempt error.');
        console.log(msg, code);
        deferred.reject(msg);
      });
      return deferred.promise;
    };

    return {
      evaluateAttempt: evaluateAttempt,
      submitAttempt: submitAttempt,
    };

  });
