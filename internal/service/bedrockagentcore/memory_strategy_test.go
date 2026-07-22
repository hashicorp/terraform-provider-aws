// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	fwschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	frameworkvalidator "github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrockagentcore "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagentcore"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestMemoryStrategyTriggerConditionsSchema(t *testing.T) {
	t.Parallel()

	ctx := acctest.Context(t)
	r, err := tfbedrockagentcore.ResourceMemoryStrategy(ctx)
	if err != nil {
		t.Fatalf("creating resource: %s", err)
	}
	var response fwresource.SchemaResponse
	r.Schema(ctx, fwresource.SchemaRequest{}, &response)
	if diags := response.Schema.ValidateImplementation(ctx); diags.HasError() {
		t.Fatalf("validating resource schema: %s", diags)
	}

	configuration, ok := response.Schema.Blocks[names.AttrConfiguration].(fwschema.ListNestedBlock)
	if !ok {
		t.Fatalf("expected configuration to be a ListNestedBlock, got %T", response.Schema.Blocks[names.AttrConfiguration])
	}
	selfManaged, ok := configuration.NestedObject.Blocks["self_managed"].(fwschema.ListNestedBlock)
	if !ok {
		t.Fatalf("expected self_managed to be a ListNestedBlock, got %T", configuration.NestedObject.Blocks["self_managed"])
	}
	triggerConditions, ok := selfManaged.NestedObject.Attributes["trigger_conditions"].(fwschema.ListAttribute)
	if !ok {
		t.Fatalf("expected trigger_conditions to be a protocol 5 compatible ListAttribute, got %T", selfManaged.NestedObject.Attributes["trigger_conditions"])
	}
	if !triggerConditions.Optional || !triggerConditions.Computed {
		t.Fatalf("expected trigger_conditions to be Optional+Computed, got Optional=%t Computed=%t", triggerConditions.Optional, triggerConditions.Computed)
	}

	elementType, ok := triggerConditions.ElementType.(types.ObjectType)
	if !ok {
		t.Fatalf("expected trigger_conditions elements to be objects, got %T", triggerConditions.ElementType)
	}
	for _, name := range []string{"message_based_trigger", "token_based_trigger", "time_based_trigger"} {
		if _, ok := elementType.AttrTypes[name]; !ok {
			t.Errorf("expected trigger_conditions object to contain %s", name)
		}
	}
}

func TestMemoryStrategyTriggerConditionsValidation(t *testing.T) {
	t.Parallel()

	ctx := acctest.Context(t)
	r, err := tfbedrockagentcore.ResourceMemoryStrategy(ctx)
	if err != nil {
		t.Fatalf("creating resource: %s", err)
	}
	var schemaResponse fwresource.SchemaResponse
	r.Schema(ctx, fwresource.SchemaRequest{}, &schemaResponse)
	configuration := schemaResponse.Schema.Blocks[names.AttrConfiguration].(fwschema.ListNestedBlock)
	selfManaged := configuration.NestedObject.Blocks["self_managed"].(fwschema.ListNestedBlock)
	attribute := selfManaged.NestedObject.Attributes["trigger_conditions"].(fwschema.ListAttribute)

	value := fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.TriggerConditionsModel{
		MessageBasedTrigger: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.MessageBasedTriggerModel{
			MessageCount: types.Int32Value(0),
		}),
		TokenBasedTrigger: fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.TokenBasedTriggerModel](ctx),
		TimeBasedTrigger:  fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.TimeBasedTriggerModel](ctx),
	})
	request := frameworkvalidator.ListRequest{
		Path:        path.Root("trigger_conditions"),
		ConfigValue: value.ListValue,
	}
	var response frameworkvalidator.ListResponse
	for _, validator := range attribute.Validators {
		validator.ValidateList(ctx, request, &response)
	}
	if !response.Diagnostics.HasError() {
		t.Fatal("expected message_count below the service minimum to be rejected")
	}
}

func TestMemoryStrategyTriggerConditionsExpand(t *testing.T) {
	t.Parallel()

	ctx := acctest.Context(t)
	model := tfbedrockagentcore.TriggerConditionsModel{
		MessageBasedTrigger: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.MessageBasedTriggerModel{
			MessageCount: types.Int32Value(12),
		}),
		TokenBasedTrigger: fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.TokenBasedTriggerModel](ctx),
		TimeBasedTrigger:  fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.TimeBasedTriggerModel](ctx),
	}

	result, diags := model.Expand(ctx)
	if diags.HasError() {
		t.Fatalf("expanding trigger conditions: %s", diags)
	}
	conditions, ok := result.([]awstypes.TriggerConditionInput)
	if !ok {
		t.Fatalf("expected []TriggerConditionInput, got %T", result)
	}
	if len(conditions) != 1 {
		t.Fatalf("expected one configured trigger condition, got %d", len(conditions))
	}
	message, ok := conditions[0].(*awstypes.TriggerConditionInputMemberMessageBasedTrigger)
	if !ok {
		t.Fatalf("expected message-based trigger, got %T", conditions[0])
	}
	if got := aws.ToInt32(message.Value.MessageCount); got != 12 {
		t.Errorf("expected message_count 12, got %d", got)
	}
}

