
"use strict";

var gulp = require("gulp-param")( require("gulp"), process.argv ),

    fs = require("fs"),
    semver = require("semver"),
    dateformat = require("dateformat"),

    git = require("gulp-git"),
    sass = require("gulp-sass"),
    bump = require("gulp-bump"),
    clean = require("gulp-clean"),
    uglify = require("gulp-uglify"),
    rename = require("gulp-rename"),
    header = require("gulp-header"),
    autoprefixer = require("gulp-autoprefixer"),

    pack = function() {

        return JSON.parse(fs.readFileSync("./package.json", "utf8"));

    },

    pkg = pack(),

    dates = {

        today: dateformat( new Date() , "yyyy-mm-dd" ),
        year: dateformat( new Date() , "yyyy" )

    },

    banner = "/*! <%= pkg.title || pkg.name %> - v<%= pkg.version %> - <%= dates.today %>\n" +
                "<%= pkg.homepage ? \"* \" + pkg.homepage + \"\\n\" : \"\" %>" +
                "* Copyright (c) <%= dates.year %> <%= pkg.author %>;" +
                " Licensed <%= pkg.license %> */\n\n",

    out = {

        js: "jquery-ui-slider-pips",
        css: "jquery-ui-slider-pips"

    };










/**
 * tasks
 *
 * to release; tasks should be run in the order:
 * "gulp bump" -> "gulp assets" -> "gulp release"
 *
 * for dev; just run
 * "gulp"
 */

gulp.task("default", ["assets"], function() {

    return gulp;

});

gulp.task("assets", ["clean", "js", "sass"], function() {

    console.log("⭐ >> Finished putting assets to /dist/" );

    return gulp;

});

/**
 * clean the dist folder (empty it)
 */

gulp.task("clean", function() {

    console.log("⭐ >> All clean and shiny! ");

    return gulp
        .src("./dist", { read: false })
        .pipe( clean() );

});


/**
 * js task is used to clean the dist folder and output
 * the minified and non-minified files.
 */

gulp.task("js", ["clean"], function() {

    var pkg = pack();

    return gulp
        .src( "./src/js/**/*.js" )

        .pipe( header( banner, { pkg: pkg, dates: dates } ))
        .pipe( rename( out.js + ".js" ) )
        .pipe( gulp.dest( "./dist" ) )

        .pipe( uglify("combined.js") )
        .pipe( header( banner, { pkg: pkg, dates: dates } ))
        .pipe( rename( out.js + ".min.js" ) )
        .pipe( gulp.dest( "./dist" ) );

});


/**
 * sass task is used to clean the dist folder and output
 * the minified and non-minified files.
 */

gulp.task("sass", ["clean"], function() {

    var pkg = pack();

    gulp
        .src("./src/**/*.scss")
        .pipe( sass({ outputStyle: "expanded" }).on("error", sass.logError ) )
        .pipe( autoprefixer("last 5 versions") )
        .pipe( header( banner, { pkg: pkg, dates: dates } ))
        .pipe( rename( out.css + ".css" ))
        .pipe( gulp.dest("./dist") );

    return gulp
        .src("./src/**/*.scss")
        .pipe( sass({ outputStyle: "compressed" }).on("error", sass.logError ) )
        .pipe( autoprefixer("last 5 versions") )
        .pipe( header( banner, { pkg: pkg, dates: dates } ))
        .pipe( rename( out.css + ".min.css" ))
        .pipe( gulp.dest("./dist") );

});







/**
 * bump task can be used like:
 *
 *     gulp bump --patch
 *     gulp bump --minor
 *     gulp bump --major
 *
 * this task will ONLY bump the version, it will not
 * release a tag, commit the code or update the assets.
 */

gulp.task("bump", function( patch, minor, major ) {
    
    var b = 
        (patch) ? "patch" :
        (minor) ? "minor" :
        (major) ? "major" :
        null;
    
    if( b ) {

        var pkg = pack(),
            oldv = pkg.version,
            newv = semver.inc( oldv , b );

        console.log("⭐ >> Bumping Version to v" + newv );

        return gulp
            .src("./*.json")
            .pipe( bump({ version: newv }) )
            .pipe( gulp.dest("./") );

    } else {

        throw new Error("\n⚠ >> Not Bumping; didn't supply bump type\n\n");

        return false;

    }

});




/**
 * release task should be used after "bump" and "assets" was run.
 * this task will create a commit, and tag it with the version in package.json
 */

gulp.task("release", ["commit", "tag"], function() {
    return gulp;
});


/**
 * commit task is used for creating a cute release icon, and committing dist files
 * to the GIT repository; all src files should already be committed.
 */

gulp.task("commit", function() {

    var pkg = pack(),
        newv = pkg.version,

        fun = "🐒 🐔 🐧 🐤 🐗 🐝 🐌 🐞 🐜 🕷 🦂 🦀 🐍 🐢 🐟 🐡 🐬 🐋 🐊 🐆 🐅 🐃 🐂 🐄 🐪 🐘 🐐 🐏 🐑 🐎 🐖 🐀 🐁 🐓 🦃 🕊 🐕 🐈 🐇 🐿 🐉 🐲".split(" ");
        fun = fun[ Math.floor(Math.random() * fun.length ) ];

    console.log("⭐ >> Committing release v" + newv );

    return gulp
        .src([
            "./*.json",
            "./dist/**/*"
        ])
        .pipe( git.add() )
        .pipe( git.commit("Release v" + newv + " ⚡" + fun + "⚡") );

});

/**
 * tag task should just tag the last commit in repository
 * with the latest version information from package.json.
 */

gulp.task("tag", ["commit"], function() {

    var pkg = pack(),
        newv = pkg.version;

    console.log("⭐ >> Creating new tag for v" + newv );

    git.tag("v" + newv, "Version " + newv, function(err) {
        if ( err ) { throw err; }
    });

    return gulp;

});

