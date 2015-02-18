'use strict';

/**
 * @ngdoc service
 * @name onlinejudgeApp.problem
 * @description
 * # problem
 * Factory in the onlinejudgeApp.
 */
angular.module('onlinejudgeApp')
  .factory('problemService', function($http, $q) {

    var getProblemSummaries = function() {
      var deferred = $q.defer();
      $http.get('/evaluator/get_problem_summaries')
      .success(function(response) {
        deferred.resolve(response);
      }).error(function(msg, code) {
        deferred.reject(msg);
        console.log(msg, code);
      });
      return deferred.promise;
    };

    var getProblemDescriptionAndInitialCode = function(problem, language) {
      var deferred = $q.defer();
      var url = '/evaluator/get_problem_details/' + problem + '/' + language;
      $http.get(url)
      .success(function(response) {
        deferred.resolve(response);
      }).error(function(msg, code) {
        deferred.reject(msg);
        console.log(msg, code);
      });
      return deferred.promise;
    };

    return {
      getProblemSummaries: getProblemSummaries,
      getProblemDescriptionAndInitialCode: getProblemDescriptionAndInitialCode,
    };
});
