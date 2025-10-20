package flex

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// Test field names and types directly
func TestDebugFieldMatching(t *testing.T) {
	// Check source struct fields
	sourceType := reflect.TypeOf(DefaultCacheBehavior{})
	t.Logf("Source struct: %s", sourceType.Name())
	for i := 0; i < sourceType.NumField(); i++ {
		field := sourceType.Field(i)
		t.Logf("  Source field[%d]: %s (type: %s)", i, field.Name, field.Type)
	}

	// Check target struct fields
	targetStruct := struct {
		TargetOriginId       types.String                      `tfsdk:"target_origin_id"`
		ViewerProtocolPolicy types.String                      `tfsdk:"viewer_protocol_policy"`
		TrustedSigners       fwtypes.ListValueOf[types.String] `tfsdk:"trusted_signers" autoflex:",wrapper=items"`
		TrustedKeyGroups     fwtypes.ListValueOf[types.String] `tfsdk:"trusted_key_groups" autoflex:",wrapper=items"`
	}{}

	targetType := reflect.TypeOf(targetStruct)
	t.Logf("Target struct: %s", targetType.Name())
	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)
		tag := field.Tag.Get("autoflex")
		t.Logf("  Target field[%d]: %s (type: %s) [autoflex: %s]", i, field.Name, field.Type, tag)
	}
}
