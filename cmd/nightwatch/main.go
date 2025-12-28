package main

import "gitlab.com/caffeinatedjack/sleepless/internal/nightwatch"

// Version and BuildTime are set at build time via ldflags
var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	nightwatch.Version = Version
	nightwatch.BuildTime = BuildTime
	nightwatch.Execute()
}
