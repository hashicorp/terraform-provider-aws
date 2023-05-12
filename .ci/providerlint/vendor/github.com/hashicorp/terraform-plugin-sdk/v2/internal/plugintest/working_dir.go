package plugintest

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/logging"
)

const (
	ConfigFileName     = "terraform_plugin_test.tf"
	ConfigFileNameJSON = ConfigFileName + ".json"
	PlanFileName       = "tfplan"
)

// WorkingDir represents a distinct working directory that can be used for
// running tests. Each test should construct its own WorkingDir by calling
// NewWorkingDir or RequireNewWorkingDir on its package's singleton
// plugintest.Helper.
type WorkingDir struct {
	h *Helper

	// baseDir is the root of the working directory tree
	baseDir string

	// configFilename is the full filename where the latest configuration
	// was stored; empty until SetConfig is called.
	configFilename string

	// tf is the instance of tfexec.Terraform used for running Terraform commands
	tf *tfexec.Terraform

	// terraformExec is a path to a terraform binary, inherited from Helper
	terraformExec string

	// reattachInfo stores the gRPC socket info required for Terraform's
	// plugin reattach functionality
	reattachInfo tfexec.ReattachInfo
}

// Close deletes the directories and files created to represent the receiving
// working directory. After this method is called, the working directory object
// is invalid and may no longer be used.
func (wd *WorkingDir) Close() error {
	return os.RemoveAll(wd.baseDir)
}

func (wd *WorkingDir) SetReattachInfo(ctx context.Context, reattachInfo tfexec.ReattachInfo) {
	logging.HelperResourceTrace(ctx, "Setting Terraform CLI reattach configuration", map[string]interface{}{"tf_reattach_config": reattachInfo})
	wd.reattachInfo = reattachInfo
}

func (wd *WorkingDir) UnsetReattachInfo() {
	wd.reattachInfo = nil
}

// GetHelper returns the Helper set on the WorkingDir.
func (wd *WorkingDir) GetHelper() *Helper {
	return wd.h
}

// SetConfig sets a new configuration for the working directory.
//
// This must be called at least once before any call to Init, Plan, Apply, or
// Destroy to establish the configuration. Any previously-set configuration is
// discarded and any saved plan is cleared.
func (wd *WorkingDir) SetConfig(ctx context.Context, cfg string) error {
	logging.HelperResourceTrace(ctx, "Setting Terraform configuration", map[string]any{logging.KeyTestTerraformConfiguration: cfg})

	outFilename := filepath.Join(wd.baseDir, ConfigFileName)
	rmFilename := filepath.Join(wd.baseDir, ConfigFileNameJSON)
	bCfg := []byte(cfg)
	if json.Valid(bCfg) {
		outFilename, rmFilename = rmFilename, outFilename
	}
	if err := os.Remove(rmFilename); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("unable to remove %q: %w", rmFilename, err)
	}
	err := os.WriteFile(outFilename, bCfg, 0700)
	if err != nil {
		return err
	}
	wd.configFilename = outFilename

	// Changing configuration invalidates any saved plan.
	err = wd.ClearPlan(ctx)
	if err != nil {
		return err
	}
	return nil
}

// ClearState deletes any Terraform state present in the working directory.
//
// Any remote objects tracked by the state are not destroyed first, so this
// will leave them dangling in the remote system.
func (wd *WorkingDir) ClearState(ctx context.Context) error {
	logging.HelperResourceTrace(ctx, "Clearing Terraform state")

	err := os.Remove(filepath.Join(wd.baseDir, "terraform.tfstate"))

	if os.IsNotExist(err) {
		logging.HelperResourceTrace(ctx, "No Terraform state to clear")
		return nil
	}

	if err != nil {
		return err
	}

	logging.HelperResourceTrace(ctx, "Cleared Terraform state")

	return nil
}

// ClearPlan deletes any saved plan present in the working directory.
func (wd *WorkingDir) ClearPlan(ctx context.Context) error {
	logging.HelperResourceTrace(ctx, "Clearing Terraform plan")

	err := os.Remove(wd.planFilename())

	if os.IsNotExist(err) {
		logging.HelperResourceTrace(ctx, "No Terraform plan to clear")
		return nil
	}

	if err != nil {
		return err
	}

	logging.HelperResourceTrace(ctx, "Cleared Terraform plan")

	return nil
}

