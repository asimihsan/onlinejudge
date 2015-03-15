'use strict';

/**
 * @ngdoc service
 * @name onlinejudgeApp.solution
 * @description
 * # solution
 * Factory in the onlinejudgeApp.
 */
angular.module('onlinejudgeApp')
  .factory('solutionService', function($http, $q, configService) {
    var getSolutions = function(problem, language) {
      var deferred = $q.defer();
      var url = configService.backendBaseUrl() + '/user_data/solution/get/' + problem + '/' + language;
      $http.get(url)
      .success(function(response) {
        deferred.resolve(response);
      }).error(function(msg, code) {
        deferred.reject(msg);
        console.log(msg, code);
      });
      return deferred.promise;
    };
    var vote = function(problemId, language, solutionId, voteType) {
      /*jshint camelcase: false */
      var deferred = $q.defer();
      var data = {
        problemId: problemId,
        language: language,
        solutionId: solutionId,
        voteType: voteType
      };
      var url = configService.backendBaseUrl() + '/user_data/solution/vote';
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
      getSolutions: getSolutions,
      vote: vote,
    };
  });