func TestMemoryStrategyTriggerConditionsExpandThroughConfiguration(t *testing.T) {
	t.Parallel()

	ctx := acctest.Context(t)
	selfManaged := tfbedrockagentcore.SelfManagedConfigurationModel{
		HistoricalContextWindowSize: types.Int32Value(10),
		InvocationConfiguration: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.InvocationConfigurationModel{
			PayloadDeliveryBucketName: types.StringValue("example-bucket"),
			TopicARN:                  fwtypes.ARNValue("arn:aws:sns:us-west-2:123456789012:example"),
		}),
		TriggerConditions: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.TriggerConditionsModel{
			MessageBasedTrigger: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.MessageBasedTriggerModel{
				MessageCount: types.Int32Value(12),
			}),
			TokenBasedTrigger: fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.TokenBasedTriggerModel](ctx),
			TimeBasedTrigger:  fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.TimeBasedTriggerModel](ctx),
		}),
	}
	configuration := tfbedrockagentcore.CustomConfigurationModel{
		Type:        fwtypes.StringEnumValue(awstypes.OverrideTypeSelfManaged),
		SelfManaged: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &selfManaged),
	}

	t.Run("create", func(t *testing.T) {
		result, diags := configuration.ExpandTo(ctx, reflect.TypeFor[awstypes.CustomConfigurationInput]())
		if diags.HasError() {
			t.Fatalf("expanding create configuration: %s", diags)
		}
		custom, ok := result.(*awstypes.CustomConfigurationInputMemberSelfManagedConfiguration)
		if !ok {
			t.Fatalf("expected self-managed create configuration, got %T", result)
		}
		assertMessageTriggerCount(t, custom.Value.TriggerConditions, 12)
	})

	t.Run("update", func(t *testing.T) {
		result, diags := configuration.ExpandTo(ctx, reflect.TypeFor[awstypes.ModifyStrategyConfiguration]())
		if diags.HasError() {
			t.Fatalf("expanding update configuration: %s", diags)
		}
		modify, ok := result.(*awstypes.ModifyStrategyConfiguration)
		if !ok {
			t.Fatalf("expected modify strategy configuration, got %T", result)
		}
		if modify.SelfManagedConfiguration == nil {
			t.Fatal("expected self-managed update configuration")
		}
		assertMessageTriggerCount(t, modify.SelfManagedConfiguration.TriggerConditions, 12)
	})

	selfManaged.TriggerConditions = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.TriggerConditionsModel{
		MessageBasedTrigger: fwtypes.NewListNestedObjectValueOfUnknown[tfbedrockagentcore.MessageBasedTriggerModel](ctx),
		TokenBasedTrigger:   fwtypes.NewListNestedObjectValueOfUnknown[tfbedrockagentcore.TokenBasedTriggerModel](ctx),
		TimeBasedTrigger:    fwtypes.NewListNestedObjectValueOfUnknown[tfbedrockagentcore.TimeBasedTriggerModel](ctx),
	})
	emptyConfiguration := tfbedrockagentcore.CustomConfigurationModel{
		Type:        fwtypes.StringEnumValue(awstypes.OverrideTypeSelfManaged),
		SelfManaged: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &selfManaged),
	}

	t.Run("create with only unknown children", func(t *testing.T) {
		result, diags := emptyConfiguration.ExpandTo(ctx, reflect.TypeFor[awstypes.CustomConfigurationInput]())
		if diags.HasError() {
			t.Fatalf("expanding create configuration: %s", diags)
		}
		custom, ok := result.(*awstypes.CustomConfigurationInputMemberSelfManagedConfiguration)
		if !ok {
			t.Fatalf("expected self-managed create configuration, got %T", result)
		}
		if custom.Value.TriggerConditions != nil {
			t.Errorf("expected unknown trigger conditions to be omitted, got %#v", custom.Value.TriggerConditions)
		}
	})

	t.Run("update with only unknown children", func(t *testing.T) {
		result, diags := emptyConfiguration.ExpandTo(ctx, reflect.TypeFor[awstypes.ModifyStrategyConfiguration]())
		if diags.HasError() {
			t.Fatalf("expanding update configuration: %s", diags)
		}
		modify, ok := result.(*awstypes.ModifyStrategyConfiguration)
		if !ok {
			t.Fatalf("expected modify strategy configuration, got %T", result)
		}
		if modify.SelfManagedConfiguration == nil {
			t.Fatal("expected self-managed update configuration")
		}
		if modify.SelfManagedConfiguration.TriggerConditions != nil {
			t.Errorf("expected unknown trigger conditions to be omitted, got %#v", modify.SelfManagedConfiguration.TriggerConditions)
		}
	})
}

func assertMessageTriggerCount(t *testing.T, conditions []awstypes.TriggerConditionInput, expected int32) {
	t.Helper()

	if len(conditions) != 1 {
		t.Fatalf("expected one configured trigger condition, got %d", len(conditions))
	}
	message, ok := conditions[0].(*awstypes.TriggerConditionInputMemberMessageBasedTrigger)
	if !ok {
		t.Fatalf("expected message-based trigger, got %T", conditions[0])
	}
	if got := aws.ToInt32(message.Value.MessageCount); got != expected {
		t.Errorf("expected message_count %d, got %d", expected, got)
	}
}

