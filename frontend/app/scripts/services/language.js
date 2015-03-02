'use strict';

/**
 * @ngdoc service
 * @name onlinejudgeApp.language
 * @description
 * # language
 * Factory in the onlinejudgeApp.
 */
angular.module('onlinejudgeApp')
  .factory('languageService', function () {
    var languages = [
      {'value': 'c', 'text': 'C'},
      {'value': 'cpp', 'text': 'C++'},
      {'value': 'java', 'text': 'Java'},
      {'value': 'javascript', 'text': 'JavaScript'},
      {'value': 'python', 'text': 'Python'},
      {'value': 'ruby', 'text': 'Ruby'},
    ];
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

    var getLanguages = function() {
      return languages;
    };
    var getIndentSizes = function() {
      return indentSizes;
    };
    var getCodemirrorModes = function() {
      return codemirrorModes;
    };

    // Public API here
    return {
      getLanguages: getLanguages,
      getIndentSizes: getIndentSizes,
      getCodemirrorModes: getCodemirrorModes,
    };
  });
