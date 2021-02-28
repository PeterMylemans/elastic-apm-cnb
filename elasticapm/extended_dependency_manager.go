package elasticapm

import (
	"fmt"
	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/paketo-buildpacks/packit/postal"
	"io"
	"os"
	"path/filepath"
)

type ExtendedDependencyManager struct {
	postal.Service
	transport cargo.Transport
}

func NewExtendedDependencyManager() ExtendedDependencyManager {
	transport := cargo.NewTransport()
	return ExtendedDependencyManager{
		Service:   postal.NewService(transport),
		transport: transport,
	}
}

func (e ExtendedDependencyManager) Copy(dependency postal.Dependency, cnbPath, layerPath string) error {
	source, err := e.transport.Drop(cnbPath, dependency.URI)
	if err != nil {
		return fmt.Errorf("failed to fetch dependency: %s", err)
	}

	validatedReader := cargo.NewValidatedReader(source, dependency.SHA256)

	err = os.MkdirAll(filepath.Dir(layerPath), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create destination directory: %s", err)
	}

	destination, err := os.Create(layerPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %s", err)
	}

	_, err = io.Copy(destination, validatedReader)
	if err != nil {
		return fmt.Errorf("failed to copy dependency: %s", err)
	}

	err = destination.Close()
	if err != nil {
		return fmt.Errorf("failed to close dependency destination: %s", err)
	}

	err = source.Close()
	if err != nil {
		return fmt.Errorf("failed to close dependency source: %s", err)
	}

	return nil
}
