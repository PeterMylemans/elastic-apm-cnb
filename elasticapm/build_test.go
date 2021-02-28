package elasticapm_test

import (
	"bytes"
	"github.com/paketo-buildpacks/packit/postal"
	"github.com/paketo-buildpacks/packit/scribe"
	"github.com/petermylemans/elastic-apm/elasticapm"
	"github.com/petermylemans/elastic-apm/elasticapm/fakes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layersDir  string
		workingDir string
		cnbDir     string

		dependencyManager *fakes.DependencyManager
		logBuffer         *bytes.Buffer

		javaAgentDependency postal.Dependency

		build packit.BuildFunc
	)

	it.Before(func() {
		var err error
		layersDir, err = ioutil.TempDir("", "layers")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = ioutil.TempDir("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		workingDir, err = ioutil.TempDir("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		dependencyManager = &fakes.DependencyManager{}
		javaAgentDependency = postal.Dependency{
			ID:   "elastic-apm-agent-java",
			Name: "Elastic APM Java Agent",
		}
		dependencyManager.ResolveCall.Returns.Dependency = javaAgentDependency

		logBuffer = bytes.NewBuffer(nil)
		logger := scribe.NewLogger(logBuffer)

		build = elasticapm.Build(dependencyManager, logger)
	})

	it.After(func() {
		Expect(os.RemoveAll(layersDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	it("returns a result that builds correctly", func() {
		result, err := build(packit.BuildContext{
			WorkingDir: workingDir,
			CNBPath:    cnbDir,
			Stack:      "some-stack",
			BuildpackInfo: packit.BuildpackInfo{
				Name:    "Some Buildpack",
				Version: "some-version",
			},
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: "elastic-apm-agent-java",
					},
				},
			},
			Layers: packit.Layers{Path: layersDir},
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers).To(HaveLen(1))
		Expect(result.Layers[0].Name).To(Equal("agent"))
		Expect(result.Layers[0].LaunchEnv["JAVA_TOOL_OPTIONS.append"]).To(ContainSubstring("-javaagent:"))
		Expect(result.Layers[0].LaunchEnv["JAVA_TOOL_OPTIONS.append"]).To(ContainSubstring("apm-java-agent.jar"))
		Expect(result.Layers[0].LaunchEnv["JAVA_TOOL_OPTIONS.delim"]).To(Equal(" "))

		Expect(dependencyManager.ResolveCall.Receives.Path).To(Equal(filepath.Join(cnbDir, "buildpack.toml")))
		Expect(dependencyManager.ResolveCall.Receives.Id).To(Equal("elastic-apm-agent-java"))
		Expect(dependencyManager.ResolveCall.Receives.Version).To(Equal("*"))
		Expect(dependencyManager.ResolveCall.Receives.Stack).To(Equal("some-stack"))

		Expect(dependencyManager.CopyCall.Receives.Dependency).To(Equal(javaAgentDependency))
		Expect(dependencyManager.CopyCall.Receives.CnbPath).To(Equal(cnbDir))
		Expect(dependencyManager.CopyCall.Receives.LayerPath).To(Equal(filepath.Join(layersDir, "agent/apm-java-agent.jar")))

	})
}
