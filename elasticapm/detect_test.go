package elasticapm_test

import (
	"github.com/petermylemans/elastic-apm/elasticapm"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/packit"
	"github.com/sclevine/spec"
	"io/ioutil"
	"os"
	"testing"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir string
		detect     packit.DetectFunc
	)

	it.Before(func() {
		var err error
		workingDir, err = ioutil.TempDir("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		detect = elasticapm.Detect()
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("when conditions for detect true are met", func() {
		it("detects", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: "elastic-apm-agent-java"},
				},
				Requires: []packit.BuildPlanRequirement{
					{Name: "elastic-apm-agent-java"},
					{Name: "jvm-application"},
				},
			}))
		})
	})
}
