package flex

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func TestDebugFlattenTrustedSigners(t *testing.T) {
	ctx := context.Background()

	// Test our custom TrustedSigners type (matching the failing test)
	customTrustedSigners := &TrustedSigners{
		Enabled:  aws.Bool(false), // This field is required by AWS SDK
		Items:    []string{},      // Empty slice
		Quantity: aws.Int32(0),
	}

	// Create target Terraform list type
	var tfTrustedSigners types.List

	// Try to flatten with XML wrapper detection
	t.Logf("Testing custom TrustedSigners type...")
	t.Logf("isXMLWrapperStruct: %v", isXMLWrapperStruct(reflect.TypeOf(*customTrustedSigners)))

	// Try to flatten
	diags := Flatten(ctx, customTrustedSigners, &tfTrustedSigners)

	// Print detailed diagnostics
	if diags.HasError() {
		for _, diag := range diags.Errors() {
			t.Logf("Error: %s - %s", diag.Summary(), diag.Detail())
		}
		t.Fatalf("Flatten failed with errors")
	}

	t.Logf("Flattened successfully: %v", tfTrustedSigners)
}

// Test the XML wrapper detection function directly
func TestXMLWrapperDetection(t *testing.T) {
	// Test our custom types
	t.Logf("TrustedSigners isXMLWrapperStruct: %v", isXMLWrapperStruct(reflect.TypeOf(TrustedSigners{})))
	t.Logf("TrustedKeyGroups isXMLWrapperStruct: %v", isXMLWrapperStruct(reflect.TypeOf(TrustedKeyGroups{})))
}

// Test flattening XML wrapper struct directly (like working examples)
func TestDebugFlattenStructWithXMLWrappers(t *testing.T) {
	ctx := context.Background()

	// Test 1: Direct XML wrapper to wrapper-tagged field (like working examples)
	trustedSignersSource := TrustedSigners{
		Enabled:  aws.Bool(false),
		Items:    []string{"signer1", "signer2"}, // Populated for this test
		Quantity: aws.Int32(2),
	}

	trustedSignersTarget := &struct {
		TrustedSigners fwtypes.ListValueOf[types.String] `tfsdk:"trusted_signers" autoflex:",wrapper=items"`
	}{}

	diags := Flatten(ctx, &trustedSignersSource, trustedSignersTarget)
	if diags.HasError() {
		for _, diag := range diags.Errors() {
			t.Logf("Direct XML wrapper Error: %s - %s", diag.Summary(), diag.Detail())
		}
		t.Fatalf("Direct XML wrapper flatten failed")
	}
	t.Logf("Direct XML wrapper flatten succeeded: %v", trustedSignersTarget.TrustedSigners)

	// Test 2: Struct containing XML wrapper fields (our real use case)
	source := &DefaultCacheBehavior{
		TargetOriginId:       aws.String("S3-my-bucket"),
		ViewerProtocolPolicy: "redirect-to-https",
		TrustedSigners: &TrustedSigners{
			Enabled:  aws.Bool(false),
			Items:    []string{"signer1"}, // Populated for this test
			Quantity: aws.Int32(1),
		},
		// Let's try with nil first to see if that works
		TrustedKeyGroups: nil,
	}

	target := &struct {
		TargetOriginId       types.String `tfsdk:"target_origin_id"`
		ViewerProtocolPolicy types.String `tfsdk:"viewer_protocol_policy"`
		TrustedSigners       types.List   `tfsdk:"trusted_signers" autoflex:",wrapper=items"`
		TrustedKeyGroups     types.List   `tfsdk:"trusted_key_groups" autoflex:",wrapper=items"`
	}{}

	diags = Flatten(ctx, source, target)
	if diags.HasError() {
		for _, diag := range diags.Errors() {
			t.Logf("Struct with XML fields Error: %s - %s", diag.Summary(), diag.Detail())
		}
		t.Fatalf("Struct with XML fields flatten failed")
	}

	t.Logf("Struct with XML fields flatten succeeded!")
	t.Logf("  TargetOriginId: %v", target.TargetOriginId)
	t.Logf("  ViewerProtocolPolicy: %v", target.ViewerProtocolPolicy)
	t.Logf("  TrustedSigners: %v", target.TrustedSigners)
	t.Logf("  TrustedKeyGroups: %v", target.TrustedKeyGroups)
}
