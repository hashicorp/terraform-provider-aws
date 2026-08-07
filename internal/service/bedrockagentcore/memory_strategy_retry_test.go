// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	frameworkvalidator "github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func TestMemoryStrategyRetryableMemoryTransitionalState(t *testing.T) {
	t.Parallel()

	err := &awstypes.ValidationException{
		Message: aws.String("Validation failed during UpdateMemory: Memory is in transitional state UPDATING. Cannot update memory."),
	}
	retryable, returnedErr := memoryStrategyRetryable(false)(err)
	if !retryable {
		t.Fatal("expected parent memory transitional state to be retryable")
	}
	if returnedErr == nil {
		t.Fatal("expected retryable error to be returned")
	}
}

func TestFlattenMemoryStrategyPreservesTriggerConditionsShape(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	target := memoryStrategyResourceModel{
		Configuration: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &customConfigurationModel{
			Type:          fwtypes.StringEnumValue(awstypes.OverrideTypeSelfManaged),
			Consolidation: fwtypes.NewListNestedObjectValueOfNull[overrideDetailsModel](ctx),
			Extraction:    fwtypes.NewListNestedObjectValueOfNull[overrideDetailsModel](ctx),
			Reflection:    fwtypes.NewListNestedObjectValueOfNull[episodicReflectionOverrideDetailsModel](ctx),
			SelfManaged: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &selfManagedConfigurationModel{
				HistoricalContextWindowSize: types.Int32Value(25),
				InvocationConfiguration: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &invocationConfigurationModel{
					PayloadDeliveryBucketName: types.StringValue("example-bucket"),
					TopicARN:                  fwtypes.ARNValue("arn:aws:sns:us-west-2:123456789012:example"),
				}),
				TriggerConditions: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &triggerConditionsModel{
					MessageBasedTrigger: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &messageBasedTriggerModel{
						MessageCount: types.Int32Value(12),
					}),
					TokenBasedTrigger: fwtypes.NewListNestedObjectValueOfNull[tokenBasedTriggerModel](ctx),
					TimeBasedTrigger:  fwtypes.NewListNestedObjectValueOfNull[timeBasedTriggerModel](ctx),
				}),
			}),
		}),
	}
	source := &awstypes.MemoryStrategy{
		Name:       aws.String("test"),
		StrategyId: aws.String("strategy-id"),
		Type:       awstypes.MemoryStrategyTypeCustom,
		Configuration: &awstypes.StrategyConfiguration{
			Type: awstypes.OverrideTypeSelfManaged,
			SelfManagedConfiguration: &awstypes.SelfManagedConfiguration{
				HistoricalContextWindowSize: aws.Int32(25),
				InvocationConfiguration: &awstypes.InvocationConfiguration{
					PayloadDeliveryBucketName: aws.String("example-bucket"),
					TopicArn:                  aws.String("arn:aws:sns:us-west-2:123456789012:example"),
				},
				TriggerConditions: []awstypes.TriggerCondition{
					&awstypes.TriggerConditionMemberMessageBasedTrigger{Value: awstypes.MessageBasedTrigger{MessageCount: aws.Int32(12)}},
					&awstypes.TriggerConditionMemberTokenBasedTrigger{Value: awstypes.TokenBasedTrigger{TokenCount: aws.Int32(5000)}},
					&awstypes.TriggerConditionMemberTimeBasedTrigger{Value: awstypes.TimeBasedTrigger{IdleSessionTimeout: aws.Int32(20)}},
				},
			},
		},
	}

	if diags := flattenMemoryStrategy(ctx, source, &target); diags.HasError() {
		t.Fatalf("flattening memory strategy: %s", diags)
	}
	configuration, diags := target.Configuration.ToPtr(ctx)
	if diags.HasError() {
		t.Fatalf("reading configuration: %s", diags)
	}
	selfManaged, diags := configuration.SelfManaged.ToPtr(ctx)
	if diags.HasError() {
		t.Fatalf("reading self-managed configuration: %s", diags)
	}
	conditions, diags := selfManaged.TriggerConditions.ToPtr(ctx)
	if diags.HasError() {
		t.Fatalf("reading trigger conditions: %s", diags)
	}
	message, diags := conditions.MessageBasedTrigger.ToPtr(ctx)
	if diags.HasError() {
		t.Fatalf("reading message trigger: %s", diags)
	}
	if got := message.MessageCount.ValueInt32(); got != 12 {
		t.Errorf("expected configured message count 12, got %d", got)
	}
	if !conditions.TokenBasedTrigger.IsNull() {
		t.Error("expected omitted token trigger to remain null")
	}
	if !conditions.TimeBasedTrigger.IsNull() {
		t.Error("expected omitted time trigger to remain null")
	}
}

func TestTriggerConditionsValidatorRejectsEmptyNestedList(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	value := fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &triggerConditionsModel{
		MessageBasedTrigger: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []messageBasedTriggerModel{}),
		TokenBasedTrigger:   fwtypes.NewListNestedObjectValueOfNull[tokenBasedTriggerModel](ctx),
		TimeBasedTrigger:    fwtypes.NewListNestedObjectValueOfNull[timeBasedTriggerModel](ctx),
	})
	request := frameworkvalidator.ListRequest{
		Path:        path.Root("trigger_conditions"),
		ConfigValue: value.ListValue,
	}
	var response frameworkvalidator.ListResponse

	triggerConditionsValidator{}.ValidateList(ctx, request, &response)
	if !response.Diagnostics.HasError() {
		t.Fatal("expected an empty nested trigger list to be rejected")
	}
}
