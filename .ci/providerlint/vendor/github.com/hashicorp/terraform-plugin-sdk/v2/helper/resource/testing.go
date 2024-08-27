// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resource

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/go-testing-interface"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/addrs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/plugintest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// flagSweep is a flag available when running tests on the command line. It
// contains a comma seperated list of regions to for the sweeper functions to
// run in.  This flag bypasses the normal Test path and instead runs functions designed to
// clean up any leaked resources a testing environment could have created. It is
// a best effort attempt, and relies on Provider authors to implement "Sweeper"
// methods for resources.

// Adding Sweeper methods with AddTestSweepers will
// construct a list of sweeper funcs to be called here. We iterate through
// regions provided by the sweep flag, and for each region we iterate through the
// tests, and exit on any errors. At time of writing, sweepers are ran
// sequentially, however they can list dependencies to be ran first. We track
// the sweepers that have been ran, so as to not run a sweeper twice for a given
// region.
//
// WARNING:
// Sweepers are designed to be destructive. You should not use the -sweep flag
// in any environment that is not strictly a test environment. Resources will be
// destroyed.

var flagSweep = flag.String("sweep", "", "List of Regions to run available Sweepers")
var flagSweepAllowFailures = flag.Bool("sweep-allow-failures", false, "Enable to allow Sweeper Tests to continue after failures")
var flagSweepRun = flag.String("sweep-run", "", "Comma seperated list of Sweeper Tests to run")
var sweeperFuncs map[string]*Sweeper

// SweeperFunc is a signature for a function that acts as a sweeper. It
// accepts a string for the region that the sweeper is to be ran in. This
// function must be able to construct a valid client for that region.
type SweeperFunc func(r string) error

type Sweeper struct {
	// Name for sweeper. Must be unique to be ran by the Sweeper Runner
	Name string

	// Dependencies list the const names of other Sweeper functions that must be ran
	// prior to running this Sweeper. This is an ordered list that will be invoked
	// recursively at the helper/resource level
	Dependencies []string

	// Sweeper function that when invoked sweeps the Provider of specific
	// resources
	F SweeperFunc
}

func init() {
	sweeperFuncs = make(map[string]*Sweeper)
}

// AddTestSweepers function adds a given name and Sweeper configuration
// pair to the internal sweeperFuncs map. Invoke this function to register a
// resource sweeper to be available for running when the -sweep flag is used
// with `go test`. Sweeper names must be unique to help ensure a given sweeper
// is only ran once per run.
func AddTestSweepers(name string, s *Sweeper) {
	if _, ok := sweeperFuncs[name]; ok {
		log.Fatalf("[ERR] Error adding (%s) to sweeperFuncs: function already exists in map", name)
	}

	sweeperFuncs[name] = s
}

// TestMain adds sweeper functionality to the "go test" command, otherwise
// tests are executed as normal. Most provider acceptance tests are written
// using the Test() function of this package, which imposes its own
// requirements and Terraform CLI behavior. Refer to that function's
// documentation for additional details.
//
// Sweepers enable infrastructure cleanup functions to be included with
// resource definitions, typically so developers can remove all resources of
// that resource type from testing infrastructure in case of failures that
// prevented the normal resource destruction behavior of acceptance tests.
// Use the AddTestSweepers() function to configure available sweepers.
//
// Sweeper flags added to the "go test" command:
//
//	-sweep: Comma-separated list of locations/regions to run available sweepers.
//	-sweep-allow-failues: Enable to allow other sweepers to run after failures.
//	-sweep-run: Comma-separated list of resource type sweepers to run. Defaults
//	        to all sweepers.
//
// Refer to the Env prefixed constants for environment variables that further
// control testing functionality.
func TestMain(m interface {
	Run() int
}) {
	flag.Parse()
	if *flagSweep != "" {
		// parse flagSweep contents for regions to run
		regions := strings.Split(*flagSweep, ",")

		// get filtered list of sweepers to run based on sweep-run flag
		sweepers := filterSweepers(*flagSweepRun, sweeperFuncs)

		if _, err := runSweepers(regions, sweepers, *flagSweepAllowFailures); err != nil {
			os.Exit(1)
		}
	} else {
		exitCode := m.Run()
		os.Exit(exitCode)
	}
}

func runSweepers(regions []string, sweepers map[string]*Sweeper, allowFailures bool) (map[string]map[string]error, error) {
	var sweeperErrorFound bool
	sweeperRunList := make(map[string]map[string]error)

	for _, region := range regions {
		region = strings.TrimSpace(region)

		var regionSweeperErrorFound bool
		regionSweeperRunList := make(map[string]error)

		start := time.Now()
		log.Printf("[DEBUG] Running Sweepers for region (%s):\n", region)
		for _, sweeper := range sweepers {
			if err := runSweeperWithRegion(region, sweeper, sweepers, regionSweeperRunList, allowFailures); err != nil {
				if allowFailures {
					continue
				}

				sweeperRunList[region] = regionSweeperRunList
				return sweeperRunList, fmt.Errorf("sweeper (%s) for region (%s) failed: %s", sweeper.Name, region, err)
			}
		}
		elapsed := time.Since(start)
		log.Printf("Completed Sweepers for region (%s) in %s", region, elapsed)

		log.Printf("Sweeper Tests for region (%s) ran successfully:\n", region)
		for sweeper, sweeperErr := range regionSweeperRunList {
			if sweeperErr == nil {
				fmt.Printf("\t- %s\n", sweeper)
			} else {
				regionSweeperErrorFound = true
			}
		}

		if regionSweeperErrorFound {
			sweeperErrorFound = true
			log.Printf("Sweeper Tests for region (%s) ran unsuccessfully:\n", region)
			for sweeper, sweeperErr := range regionSweeperRunList {
				if sweeperErr != nil {
					fmt.Printf("\t- %s: %s\n", sweeper, sweeperErr)
				}
			}
		}

		sweeperRunList[region] = regionSweeperRunList
	}

	if sweeperErrorFound {
		return sweeperRunList, errors.New("at least one sweeper failed")
	}

	return sweeperRunList, nil
}

// filterSweepers takes a comma seperated string listing the names of sweepers
// to be ran, and returns a filtered set from the list of all of sweepers to
// run based on the names given.
func filterSweepers(f string, source map[string]*Sweeper) map[string]*Sweeper {
	filterSlice := strings.Split(strings.ToLower(f), ",")
	if len(filterSlice) == 1 && filterSlice[0] == "" {
		// if the filter slice is a single element of "" then no sweeper list was
		// given, so just return the full list
		return source
	}

	sweepers := make(map[string]*Sweeper)
	for name := range source {
		for _, s := range filterSlice {
			if strings.Contains(strings.ToLower(name), s) {
				for foundName, foundSweeper := range filterSweeperWithDependencies(name, source) {
					sweepers[foundName] = foundSweeper
				}
			}
		}
	}
	return sweepers
}

