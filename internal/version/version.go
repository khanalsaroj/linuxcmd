// Package version holds the build-time version string. It is overridden
// via -ldflags at release build time (see scripts/build.ps1 and
// .github/workflows/release.yml); local `go build` without that flag
// leaves it at "dev".
package version

var Version = "dev"

func String() string {
	return Version
}
