package resource

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/logging"
)

// testStepValidateRequest contains data for the (TestStep).validate() method.
type testStepValidateRequest struct {
	// StepNumber is the index of the TestStep in the TestCase.Steps.
	StepNumber int

	// TestCaseHasProviders is enabled if the TestCase has set any of
	// ExternalProviders, ProtoV5ProviderFactories, ProtoV6ProviderFactories,
	// or ProviderFactories.
	TestCaseHasProviders bool
}

// hasProviders returns true if the TestStep has set any of the
// ExternalProviders, ProtoV5ProviderFactories, ProtoV6ProviderFactories, or
// ProviderFactories fields.
func (s TestStep) hasProviders(_ context.Context) bool {
	if len(s.ExternalProviders) > 0 {
		return true
	}

	if len(s.ProtoV5ProviderFactories) > 0 {
		return true
	}

	if len(s.ProtoV6ProviderFactories) > 0 {
		return true
	}

	if len(s.ProviderFactories) > 0 {
		return true
	}

	return false
}

// validate ensures the TestStep is valid based on the following criteria:
//
//   - Config or ImportState is set.
//   - Providers are not specified (ExternalProviders,
//     ProtoV5ProviderFactories, ProtoV6ProviderFactories, ProviderFactories)
//     if specified at the TestCase level.
//   - Providers are specified (ExternalProviders, ProtoV5ProviderFactories,
//     ProtoV6ProviderFactories, ProviderFactories) if not specified at the
//     TestCase level.
//   - No overlapping ExternalProviders and ProviderFactories entries
//   - ResourceName is not empty when ImportState is true, ImportStateIdFunc
//     is not set, and ImportStateId is not set.
func (s TestStep) validate(ctx context.Context, req testStepValidateRequest) error {
	ctx = logging.TestStepNumberContext(ctx, req.StepNumber)

	logging.HelperResourceTrace(ctx, "Validating TestStep")

	if s.Config == "" && !s.ImportState {
		err := fmt.Errorf("TestStep missing Config or ImportState")
		logging.HelperResourceError(ctx, "TestStep validation error", map[string]interface{}{logging.KeyError: err})
		return err
	}

	for name := range s.ExternalProviders {
		if _, ok := s.ProviderFactories[name]; ok {
			err := fmt.Errorf("TestStep provider %q set in both ExternalProviders and ProviderFactories", name)
			logging.HelperResourceError(ctx, "TestStep validation error", map[string]interface{}{logging.KeyError: err})
			return err
		}
	}

	hasProviders := s.hasProviders(ctx)

	if req.TestCaseHasProviders && hasProviders {
		err := fmt.Errorf("Providers must only be specified either at the TestCase or TestStep level")
		logging.HelperResourceError(ctx, "TestStep validation error", map[string]interface{}{logging.KeyError: err})
		return err
	}

	if !req.TestCaseHasProviders && !hasProviders {
		err := fmt.Errorf("Providers must be specified at the TestCase level or in all TestStep")
		logging.HelperResourceError(ctx, "TestStep validation error", map[string]interface{}{logging.KeyError: err})
		return err
	}

	if s.ImportState {
		if s.ImportStateId == "" && s.ImportStateIdFunc == nil && s.ResourceName == "" {
			err := fmt.Errorf("TestStep ImportState must be specified with ImportStateId, ImportStateIdFunc, or ResourceName")
			logging.HelperResourceError(ctx, "TestStep validation error", map[string]interface{}{logging.KeyError: err})
			return err
		}
	}

	return nil
}