func TestMemoryStrategyTriggerConditionsFlatten(t *testing.T) {
	t.Parallel()

	ctx := acctest.Context(t)
	conditions := []awstypes.TriggerCondition{
		&awstypes.TriggerConditionMemberMessageBasedTrigger{
			Value: awstypes.MessageBasedTrigger{MessageCount: aws.Int32(6)},
		},
		&awstypes.TriggerConditionMemberTokenBasedTrigger{
			Value: awstypes.TokenBasedTrigger{TokenCount: aws.Int32(5000)},
		},
		&awstypes.TriggerConditionMemberTimeBasedTrigger{
			Value: awstypes.TimeBasedTrigger{IdleSessionTimeout: aws.Int32(20)},
		},
	}

	var model tfbedrockagentcore.TriggerConditionsModel
	if diags := model.Flatten(ctx, conditions); diags.HasError() {
		t.Fatalf("flattening trigger conditions: %s", diags)
	}

	message, diags := model.MessageBasedTrigger.ToPtr(ctx)
	if diags.HasError() {
		t.Fatalf("reading message-based trigger: %s", diags)
	}
	token, diags := model.TokenBasedTrigger.ToPtr(ctx)
	if diags.HasError() {
		t.Fatalf("reading token-based trigger: %s", diags)
	}
	timeBased, diags := model.TimeBasedTrigger.ToPtr(ctx)
	if diags.HasError() {
		t.Fatalf("reading time-based trigger: %s", diags)
	}

	if got := message.MessageCount.ValueInt32(); got != 6 {
		t.Errorf("expected message_count 6, got %d", got)
	}
	if got := token.TokenCount.ValueInt32(); got != 5000 {
		t.Errorf("expected token_count 5000, got %d", got)
	}
	if got := timeBased.IdleSessionTimeout.ValueInt32(); got != 20 {
		t.Errorf("expected idle_session_timeout 20, got %d", got)
	}
}

func TestMemoryStrategyTriggerConditionsFlattenThroughConfiguration(t *testing.T) {
	t.Parallel()

	ctx := acctest.Context(t)
	configuration := awstypes.StrategyConfiguration{
		Type: awstypes.OverrideTypeSelfManaged,
		SelfManagedConfiguration: &awstypes.SelfManagedConfiguration{
			HistoricalContextWindowSize: aws.Int32(10),
			InvocationConfiguration: &awstypes.InvocationConfiguration{
				PayloadDeliveryBucketName: aws.String("example-bucket"),
				TopicArn:                  aws.String("arn:aws:sns:us-west-2:123456789012:example"),
			},
			TriggerConditions: []awstypes.TriggerCondition{
				&awstypes.TriggerConditionMemberMessageBasedTrigger{
					Value: awstypes.MessageBasedTrigger{MessageCount: aws.Int32(6)},
				},
				&awstypes.TriggerConditionMemberTokenBasedTrigger{
					Value: awstypes.TokenBasedTrigger{TokenCount: aws.Int32(5000)},
				},
				&awstypes.TriggerConditionMemberTimeBasedTrigger{
					Value: awstypes.TimeBasedTrigger{IdleSessionTimeout: aws.Int32(20)},
				},
			},
		},
	}

	var model tfbedrockagentcore.CustomConfigurationModel
	if diags := model.Flatten(ctx, configuration); diags.HasError() {
		t.Fatalf("flattening strategy configuration: %s", diags)
	}
	selfManaged, diags := model.SelfManaged.ToPtr(ctx)
	if diags.HasError() {
		t.Fatalf("reading self-managed configuration: %s", diags)
	}
	conditions, diags := selfManaged.TriggerConditions.ToPtr(ctx)
	if diags.HasError() {
		t.Fatalf("reading trigger conditions: %s", diags)
	}

	if got := conditions.MessageBasedTrigger.IsNull(); got {
		t.Error("expected message-based trigger to be populated")
	}
	if got := conditions.TokenBasedTrigger.IsNull(); got {
		t.Error("expected token-based trigger to be populated")
	}
	if got := conditions.TimeBasedTrigger.IsNull(); got {
		t.Error("expected time-based trigger to be populated")
	}
}