// filterSweeperWithDependencies recursively returns sweeper and all dependencies.
// Since filterSweepers performs fuzzy matching, this function is used
// to perform exact sweeper and dependency lookup.
func filterSweeperWithDependencies(name string, source map[string]*Sweeper) map[string]*Sweeper {
	result := make(map[string]*Sweeper)

	currentSweeper, ok := source[name]
	if !ok {
		log.Printf("[WARN] Sweeper has dependency (%s), but that sweeper was not found", name)
		return result
	}

	result[name] = currentSweeper

	for _, dependency := range currentSweeper.Dependencies {
		for foundName, foundSweeper := range filterSweeperWithDependencies(dependency, source) {
			result[foundName] = foundSweeper
		}
	}

	return result
}

// runSweeperWithRegion recieves a sweeper and a region, and recursively calls
// itself with that region for every dependency found for that sweeper. If there
// are no dependencies, invoke the contained sweeper fun with the region, and
// add the success/fail status to the sweeperRunList.
func runSweeperWithRegion(region string, s *Sweeper, sweepers map[string]*Sweeper, sweeperRunList map[string]error, allowFailures bool) error {
	for _, dep := range s.Dependencies {
		depSweeper, ok := sweepers[dep]

		if !ok {
			log.Printf("[ERROR] Sweeper (%s) has dependency (%s), but that sweeper was not found", s.Name, dep)
			return fmt.Errorf("sweeper (%s) has dependency (%s), but that sweeper was not found", s.Name, dep)
		}

		log.Printf("[DEBUG] Sweeper (%s) has dependency (%s), running..", s.Name, dep)
		err := runSweeperWithRegion(region, depSweeper, sweepers, sweeperRunList, allowFailures)

		if err != nil {
			if allowFailures {
				log.Printf("[ERROR] Error running Sweeper (%s) in region (%s): %s", depSweeper.Name, region, err)
				continue
			}

			return err
		}
	}

	if _, ok := sweeperRunList[s.Name]; ok {
		log.Printf("[DEBUG] Sweeper (%s) already ran in region (%s)", s.Name, region)
		return nil
	}

	log.Printf("[DEBUG] Running Sweeper (%s) in region (%s)", s.Name, region)

	start := time.Now()
	runE := s.F(region)
	elapsed := time.Since(start)

	log.Printf("[DEBUG] Completed Sweeper (%s) in region (%s) in %s", s.Name, region, elapsed)

	sweeperRunList[s.Name] = runE

	if runE != nil {
		log.Printf("[ERROR] Error running Sweeper (%s) in region (%s): %s", s.Name, region, runE)
	}

	return runE
}

// Deprecated: Use EnvTfAcc instead.
const TestEnvVar = EnvTfAcc

// TestCheckFunc is the callback type used with acceptance tests to check
// the state of a resource. The state passed in is the latest state known,
// or in the case of being after a destroy, it is the last known state when
// it was created.
type TestCheckFunc func(*terraform.State) error

// ImportStateCheckFunc is the check function for ImportState tests
type ImportStateCheckFunc func([]*terraform.InstanceState) error

// ImportStateIdFunc is an ID generation function to help with complex ID
// generation for ImportState tests.
type ImportStateIdFunc func(*terraform.State) (string, error)

// ErrorCheckFunc is a function providers can use to handle errors.
type ErrorCheckFunc func(error) error

// TestCase is a single acceptance test case used to test the apply/destroy
// lifecycle of a resource in a specific configuration.
//
// When the destroy plan is executed, the config from the last TestStep
// is used to plan it.
//
// Refer to the Env prefixed constants for environment variables that further
// control testing functionality.
type TestCase struct {
	// IsUnitTest allows a test to run regardless of the TF_ACC
	// environment variable. This should be used with care - only for
	// fast tests on local resources (e.g. remote state with a local
	// backend) but can be used to increase confidence in correct
	// operation of Terraform without waiting for a full acctest run.
	IsUnitTest bool

	// PreCheck, if non-nil, will be called before any test steps are
	// executed. It will only be executed in the case that the steps
	// would run, so it can be used for some validation before running
	// acceptance tests, such as verifying that keys are setup.
	PreCheck func()

	// ProviderFactories can be specified for the providers that are valid.
	//
	// This can also be specified at the TestStep level to enable per-step
	// differences in providers, however all provider specifications must
	// be done either at the TestCase level or TestStep level, otherwise the
	// testing framework will raise an error and fail the test.
	//
	// These are the providers that can be referenced within the test. Each key
	// is an individually addressable provider. Typically you will only pass a
	// single value here for the provider you are testing. Aliases are not
	// supported by the test framework, so to use multiple provider instances,
	// you should add additional copies to this map with unique names. To set
	// their configuration, you would reference them similar to the following:
	//
	//  provider "my_factory_key" {
	//    # ...
	//  }
	//
	//  resource "my_resource" "mr" {
	//    provider = my_factory_key
	//
	//    # ...
	//  }
	ProviderFactories map[string]func() (*schema.Provider, error)

	// ProtoV5ProviderFactories serves the same purpose as ProviderFactories,
	// but for protocol v5 providers defined using the terraform-plugin-go
	// ProviderServer interface.
	//
	// This can also be specified at the TestStep level to enable per-step
	// differences in providers, however all provider specifications must
	// be done either at the TestCase level or TestStep level, otherwise the
	// testing framework will raise an error and fail the test.
	ProtoV5ProviderFactories map[string]func() (tfprotov5.ProviderServer, error)

	// ProtoV6ProviderFactories serves the same purpose as ProviderFactories,
	// but for protocol v6 providers defined using the terraform-plugin-go
	// ProviderServer interface.
	// The version of Terraform used in acceptance testing must be greater
	// than or equal to v0.15.4 to use ProtoV6ProviderFactories.
	//
	// This can also be specified at the TestStep level to enable per-step
	// differences in providers, however all provider specifications must
	// be done either at the TestCase level or TestStep level, otherwise the
	// testing framework will raise an error and fail the test.
	ProtoV6ProviderFactories map[string]func() (tfprotov6.ProviderServer, error)

	// Providers is the ResourceProvider that will be under test.
	//
	// Deprecated: Providers is deprecated, please use ProviderFactories
	Providers map[string]*schema.Provider

	// ExternalProviders are providers the TestCase relies on that should
	// be downloaded from the registry during init.
	//
	// This can also be specified at the TestStep level to enable per-step
	// differences in providers, however all provider specifications must
	// be done either at the TestCase level or TestStep level, otherwise the
	// testing framework will raise an error and fail the test.
	//
	// This is generally unnecessary to set at the TestCase level, however
	// it has existing in the testing framework prior to the introduction of
	// TestStep level specification and was only necessary for performing
	// import testing where the configuration contained a provider outside the
	// one under test.
	ExternalProviders map[string]ExternalProvider

	// PreventPostDestroyRefresh can be set to true for cases where data sources
	// are tested alongside real resources
	PreventPostDestroyRefresh bool

	// CheckDestroy is called after the resource is finally destroyed
	// to allow the tester to test that the resource is truly gone.
	CheckDestroy TestCheckFunc

	// ErrorCheck allows providers the option to handle errors such as skipping
	// tests based on certain errors.
	ErrorCheck ErrorCheckFunc

	// Steps are the apply sequences done within the context of the
	// same state. Each step can have its own check to verify correctness.
	Steps []TestStep

	// IDRefreshName is the name of the resource to check during ID-only
	// refresh testing, which ensures that a resource can be refreshed solely
	// by its identifier. This will default to the first non-nil primary
	// resource in the state. It runs every TestStep.
	//
	// While not deprecated, most resource tests should instead prefer using
	// TestStep.ImportState based testing as it works with multiple attribute
	// identifiers and also verifies resource import functionality.
	IDRefreshName string

	// IDRefreshIgnore is a list of configuration keys that will be ignored
	// during ID-only refresh testing.
	IDRefreshIgnore []string
}

