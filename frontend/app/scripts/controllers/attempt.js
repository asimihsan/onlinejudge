'use strict';

/*globals marked */

/**
 * @ngdoc function
 * @name onlinejudgeApp.controller:AttemptCtrl
 * @description
 * # AttemptCtrl
 * Controller of the onlinejudgeApp
 */
angular.module('onlinejudgeApp')
  .controller('AttemptCtrl', function ($scope, $state, $stateParams, problemService, evaluateService, languageService) {
    var indentSizes = languageService.getIndentSizes();
    var codemirrorModes = languageService.getCodemirrorModes();
    function setupEditor(language) {
      $scope.editorOptions = {
        lineWrapping : true,
        lineNumbers: true,
        theme: 'solarized',
        mode: codemirrorModes[language] || 'text/x-csrc',
        tabSize: indentSizes[language] || 4,
        indentUnit: indentSizes[language] || 4,
        extraKeys: {
          Tab: function(cm) {
            var spaces = new Array(cm.getOption('indentUnit') + 1).join(' ');
            cm.replaceSelection(spaces);
          }
        },
      };
    }

    $scope.problemId = $stateParams.problemId;
    $scope.language = $stateParams.language;
    $scope.state = $state;

    $scope.problem = null;
    $scope.description = '';
    $scope.descriptionRendered = false;
    $scope.initialCode = '';
    problemService.getProblemDescriptionAndInitialCode($scope.problemId, $scope.language)
      .then(function(problem) {
        /*jshint camelcase: false */
        $scope.problem = problem;
        $scope.description = marked(
          $scope.problem.description[$scope.language].markdown,
          {
            sanitize: true,
            smartypants: true
          });
        $scope.descriptionRendered = true;
        $scope.initialCode = $scope.problem.initial_code[$scope.language].code;
      });
    setupEditor($scope.language);
    $scope.tabData = [
      {
        heading: 'Description',
        route: 'attempt.description',
        params: {
          description: $scope.description
        }
      },
      {
        heading: 'Code',
        route: 'attempt.code',
        params: {
          problem: $scope.problem
        }
      }
    ];

    $scope.output = '';
    $scope.checkCode = function(problemId, language, code) {
      $scope.checkCodeLoading = true;
      evaluateService.evaluateAttempt(problemId, language, code)
        .then(function(result) {
          $scope.output = result.output;
          $scope.checkCodeLoading = false;
        });
    };
    $scope.clearOutput = function() {
      $scope.output = '';
    };
    $scope.clearCode = function() {
      $scope.initialCode = '';
    };
    $scope.submitCode = function(problemId, language, code) {
      $scope.submitCodeLoading = true;
      evaluateService.submitAttempt(problemId, language, code)
        .then(function(result) {
          console.log('AttemptCtrl call to evaluateService.submitAttempt successful.');
          console.log(result);
          if (result.success === true) {
            $scope.output = '<your code passed the tests, and has been submitted! check the solutions section>';
          } else {
            $scope.output = '<your code did not pass the tests. try "check code" to debug>';
          }
          $scope.submitCodeLoading = false;
        }, function(result) {
          console.log('AttemptCtrl call to evaluateService.submitAttempt failed.');
          console.log(result);
          $scope.submitCodeLoading = false;
        });
    };
  });