func TestAccBedrockAgentCoreMemoryStrategy_standard(t *testing.T) {
	ctx := acctest.Context(t)
	var m awstypes.MemoryStrategy
	rName := randomMemoryName(t)
	resourceName := "aws_bedrockagentcore_memory_strategy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckMemories(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemoryStrategyDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Setup: Create memory with execution role (needed for EPISODIC steps)
			{
				Config: testAccMemoryConfig_memoryExecutionRole(rName),
			},
			// Step 1: Create episodic strategy
			{
				Config: testAccMemoryStrategyConfig_withExecutionRole(rName, "EPISODIC", "Episodic strategy", "/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "EPISODIC"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Episodic strategy"),
					resource.TestCheckResourceAttr(resourceName, "namespaces.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "memory_strategy_id"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			// Step 2: Update episodic description (in-place)
			{
				Config: testAccMemoryStrategyConfig_withExecutionRole(rName, "EPISODIC", "Updated episodic strategy", "/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "EPISODIC"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Updated episodic strategy"),
					resource.TestCheckResourceAttr(resourceName, "namespaces.#", "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			// Step 3: Change type episodic→semantic (replacement)
			{
				Config: testAccMemoryStrategyConfig_withExecutionRole(rName, "SEMANTIC", names.AttrDescription, "default"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "SEMANTIC"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "namespaces.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "namespaces.*", "default"),
					resource.TestCheckResourceAttrSet(resourceName, "memory_strategy_id"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
			// Step 4: Update description + namespace (in-place)
			{
				Config: testAccMemoryStrategyConfig_withExecutionRole(rName, "SEMANTIC", "Updated description", "custom"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "namespaces.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "namespaces.*", "custom"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			// Step 5: Change type semantic→user_preference (replacement)
			{
				Config: testAccMemoryStrategyConfig_withExecutionRole(rName, "USER_PREFERENCE", "User preference strategy", "preferences"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "USER_PREFERENCE"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "User preference strategy"),
					resource.TestCheckResourceAttr(resourceName, "namespaces.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "namespaces.*", "preferences"),
					resource.TestCheckResourceAttrSet(resourceName, "memory_strategy_id"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
			// Step 6: Try to create ANOTHER user_preference strategy → should ERROR
			{
				Config:      testAccMemoryStrategyConfig_duplicateType(rName, "USER_PREFERENCE"),
				ExpectError: regexache.MustCompile("Found multiple strategies of type"),
			},
			// Step 7: Import test - verify composite ID import works
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccMemoryStrategyImportStateIDFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "memory_strategy_id",
				ImportStateVerifyIgnore:              []string{"memory_execution_role_arn"},
			},
		},
	})
}

func TestAccBedrockAgentCoreMemoryStrategy_custom(t *testing.T) {
	ctx := acctest.Context(t)
	var m awstypes.MemoryStrategy
	rName := randomMemoryName(t)
	resourceName := "aws_bedrockagentcore_memory_strategy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckMemories(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemoryStrategyDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Setup: Create memory
			{
				Config: testAccMemoryConfig_memoryExecutionRole(rName),
			},
			// Step 1: CUSTOM type with missing configuration block → ValidateConfig error
			{
				Config:      testAccMemoryStrategyConfig_customInvalid(rName),
				ExpectError: regexache.MustCompile("When type is `CUSTOM`, the configuration block is required"),
			},
			// Step 2: Create CUSTOM strategy with consolidation block
			{
				Config: testAccMemoryStrategyConfig_customConsolidationOnly(rName, "SEMANTIC_OVERRIDE", "Focus on semantic relationships", "us.anthropic.claude-haiku-4-5-20251001-v1:0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.type", "SEMANTIC_OVERRIDE"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.consolidation.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.consolidation.0.append_to_prompt", "Focus on semantic relationships"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.consolidation.0.model_id", "us.anthropic.claude-haiku-4-5-20251001-v1:0"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.extraction.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "memory_strategy_id"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			// Step 3: Add extraction block and update consolidation properties (same override type)
			{
				Config: testAccMemoryStrategyConfig_custom(rName, "SEMANTIC_OVERRIDE", "Updated semantic consolidation", "us.anthropic.claude-sonnet-4-5-20250929-v1:0", "Extract semantic meaning", "us.anthropic.claude-haiku-4-5-20251001-v1:0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.type", "SEMANTIC_OVERRIDE"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.consolidation.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.consolidation.0.append_to_prompt", "Updated semantic consolidation"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.consolidation.0.model_id", "us.anthropic.claude-sonnet-4-5-20250929-v1:0"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.extraction.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.extraction.0.append_to_prompt", "Extract semantic meaning"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.extraction.0.model_id", "us.anthropic.claude-haiku-4-5-20251001-v1:0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			// Step 4: Try to remove consolidation block → should ERROR
			{
				Config:      testAccMemoryStrategyConfig_customExtractionOnly(rName, "SEMANTIC_OVERRIDE", "Extract semantic meaning", "us.anthropic.claude-haiku-4-5-20251001-v1:0"),
				ExpectError: regexache.MustCompile("(?s)Removing the previously configured .consolidation. block is not\\s+allowed"),
			},
			//// Step 5: Change override type → should replace resource
			{
				Config: testAccMemoryStrategyConfig_custom(rName, "USER_PREFERENCE_OVERRIDE", "Store user preferences", "us.anthropic.claude-sonnet-4-5-20250929-v1:0", "Extract user preferences", "us.anthropic.claude-haiku-4-5-20251001-v1:0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.type", "USER_PREFERENCE_OVERRIDE"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.consolidation.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.consolidation.0.append_to_prompt", "Store user preferences"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.consolidation.0.model_id", "us.anthropic.claude-sonnet-4-5-20250929-v1:0"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.extraction.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.extraction.0.append_to_prompt", "Extract user preferences"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.extraction.0.model_id", "us.anthropic.claude-haiku-4-5-20251001-v1:0"),
					resource.TestCheckResourceAttrSet(resourceName, "memory_strategy_id"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
			//// Step 6: SUMMARY_OVERRIDE with extraction block → ValidateConfig error
			{
				Config:      testAccMemoryStrategyConfig_custom(rName, "SUMMARY_OVERRIDE", "Summary consolidation", "us.anthropic.claude-sonnet-4-5-20250929-v1:0", "Summary extraction", "us.anthropic.claude-haiku-4-5-20251001-v1:0"),
				ExpectError: regexache.MustCompile("(?s)When\\s+configuration\\s+type\\s+is\\s+`SUMMARY_OVERRIDE`,\\s+the\\s+extraction\\s+block\\s+cannot\\s+be\\s+defined"),
			},
			//// Step 7: SUMMARY_OVERRIDE with no extraction block → should succeed
			{
				Config: testAccMemoryStrategyConfig_customConsolidationOnly(rName, "SUMMARY_OVERRIDE", "Summary consolidation only", "us.anthropic.claude-sonnet-4-5-20250929-v1:0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.type", "SUMMARY_OVERRIDE"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.consolidation.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.consolidation.0.append_to_prompt", "Summary consolidation only"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.consolidation.0.model_id", "us.anthropic.claude-sonnet-4-5-20250929-v1:0"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.extraction.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "memory_strategy_id"),
				),
			},
			//Step 8: Add 6 more CUSTOM strategies → should ERROR on too many
			{
				Config:      testAccMemoryStrategyConfig_customTooMany(rName, "USER_PREFERENCE_OVERRIDE", "Store user preferences", "us.anthropic.claude-sonnet-4-5-20250929-v1:0", "Extract user preferences", "us.anthropic.claude-haiku-4-5-20251001-v1:0"),
				ExpectError: regexache.MustCompile("(?s)Resource limit exceeded for memory strategies"),
			},
			// Step 9: Import test
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccMemoryStrategyImportStateIDFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "memory_strategy_id",
				ImportStateVerifyIgnore:              []string{"memory_execution_role_arn"},
			},
		},
	})
}

func TestAccBedrockAgentCoreMemoryStrategy_selfManaged(t *testing.T) {
	ctx := acctest.Context(t)
	var m awstypes.MemoryStrategy
	rName := randomMemoryName(t)
	resourceName := "aws_bedrockagentcore_memory_strategy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckMemories(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemoryStrategyDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Setup: memory + SNS topic + S3 bucket + execution role permissions
			{
				Config: testAccMemoryStrategyConfig_selfManagedBase(rName),
			},
			// Step 1: Create with trigger_conditions omitted. The service supplies all defaults.
			{
				Config: testAccMemoryStrategyConfig_selfManaged(rName, 10, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.type", "SELF_MANAGED"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.self_managed.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.self_managed.0.historical_context_window_size", "10"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.self_managed.0.invocation_configuration.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "configuration.0.self_managed.0.invocation_configuration.0.topic_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "configuration.0.self_managed.0.invocation_configuration.0.payload_delivery_bucket_name"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.self_managed.0.trigger_conditions.0.message_based_trigger.0.message_count", "6"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.self_managed.0.trigger_conditions.0.token_based_trigger.0.token_count", "5000"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.self_managed.0.trigger_conditions.0.time_based_trigger.0.idle_session_timeout", "20"),
					resource.TestCheckResourceAttrSet(resourceName, "memory_strategy_id"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			// Step 2: Partially configure only the message threshold. The service
			// still applies defaults for omitted conditions, while Terraform state
			// preserves the configured shape required by protocol version 5.
			{
				Config: testAccMemoryStrategyConfig_selfManaged(rName, 25, testAccMemoryStrategyTriggerConditionsMessageOnly(12)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.self_managed.0.historical_context_window_size", "25"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.self_managed.0.trigger_conditions.0.message_based_trigger.0.message_count", "12"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.self_managed.0.trigger_conditions.0.token_based_trigger.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.self_managed.0.trigger_conditions.0.time_based_trigger.#", "0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			// Step 3: Update every supported threshold in place.
			{
				Config: testAccMemoryStrategyConfig_selfManaged(rName, 25, testAccMemoryStrategyTriggerConditions(18, 6000, 30)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.self_managed.0.trigger_conditions.0.message_based_trigger.0.message_count", "18"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.self_managed.0.trigger_conditions.0.token_based_trigger.0.token_count", "6000"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.self_managed.0.trigger_conditions.0.time_based_trigger.0.idle_session_timeout", "30"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			// Step 4: Import retains the normalized trigger conditions returned by the service.
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccMemoryStrategyImportStateIDFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "memory_strategy_id",
				ImportStateVerifyIgnore:              []string{"memory_execution_role_arn"},
			},
		},
	})
}

func TestAccBedrockAgentCoreMemoryStrategy_selfManagedInvalidType(t *testing.T) {
	ctx := acctest.Context(t)
	rName := randomMemoryName(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckMemories(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemoryStrategyDestroy(ctx, t),
		Steps: []resource.TestStep{
			// self_managed under a non-SELF_MANAGED type → ValidateConfig error.
			// Isolated in its own test so the plan-time validation error leaves no
			// resource to destroy.
			{
				Config:      testAccMemoryStrategyConfig_selfManagedInvalidType(rName),
				ExpectError: regexache.MustCompile(`self_managed block is only valid`),
			},
		},
	})
}

// TestAccBedrockAgentCoreMemoryStrategy_rename verifies that changing `name` forces
// replacement. The service's ModifyMemoryStrategyInput has no Name field, so a rename
// cannot be sent — an in-place update would leave the server name unchanged and
// produce "inconsistent result after apply".
func TestAccBedrockAgentCoreMemoryStrategy_rename(t *testing.T) {
	ctx := acctest.Context(t)
	var m awstypes.MemoryStrategy
	rName := randomMemoryName(t)
	rNameNew := randomMemoryName(t)
	resourceName := "aws_bedrockagentcore_memory_strategy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckMemories(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemoryStrategyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMemoryStrategyConfig_basic(rName, "SEMANTIC", "Rename test", "default"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccMemoryStrategyConfig_basic(rNameNew, "SEMANTIC", "Rename test", "default"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameNew),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
		},
	})
}

// TestAccBedrockAgentCoreMemoryStrategy_descriptionClearConverges verifies that
// removing `description` from configuration after it was set does not produce
// "inconsistent result after apply" or a perpetual diff. The PATCH API ignores
// a nil Description and retains the prior value; Optional+Computed absorbs it so
// the resource still converges (documented limitation: a description cannot be
// cleared once set via this API).
func TestAccBedrockAgentCoreMemoryStrategy_descriptionClearConverges(t *testing.T) {
	ctx := acctest.Context(t)
	var m awstypes.MemoryStrategy
	rName := randomMemoryName(t)
	resourceName := "aws_bedrockagentcore_memory_strategy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckMemories(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemoryStrategyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMemoryStrategyConfig_basic(rName, "SEMANTIC", "Initial description", "default"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Initial description"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				// Remove description from config. The API retains the prior value; Optional+Computed
				// absorbs it, so this must apply cleanly. The step's implicit post-apply refresh
				// asserts no "inconsistent result" and no perpetual diff.
				Config: testAccMemoryStrategyConfig_noDescription(rName, "SEMANTIC", "default"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					// Server retains the original description; state absorbs it.
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Initial description"),
				),
			},
		},
	})
}

func TestAccBedrockAgentCoreMemoryStrategy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var m awstypes.MemoryStrategy
	rName := randomMemoryName(t)
	resourceName := "aws_bedrockagentcore_memory_strategy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckMemories(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemoryStrategyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMemoryStrategyConfig_basic(rName, "SEMANTIC", "Example Description", "default"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrockagentcore.ResourceMemoryStrategy, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckMemoryStrategyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_memory_strategy" {
				continue
			}

			_, err := tfbedrockagentcore.FindMemoryStrategyByTwoPartKey(ctx, conn, rs.Primary.Attributes["memory_id"], rs.Primary.Attributes["memory_strategy_id"])
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Agent Core Memory Strategy %s,%s still exists", rs.Primary.Attributes["memory_id"], rs.Primary.Attributes["memory_strategy_id"])
		}

		return nil
	}
}

func testAccCheckMemoryStrategyExists(ctx context.Context, t *testing.T, n string, v *awstypes.MemoryStrategy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		resp, err := tfbedrockagentcore.FindMemoryStrategyByTwoPartKey(ctx, conn, rs.Primary.Attributes["memory_id"], rs.Primary.Attributes["memory_strategy_id"])
		if err != nil {
			return err
		}

		*v = *resp

		return nil
	}
}

func testAccMemoryStrategyImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return acctest.AttrsImportStateIdFunc(resourceName, ",", "memory_id", "memory_strategy_id")
}

func testAccMemoryStrategyConfig_basic(rName, strategyType, description, namespace string) string {
	return acctest.ConfigCompose(testAccMemoryConfig_basic(rName), fmt.Sprintf(`	
resource "aws_bedrockagentcore_memory_strategy" "test" {
  name        = %[1]q
  memory_id   = aws_bedrockagentcore_memory.test.id
  type        = %[2]q
  description = %[3]q
  namespaces  = [%[4]q]
}
`, rName, strategyType, description, namespace))
}

func testAccMemoryStrategyConfig_noDescription(rName, strategyType, namespace string) string {
	return acctest.ConfigCompose(testAccMemoryConfig_basic(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_memory_strategy" "test" {
  name       = %[1]q
  memory_id  = aws_bedrockagentcore_memory.test.id
  type       = %[2]q
  namespaces = [%[3]q]
}
`, rName, strategyType, namespace))
}

func testAccMemoryStrategyConfig_withExecutionRole(rName, strategyType, description, namespace string) string {
	return acctest.ConfigCompose(testAccMemoryConfig_memoryExecutionRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_memory_strategy" "test" {
  name                      = %[1]q
  memory_id                 = aws_bedrockagentcore_memory.test.id
  memory_execution_role_arn = aws_bedrockagentcore_memory.test.memory_execution_role_arn
  type                      = %[2]q
  description               = %[3]q
  namespaces                = [%[4]q]
}
`, rName, strategyType, description, namespace))
}