// ExternalProvider holds information about third-party providers that should
// be downloaded by Terraform as part of running the test step.
type ExternalProvider struct {
	VersionConstraint string // the version constraint for the provider
	Source            string // the provider source
}

// TestStep is a single apply sequence of a test, done within the
// context of a state.
//
// Multiple TestSteps can be sequenced in a Test to allow testing
// potentially complex update logic. In general, simply create/destroy
// tests will only need one step.
//
// Refer to the Env prefixed constants for environment variables that further
// control testing functionality.
type TestStep struct {
	// ResourceName should be set to the name of the resource
	// that is being tested. Example: "aws_instance.foo". Various test
	// modes use this to auto-detect state information.
	//
	// This is only required if the test mode settings below say it is
	// for the mode you're using.
	ResourceName string

	// PreConfig is called before the Config is applied to perform any per-step
	// setup that needs to happen. This is called regardless of "test mode"
	// below.
	PreConfig func()

	// Taint is a list of resource addresses to taint prior to the execution of
	// the step. Be sure to only include this at a step where the referenced
	// address will be present in state, as it will fail the test if the resource
	// is missing.
	//
	// This option is ignored on ImportState tests, and currently only works for
	// resources in the root module path.
	Taint []string

	//---------------------------------------------------------------
	// Test modes. One of the following groups of settings must be
	// set to determine what the test step will do. Ideally we would've
	// used Go interfaces here but there are now hundreds of tests we don't
	// want to re-type so instead we just determine which step logic
	// to run based on what settings below are set.
	//---------------------------------------------------------------

	//---------------------------------------------------------------
	// Plan, Apply testing
	//---------------------------------------------------------------

	// Config a string of the configuration to give to Terraform. If this
	// is set, then the TestCase will execute this step with the same logic
	// as a `terraform apply`.
	//
	// JSON Configuration Syntax can be used and is assumed whenever Config
	// contains valid JSON.
	Config string

	// Check is called after the Config is applied. Use this step to
	// make your own API calls to check the status of things, and to
	// inspect the format of the ResourceState itself.
	//
	// If an error is returned, the test will fail. In this case, a
	// destroy plan will still be attempted.
	//
	// If this is nil, no check is done on this step.
	Check TestCheckFunc

	// Destroy will create a destroy plan if set to true.
	Destroy bool

	// ExpectNonEmptyPlan can be set to true for specific types of tests that are
	// looking to verify that a diff occurs
	ExpectNonEmptyPlan bool

	// ExpectError allows the construction of test cases that we expect to fail
	// with an error. The specified regexp must match against the error for the
	// test to pass.
	ExpectError *regexp.Regexp

	// PlanOnly can be set to only run `plan` with this configuration, and not
	// actually apply it. This is useful for ensuring config changes result in
	// no-op plans
	PlanOnly bool

	// PreventDiskCleanup can be set to true for testing terraform modules which
	// require access to disk at runtime. Note that this will leave files in the
	// temp folder
	PreventDiskCleanup bool

	// PreventPostDestroyRefresh can be set to true for cases where data sources
	// are tested alongside real resources
	PreventPostDestroyRefresh bool

	// SkipFunc enables skipping the TestStep, based on environment criteria.
	// For example, this can prevent running certain steps that may be runtime
	// platform or API configuration dependent.
	//
	// Return true with no error to skip the test step. The error return
	// should be used to signify issues that prevented the function from
	// completing as expected.
	//
	// SkipFunc is called after PreConfig but before applying the Config.
	SkipFunc func() (bool, error)

	//---------------------------------------------------------------
	// ImportState testing
	//---------------------------------------------------------------

	// ImportState, if true, will test the functionality of ImportState
	// by importing the resource with ResourceName (must be set) and the
	// ID of that resource.
	ImportState bool

	// ImportStateId is the ID to perform an ImportState operation with.
	// This is optional. If it isn't set, then the resource ID is automatically
	// determined by inspecting the state for ResourceName's ID.
	ImportStateId string

	// ImportStateIdPrefix is the prefix added in front of ImportStateId.
	// This can be useful in complex import cases, where more than one
	// attribute needs to be passed on as the Import ID. Mainly in cases
	// where the ID is not known, and a known prefix needs to be added to
	// the unset ImportStateId field.
	ImportStateIdPrefix string

	// ImportStateIdFunc is a function that can be used to dynamically generate
	// the ID for the ImportState tests. It is sent the state, which can be
	// checked to derive the attributes necessary and generate the string in the
	// desired format.
	ImportStateIdFunc ImportStateIdFunc

	// ImportStateCheck checks the results of ImportState. It should be
	// used to verify that the resulting value of ImportState has the
	// proper resources, IDs, and attributes.
	//
	// Prefer ImportStateVerify over ImportStateCheck, unless the resource
	// import explicitly is expected to create multiple resources (not a
	// recommended resource implementation) or if attributes are imported with
	// syntactically different but semantically/functionally equivalent values
	// where special logic is needed.
	//
	// Terraform versions 1.3 and later can include data source states during
	// import, which the testing framework will skip to prevent the need for
	// Terraform version specific logic in provider testing.
	ImportStateCheck ImportStateCheckFunc

	// ImportStateVerify, if true, will also check that the state values
	// that are finally put into the state after import match for all the
	// IDs returned by the Import.  Note that this checks for strict equality
	// and does not respect DiffSuppressFunc or CustomizeDiff.
	//
	// ImportStateVerifyIgnore is a list of prefixes of fields that should
	// not be verified to be equal. These can be set to ephemeral fields or
	// fields that can't be refreshed and don't matter.
	ImportStateVerify       bool
	ImportStateVerifyIgnore []string

	// ImportStatePersist, if true, will update the persisted state with the
	// state generated by the import operation (i.e., terraform import). When
	// false (default) the state generated by the import operation is discarded
	// at the end of the test step that is verifying import behavior.
	ImportStatePersist bool

	//---------------------------------------------------------------
	// RefreshState testing
	//---------------------------------------------------------------

	// RefreshState, if true, will test the functionality of `terraform
	// refresh` by refreshing the state, running any checks against the
	// refreshed state, and running a plan to verify against unexpected plan
	// differences.
	//
	// If the refresh is expected to result in a non-empty plan
	// ExpectNonEmptyPlan should be set to true in the same TestStep.
	//
	// RefreshState cannot be the first TestStep and, it is mutually exclusive
	// with ImportState.
	RefreshState bool

	// ProviderFactories can be specified for the providers that are valid for
	// this TestStep. When providers are specified at the TestStep level, all
	// TestStep within a TestCase must declare providers.
	//
	// This can also be specified at the TestCase level for all TestStep,
	// however all provider specifications must be done either at the TestCase
	// level or TestStep level, otherwise the testing framework will raise an
	// error and fail the test.
	//
	// These are the providers that can be referenced within the test. Each key
	// is an individually addressable provider. Typically you will only pass a
	// single value here for the provider you are testing. Aliases are not
	// supported by the test framework, so to use multiple provider instances,
	// you should add additional copies to this map with unique names. To set
	// their configuration, you would reference them similar to the following:
	//
	//  provider "my_factory_key" {
	//    # ...
	//  }
	//
	//  resource "my_resource" "mr" {
	//    provider = my_factory_key
	//
	//    # ...
	//  }
	ProviderFactories map[string]func() (*schema.Provider, error)

	// ProtoV5ProviderFactories serves the same purpose as ProviderFactories,
	// but for protocol v5 providers defined using the terraform-plugin-go
	// ProviderServer interface. When providers are specified at the TestStep
	// level, all TestStep within a TestCase must declare providers.
	//
	// This can also be specified at the TestCase level for all TestStep,
	// however all provider specifications must be done either at the TestCase
	// level or TestStep level, otherwise the testing framework will raise an
	// error and fail the test.
	ProtoV5ProviderFactories map[string]func() (tfprotov5.ProviderServer, error)

	// ProtoV6ProviderFactories serves the same purpose as ProviderFactories,
	// but for protocol v6 providers defined using the terraform-plugin-go
	// ProviderServer interface.
	// The version of Terraform used in acceptance testing must be greater
	// than or equal to v0.15.4 to use ProtoV6ProviderFactories. When providers
	// are specified at the TestStep level, all TestStep within a TestCase must
	// declare providers.
	//
	// This can also be specified at the TestCase level for all TestStep,
	// however all provider specifications must be done either at the TestCase
	// level or TestStep level, otherwise the testing framework will raise an
	// error and fail the test.
	ProtoV6ProviderFactories map[string]func() (tfprotov6.ProviderServer, error)

	// ExternalProviders are providers the TestStep relies on that should
	// be downloaded from the registry during init. When providers are
	// specified at the TestStep level, all TestStep within a TestCase must
	// declare providers.
	//
	// This can also be specified at the TestCase level for all TestStep,
	// however all provider specifications must be done either at the TestCase
	// level or TestStep level, otherwise the testing framework will raise an
	// error and fail the test.
	//
	// Outside specifying an earlier version of the provider under test,
	// typically for state upgrader testing, this is generally only necessary
	// for performing import testing where the prior TestStep configuration
	// contained a provider outside the one under test.
	ExternalProviders map[string]ExternalProvider
}

