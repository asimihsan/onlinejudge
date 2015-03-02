'use strict';

/**
 * @ngdoc function
 * @name onlinejudgeApp.controller:SolutionCtrl
 * @description
 * # SolutionCtrl
 * Controller of the onlinejudgeApp
 */
angular.module('onlinejudgeApp')
  .controller('SolutionCtrl', function ($scope, languageService, problemService) {
    $scope.data = {
      selectedLanguage: null,
      problems: [],
    };
    $scope.languages = languageService.getLanguages();
    $scope.languageSelected = function(language) {
      $scope.data.selectedLanguage = language;
      problemService.getProblemSummaries()
        .then(function(problems) {
          problems = _.sortBy(problems, function(problem) {
              return problem.title;
          });
          problems = _.filter(problems, function(problem) {
              return _.includes(problem.supported_languages, $scope.data.selectedLanguage);
          });
          $scope.data.problems = problems;
        });
    };
  });
