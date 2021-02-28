package main

import (
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/scribe"
	"github.com/petermylemans/elastic-apm/elasticapm"
	"os"
)

func main() {
	dependencyManager := elasticapm.NewExtendedDependencyManager()
	logger := scribe.NewLogger(os.Stdout)

	packit.Run(elasticapm.Detect(), elasticapm.Build(dependencyManager, logger))
}