func testAccMemoryStrategyConfig_duplicateType(rName string, strategyType string) string {
	namespace := "default"
	duplicateNamespace := "duplicate"
	if strategyType == "EPISODIC" {
		namespace = "/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}"
		duplicateNamespace = "/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}"
	}
	return acctest.ConfigCompose(testAccMemoryStrategyConfig_withExecutionRole(rName, strategyType, "Strategy for duplicate test", namespace), fmt.Sprintf(`	
resource "aws_bedrockagentcore_memory_strategy" "test2" {
  name                      = "%[1]s_duplicate"
  memory_id                 = aws_bedrockagentcore_memory.test.id
  memory_execution_role_arn = aws_bedrockagentcore_memory.test.memory_execution_role_arn
  type                      = %[2]q
  description               = "Duplicate strategy"
  namespaces                = [%[3]q]
}
`, rName, strategyType, duplicateNamespace))
}

func testAccMemoryStrategyConfig_custom(rName, overrideType, consolidationPrompt, consolidationModel, extractionPrompt, extractionModel string) string {
	return acctest.ConfigCompose(testAccMemoryConfig_memoryExecutionRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_memory_strategy" "test" {
  name                      = %[1]q
  memory_id                 = aws_bedrockagentcore_memory.test.id
  memory_execution_role_arn = aws_bedrockagentcore_memory.test.memory_execution_role_arn
  type                      = "CUSTOM"
  description               = "Test custom strategy"
  namespaces                = ["{sessionId}"]

  configuration {
    type = %[2]q
    consolidation {
      append_to_prompt = %[3]q
      model_id         = %[4]q
    }
    extraction {
      append_to_prompt = %[5]q
      model_id         = %[6]q
    }
  }
}
`, rName, overrideType, consolidationPrompt, consolidationModel, extractionPrompt, extractionModel))
}

func testAccMemoryStrategyConfig_customConsolidationOnly(rName, overrideType, consolidationPrompt, consolidationModel string) string {
	return acctest.ConfigCompose(testAccMemoryConfig_memoryExecutionRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_memory_strategy" "test" {
  name                      = %[1]q
  memory_id                 = aws_bedrockagentcore_memory.test.id
  memory_execution_role_arn = aws_bedrockagentcore_memory.test.memory_execution_role_arn
  type                      = "CUSTOM"
  description               = "Test custom strategy"
  namespaces                = ["{sessionId}"]

  configuration {
    type = %[2]q
    consolidation {
      append_to_prompt = %[3]q
      model_id         = %[4]q
    }
  }
}
`, rName, overrideType, consolidationPrompt, consolidationModel))
}

