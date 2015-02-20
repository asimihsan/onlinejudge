'use strict';

/**
 * @ngdoc function
 * @name onlinejudgeApp.controller:AboutCtrl
 * @description
 * # AboutCtrl
 * Controller of the onlinejudgeApp
 */
angular.module('onlinejudgeApp')
  .controller('AboutCtrl', function ($scope, $state, $rootScope) {
    // hack. should have a controller to handle nav bar
    $rootScope.state = $state;
  });
