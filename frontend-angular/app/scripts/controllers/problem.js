'use strict';

/**
 * @ngdoc function
 * @name onlinejudgeApp.controller:ProblemCtrl
 * @description
 * # ProblemCtrl
 * Controller of the onlinejudgeApp
 */
angular.module('onlinejudgeApp')
  .controller('ProblemCtrl', function ($scope, $state, $rootScope, problemService) {
    // hack. should have a controller to handle nav bar
    $rootScope.state = $state;
    
    $scope.language = null;
    $scope.languages = [
      {'value': 'c', 'text': 'C'},
      {'value': 'cpp', 'text': 'C++'},
      {'value': 'java', 'text': 'Java'},
      {'value': 'javascript', 'text': 'JavaScript'},
      {'value': 'python', 'text': 'Python'},
      {'value': 'ruby', 'text': 'Ruby'},
    ];
    $scope.problem = null;
    $scope.problems = [];

    $scope.changeLanguage = function() {
      problemService.getProblemSummaries()
        .then(function(problems) {
          problems = _.sortBy(problems, function(problem) {
              return problem.title;
          });
          problems = _.filter(problems, function(problem) {
              return _.includes(problem.supported_languages, $scope.language.value);
          });
          $scope.problems = problems;
          $scope.problem = null;
        });
    };
    $scope.changeProblem = function() {
      console.log('selected problem: ' + $scope.problem.id);
    };
  });
