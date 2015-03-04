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
    'ui.codemirror',
    'angular-ladda',
    'hljs',
  ])
  // allow DI for use in controllers, unit tests
  .constant('_', window._)
  // use in views, ng-repeat="x in _.range(3)"
  .run(function ($rootScope, $state) {
     $rootScope._ = window._;
     $rootScope.state = $state;
  })
  .config(function ($stateProvider, $urlRouterProvider, hljsServiceProvider) {
    hljsServiceProvider.setOptions({
    });
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
      .state('attempt', {
        url: '/problem/{language:[a-z0-9_]+}/{problemId:[a-z0-9_]+}',
        templateUrl: 'views/attempt.html',
        controller: 'AttemptCtrl',
      })
      .state('attempt.description', {
        url: '/description',
        templateUrl: 'views/attempt-description.html',
      })
      .state('attempt.code', {
        url: '/code',
        templateUrl: 'views/attempt-code.html',
      })
      .state('solution', {
        url: '/solution',
        templateUrl: 'views/solution.html',
        controller: 'SolutionCtrl',
      })
      .state('solution.languageSelected', {
        url: '/{language:[a-z0-9_]+}',
        templateUrl: 'views/solution.html'
      })
      .state('solutionDetail', {
        url: '/solution/{language:[a-z0-9_]+}/{problemId:[a-z0-9_]+}',
        templateUrl: 'views/solution-detail.html',
        controller: 'SolutionDetailCtrl',
      })
      ;
    $urlRouterProvider
      .otherwise('/');
  });
