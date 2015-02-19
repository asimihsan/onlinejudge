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
  .controller('AttemptCtrl', function ($scope, $state, $stateParams, problemService, evaluateService) {
    var indentSizes = {
      'c': 4,
      'cpp': 4,
      'java': 4,
      'javascript': 2,
      'python': 4,
      'ruby': 2,
    };
    var codemirrorModes = {
      'c': 'text/x-csrc',
      'cpp': 'text/x-c++src',
      'java': 'text/x-java',
      'javascript': 'javascript',
      'python': 'python',
      'ruby': 'ruby',
    };
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
    $scope.initialCode = '';
    problemService.getProblemDescriptionAndInitialCode($scope.problemId, $scope.language)
      .then(function(problem) {
        /*jshint camelcase: false */
        $scope.problem = problem;
        $scope.description = marked($scope.problem.description[$scope.language].markdown);
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
    $scope.submitCode = function(problemId, language, code) {
      $scope.submitCodeLoading = true;
      evaluateService.evaluateAttempt(problemId, language, code)
        .then(function(result) {
          $scope.output = result.output;
          $scope.submitCodeLoading = false;
        });
    };
    $scope.clearOutput = function() {
      $scope.output = '';
    };
    $scope.clearCode = function() {
      $scope.initialCode = '';
    };
  });
