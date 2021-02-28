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

func testJavaReusesLayers(t *testing.T, context spec.G, it spec.S) {
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

	context("when rebuilding a java app", func() {
		var (
			imageIDs     map[string]struct{}
			containerIDs map[string]struct{}
			name         string
			source       string
		)

		it.Before(func() {
			var err error
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

		it("reuses the layers", func() {
			var (
				err         error
				logs        fmt.Stringer
				firstImage  occam.Image
				secondImage occam.Image
				container   occam.Container
			)

			source, err = occam.Source(filepath.Join("testdata", "default_app"))
			Expect(err).NotTo(HaveOccurred())

			firstImage, logs, err = pack.WithNoColor().Build.
				WithBuildpacks(
					"gcr.io/paketo-buildpacks/java:latest",
					buildpack,
				).
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred(), logs.String())
			imageIDs[firstImage.ID] = struct{}{}

			secondImage, logs, err = pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					"gcr.io/paketo-buildpacks/java:latest",
					buildpack,
				).
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred(), logs.String())
			imageIDs[secondImage.ID] = struct{}{}

			container, err = docker.Container.Run.
				WithPublish("8080").
				Execute(secondImage.ID)
			Expect(err).NotTo(HaveOccurred())
			containerIDs[container.ID] = struct{}{}

			Eventually(container).Should(BeAvailable())

			logs, err = docker.Container.Logs.Execute(container.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(logs.String()).To(ContainSubstring("Tracer switched to RUNNING state"))

			Expect(secondImage.Buildpacks[0].Layers["agent"].Metadata["built_at"]).NotTo(Equal(firstImage.Buildpacks[0].Layers["agent"].Metadata["built_at"]))
		})
	})
}
