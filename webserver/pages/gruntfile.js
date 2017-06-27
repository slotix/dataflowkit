module.exports = function(grunt) {
  grunt.initConfig({
    pkg: grunt.file.readJSON('package.json'),

    meta: {
      css_dev_path: 'css',
      less_dev_path: 'less',

      less_shared_path: '../../less'
    },

    less: {
      dev: {
        options: {
          compress: true,
          dumpLineNumbers: true,
          sourceMap: true,
          sourceMapFilename: '<%= meta.css_dev_path %>/mvpready-landing.min.css.map',
          sourceMapRootpath: '/'
        },
        files: { '<%= meta.css_dev_path %>/mvpready-landing.css': '<%= meta.less_dev_path %>/mvpready-landing.less' }
      }
    },

    cssmin: {
      dev: {
        files: {
          '<%= meta.css_dev_path %>/mvpready-landing.min.css': [ '<%= meta.css_dev_path %>/mvpready-landing.css' ]
        }
      }
    },

    watch: {
      dev: {
        files: ['<%= meta.less_shared_path %>/**/*.less', '<%= meta.less_dev_path %>/mvpready-landing.less'],
        tasks: ['less:dev', 'cssmin:dev']      
      }
    }

  });

  grunt.loadNpmTasks('grunt-contrib-less');
  grunt.loadNpmTasks('grunt-contrib-watch');
  grunt.loadNpmTasks('grunt-contrib-cssmin');

  grunt.registerTask ('default', ['watch:dev']);
}