func testAccMemoryStrategyConfig_customExtractionOnly(rName, overrideType, extractionPrompt, extractionModel string) string {
	return acctest.ConfigCompose(testAccMemoryConfig_memoryExecutionRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_memory_strategy" "test" {
  name                      = %[1]q
  memory_id                 = aws_bedrockagentcore_memory.test.id
  memory_execution_role_arn = aws_bedrockagentcore_memory.test.memory_execution_role_arn
  type                      = "CUSTOM"
  description               = "Test custom strategy"
  namespaces                = ["{sessionId}"]

  configuration {
    type = %[2]q
    extraction {
      append_to_prompt = %[3]q
      model_id         = %[4]q
    }
  }
}
`, rName, overrideType, extractionPrompt, extractionModel))
}

func testAccMemoryStrategyConfig_customDummy(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_memory_strategy" "%[1]s" {
  name                      = %[1]q
  memory_id                 = aws_bedrockagentcore_memory.test.id
  memory_execution_role_arn = aws_bedrockagentcore_memory.test.memory_execution_role_arn
  type                      = "CUSTOM"
  description               = "Test custom strategy"
  namespaces                = ["{sessionId}"]

  configuration {
    type = "SEMANTIC_OVERRIDE"
    extraction {
      append_to_prompt = "test"
      model_id         = "us.anthropic.claude-haiku-4-5-20251001-v1:0"
    }
  }
  # Preventing the main strategy from being lost due to "too many resources" error
  depends_on = [aws_bedrockagentcore_memory_strategy.test]
}
`, rName)
}

func testAccMemoryStrategyConfig_customTooMany(rName, overrideType, consolidationPrompt, consolidationModel, extractionPrompt, extractionModel string) string {
	compose := acctest.ConfigCompose(testAccMemoryStrategyConfig_custom(rName, overrideType, consolidationPrompt, consolidationModel, extractionPrompt, extractionModel),
		testAccMemoryStrategyConfig_customDummy(rName+"_2"),
		testAccMemoryStrategyConfig_customDummy(rName+"_3"),
		testAccMemoryStrategyConfig_customDummy(rName+"_4"),
		testAccMemoryStrategyConfig_customDummy(rName+"_5"),
		testAccMemoryStrategyConfig_customDummy(rName+"_6"),
		testAccMemoryStrategyConfig_customDummy(rName+"_7"))
	return compose
}

