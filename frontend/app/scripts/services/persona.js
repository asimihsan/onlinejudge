'use strict';

/**
 * @ngdoc service
 * @name onlinejudgeApp.persona
 * @description
 * # persona
 * Factory in the onlinejudgeApp.
 */
angular.module('onlinejudgeApp')
  .factory('personaService', function($http, $q, $location, configService) {
    var user = {
      loggedIn: false,
      id: null,
      email: null,
    };
    var getId = function() { return user.id; };
    var getEmail = function() { return user.email; };
    var getLoggedIn = function() { return user.loggedIn; };

    var isAuthenticated = function() {
      var deferred = $q.defer();
      $http.post(configService.backendBaseUrl() + '/user_data/auth/check').
        success(function(data, status) {
          console.log('auth check response success. status: ' + status);
          console.log(data);
          user.loggedIn = true;
          user.id = data.id;
          user.email = data.email;
          deferred.resolve(true);
        }).
        error(function(data, status) {
          console.log('auth check response error. status: ' + status);
          console.log(data);
          user.loggedIn = false;
          user.id = null;
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
        $http.post(configService.backendBaseUrl() + '/user_data/auth/login', data).
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
        $http.post(configService.backendBaseUrl() + '/user_data/auth/logout')
          .then(function() {
            window.location = '/auth/login';
          });
      }
    });

    return {
      isAuthenticated: isAuthenticated,
      login: login,
      getId: getId,
      getEmail: getEmail,
      getLoggedIn: getLoggedIn,
    };
  });
