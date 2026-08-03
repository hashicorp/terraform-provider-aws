// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/document"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfsmithy "github.com/hashicorp/terraform-provider-aws/internal/smithy"
)

func TestHarnessOpenAIModelConfigExpand_additionalParams(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	m := testHarnessOpenAIModelConfiguration(t, `{"reasoning_effort":"high"}`)

	v, diags := m.Expand(ctx)
	if diags.HasError() {
		t.Fatalf("Expand diags: %v", diags)
	}

	member, ok := v.(*awstypes.HarnessModelConfigurationMemberOpenAiModelConfig)
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

	var flat harnessModelConfigurationModel
	if d := flat.Flatten(ctx, *member); d.HasError() {
		t.Fatalf("Flatten diags: %v", d)
	}
	elements := flat.OpenAiModelConfig.Elements()
	if len(elements) != 1 {
		t.Fatalf("unexpected flattened OpenAI model config element count: %d", len(elements))
	}
	object, ok := elements[0].(fwtypes.ObjectValueOf[harnessOpenAIModelConfigModel])
	if !ok {
		t.Fatalf("unexpected flattened OpenAI model config type %T", elements[0])
	}
	additionalParams, ok := object.Attributes()["additional_params"].(fwtypes.SmithyJSON[document.Interface])
	if !ok {
		t.Fatalf("unexpected flattened additional_params type %T", object.Attributes()["additional_params"])
	}
	if got := additionalParams.ValueString(); got != `{"reasoning_effort":"high"}` {
		t.Fatalf("unexpected flattened AdditionalParams: %s", got)
	}
}

func TestHarnessOpenAIModelConfigExpand_invalidAdditionalParams(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	m := testHarnessOpenAIModelConfiguration(t, `{`)

	_, diags := m.Expand(ctx)
	if !diags.HasError() {
		t.Fatal("expected invalid additional_params JSON diagnostic")
	}
}

func TestHarnessOpenAIModelConfigFlatten_additionalParamsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	member := awstypes.HarnessModelConfigurationMemberOpenAiModelConfig{
		Value: awstypes.HarnessOpenAiModelConfig{
			AdditionalParams: document.NewLazyDocument(awstypes.HarnessOpenAiModelConfig{}),
		},
	}

	var flat harnessModelConfigurationModel
	diags := flat.Flatten(ctx, member)
	if !diags.HasError() {
		t.Fatal("expected additional_params document serialization diagnostic")
	}
	if got := diags.Errors()[0].Summary(); got != "reading Smithy document" {
		t.Fatalf("unexpected diagnostic summary: %s", got)
	}
	if detail := diags.Errors()[0].Detail(); !strings.Contains(detail, "unsupported type") {
		t.Fatalf("unexpected diagnostic detail: %s", detail)
	}
}

func TestHarnessGeminiModelConfigExpand_additionalParams(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	m := testHarnessGeminiModelConfiguration(t, `{"thinking_budget":1024}`)

	v, diags := m.Expand(ctx)
	if diags.HasError() {
		t.Fatalf("Expand diags: %v", diags)
	}

	member, ok := v.(*awstypes.HarnessModelConfigurationMemberGeminiModelConfig)
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
	if json != `{"thinking_budget":1024}` {
		t.Fatalf("unexpected AdditionalParams JSON: %s", json)
	}

	var flat harnessModelConfigurationModel
	if d := flat.Flatten(ctx, *member); d.HasError() {
		t.Fatalf("Flatten diags: %v", d)
	}
	elements := flat.GeminiModelConfig.Elements()
	if len(elements) != 1 {
		t.Fatalf("unexpected flattened Gemini model config element count: %d", len(elements))
	}
	object, ok := elements[0].(fwtypes.ObjectValueOf[harnessGeminiModelConfigModel])
	if !ok {
		t.Fatalf("unexpected flattened Gemini model config type %T", elements[0])
	}
	additionalParams, ok := object.Attributes()["additional_params"].(fwtypes.SmithyJSON[document.Interface])
	if !ok {
		t.Fatalf("unexpected flattened additional_params type %T", object.Attributes()["additional_params"])
	}
	if got := additionalParams.ValueString(); got != `{"thinking_budget":1024}` {
		t.Fatalf("unexpected flattened AdditionalParams: %s", got)
	}
}

func TestHarnessGeminiModelConfigExpand_invalidAdditionalParams(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	m := testHarnessGeminiModelConfiguration(t, `{`)

	_, diags := m.Expand(ctx)
	if !diags.HasError() {
		t.Fatal("expected invalid additional_params JSON diagnostic")
	}
}