// ParallelTest performs an acceptance test on a resource, allowing concurrency
// with other ParallelTest. The number of concurrent tests is controlled by the
// "go test" command -parallel flag.
//
// Tests will fail if they do not properly handle conditions to allow multiple
// tests to occur against the same resource or service (e.g. random naming).
//
// Test() function requirements and documentation also apply to this function.
func ParallelTest(t testing.T, c TestCase) {
	t.Helper()
	t.Parallel()
	Test(t, c)
}

// Test performs an acceptance test on a resource.
//
// Tests are not run unless an environmental variable "TF_ACC" is
// set to some non-empty value. This is to avoid test cases surprising
// a user by creating real resources.
//
// Tests will fail unless the verbose flag (`go test -v`, or explicitly
// the "-test.v" flag) is set. Because some acceptance tests take quite
// long, we require the verbose flag so users are able to see progress
// output.
//
// Use the ParallelTest() function to automatically set (*testing.T).Parallel()
// to enable testing concurrency. Use the UnitTest() function to automatically
// set the TestCase type IsUnitTest field.
//
// This function will automatically find or install Terraform CLI into a
// temporary directory, based on the following behavior:
//
//   - If the TF_ACC_TERRAFORM_PATH environment variable is set, that
//     Terraform CLI binary is used if found and executable. If not found or
//     executable, an error will be returned unless the
//     TF_ACC_TERRAFORM_VERSION environment variable is also set.
//   - If the TF_ACC_TERRAFORM_VERSION environment variable is set, install
//     and use that Terraform CLI version.
//   - If both the TF_ACC_TERRAFORM_PATH and TF_ACC_TERRAFORM_VERSION
//     environment variables are unset, perform a lookup for the Terraform
//     CLI binary based on the operating system PATH. If not found, the
//     latest available Terraform CLI binary is installed.
//
// Refer to the Env prefixed constants for additional details about these
// environment variables, and others, that control testing functionality.
func Test(t testing.T, c TestCase) {
	t.Helper()

	ctx := context.Background()
	ctx = logging.InitTestContext(ctx, t)

	err := c.validate(ctx)

	if err != nil {
		logging.HelperResourceError(ctx,
			"Test validation error",
			map[string]interface{}{logging.KeyError: err},
		)
		t.Fatalf("Test validation error: %s", err)
	}

	// We only run acceptance tests if an env var is set because they're
	// slow and generally require some outside configuration. You can opt out
	// of this with OverrideEnvVar on individual TestCases.
	if os.Getenv(EnvTfAcc) == "" && !c.IsUnitTest {
		t.Skip(fmt.Sprintf(
			"Acceptance tests skipped unless env '%s' set",
			EnvTfAcc))
		return
	}

	// Copy any explicitly passed providers to factories, this is for backwards compatibility.
	if len(c.Providers) > 0 {
		c.ProviderFactories = map[string]func() (*schema.Provider, error){}

		for name, p := range c.Providers {
			prov := p
			c.ProviderFactories[name] = func() (*schema.Provider, error) { //nolint:unparam // required signature
				return prov, nil
			}
		}
	}

	logging.HelperResourceDebug(ctx, "Starting TestCase")

	// Run the PreCheck if we have it.
	// This is done after the auto-configure to allow providers
	// to override the default auto-configure parameters.
	if c.PreCheck != nil {
		logging.HelperResourceDebug(ctx, "Calling TestCase PreCheck")

		c.PreCheck()

		logging.HelperResourceDebug(ctx, "Called TestCase PreCheck")
	}

	sourceDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Error getting working dir: %s", err)
	}
	helper := plugintest.AutoInitProviderHelper(ctx, sourceDir)
	defer func(helper *plugintest.Helper) {
		err := helper.Close()
		if err != nil {
			logging.HelperResourceError(ctx, "Unable to clean up temporary test files", map[string]interface{}{logging.KeyError: err})
		}
	}(helper)

	runNewTest(ctx, t, c, helper)

	logging.HelperResourceDebug(ctx, "Finished TestCase")
}

