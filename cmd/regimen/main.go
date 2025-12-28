package main

import "gitlab.com/caffeinatedjack/sleepless/internal/regimen"

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	regimen.Version = Version
	regimen.BuildTime = BuildTime
	regimen.Execute()
}
