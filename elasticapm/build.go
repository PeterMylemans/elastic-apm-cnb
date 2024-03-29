package elasticapm

import (
	"fmt"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/postal"
	"github.com/paketo-buildpacks/packit/scribe"
	"path/filepath"
)

//go:generate faux -i DependencyManager -o fakes/dependency_manager.go
type DependencyManager interface {
	Resolve(path, id, version, stack string) (postal.Dependency, error)
	Install(dependency postal.Dependency, cnbPath, layerPath string) error
	Copy(dependency postal.Dependency, cnbPath, layerPath string) error
}

func Build(dependencies DependencyManager, logger scribe.Logger) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		entry := context.Plan.Entries[0]

		dependency, err := dependencies.Resolve(filepath.Join(context.CNBPath, "buildpack.toml"), entry.Name, "*", context.Stack)
		if err != nil {
			return packit.BuildResult{}, err
		}

		bom := packit.BuildpackPlan{
			Entries: []packit.BuildpackPlanEntry{
				{
					Name: dependency.ID,
					Metadata: map[string]interface{}{
						"licenses": []string{},
						"name":     dependency.Name,
						"sha256":   dependency.SHA256,
						"stacks":   dependency.Stacks,
						"uri":      dependency.URI,
						"version":  dependency.Version,
					},
				},
			},
		}

		layer, err := context.Layers.Get("agent")
		if err != nil {
			return packit.BuildResult{}, err
		}

		cachedSHA, ok := layer.Metadata["dependency-sha"].(string)
		if ok && cachedSHA == dependency.SHA256 {
			logger.Process("Reusing cached layer")
			logger.Break()

			return packit.BuildResult{
				Plan:   bom,
				Layers: []packit.Layer{layer},
			}, nil
		}

		layer, err = layer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Process("Creating java agent layer")

		layer.Launch = true
		layer.Cache = true
		layer.Metadata = map[string]interface{}{
			"dependency-sha": dependency.SHA256,
		}

		path := filepath.Join(layer.Path, "apm-java-agent.jar")

		logger.Subprocess("Creating %s", path)
		err = dependencies.Copy(dependency, context.CNBPath, path)
		if err != nil {
			return packit.BuildResult{}, err
		}

		layer.LaunchEnv.Append("JAVA_TOOL_OPTIONS", fmt.Sprintf("-javaagent:%s", path), " ")

		logger.Break()
		return packit.BuildResult{
			Plan:   bom,
			Layers: []packit.Layer{layer},
		}, nil
	}
}