// UnitTest is a helper to force the acceptance testing harness to run in the
// normal unit test suite. This should only be used for resource that don't
// have any external dependencies.
//
// Test() function requirements and documentation also apply to this function.
func UnitTest(t testing.T, c TestCase) {
	t.Helper()

	c.IsUnitTest = true
	Test(t, c)
}

func testResource(c TestStep, state *terraform.State) (*terraform.ResourceState, error) {
	for _, m := range state.Modules {
		if len(m.Resources) > 0 {
			if v, ok := m.Resources[c.ResourceName]; ok {
				return v, nil
			}
		}
	}

	return nil, fmt.Errorf(
		"Resource specified by ResourceName couldn't be found: %s", c.ResourceName)
}

// ComposeTestCheckFunc lets you compose multiple TestCheckFuncs into
// a single TestCheckFunc.
//
// As a user testing their provider, this lets you decompose your checks
// into smaller pieces more easily.
//
// ComposeTestCheckFunc returns immediately on the first TestCheckFunc error.
// To aggregrate all errors, use ComposeAggregateTestCheckFunc instead.
func ComposeTestCheckFunc(fs ...TestCheckFunc) TestCheckFunc {
	return func(s *terraform.State) error {
		for i, f := range fs {
			if err := f(s); err != nil {
				return fmt.Errorf("Check %d/%d error: %s", i+1, len(fs), err)
			}
		}

		return nil
	}
}

// ComposeAggregateTestCheckFunc lets you compose multiple TestCheckFuncs into
// a single TestCheckFunc.
//
// As a user testing their provider, this lets you decompose your checks
// into smaller pieces more easily.
//
// Unlike ComposeTestCheckFunc, ComposeAggergateTestCheckFunc runs _all_ of the
// TestCheckFuncs and aggregates failures.
func ComposeAggregateTestCheckFunc(fs ...TestCheckFunc) TestCheckFunc {
	return func(s *terraform.State) error {
		var result []error

		for i, f := range fs {
			if err := f(s); err != nil {
				result = append(result, fmt.Errorf("Check %d/%d error: %w", i+1, len(fs), err))
			}
		}

		return errors.Join(result...)
	}
}

// TestCheckResourceAttrSet ensures any value exists in the state for the
// given name and key combination. The opposite of this TestCheckFunc is
// TestCheckNoResourceAttr. State value checking is only recommended for
// testing Computed attributes and attribute defaults.
//
// Use this as a last resort when a more specific TestCheckFunc cannot be
// implemented, such as:
//
//   - TestCheckResourceAttr: Equality checking of non-TypeSet state value.
//   - TestCheckResourceAttrPair: Equality checking of non-TypeSet state
//     value, based on another state value.
//   - TestCheckTypeSet*: Equality checking of TypeSet state values.
//   - TestMatchResourceAttr: Regular expression checking of non-TypeSet
//     state value.
//   - TestMatchTypeSet*: Regular expression checking on TypeSet state values.
//
// For managed resources, the name parameter is combination of the resource
// type, a period (.), and the name label. The name for the below example
// configuration would be "myprovider_thing.example".
//
//	resource "myprovider_thing" "example" { ... }
//
// For data sources, the name parameter is a combination of the keyword "data",
// a period (.), the data source type, a period (.), and the name label. The
// name for the below example configuration would be
// "data.myprovider_thing.example".
//
//	data "myprovider_thing" "example" { ... }
//
// The key parameter is an attribute path in Terraform CLI 0.11 and earlier
// "flatmap" syntax. Keys start with the attribute name of a top-level
// attribute. Use the following special key syntax to inspect underlying
// values of a list or map attribute:
//
//   - .{NUMBER}: List value at index, e.g. .0 to inspect the first element
//   - .{KEY}: Map value at key, e.g. .example to inspect the example key
//     value
//
// While it is possible to check nested attributes under list and map
// attributes using the special key syntax, checking a list, map, or set
// attribute directly is not supported. Use TestCheckResourceAttr with
// the special .# or .% key syntax for those situations instead.
func TestCheckResourceAttrSet(name, key string) TestCheckFunc {
	return checkIfIndexesIntoTypeSet(key, func(s *terraform.State) error {
		is, err := primaryInstanceState(s, name)
		if err != nil {
			return err
		}

		return testCheckResourceAttrSet(is, name, key)
	})
}

// TestCheckModuleResourceAttrSet - as per TestCheckResourceAttrSet but with
// support for non-root modules
func TestCheckModuleResourceAttrSet(mp []string, name string, key string) TestCheckFunc {
	mpt := addrs.Module(mp).UnkeyedInstanceShim()
	return checkIfIndexesIntoTypeSet(key, func(s *terraform.State) error {
		is, err := modulePathPrimaryInstanceState(s, mpt, name)
		if err != nil {
			return err
		}

		return testCheckResourceAttrSet(is, name, key)
	})
}

func testCheckResourceAttrSet(is *terraform.InstanceState, name string, key string) error {
	val, ok := is.Attributes[key]

	if ok && val != "" {
		return nil
	}

	if _, ok := is.Attributes[key+".#"]; ok {
		return fmt.Errorf(
			"%s: list or set attribute '%s' must be checked by element count key (%s) or element value keys (e.g. %s). Set element value checks should use TestCheckTypeSet functions instead.",
			name,
			key,
			key+".#",
			key+".0",
		)
	}

	if _, ok := is.Attributes[key+".%"]; ok {
		return fmt.Errorf(
			"%s: map attribute '%s' must be checked by element count key (%s) or element value keys (e.g. %s).",
			name,
			key,
			key+".%",
			key+".examplekey",
		)
	}

	return fmt.Errorf("%s: Attribute '%s' expected to be set", name, key)
}

