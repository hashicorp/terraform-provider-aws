package resource

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/davecgh/go-spew/spew"
	tfjson "github.com/hashicorp/terraform-json"
	testing "github.com/mitchellh/go-testing-interface"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/plugintest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func runPostTestDestroy(ctx context.Context, t testing.T, c TestCase, wd *plugintest.WorkingDir, factories map[string]func() (*schema.Provider, error), v5factories map[string]func() (tfprotov5.ProviderServer, error), v6factories map[string]func() (tfprotov6.ProviderServer, error), statePreDestroy *terraform.State) error {
	t.Helper()

	err := runProviderCommand(ctx, t, func() error {
		return wd.Destroy(ctx)
	}, wd, providerFactories{
		legacy:  factories,
		protov5: v5factories,
		protov6: v6factories})
	if err != nil {
		return err
	}

	if c.CheckDestroy != nil {
		logging.HelperResourceTrace(ctx, "Using TestCase CheckDestroy")
		logging.HelperResourceDebug(ctx, "Calling TestCase CheckDestroy")

		if err := c.CheckDestroy(statePreDestroy); err != nil {
			return err
		}

		logging.HelperResourceDebug(ctx, "Called TestCase CheckDestroy")
	}

	return nil
}

func runNewTest(ctx context.Context, t testing.T, c TestCase, helper *plugintest.Helper) {
	t.Helper()

	spewConf := spew.NewDefaultConfig()
	spewConf.SortKeys = true
	wd := helper.RequireNewWorkingDir(ctx, t)

	ctx = logging.TestTerraformPathContext(ctx, wd.GetHelper().TerraformExecPath())
	ctx = logging.TestWorkingDirectoryContext(ctx, wd.GetHelper().WorkingDirectory())

	defer func() {
		var statePreDestroy *terraform.State
		var err error
		err = runProviderCommand(ctx, t, func() error {
			statePreDestroy, err = getState(ctx, t, wd)
			if err != nil {
				return err
			}
			return nil
		}, wd, providerFactories{
			legacy:  c.ProviderFactories,
			protov5: c.ProtoV5ProviderFactories,
			protov6: c.ProtoV6ProviderFactories})
		if err != nil {
			t.Fatalf("Error retrieving state, there may be dangling resources: %s", err.Error())
			return
		}

		if !stateIsEmpty(statePreDestroy) {
			err := runPostTestDestroy(ctx, t, c, wd, c.ProviderFactories, c.ProtoV5ProviderFactories, c.ProtoV6ProviderFactories, statePreDestroy)
			if err != nil {
				t.Fatalf("Error running post-test destroy, there may be dangling resources: %s", err.Error())
			}
		}

		wd.Close()
	}()

	providerCfg, err := testProviderConfig(c)
	if err != nil {
		t.Fatal(err)
	}

	err = wd.SetConfig(ctx, providerCfg)
	if err != nil {
		t.Fatalf("Error setting test config: %s", err)
	}
	err = runProviderCommand(ctx, t, func() error {
		return wd.Init(ctx)
	}, wd, providerFactories{
		legacy:  c.ProviderFactories,
		protov5: c.ProtoV5ProviderFactories,
		protov6: c.ProtoV6ProviderFactories})
	if err != nil {
		t.Fatalf("Error running init: %s", err.Error())
		return
	}

	logging.HelperResourceDebug(ctx, "Starting TestSteps")

	// use this to track last step succesfully applied
	// acts as default for import tests
	var appliedCfg string

	for i, step := range c.Steps {
		ctx = logging.TestStepNumberContext(ctx, i+1)

		logging.HelperResourceDebug(ctx, "Starting TestStep")

		if step.PreConfig != nil {
			logging.HelperResourceDebug(ctx, "Calling TestStep PreConfig")
			step.PreConfig()
			logging.HelperResourceDebug(ctx, "Called TestStep PreConfig")
		}

		if step.SkipFunc != nil {
			logging.HelperResourceDebug(ctx, "Calling TestStep SkipFunc")

			skip, err := step.SkipFunc()
			if err != nil {
				t.Fatal(err)
			}

			logging.HelperResourceDebug(ctx, "Called TestStep SkipFunc")

			if skip {
				t.Logf("Skipping step %d/%d due to SkipFunc", i+1, len(c.Steps))
				logging.HelperResourceWarn(ctx, "Skipping TestStep due to SkipFunc")
				continue
			}
		}

		if step.ImportState {
			logging.HelperResourceTrace(ctx, "TestStep is ImportState mode")

			err := testStepNewImportState(ctx, t, c, helper, wd, step, appliedCfg)
			if step.ExpectError != nil {
				logging.HelperResourceDebug(ctx, "Checking TestStep ExpectError")
				if err == nil {
					t.Fatalf("Step %d/%d error running import: expected an error but got none", i+1, len(c.Steps))
				}
				if !step.ExpectError.MatchString(err.Error()) {
					t.Fatalf("Step %d/%d error running import, expected an error with pattern (%s), no match on: %s", i+1, len(c.Steps), step.ExpectError.String(), err)
				}
			} else {
				if err != nil && c.ErrorCheck != nil {
					logging.HelperResourceDebug(ctx, "Calling TestCase ErrorCheck")
					err = c.ErrorCheck(err)
				}
				if err != nil {
					t.Fatalf("Step %d/%d error running import: %s", i+1, len(c.Steps), err)
				}
			}

			logging.HelperResourceDebug(ctx, "Finished TestStep")

			continue
		}

		if step.Config != "" {
			logging.HelperResourceTrace(ctx, "TestStep is Config mode")

			err := testStepNewConfig(ctx, t, c, wd, step)
			if step.ExpectError != nil {
				logging.HelperResourceDebug(ctx, "Checking TestStep ExpectError")

				if err == nil {
					t.Fatalf("Step %d/%d, expected an error but got none", i+1, len(c.Steps))
				}
				if !step.ExpectError.MatchString(err.Error()) {
					t.Fatalf("Step %d/%d, expected an error with pattern, no match on: %s", i+1, len(c.Steps), err)
				}
			} else {
				if err != nil && c.ErrorCheck != nil {
					logging.HelperResourceDebug(ctx, "Calling TestCase ErrorCheck")

					err = c.ErrorCheck(err)

					logging.HelperResourceDebug(ctx, "Called TestCase ErrorCheck")
				}
				if err != nil {
					t.Fatalf("Step %d/%d error: %s", i+1, len(c.Steps), err)
				}
			}

			appliedCfg = step.Config

			logging.HelperResourceDebug(ctx, "Finished TestStep")

			continue
		}

		t.Fatalf("Step %d/%d, unsupported test mode", i+1, len(c.Steps))
	}
}

