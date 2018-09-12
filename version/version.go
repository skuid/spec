package version

import "runtime"

/*
version creates exportable variables for the current commit hash and the golang runtime version.

The default value for commit is "HEAD", but should be overwritten at compile
time using the go link flag "-X" like so:

	go build -ldflags="-X github.com/skuid/spec/version.Commit=`git rev-parse --short HEAD`"
*/

var (
	Commit    = "HEAD"            // Commit is used to surface which commit is at the HEAD
	GoVersion = runtime.Version() // GoVersion is used to surface which version of golang is being used
)