// TestCheckResourceAttr ensures a specific value is stored in state for the
// given name and key combination. State value checking is only recommended for
// testing Computed attributes and attribute defaults.
//
// For managed resources, the name parameter is combination of the resource
// type, a period (.), and the name label. The name for the below example
// configuration would be "myprovider_thing.example".
//
//	resource "myprovider_thing" "example" { ... }
//
// For data sources, the name parameter is a combination of the keyword "data",
// a period (.), the data source type, a period (.), and the name label. The
// name for the below example configuration would be
// "data.myprovider_thing.example".
//
//	data "myprovider_thing" "example" { ... }
//
// The key parameter is an attribute path in Terraform CLI 0.11 and earlier
// "flatmap" syntax. Keys start with the attribute name of a top-level
// attribute. Use the following special key syntax to inspect list, map, and
// set attributes:
//
//   - .{NUMBER}: List value at index, e.g. .0 to inspect the first element.
//     Use the TestCheckTypeSet* and TestMatchTypeSet* functions instead
//     for sets.
//   - .{KEY}: Map value at key, e.g. .example to inspect the example key
//     value.
//   - .#: Number of elements in list or set.
//   - .%: Number of elements in map.
//
// The value parameter is the stringified data to check at the given key. Use
// the following attribute type rules to set the value:
//
//   - Boolean: "false" or "true".
//   - Float/Integer: Stringified number, such as "1.2" or "123".
//   - String: No conversion necessary.
func TestCheckResourceAttr(name, key, value string) TestCheckFunc {
	return checkIfIndexesIntoTypeSet(key, func(s *terraform.State) error {
		is, err := primaryInstanceState(s, name)
		if err != nil {
			return err
		}

		return testCheckResourceAttr(is, name, key, value)
	})
}

// TestCheckModuleResourceAttr - as per TestCheckResourceAttr but with
// support for non-root modules
func TestCheckModuleResourceAttr(mp []string, name string, key string, value string) TestCheckFunc {
	mpt := addrs.Module(mp).UnkeyedInstanceShim()
	return checkIfIndexesIntoTypeSet(key, func(s *terraform.State) error {
		is, err := modulePathPrimaryInstanceState(s, mpt, name)
		if err != nil {
			return err
		}

		return testCheckResourceAttr(is, name, key, value)
	})
}

func testCheckResourceAttr(is *terraform.InstanceState, name string, key string, value string) error {
	v, ok := is.Attributes[key]

	if !ok {
		// Empty containers may be elided from the state.
		// If the intent here is to check for an empty container, allow the key to
		// also be non-existent.
		if value == "0" && (strings.HasSuffix(key, ".#") || strings.HasSuffix(key, ".%")) {
			return nil
		}

		if _, ok := is.Attributes[key+".#"]; ok {
			return fmt.Errorf(
				"%s: list or set attribute '%s' must be checked by element count key (%s) or element value keys (e.g. %s). Set element value checks should use TestCheckTypeSet functions instead.",
				name,
				key,
				key+".#",
				key+".0",
			)
		}

		if _, ok := is.Attributes[key+".%"]; ok {
			return fmt.Errorf(
				"%s: map attribute '%s' must be checked by element count key (%s) or element value keys (e.g. %s).",
				name,
				key,
				key+".%",
				key+".examplekey",
			)
		}

		return fmt.Errorf("%s: Attribute '%s' not found", name, key)
	}

	if v != value {
		return fmt.Errorf(
			"%s: Attribute '%s' expected %#v, got %#v",
			name,
			key,
			value,
			v)
	}

	return nil
}

// CheckResourceAttrWithFunc is the callback type used to apply a custom checking logic
// when using TestCheckResourceAttrWith and a value is found for the given name and key.
//
// When this function returns an error, TestCheckResourceAttrWith will fail the check.
type CheckResourceAttrWithFunc func(value string) error

// TestCheckResourceAttrWith ensures a value stored in state for the
// given name and key combination, is checked against a custom logic.
// State value checking is only recommended for testing Computed attributes
// and attribute defaults.
//
// For managed resources, the name parameter is combination of the resource
// type, a period (.), and the name label. The name for the below example
// configuration would be "myprovider_thing.example".
//
//	resource "myprovider_thing" "example" { ... }
//
// For data sources, the name parameter is a combination of the keyword "data",
// a period (.), the data source type, a period (.), and the name label. The
// name for the below example configuration would be
// "data.myprovider_thing.example".
//
//	data "myprovider_thing" "example" { ... }
//
// The key parameter is an attribute path in Terraform CLI 0.11 and earlier
// "flatmap" syntax. Keys start with the attribute name of a top-level
// attribute. Use the following special key syntax to inspect list, map, and
// set attributes:
//
//   - .{NUMBER}: List value at index, e.g. .0 to inspect the first element.
//     Use the TestCheckTypeSet* and TestMatchTypeSet* functions instead
//     for sets.
//   - .{KEY}: Map value at key, e.g. .example to inspect the example key
//     value.
//   - .#: Number of elements in list or set.
//   - .%: Number of elements in map.
//
// The checkValueFunc parameter is a CheckResourceAttrWithFunc,
// and it's provided with the attribute value to apply a custom checking logic,
// if it was found in the state. The function must return an error for the
// check to fail, or `nil` to succeed.
func TestCheckResourceAttrWith(name, key string, checkValueFunc CheckResourceAttrWithFunc) TestCheckFunc {
	return checkIfIndexesIntoTypeSet(key, func(s *terraform.State) error {
		is, err := primaryInstanceState(s, name)
		if err != nil {
			return err
		}

		err = testCheckResourceAttrSet(is, name, key)
		if err != nil {
			return err
		}

		err = checkValueFunc(is.Attributes[key])
		if err != nil {
			return fmt.Errorf("%s: Attribute %q value: %w", name, key, err)
		}

		return nil
	})
}

// TestCheckNoResourceAttr ensures no value exists in the state for the
// given name and key combination. The opposite of this TestCheckFunc is
// TestCheckResourceAttrSet. State value checking is only recommended for
// testing Computed attributes and attribute defaults.
//
// For managed resources, the name parameter is combination of the resource
// type, a period (.), and the name label. The name for the below example
// configuration would be "myprovider_thing.example".
//
//	resource "myprovider_thing" "example" { ... }
//
// For data sources, the name parameter is a combination of the keyword "data",
// a period (.), the data source type, a period (.), and the name label. The
// name for the below example configuration would be
// "data.myprovider_thing.example".
//
//	data "myprovider_thing" "example" { ... }
//
// The key parameter is an attribute path in Terraform CLI 0.11 and earlier
// "flatmap" syntax. Keys start with the attribute name of a top-level
// attribute. Use the following special key syntax to inspect underlying
// values of a list or map attribute:
//
//   - .{NUMBER}: List value at index, e.g. .0 to inspect the first element.
//   - .{KEY}: Map value at key, e.g. .example to inspect the example key
//     value.
//
// While it is possible to check nested attributes under list and map
// attributes using the special key syntax, checking a list, map, or set
// attribute directly is not supported. Use TestCheckResourceAttr with
// the special .# or .% key syntax for those situations instead.
func TestCheckNoResourceAttr(name, key string) TestCheckFunc {
	return checkIfIndexesIntoTypeSet(key, func(s *terraform.State) error {
		is, err := primaryInstanceState(s, name)
		if err != nil {
			return err
		}

		return testCheckNoResourceAttr(is, name, key)
	})
}

