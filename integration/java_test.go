package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testJava(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually
		pack       occam.Pack
		docker     occam.Docker
	)

	it.Before(func() {
		pack = occam.NewPack()
		docker = occam.NewDocker()
	})

	context("when building a java app", func() {
		var (
			imageIDs     map[string]struct{}
			containerIDs map[string]struct{}
			image        occam.Image
			container    occam.Container
			name         string
			source       string
			err          error
			logs         fmt.Stringer
		)

		it.Before(func() {
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())
			imageIDs = map[string]struct{}{}
			containerIDs = map[string]struct{}{}
		})

		it.After(func() {
			for id := range containerIDs {
				Expect(docker.Container.Remove.Execute(id)).To(Succeed())
			}

			for id := range imageIDs {
				Expect(docker.Image.Remove.Execute(id)).To(Succeed())
			}
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		it("builds an oci image that has the java agent injected", func() {
			source, err = occam.Source(filepath.Join("testdata", "default_app"))
			Expect(err).NotTo(HaveOccurred())

			image, logs, err = pack.WithNoColor().Build.
				WithBuildpacks(
					"gcr.io/paketo-buildpacks/java:latest",
					buildpack,
				).
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred(), logs)
			imageIDs[image.ID] = struct{}{}

			container, err = docker.Container.Run.
				WithPublish("8080").
				Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())
			containerIDs[container.ID] = struct{}{}

			Eventually(container).Should(BeAvailable())

			logs, err = docker.Container.Logs.Execute(container.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(logs).To(ContainSubstring("Tracer switched to RUNNING state"))

			image, logs, err = pack.WithNoColor().Build.
				WithBuildpacks(
					"gcr.io/paketo-buildpacks/java:latest",
					buildpack,
				).
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred(), logs)
			imageIDs[image.ID] = struct{}{}
			Expect(logs).To(ContainLines(
				"Elastic APM Buildpack 1.2.3",
				"  Reusing cached layer",
			))

			container, err = docker.Container.Run.
				WithPublish("8080").
				Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())
			containerIDs[container.ID] = struct{}{}

			Eventually(container).Should(BeAvailable())

			logs, err = docker.Container.Logs.Execute(container.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(logs).To(ContainSubstring("Tracer switched to RUNNING state"))
		})
	})
}
