package main

import "gitlab.com/caffeinatedjack/sleepless/internal/doze"

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	doze.Version = Version
	doze.BuildTime = BuildTime
	doze.Execute()
}