// TestCheckModuleNoResourceAttr - as per TestCheckNoResourceAttr but with
// support for non-root modules
func TestCheckModuleNoResourceAttr(mp []string, name string, key string) TestCheckFunc {
	mpt := addrs.Module(mp).UnkeyedInstanceShim()
	return checkIfIndexesIntoTypeSet(key, func(s *terraform.State) error {
		is, err := modulePathPrimaryInstanceState(s, mpt, name)
		if err != nil {
			return err
		}

		return testCheckNoResourceAttr(is, name, key)
	})
}

func testCheckNoResourceAttr(is *terraform.InstanceState, name string, key string) error {
	v, ok := is.Attributes[key]

	// Empty containers may sometimes be included in the state.
	// If the intent here is to check for an empty container, allow the value to
	// also be "0".
	if v == "0" && (strings.HasSuffix(key, ".#") || strings.HasSuffix(key, ".%")) {
		return nil
	}

	if ok {
		return fmt.Errorf("%s: Attribute '%s' found when not expected", name, key)
	}

	if _, ok := is.Attributes[key+".#"]; ok {
		return fmt.Errorf(
			"%s: list or set attribute '%s' must be checked by element count key (%s) or element value keys (e.g. %s). Set element value checks should use TestCheckTypeSet functions instead.",
			name,
			key,
			key+".#",
			key+".0",
		)
	}

	if _, ok := is.Attributes[key+".%"]; ok {
		return fmt.Errorf(
			"%s: map attribute '%s' must be checked by element count key (%s) or element value keys (e.g. %s).",
			name,
			key,
			key+".%",
			key+".examplekey",
		)
	}

	return nil
}

// TestMatchResourceAttr ensures a value matching a regular expression is
// stored in state for the given name and key combination. State value checking
// is only recommended for testing Computed attributes and attribute defaults.
//
// For managed resources, the name parameter is combination of the resource
// type, a period (.), and the name label. The name for the below example
// configuration would be "myprovider_thing.example".
//
//	resource "myprovider_thing" "example" { ... }
//
// For data sources, the name parameter is a combination of the keyword "data",
// a period (.), the data source type, a period (.), and the name label. The
// name for the below example configuration would be
// "data.myprovider_thing.example".
//
//	data "myprovider_thing" "example" { ... }
//
// The key parameter is an attribute path in Terraform CLI 0.11 and earlier
// "flatmap" syntax. Keys start with the attribute name of a top-level
// attribute. Use the following special key syntax to inspect list, map, and
// set attributes:
//
//   - .{NUMBER}: List value at index, e.g. .0 to inspect the first element.
//     Use the TestCheckTypeSet* and TestMatchTypeSet* functions instead
//     for sets.
//   - .{KEY}: Map value at key, e.g. .example to inspect the example key
//     value.
//   - .#: Number of elements in list or set.
//   - .%: Number of elements in map.
//
// The value parameter is a compiled regular expression. A typical pattern is
// using the regexp.MustCompile() function, which will automatically ensure the
// regular expression is supported by the Go regular expression handlers during
// compilation.
func TestMatchResourceAttr(name, key string, r *regexp.Regexp) TestCheckFunc {
	return checkIfIndexesIntoTypeSet(key, func(s *terraform.State) error {
		is, err := primaryInstanceState(s, name)
		if err != nil {
			return err
		}

		return testMatchResourceAttr(is, name, key, r)
	})
}

// TestModuleMatchResourceAttr - as per TestMatchResourceAttr but with
// support for non-root modules
func TestModuleMatchResourceAttr(mp []string, name string, key string, r *regexp.Regexp) TestCheckFunc {
	mpt := addrs.Module(mp).UnkeyedInstanceShim()
	return checkIfIndexesIntoTypeSet(key, func(s *terraform.State) error {
		is, err := modulePathPrimaryInstanceState(s, mpt, name)
		if err != nil {
			return err
		}

		return testMatchResourceAttr(is, name, key, r)
	})
}

func testMatchResourceAttr(is *terraform.InstanceState, name string, key string, r *regexp.Regexp) error {
	if !r.MatchString(is.Attributes[key]) {
		return fmt.Errorf(
			"%s: Attribute '%s' didn't match %q, got %#v",
			name,
			key,
			r.String(),
			is.Attributes[key])
	}

	return nil
}

// TestCheckResourceAttrPtr is like TestCheckResourceAttr except the
// value is a pointer so that it can be updated while the test is running.
// It will only be dereferenced at the point this step is run.
//
// Refer to the TestCheckResourceAttr documentation for more information about
// setting the name, key, and value parameters.
func TestCheckResourceAttrPtr(name string, key string, value *string) TestCheckFunc {
	return func(s *terraform.State) error {
		return TestCheckResourceAttr(name, key, *value)(s)
	}
}

// TestCheckModuleResourceAttrPtr - as per TestCheckResourceAttrPtr but with
// support for non-root modules
func TestCheckModuleResourceAttrPtr(mp []string, name string, key string, value *string) TestCheckFunc {
	return func(s *terraform.State) error {
		return TestCheckModuleResourceAttr(mp, name, key, *value)(s)
	}
}

// TestCheckResourceAttrPair ensures value equality in state between the first
// given name and key combination and the second name and key combination.
// State value checking is only recommended for testing Computed attributes
// and attribute defaults.
//
// For managed resources, the name parameter is combination of the resource
// type, a period (.), and the name label. The name for the below example
// configuration would be "myprovider_thing.example".
//
//	resource "myprovider_thing" "example" { ... }
//
// For data sources, the name parameter is a combination of the keyword "data",
// a period (.), the data source type, a period (.), and the name label. The
// name for the below example configuration would be
// "data.myprovider_thing.example".
//
//	data "myprovider_thing" "example" { ... }
//
// The first and second names may use any combination of managed resources
// and/or data sources.
//
// The key parameter is an attribute path in Terraform CLI 0.11 and earlier
// "flatmap" syntax. Keys start with the attribute name of a top-level
// attribute. Use the following special key syntax to inspect list, map, and
// set attributes:
//
//   - .{NUMBER}: List value at index, e.g. .0 to inspect the first element.
//     Use the TestCheckTypeSet* and TestMatchTypeSet* functions instead
//     for sets.
//   - .{KEY}: Map value at key, e.g. .example to inspect the example key
//     value.
//   - .#: Number of elements in list or set.
//   - .%: Number of elements in map.
func TestCheckResourceAttrPair(nameFirst, keyFirst, nameSecond, keySecond string) TestCheckFunc {
	return checkIfIndexesIntoTypeSetPair(keyFirst, keySecond, func(s *terraform.State) error {
		isFirst, err := primaryInstanceState(s, nameFirst)
		if err != nil {
			return err
		}

		isSecond, err := primaryInstanceState(s, nameSecond)
		if err != nil {
			return err
		}

		return testCheckResourceAttrPair(isFirst, nameFirst, keyFirst, isSecond, nameSecond, keySecond)
	})
}

