// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfsmithy "github.com/hashicorp/terraform-provider-aws/internal/smithy"
)

// Builds the bedrock_model_config value through the reflected object type
// (which yields SmithyJSONType with a nil document constructor, as after a
// plan round-trip) and round-trips it through Expand and Flatten.
// Regression test for a panic (reflect: call of reflect.Value.Set on zero
// Value) when additional_params was expanded via generic autoflex.
func TestHarnessBedrockModelConfigExpand_additionalParams(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	listType := fwtypes.NewListNestedObjectTypeOf[harnessBedrockModelConfigModel](ctx)
	tfListType := listType.TerraformType(ctx)
	tfObjType := tfListType.(tftypes.List).ElementType.(tftypes.Object)

	rawObj := tftypes.NewValue(tfObjType, map[string]tftypes.Value{
		"additional_params": tftypes.NewValue(tftypes.String, `{"reasoning_effort":"high"}`),
		"max_tokens":        tftypes.NewValue(tftypes.Number, 1024),
		"model_id":          tftypes.NewValue(tftypes.String, "anthropic.claude-3-5-sonnet-20241022-v2:0"),
		"temperature":       tftypes.NewValue(tftypes.Number, nil),
		"top_p":             tftypes.NewValue(tftypes.Number, nil),
	})
	rawList := tftypes.NewValue(tfListType, []tftypes.Value{rawObj})

	listVal, err := listType.ValueFromTerraform(ctx, rawList)
	if err != nil {
		t.Fatalf("ValueFromTerraform: %s", err)
	}

	m := harnessModelConfigurationModel{
		BedrockModelConfig: listVal.(fwtypes.ListNestedObjectValueOf[harnessBedrockModelConfigModel]),
		GeminiModelConfig:  fwtypes.NewListNestedObjectValueOfNull[harnessGeminiModelConfigModel](ctx),
		OpenAiModelConfig:  fwtypes.NewListNestedObjectValueOfNull[harnessOpenAIModelConfigModel](ctx),
	}

	// This panicked before the fix: reflect: call of reflect.Value.Set on zero Value.
	v, diags := m.Expand(ctx)
	if diags.HasError() {
		t.Fatalf("Expand diags: %v", diags)
	}

	member, ok := v.(*awstypes.HarnessModelConfigurationMemberBedrockModelConfig)
	if !ok {
		t.Fatalf("unexpected type %T", v)
	}
	if member.Value.AdditionalParams == nil {
		t.Fatal("AdditionalParams not set on API value")
	}
	json, err := tfsmithy.DocumentToJSONString(member.Value.AdditionalParams)
	if err != nil {
		t.Fatalf("DocumentToJSONString: %s", err)
	}
	if json != `{"reasoning_effort":"high"}` {
		t.Fatalf("unexpected AdditionalParams JSON: %s", json)
	}

	// Round-trip: Flatten the API value back and check the string survives.
	var flat harnessModelConfigurationModel
	if d := flat.Flatten(ctx, *member); d.HasError() {
		t.Fatalf("Flatten diags: %v", d)
	}
	data, d := flat.BedrockModelConfig.ToPtr(ctx)
	if d.HasError() {
		t.Fatalf("ToPtr diags: %v", d)
	}
	if got := data.AdditionalParams.ValueString(); got != `{"reasoning_effort":"high"}` {
		t.Fatalf("unexpected flattened AdditionalParams: %s", got)
	}
}
