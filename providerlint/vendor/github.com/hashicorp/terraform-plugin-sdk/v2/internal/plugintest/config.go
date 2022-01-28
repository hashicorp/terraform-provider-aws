package plugintest

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hashicorp/go-version"
	install "github.com/hashicorp/hc-install"
	"github.com/hashicorp/hc-install/checkpoint"
	"github.com/hashicorp/hc-install/fs"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/hc-install/src"
)

// Config is used to configure the test helper. In most normal test programs
// the configuration is discovered automatically by an Init* function using
// DiscoverConfig, but this is exposed so that more complex scenarios can be
// implemented by direct configuration.
type Config struct {
	SourceDir          string
	TerraformExec      string
	execTempDir        string
	PreviousPluginExec string
}

// DiscoverConfig uses environment variables and other means to automatically
// discover a reasonable test helper configuration.
func DiscoverConfig(sourceDir string) (*Config, error) {
	tfVersion := strings.TrimPrefix(os.Getenv("TF_ACC_TERRAFORM_VERSION"), "v")
	tfPath := os.Getenv("TF_ACC_TERRAFORM_PATH")

	tempDir := os.Getenv("TF_ACC_TEMP_DIR")
	tfDir, err := ioutil.TempDir(tempDir, "plugintest-terraform")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	var sources []src.Source
	switch {
	case tfPath != "":
		sources = append(sources, &fs.AnyVersion{
			ExactBinPath: tfPath,
		})
	case tfVersion != "":
		tfVersion, err := version.NewVersion(tfVersion)

		if err != nil {
			return nil, fmt.Errorf("invalid Terraform version: %w", err)
		}

		sources = append(sources, &releases.ExactVersion{
			InstallDir: tfDir,
			Product:    product.Terraform,
			Version:    tfVersion,
		})
	default:
		sources = append(sources, &fs.AnyVersion{
			Product: &product.Terraform,
		})
		sources = append(sources, &checkpoint.LatestVersion{
			InstallDir: tfDir,
			Product:    product.Terraform,
		})
	}

	installer := install.NewInstaller()
	tfExec, err := installer.Ensure(context.Background(), sources)
	if err != nil {
		return nil, err
	}

	return &Config{
		SourceDir:     sourceDir,
		TerraformExec: tfExec,
		execTempDir:   tfDir,
	}, nil
}
