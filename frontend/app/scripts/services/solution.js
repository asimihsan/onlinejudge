'use strict';

/**
 * @ngdoc service
 * @name onlinejudgeApp.solution
 * @description
 * # solution
 * Factory in the onlinejudgeApp.
 */
angular.module('onlinejudgeApp')
  .factory('solutionService', function($http, $q) {
    var getSolutions = function(problem, language) {
      var deferred = $q.defer();
      var url = '/user_data/solution/get/' + problem + '/' + language;
      $http.get(url)
      .success(function(response) {
        deferred.resolve(response);
      }).error(function(msg, code) {
        deferred.reject(msg);
        console.log(msg, code);
      });
      return deferred.promise;
    };
    var vote = function(solutionId, voteType) {
      /*jshint camelcase: false */
      var deferred = $q.defer();
      var url = '/user_data/solution/vote/' + solutionId + '/' + voteType;
      $http.post(url)
      .success(function(response) {
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
