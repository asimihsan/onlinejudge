'use strict';

/* global _ */

/**
 * @ngdoc function
 * @name onlinejudgeApp.controller:ProblemCtrl
 * @description
 * # ProblemCtrl
 * Controller of the onlinejudgeApp
 */
angular.module('onlinejudgeApp')
  .controller('ProblemCtrl', function ($scope, languageService, problemService) {
    $scope.data = {
      selectedLanguage: null,
      problems: [],
    };
    $scope.languages = languageService.getLanguages();
    $scope.languageValueToText = languageService.getLanguageValueToText();
    $scope.languageSelected = function(language) {
      $scope.data.selectedLanguage = language;
      problemService.getProblemSummaries()
        .then(function(problems) {
          problems = _.sortBy(problems, function(problem) {
              return problem.title;
          });
          problems = _.filter(problems, function(problem) {
              /*jshint camelcase: false */
              return _.includes(problem.supported_languages, $scope.data.selectedLanguage);
          });
          $scope.data.problems = problems;
          $scope.data.problemsByCategory = _.groupBy(problems, 'category');
          $scope.data.categories = _.sortBy(_.unique(_.map(problems, function(o) { return o.category; })));
        });
    };
    $scope.clearSelectedLanguage = function() {
      $scope.data.selectedLanguage = null;
      $scope.data.problems = [];
      $scope.data.problemsByCategory = {};
      $scope.data.categories = [];
    };
    $scope.$on('$stateChangeSuccess', function(event, toState, toParams, fromState, fromParams){
      if ($scope.state.current.name === 'problem.languageSelected') {
        console.log('changed state, language is chosen as: ' + toParams.language);
        $scope.languageSelected(toParams.language);
      } else {
        console.log('changed state, language is now not selected');
        $scope.clearSelectedLanguage();
      }
    });
  });
