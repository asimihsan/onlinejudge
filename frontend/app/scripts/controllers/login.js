'use strict';

/**
 * @ngdoc function
 * @name onlinejudgeApp.controller:LoginCtrl
 * @description
 * # LoginCtrl
 * Controller of the onlinejudgeApp
 */
angular.module('onlinejudgeApp')
  .controller('LoginCtrl', function ($scope, personaService) {
    $scope.login = function() {
      personaService.login();
    };
  });