var errWorkingDirSetConfigNotCalled = fmt.Errorf("must call SetConfig before Init")

// Init runs "terraform init" for the given working directory, forcing Terraform
// to use the current version of the plugin under test.
func (wd *WorkingDir) Init(ctx context.Context) error {
	if wd.configFilename == "" {
		return errWorkingDirSetConfigNotCalled
	}
	if _, err := os.Stat(wd.configFilename); err != nil {
		return errWorkingDirSetConfigNotCalled
	}

	logging.HelperResourceTrace(ctx, "Calling Terraform CLI init command")

	// -upgrade=true is required for per-TestStep provider version changes
	// e.g. TestTest_TestStep_ExternalProviders_DifferentVersions
	err := wd.tf.Init(context.Background(), tfexec.Reattach(wd.reattachInfo), tfexec.Upgrade(true))

	logging.HelperResourceTrace(ctx, "Called Terraform CLI init command")

	return err
}

func (wd *WorkingDir) planFilename() string {
	return filepath.Join(wd.baseDir, PlanFileName)
}

// CreatePlan runs "terraform plan" to create a saved plan file, which if successful
// will then be used for the next call to Apply.
func (wd *WorkingDir) CreatePlan(ctx context.Context) error {
	logging.HelperResourceTrace(ctx, "Calling Terraform CLI plan command")

	hasChanges, err := wd.tf.Plan(context.Background(), tfexec.Reattach(wd.reattachInfo), tfexec.Refresh(false), tfexec.Out(PlanFileName))

	logging.HelperResourceTrace(ctx, "Called Terraform CLI plan command")

	if err != nil {
		return err
	}

	if !hasChanges {
		logging.HelperResourceTrace(ctx, "Created plan with no changes")

		return nil
	}

	stdout, err := wd.SavedPlanRawStdout(ctx)

	if err != nil {
		return fmt.Errorf("error retrieving formatted plan output: %w", err)
	}

	logging.HelperResourceTrace(ctx, "Created plan with changes", map[string]any{logging.KeyTestTerraformPlan: stdout})

	return nil
}

// CreateDestroyPlan runs "terraform plan -destroy" to create a saved plan
// file, which if successful will then be used for the next call to Apply.
func (wd *WorkingDir) CreateDestroyPlan(ctx context.Context) error {
	logging.HelperResourceTrace(ctx, "Calling Terraform CLI plan -destroy command")

	hasChanges, err := wd.tf.Plan(context.Background(), tfexec.Reattach(wd.reattachInfo), tfexec.Refresh(false), tfexec.Out(PlanFileName), tfexec.Destroy(true))

	logging.HelperResourceTrace(ctx, "Called Terraform CLI plan -destroy command")

	if err != nil {
		return err
	}

	if !hasChanges {
		logging.HelperResourceTrace(ctx, "Created destroy plan with no changes")

		return nil
	}

	stdout, err := wd.SavedPlanRawStdout(ctx)

	if err != nil {
		return fmt.Errorf("error retrieving formatted plan output: %w", err)
	}

	logging.HelperResourceTrace(ctx, "Created destroy plan with changes", map[string]any{logging.KeyTestTerraformPlan: stdout})

	return nil
}

// Apply runs "terraform apply". If CreatePlan has previously completed
// successfully and the saved plan has not been cleared in the meantime then
// this will apply the saved plan. Otherwise, it will implicitly create a new
// plan and apply it.
func (wd *WorkingDir) Apply(ctx context.Context) error {
	args := []tfexec.ApplyOption{tfexec.Reattach(wd.reattachInfo), tfexec.Refresh(false)}
	if wd.HasSavedPlan() {
		args = append(args, tfexec.DirOrPlan(PlanFileName))
	}

	logging.HelperResourceTrace(ctx, "Calling Terraform CLI apply command")

	err := wd.tf.Apply(context.Background(), args...)

	logging.HelperResourceTrace(ctx, "Called Terraform CLI apply command")

	return err
}

