'use strict';

/**
 * @ngdoc service
 * @name onlinejudgeApp.persona
 * @description
 * # persona
 * Factory in the onlinejudgeApp.
 */
angular.module('onlinejudgeApp')
  .factory('personaService', function($http, $q, $location) {
    var user = {
      loggedIn: false,
      email: null,
    };
    var isAuthenticated = function() {
      var deferred = $q.defer();
      $http.post('/auth/check').
        success(function(data, status) {
          console.log('auth check response success. status: ' + status);
          console.log(data);
          user.loggedIn = true;
          user.email = data.email;
          deferred.resolve(true);
        }).
        error(function(data, status) {
          console.log('auth check response error. status: ' + status);
          console.log(data);
          user.loggedIn = false;
          user.email = null;
          deferred.resolve(false);
        });
        return deferred.promise;
    };
    var login = function() {
      navigator.id.request();
    };
    navigator.id.watch({
      loggedInUser: null,
      onlogin: function(assertion) {
        var deferred = $q.defer();
        var data = {
          assertion: assertion,
          host: $location.host(),
          port: $location.port(),
        };
        console.log('data');
        console.log(data);
        $http.post('/auth/login', data).
          success(function(data, status) {
            console.log('persona login response is success, status: ' + status);
            console.log(data);
            deferred.resolve(data.email);
            $location.path('/problem/python');
          }).
          error(function(data, status) {
            console.log('persona login response is failure. status: ' + status);
            console.log(data);
            deferred.reject();
          });
      },
      onlogout: function($http) {
        $http.post('/auth/logout')
          .then(function() {
            window.location = '/auth/login';
          });
      }
    });

    return {
      isAuthenticated: isAuthenticated,
      login: login,
    };
  });