func testAccMemoryStrategyConfig_customInvalid(rName string) string {
	return acctest.ConfigCompose(testAccMemoryConfig_memoryExecutionRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_memory_strategy" "test" {
  name                      = %[1]q
  memory_execution_role_arn = aws_bedrockagentcore_memory.test.memory_execution_role_arn
  memory_id                 = aws_bedrockagentcore_memory.test.id
  type                      = "CUSTOM"
  description               = "Test custom strategy"
  namespaces                = ["default"]
}
`, rName))
}

// testAccMemoryStrategyConfig_selfManagedBase provisions the memory plus the SNS
// topic and S3 bucket that a self-managed strategy's invocation_configuration
// requires, granting the memory execution role permission to use them.
func testAccMemoryStrategyConfig_selfManagedBase(rName string) string {
	return acctest.ConfigCompose(testAccMemoryConfig_baseIAMRole(rName), fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_s3_bucket" "test" {
  bucket_prefix = "tf-acc-test-agentcore"
  force_destroy = true
}

resource "aws_iam_role_policy" "test_self_managed" {
  role = aws_iam_role.test.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sns:Publish",
          "sns:GetTopicAttributes",
          "sns:Subscribe",
        ]
        Resource = aws_sns_topic.test.arn
      },
      {
        Effect = "Allow"
        Action = [
          "s3:PutObject",
          "s3:GetObject",
          "s3:ListBucket",
          "s3:GetBucketLocation",
          "s3:DeleteObject",
        ]
        Resource = [
          aws_s3_bucket.test.arn,
          "${aws_s3_bucket.test.arn}/*",
        ]
      },
    ]
  })
}

resource "aws_sns_topic_policy" "test" {
  arn = aws_sns_topic.test.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "bedrock-agentcore.amazonaws.com" }
      Action    = ["sns:Publish"]
      Resource  = aws_sns_topic.test.arn
    }]
  })
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "bedrock-agentcore.amazonaws.com" }
      Action = [
        "s3:PutObject",
        "s3:GetObject",
        "s3:ListBucket",
      ]
      Resource = [
        aws_s3_bucket.test.arn,
        "${aws_s3_bucket.test.arn}/*",
      ]
    }]
  })
}

resource "aws_bedrockagentcore_memory" "test" {
  name                      = %[1]q
  event_expiry_duration     = 7
  memory_execution_role_arn = aws_iam_role.test.arn

  depends_on = [aws_iam_role_policy.test_self_managed]
}
`, rName))
}

func testAccMemoryStrategyConfig_selfManaged(rName string, windowSize int, triggerConditions string) string {
	return acctest.ConfigCompose(testAccMemoryStrategyConfig_selfManagedBase(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_memory_strategy" "test" {
  name                      = %[1]q
  memory_id                 = aws_bedrockagentcore_memory.test.id
  memory_execution_role_arn = aws_bedrockagentcore_memory.test.memory_execution_role_arn
  type                      = "CUSTOM"
  description               = "Test self-managed strategy"

  configuration {
    type = "SELF_MANAGED"

    self_managed {
      historical_context_window_size = %[2]d

      %[3]s

      invocation_configuration {
        topic_arn                    = aws_sns_topic.test.arn
        payload_delivery_bucket_name = aws_s3_bucket.test.bucket
      }
    }
  }

  depends_on = [aws_iam_role_policy.test_self_managed, aws_sns_topic_policy.test, aws_s3_bucket_policy.test]
}
`, rName, windowSize, triggerConditions))
}

func testAccMemoryStrategyTriggerConditionsMessageOnly(messageCount int) string {
	return fmt.Sprintf(`
trigger_conditions {
  message_based_trigger {
    message_count = %d
  }
}
`, messageCount)
}

func testAccMemoryStrategyTriggerConditions(messageCount, tokenCount, idleSessionTimeout int) string {
	return fmt.Sprintf(`
trigger_conditions {
  message_based_trigger {
    message_count = %d
  }
  token_based_trigger {
    token_count = %d
  }
  time_based_trigger {
    idle_session_timeout = %d
  }
}
`, messageCount, tokenCount, idleSessionTimeout)
}

func testAccMemoryStrategyConfig_selfManagedInvalidType(rName string) string {
	return acctest.ConfigCompose(testAccMemoryStrategyConfig_selfManagedBase(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_memory_strategy" "test" {
  name                      = %[1]q
  memory_id                 = aws_bedrockagentcore_memory.test.id
  memory_execution_role_arn = aws_bedrockagentcore_memory.test.memory_execution_role_arn
  type                      = "CUSTOM"
  description               = "Test self-managed invalid type"
  namespaces                = ["{sessionId}"]

  configuration {
    type = "SEMANTIC_OVERRIDE"

    self_managed {
      invocation_configuration {
        topic_arn                    = aws_sns_topic.test.arn
        payload_delivery_bucket_name = aws_s3_bucket.test.bucket
      }
    }
  }
}
`, rName))
}