// TestCheckModuleResourceAttrPair - as per TestCheckResourceAttrPair but with
// support for non-root modules
func TestCheckModuleResourceAttrPair(mpFirst []string, nameFirst string, keyFirst string, mpSecond []string, nameSecond string, keySecond string) TestCheckFunc {
	mptFirst := addrs.Module(mpFirst).UnkeyedInstanceShim()
	mptSecond := addrs.Module(mpSecond).UnkeyedInstanceShim()
	return checkIfIndexesIntoTypeSetPair(keyFirst, keySecond, func(s *terraform.State) error {
		isFirst, err := modulePathPrimaryInstanceState(s, mptFirst, nameFirst)
		if err != nil {
			return err
		}

		isSecond, err := modulePathPrimaryInstanceState(s, mptSecond, nameSecond)
		if err != nil {
			return err
		}

		return testCheckResourceAttrPair(isFirst, nameFirst, keyFirst, isSecond, nameSecond, keySecond)
	})
}

func testCheckResourceAttrPair(isFirst *terraform.InstanceState, nameFirst string, keyFirst string, isSecond *terraform.InstanceState, nameSecond string, keySecond string) error {
	if nameFirst == nameSecond && keyFirst == keySecond {
		return fmt.Errorf(
			"comparing self: resource %s attribute %s",
			nameFirst,
			keyFirst,
		)
	}

	vFirst, okFirst := isFirst.Attributes[keyFirst]
	vSecond, okSecond := isSecond.Attributes[keySecond]

	// Container count values of 0 should not be relied upon, and not reliably
	// maintained by helper/schema. For the purpose of tests, consider unset and
	// 0 to be equal.
	if len(keyFirst) > 2 && len(keySecond) > 2 && keyFirst[len(keyFirst)-2:] == keySecond[len(keySecond)-2:] &&
		(strings.HasSuffix(keyFirst, ".#") || strings.HasSuffix(keyFirst, ".%")) {
		// they have the same suffix, and it is a collection count key.
		if vFirst == "0" || vFirst == "" {
			okFirst = false
		}
		if vSecond == "0" || vSecond == "" {
			okSecond = false
		}
	}

	if okFirst != okSecond {
		if !okFirst {
			return fmt.Errorf("%s: Attribute %q not set, but %q is set in %s as %q", nameFirst, keyFirst, keySecond, nameSecond, vSecond)
		}
		return fmt.Errorf("%s: Attribute %q is %q, but %q is not set in %s", nameFirst, keyFirst, vFirst, keySecond, nameSecond)
	}
	if !(okFirst || okSecond) {
		// If they both don't exist then they are equally unset, so that's okay.
		return nil
	}

	if vFirst != vSecond {
		return fmt.Errorf(
			"%s: Attribute '%s' expected %#v, got %#v",
			nameFirst,
			keyFirst,
			vSecond,
			vFirst)
	}

	return nil
}

// TestCheckOutput checks an output in the Terraform configuration
func TestCheckOutput(name, value string) TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Outputs[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Value != value {
			return fmt.Errorf(
				"Output '%s': expected %#v, got %#v",
				name,
				value,
				rs)
		}

		return nil
	}
}

func TestMatchOutput(name string, r *regexp.Regexp) TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Outputs[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if !r.MatchString(rs.Value.(string)) {
			return fmt.Errorf(
				"Output '%s': %#v didn't match %q",
				name,
				rs,
				r.String())
		}

		return nil
	}
}

// modulePrimaryInstanceState returns the instance state for the given resource
// name in a ModuleState
func modulePrimaryInstanceState(ms *terraform.ModuleState, name string) (*terraform.InstanceState, error) {
	rs, ok := ms.Resources[name]
	if !ok {
		return nil, fmt.Errorf("Not found: %s in %s", name, ms.Path)
	}

	is := rs.Primary
	if is == nil {
		return nil, fmt.Errorf("No primary instance: %s in %s", name, ms.Path)
	}

	return is, nil
}

// modulePathPrimaryInstanceState returns the primary instance state for the
// given resource name in a given module path.
func modulePathPrimaryInstanceState(s *terraform.State, mp addrs.ModuleInstance, name string) (*terraform.InstanceState, error) {
	ms := s.ModuleByPath(mp)
	if ms == nil {
		return nil, fmt.Errorf("No module found at: %s", mp)
	}

	return modulePrimaryInstanceState(ms, name)
}

// primaryInstanceState returns the primary instance state for the given
// resource name in the root module.
func primaryInstanceState(s *terraform.State, name string) (*terraform.InstanceState, error) {
	ms := s.RootModule()
	return modulePrimaryInstanceState(ms, name)
}

// indexesIntoTypeSet is a heuristic to try and identify if a flatmap style
// string address uses a precalculated TypeSet hash, which are integers and
// typically are large and obviously not a list index
func indexesIntoTypeSet(key string) bool {
	for _, part := range strings.Split(key, ".") {
		if i, err := strconv.Atoi(part); err == nil && i > 100 {
			return true
		}
	}
	return false
}

func checkIfIndexesIntoTypeSet(key string, f TestCheckFunc) TestCheckFunc {
	return func(s *terraform.State) error {
		err := f(s)
		if err != nil && s.IsBinaryDrivenTest && indexesIntoTypeSet(key) {
			return fmt.Errorf("Error in test check: %s\nTest check address %q likely indexes into TypeSet\nThis is currently not possible in the SDK", err, key)
		}
		return err
	}
}

func checkIfIndexesIntoTypeSetPair(keyFirst, keySecond string, f TestCheckFunc) TestCheckFunc {
	return func(s *terraform.State) error {
		err := f(s)
		if err != nil && s.IsBinaryDrivenTest && (indexesIntoTypeSet(keyFirst) || indexesIntoTypeSet(keySecond)) {
			return fmt.Errorf("Error in test check: %s\nTest check address %q or %q likely indexes into TypeSet\nThis is currently not possible in the SDK", err, keyFirst, keySecond)
		}
		return err
	}
}