// Destroy runs "terraform destroy". It does not consider or modify any saved
// plan, and is primarily for cleaning up at the end of a test run.
//
// If destroy fails then remote objects might still exist, and continue to
// exist after a particular test is concluded.
func (wd *WorkingDir) Destroy(ctx context.Context) error {
	logging.HelperResourceTrace(ctx, "Calling Terraform CLI destroy command")

	err := wd.tf.Destroy(context.Background(), tfexec.Reattach(wd.reattachInfo), tfexec.Refresh(false))

	logging.HelperResourceTrace(ctx, "Called Terraform CLI destroy command")

	return err
}

// HasSavedPlan returns true if there is a saved plan in the working directory. If
// so, a subsequent call to Apply will apply that saved plan.
func (wd *WorkingDir) HasSavedPlan() bool {
	_, err := os.Stat(wd.planFilename())
	return err == nil
}

// SavedPlan returns an object describing the current saved plan file, if any.
//
// If no plan is saved or if the plan file cannot be read, SavedPlan returns
// an error.
func (wd *WorkingDir) SavedPlan(ctx context.Context) (*tfjson.Plan, error) {
	if !wd.HasSavedPlan() {
		return nil, fmt.Errorf("there is no current saved plan")
	}

	logging.HelperResourceTrace(ctx, "Calling Terraform CLI show command for JSON plan")

	plan, err := wd.tf.ShowPlanFile(context.Background(), wd.planFilename(), tfexec.Reattach(wd.reattachInfo))

	logging.HelperResourceTrace(ctx, "Calling Terraform CLI show command for JSON plan")

	return plan, err
}

// SavedPlanRawStdout returns a human readable stdout capture of the current saved plan file, if any.
//
// If no plan is saved or if the plan file cannot be read, SavedPlanRawStdout returns
// an error.
func (wd *WorkingDir) SavedPlanRawStdout(ctx context.Context) (string, error) {
	if !wd.HasSavedPlan() {
		return "", fmt.Errorf("there is no current saved plan")
	}

	logging.HelperResourceTrace(ctx, "Calling Terraform CLI show command for stdout plan")

	stdout, err := wd.tf.ShowPlanFileRaw(context.Background(), wd.planFilename(), tfexec.Reattach(wd.reattachInfo))

	logging.HelperResourceTrace(ctx, "Called Terraform CLI show command for stdout plan")

	if err != nil {
		return "", err
	}

	return stdout, nil
}

// State returns an object describing the current state.
//

// If the state cannot be read, State returns an error.
func (wd *WorkingDir) State(ctx context.Context) (*tfjson.State, error) {
	logging.HelperResourceTrace(ctx, "Calling Terraform CLI show command for JSON state")

	state, err := wd.tf.Show(context.Background(), tfexec.Reattach(wd.reattachInfo))

	logging.HelperResourceTrace(ctx, "Called Terraform CLI show command for JSON state")

	return state, err
}

// Import runs terraform import
func (wd *WorkingDir) Import(ctx context.Context, resource, id string) error {
	logging.HelperResourceTrace(ctx, "Calling Terraform CLI import command")

	err := wd.tf.Import(context.Background(), resource, id, tfexec.Config(wd.baseDir), tfexec.Reattach(wd.reattachInfo))

	logging.HelperResourceTrace(ctx, "Called Terraform CLI import command")

	return err
}

// Taint runs terraform taint
func (wd *WorkingDir) Taint(ctx context.Context, address string) error {
	logging.HelperResourceTrace(ctx, "Calling Terraform CLI taint command")

	err := wd.tf.Taint(context.Background(), address)

	logging.HelperResourceTrace(ctx, "Called Terraform CLI taint command")

	return err
}

// Refresh runs terraform refresh
func (wd *WorkingDir) Refresh(ctx context.Context) error {
	logging.HelperResourceTrace(ctx, "Calling Terraform CLI refresh command")

	err := wd.tf.Refresh(context.Background(), tfexec.Reattach(wd.reattachInfo))

	logging.HelperResourceTrace(ctx, "Called Terraform CLI refresh command")

	return err
}

// Schemas returns an object describing the provider schemas.
//
// If the schemas cannot be read, Schemas returns an error.
func (wd *WorkingDir) Schemas(ctx context.Context) (*tfjson.ProviderSchemas, error) {
	logging.HelperResourceTrace(ctx, "Calling Terraform CLI providers schema command")

	providerSchemas, err := wd.tf.ProvidersSchema(context.Background())

	logging.HelperResourceTrace(ctx, "Called Terraform CLI providers schema command")

	return providerSchemas, err
}
