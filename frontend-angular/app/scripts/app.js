'use strict';

/**
 * @ngdoc overview
 * @name onlinejudgeApp
 * @description
 * # onlinejudgeApp
 *
 * Main module of the application.
 */
angular
  .module('onlinejudgeApp', [
    'ngAnimate',
    'ngAria',
    'ngCookies',
    'ngMessages',
    'ngResource',
    'ngRoute',
    'ngSanitize',
    'ngTouch',
    'ui.router',
    'ui.router.tabs',
  ])
  .config(function ($stateProvider, $urlRouterProvider) {
    $stateProvider
      .state('prelogin', {
        url: '/',
        templateUrl: 'views/prelogin.html',
        controller: 'PreLoginCtrl'
      })
      .state('login', {
        url: '/auth/login',
        templateUrl: 'views/login.html',
        controller: 'LoginCtrl'
      })
      .state('about', {
        url: '/about',
        templateUrl: 'views/about.html',
        controller: 'AboutCtrl',
      })
      .state('problem', {
        url: '/problem',
        templateUrl: 'views/problem.html',
        controller: 'ProblemCtrl',
      })
      ;
    $urlRouterProvider
      .otherwise('/');
  });
