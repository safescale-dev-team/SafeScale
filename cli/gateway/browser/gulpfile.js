"use strict";

var gulp = require('gulp');

var path = require('path');

var bower = require('gulp-bower');
var exit = require('gulp-exit');
var gprocess = require('gulp-process');
var shell = require('gulp-shell');
var jasmineBrowser = require('gulp-jasmine-browser');
var webpack = require('webpack-stream');

gulp.task('bower', function () {
  return bower();
});

gulp.task('serve-server', [], function () {
  gprocess.start('server', '../gateway', [ '-endpoint', 'localhost:50051'
  ]);
  gulp.watch('../gateway', ['serve-server']);
});

gulp.task('backends', ['serve-server']);

var specFiles = ['*.spec.js'];
gulp.task('test', ['backends'], function (done) {
  return gulp.src(specFiles)
    .pipe(webpack({ output: { filename: 'spec.js' } }))
    .pipe(jasmineBrowser.specRunner({
      console: true,
      sourceMappedStacktrace: true,
    }))
    .pipe(jasmineBrowser.headless({
      findOpenPort: true,
      catch: true,
      throwFailures: true,
    }))
    .on('error', function (err) {
      done(err);
      process.exit(1);
    })
    .pipe(exit());
});

gulp.task('serve', ['backends'], function (done) {
  var JasminePlugin = require('gulp-jasmine-browser/webpack/jasmine-plugin');
  var plugin = new JasminePlugin();

  return gulp.src(specFiles)
    .pipe(webpack({
      output: { filename: 'spec.js' },
      watch: true,
      plugins: [plugin],
    }))
    .pipe(jasmineBrowser.specRunner({
      sourceMappedStacktrace: true,
    }))
    .pipe(jasmineBrowser.server({
      port: 8000,
      whenReady: plugin.whenReady,
    }));
});

gulp.task('default', ['test']);