func getState(ctx context.Context, t testing.T, wd *plugintest.WorkingDir) (*terraform.State, error) {
	t.Helper()

	jsonState, err := wd.State(ctx)
	if err != nil {
		return nil, err
	}
	state, err := shimStateFromJson(jsonState)
	if err != nil {
		t.Fatal(err)
	}
	return state, nil
}

func stateIsEmpty(state *terraform.State) bool {
	return state.Empty() || !state.HasResources()
}

func planIsEmpty(plan *tfjson.Plan) bool {
	for _, rc := range plan.ResourceChanges {
		for _, a := range rc.Change.Actions {
			if a != tfjson.ActionNoop {
				return false
			}
		}
	}
	return true
}

func testIDRefresh(ctx context.Context, t testing.T, c TestCase, wd *plugintest.WorkingDir, step TestStep, r *terraform.ResourceState) error {
	t.Helper()

	spewConf := spew.NewDefaultConfig()
	spewConf.SortKeys = true

	// Build the state. The state is just the resource with an ID. There
	// are no attributes. We only set what is needed to perform a refresh.
	state := terraform.NewState()
	state.RootModule().Resources = make(map[string]*terraform.ResourceState)
	state.RootModule().Resources[c.IDRefreshName] = &terraform.ResourceState{}

	// Temporarily set the config to a minimal provider config for the refresh
	// test. After the refresh we can reset it.
	cfg, err := testProviderConfig(c)
	if err != nil {
		return err
	}
	err = wd.SetConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("Error setting import test config: %s", err)
	}
	defer func() {
		err = wd.SetConfig(ctx, step.Config)
		if err != nil {
			t.Fatalf("Error resetting test config: %s", err)
		}
	}()

	// Refresh!
	err = runProviderCommand(ctx, t, func() error {
		err = wd.Refresh(ctx)
		if err != nil {
			t.Fatalf("Error running terraform refresh: %s", err)
		}
		state, err = getState(ctx, t, wd)
		if err != nil {
			return err
		}
		return nil
	}, wd, providerFactories{
		legacy:  c.ProviderFactories,
		protov5: c.ProtoV5ProviderFactories,
		protov6: c.ProtoV6ProviderFactories})
	if err != nil {
		return err
	}

	// Verify attribute equivalence.
	actualR := state.RootModule().Resources[c.IDRefreshName]
	if actualR == nil {
		return fmt.Errorf("Resource gone!")
	}
	if actualR.Primary == nil {
		return fmt.Errorf("Resource has no primary instance")
	}
	actual := actualR.Primary.Attributes
	expected := r.Primary.Attributes

	if len(c.IDRefreshIgnore) > 0 {
		logging.HelperResourceTrace(ctx, fmt.Sprintf("Using TestCase IDRefreshIgnore: %v", c.IDRefreshIgnore))
	}

	// Remove fields we're ignoring
	for _, v := range c.IDRefreshIgnore {
		for k := range actual {
			if strings.HasPrefix(k, v) {
				delete(actual, k)
			}
		}
		for k := range expected {
			if strings.HasPrefix(k, v) {
				delete(expected, k)
			}
		}
	}

	if !reflect.DeepEqual(actual, expected) {
		// Determine only the different attributes
		for k, v := range expected {
			if av, ok := actual[k]; ok && v == av {
				delete(expected, k)
				delete(actual, k)
			}
		}

		spewConf := spew.NewDefaultConfig()
		spewConf.SortKeys = true
		return fmt.Errorf(
			"Attributes not equivalent. Difference is shown below. Top is actual, bottom is expected."+
				"\n\n%s\n\n%s",
			spewConf.Sdump(actual), spewConf.Sdump(expected))
	}

	return nil
}