func TestHarnessGeminiModelConfigFlatten_additionalParamsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	member := awstypes.HarnessModelConfigurationMemberGeminiModelConfig{
		Value: awstypes.HarnessGeminiModelConfig{
			AdditionalParams: document.NewLazyDocument(awstypes.HarnessGeminiModelConfig{}),
		},
	}

	var flat harnessModelConfigurationModel
	diags := flat.Flatten(ctx, member)
	if !diags.HasError() {
		t.Fatal("expected additional_params document serialization diagnostic")
	}
	if got := diags.Errors()[0].Summary(); got != "reading Smithy document" {
		t.Fatalf("unexpected diagnostic summary: %s", got)
	}
	if detail := diags.Errors()[0].Detail(); !strings.Contains(detail, "unsupported type") {
		t.Fatalf("unexpected diagnostic detail: %s", detail)
	}
}

func testHarnessOpenAIModelConfiguration(t *testing.T, additionalParams string) harnessModelConfigurationModel {
	t.Helper()
	ctx := context.Background()
	listType := fwtypes.NewListNestedObjectTypeOf[harnessOpenAIModelConfigModel](ctx)
	tfListType := listType.TerraformType(ctx)
	tfObjType := tfListType.(tftypes.List).ElementType.(tftypes.Object)
	if _, ok := tfObjType.AttributeTypes["additional_params"]; !ok {
		t.Fatal("additional_params missing from OpenAI model config Terraform type")
	}

	rawObj := tftypes.NewValue(tfObjType, map[string]tftypes.Value{
		"additional_params": tftypes.NewValue(tftypes.String, additionalParams),
		"api_key_arn":       tftypes.NewValue(tftypes.String, testHarnessAPIKeyARN()),
		"max_tokens":        tftypes.NewValue(tftypes.Number, 1024),
		"model_id":          tftypes.NewValue(tftypes.String, "gpt-5"),
		"temperature":       tftypes.NewValue(tftypes.Number, nil),
		"top_p":             tftypes.NewValue(tftypes.Number, nil),
	})
	rawList := tftypes.NewValue(tfListType, []tftypes.Value{rawObj})

	listVal, err := listType.ValueFromTerraform(ctx, rawList)
	if err != nil {
		t.Fatalf("ValueFromTerraform: %s", err)
	}

	return harnessModelConfigurationModel{
		BedrockModelConfig: fwtypes.NewListNestedObjectValueOfNull[harnessBedrockModelConfigModel](ctx),
		GeminiModelConfig:  fwtypes.NewListNestedObjectValueOfNull[harnessGeminiModelConfigModel](ctx),
		LiteLlmModelConfig: fwtypes.NewListNestedObjectValueOfNull[harnessLiteLLMModelConfigModel](ctx),
		OpenAiModelConfig:  listVal.(fwtypes.ListNestedObjectValueOf[harnessOpenAIModelConfigModel]),
	}
}

func testHarnessGeminiModelConfiguration(t *testing.T, additionalParams string) harnessModelConfigurationModel {
	t.Helper()
	ctx := context.Background()
	listType := fwtypes.NewListNestedObjectTypeOf[harnessGeminiModelConfigModel](ctx)
	tfListType := listType.TerraformType(ctx)
	tfObjType := tfListType.(tftypes.List).ElementType.(tftypes.Object)
	if _, ok := tfObjType.AttributeTypes["additional_params"]; !ok {
		t.Fatal("additional_params missing from Gemini model config Terraform type")
	}

	rawObj := tftypes.NewValue(tfObjType, map[string]tftypes.Value{
		"additional_params": tftypes.NewValue(tftypes.String, additionalParams),
		"api_key_arn":       tftypes.NewValue(tftypes.String, testHarnessAPIKeyARN()),
		"max_tokens":        tftypes.NewValue(tftypes.Number, 1024),
		"model_id":          tftypes.NewValue(tftypes.String, "gemini-2.5-pro"),
		"temperature":       tftypes.NewValue(tftypes.Number, nil),
		"top_k":             tftypes.NewValue(tftypes.Number, nil),
		"top_p":             tftypes.NewValue(tftypes.Number, nil),
	})
	rawList := tftypes.NewValue(tfListType, []tftypes.Value{rawObj})

	listVal, err := listType.ValueFromTerraform(ctx, rawList)
	if err != nil {
		t.Fatalf("ValueFromTerraform: %s", err)
	}

	return harnessModelConfigurationModel{
		BedrockModelConfig: fwtypes.NewListNestedObjectValueOfNull[harnessBedrockModelConfigModel](ctx),
		GeminiModelConfig:  listVal.(fwtypes.ListNestedObjectValueOf[harnessGeminiModelConfigModel]),
		LiteLlmModelConfig: fwtypes.NewListNestedObjectValueOfNull[harnessLiteLLMModelConfigModel](ctx),
		OpenAiModelConfig:  fwtypes.NewListNestedObjectValueOfNull[harnessOpenAIModelConfigModel](ctx),
	}
}

func testHarnessAPIKeyARN() string {
	return arn.ARN{
		Partition: "aws",
		Service:   "bedrock-agentcore",
		Region:    "us-east-1",    //lintignore:AWSAT003
		AccountID: "123456789012", // nosemgrep:ci.literal-12Digit-string-test-constant
		Resource:  "api-key/test",
	}.String()
}
