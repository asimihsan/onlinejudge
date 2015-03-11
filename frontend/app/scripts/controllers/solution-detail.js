'use strict';

/**
 * @ngdoc function
 * @name onlinejudgeApp.controller:SolutionDetailCtrl
 * @description
 * # SolutionDetailCtrl
 * Controller of the onlinejudgeApp
 */
angular.module('onlinejudgeApp')
  .controller('SolutionDetailCtrl', function ($scope, $state, $stateParams, solutionService, languageService) {
    $scope.data = {
      solutions: [],
      problemId: $stateParams.problemId,
      language: $stateParams.language,
      state: $state,
      languageValueToText: languageService.getLanguageValueToText(),
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
    $scope.vote = function(solutionId, voteType) {
      console.log('vote. solutionId: ' + solutionId + ', voteType: ' + voteType);
      solutionService.vote(solutionId, voteType);
    };
  });
