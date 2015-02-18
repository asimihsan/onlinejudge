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
      $http.post('/auth/check')
        .then(function(response) {
          console.log('auth check response');
          console.log(response);
          if (response.statusText === 'OK' && response.data !== '') {
            user.loggedIn = true;
            user.email = response.data;
            deferred.resolve(true);
          }
          else {
            deferred.resolve(false);
          }
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
        $http.post('/auth/login', data)
          .then(function(response) {
            console.log('response');
            console.log(response);
            if (response.data.status !== 'okay') {
              deferred.reject(response.data.reason);
            } else {
              deferred.resolve(response.data.email);
              $location.path('/problem/python');
            }
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
