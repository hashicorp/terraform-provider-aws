// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func TestRuntimeTargetConfigurationModel_schemaRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	const payload = `{"openapi":"3.0.0","info":{"title":"Test API","version":"1.0.0"},"paths":{}}`

	runtimeARN := arn.ARN{
		Partition: "aws",
		Service:   "bedrock-agentcore",
		Region:    "us-east-1", //lintignore:AWSAT003
		AccountID: "123456789012", // nosemgrep:ci.literal-12Digit-string-test-constant
		Resource:  "runtime/test",
	}.String()
	runtimeModel := runtimeTargetConfigurationModel{
		ARN:       fwtypes.ARNValue(runtimeARN),
		Qualifier: types.StringValue("DEFAULT"),
		Schema: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &apiSchemaConfigurationModel{
			InlinePayload: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &inlinePayloadModel{
				Payload: types.StringValue(payload),
			}),
			S3: fwtypes.NewListNestedObjectValueOfNull[s3ConfigurationModel](ctx),
		}),
	}
	model := httpTargetConfigurationModel{
		AgentcoreRuntime: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &runtimeModel),
		Passthrough:      fwtypes.NewListNestedObjectValueOfNull[passthroughTargetConfigurationModel](ctx),
	}

	v, diags := model.Expand(ctx)
	if diags.HasError() {
		t.Fatalf("Expand diags: %v", diags)
	}
	member, ok := v.(*awstypes.HttpTargetConfigurationMemberAgentcoreRuntime)
	if !ok {
		t.Fatalf("unexpected type %T", v)
	}
	runtime := &member.Value
	if runtime.Schema == nil {
		t.Fatal("Schema not set on API value")
	}
	inline, ok := runtime.Schema.Source.(*awstypes.ApiSchemaConfigurationMemberInlinePayload)
	if !ok {
		t.Fatalf("unexpected schema source type %T", runtime.Schema.Source)
	}
	if inline.Value != payload {
		t.Fatalf("unexpected schema payload: %s", inline.Value)
	}

	var flat httpTargetConfigurationModel
	if d := flat.Flatten(ctx, *member); d.HasError() {
		t.Fatalf("Flatten diags: %v", d)
	}
	flatRuntime, d := flat.AgentcoreRuntime.ToPtr(ctx)
	if d.HasError() {
		t.Fatalf("AgentcoreRuntime ToPtr diags: %v", d)
	}
	schemaModel, d := flatRuntime.Schema.ToPtr(ctx)
	if d.HasError() {
		t.Fatalf("Schema ToPtr diags: %v", d)
	}
	inlineModel, d := schemaModel.InlinePayload.ToPtr(ctx)
	if d.HasError() {
		t.Fatalf("InlinePayload ToPtr diags: %v", d)
	}
	if got := inlineModel.Payload.ValueString(); got != payload {
		t.Fatalf("unexpected flattened schema payload: %s", got)
	}
}
