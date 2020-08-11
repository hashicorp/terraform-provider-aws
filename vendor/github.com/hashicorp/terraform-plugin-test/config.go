package tftest

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-exec/tfinstall"
)

// Config is used to configure the test helper. In most normal test programs
// the configuration is discovered automatically by an Init* function using
// DiscoverConfig, but this is exposed so that more complex scenarios can be
// implemented by direct configuration.
type Config struct {
	PluginName         string
	SourceDir          string
	TerraformExec      string
	execTempDir        string
	CurrentPluginExec  string
	PreviousPluginExec string
}

// DiscoverConfig uses environment variables and other means to automatically
// discover a reasonable test helper configuration.
func DiscoverConfig(pluginName string, sourceDir string) (*Config, error) {
	tfVersion := os.Getenv("TF_ACC_TERRAFORM_VERSION")
	tfPath := os.Getenv("TF_ACC_TERRAFORM_PATH")

	tempDir := os.Getenv("TF_ACC_TEMP_DIR")
	tfDir, err := ioutil.TempDir(tempDir, "tftest-terraform")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	finders := []tfinstall.ExecPathFinder{}
	switch {
	case tfPath != "":
		finders = append(finders, tfinstall.ExactPath(tfPath))
	case tfVersion != "":
		finders = append(finders, tfinstall.ExactVersion(tfVersion, tfDir))
	default:
		finders = append(finders, tfinstall.LookPath(), tfinstall.LatestVersion(tfDir, true))
	}
	tfExec, err := tfinstall.Find(finders...)
	if err != nil {
		return nil, err
	}

	prevExec := os.Getenv("TF_ACC_PREVIOUS_EXEC")
	if prevExec != "" {
		if info, err := os.Stat(prevExec); err != nil {
			return nil, fmt.Errorf("TF_ACC_PREVIOUS_EXEC of %s cannot be used: %s", prevExec, err)
		} else if info.IsDir() {
			return nil, fmt.Errorf("TF_ACC_PREVIOUS_EXEC of %s is directory, not file", prevExec)
		}
	}

	absPluginExecPath, err := filepath.Abs(os.Args[0])
	if err != nil {
		return nil, fmt.Errorf("could not resolve plugin exec path %s: %s", os.Args[0], err)
	}

	return &Config{
		PluginName:         pluginName,
		SourceDir:          sourceDir,
		TerraformExec:      tfExec,
		execTempDir:        tfDir,
		CurrentPluginExec:  absPluginExecPath,
		PreviousPluginExec: os.Getenv("TF_ACC_PREVIOUS_EXEC"),
	}, nil
}
