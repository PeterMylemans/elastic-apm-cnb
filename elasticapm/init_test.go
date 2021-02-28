package elasticapm_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitElasticApm(t *testing.T) {
	suite := spec.New("elastic-apm", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Build", testBuild)
	suite("Detect", testDetect)
	suite.Run(t)
}
