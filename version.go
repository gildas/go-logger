package logger

// commit contains the current git commit and is set in the build.sh script
var commit string

// VERSION is the version of this application
var VERSION = "1.7.5" + commit
