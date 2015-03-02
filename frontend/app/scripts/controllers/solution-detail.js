'use strict';

/**
 * @ngdoc function
 * @name onlinejudgeApp.controller:SolutionDetailCtrl
 * @description
 * # SolutionDetailCtrl
 * Controller of the onlinejudgeApp
 */
angular.module('onlinejudgeApp')
  .controller('SolutionDetailCtrl', function ($scope, $state, $stateParams, solutionService) {
    $scope.data = {
      solutions: [],
      problemId: $stateParams.problemId,
      language: $stateParams.language,
      state: $state,
    };
    solutionService.getSolutions($scope.data.problemId, $scope.data.language)
      .then(function(response) {
        console.log('solutionService.getSolutions() succeeded.');
        console.log(response);
        $scope.data.solutions = response.solutions;
      }, function(response) {
        console.log('solutionService.getSolutions() failed.');
        console.log(response);
      });
  });
