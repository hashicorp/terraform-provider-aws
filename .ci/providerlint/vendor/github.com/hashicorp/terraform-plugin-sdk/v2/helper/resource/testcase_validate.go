// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resource

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/logging"
)

// hasProviders returns true if the TestCase has set any of the
// ExternalProviders, ProtoV5ProviderFactories, ProtoV6ProviderFactories,
// ProviderFactories, or Providers fields.
func (c TestCase) hasProviders(_ context.Context) bool {
	if len(c.ExternalProviders) > 0 {
		return true
	}

	if len(c.ProtoV5ProviderFactories) > 0 {
		return true
	}

	if len(c.ProtoV6ProviderFactories) > 0 {
		return true
	}

	if len(c.ProviderFactories) > 0 {
		return true
	}

	if len(c.Providers) > 0 {
		return true
	}

	return false
}

// validate ensures the TestCase is valid based on the following criteria:
//
//   - No overlapping ExternalProviders and Providers entries
//   - No overlapping ExternalProviders and ProviderFactories entries
//   - TestStep validations performed by the (TestStep).validate() method.
func (c TestCase) validate(ctx context.Context) error {
	logging.HelperResourceTrace(ctx, "Validating TestCase")

	if len(c.Steps) == 0 {
		err := fmt.Errorf("TestCase missing Steps")
		logging.HelperResourceError(ctx, "TestCase validation error", map[string]interface{}{logging.KeyError: err})
		return err
	}

	for name := range c.ExternalProviders {
		if _, ok := c.Providers[name]; ok {
			err := fmt.Errorf("TestCase provider %q set in both ExternalProviders and Providers", name)
			logging.HelperResourceError(ctx, "TestCase validation error", map[string]interface{}{logging.KeyError: err})
			return err
		}

		if _, ok := c.ProviderFactories[name]; ok {
			err := fmt.Errorf("TestCase provider %q set in both ExternalProviders and ProviderFactories", name)
			logging.HelperResourceError(ctx, "TestCase validation error", map[string]interface{}{logging.KeyError: err})
			return err
		}
	}

	testCaseHasProviders := c.hasProviders(ctx)

	for stepIndex, step := range c.Steps {
		stepNumber := stepIndex + 1 // Use 1-based index for humans
		stepValidateReq := testStepValidateRequest{
			StepNumber:           stepNumber,
			TestCaseHasProviders: testCaseHasProviders,
		}

		err := step.validate(ctx, stepValidateReq)

		if err != nil {
			err := fmt.Errorf("TestStep %d/%d validation error: %w", stepNumber, len(c.Steps), err)
			logging.HelperResourceError(ctx, "TestCase validation error", map[string]interface{}{logging.KeyError: err})
			return err
		}
	}

	return nil
}
