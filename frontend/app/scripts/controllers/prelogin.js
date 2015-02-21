'use strict';

/**
 * @ngdoc function
 * @name onlinejudgeApp.controller:PreloginCtrl
 * @description
 * # PreloginCtrl
 * Controller of the onlinejudgeApp
 */
angular.module('onlinejudgeApp')
  .controller('PreLoginCtrl', function ($scope, $location, personaService) {
    console.log('checking if user is authenticated...');
    var promise = personaService.isAuthenticated();
    promise.then(function(result) {
        console.log('finished checking if user is authenticated.');
        if (result === true) {
            console.log('user is authenticated.');
            $location.path('/problem');
        } else {
            $location.path('/auth/login');
        }
    });
  });
