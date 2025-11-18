package flex

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Test single field flattening
func TestDebugSingleFieldFlatten(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Simple source struct with just one XML wrapper field
	source := &struct {
		TrustedSigners *TrustedSigners
	}{
		TrustedSigners: &TrustedSigners{
			Enabled:  aws.Bool(false),
			Items:    []string{"signer1"},
			Quantity: aws.Int32(1),
		},
	}

	// Simple target struct with wrapper tag
	target := &struct {
		TrustedSigners types.List `tfsdk:"trusted_signers" autoflex:",wrapper=items"`
	}{}

	t.Logf("Source: %+v", source)
	t.Logf("Target: %+v", target)

	// Try to flatten
	diags := Flatten(ctx, source, target)

	// Print detailed diagnostics
	if diags.HasError() {
		for _, diag := range diags.Errors() {
			t.Logf("Error: %s - %s", diag.Summary(), diag.Detail())
		}
		t.Fatalf("Single field flatten failed")
	}

	t.Logf("Single field flatten succeeded: %v", target.TrustedSigners)
}
