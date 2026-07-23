// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfbedrockagentcore "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagentcore"
)

// TestMemoryStrategyResourceModelExpand verifies that fwflex.Expand correctly
// converts a memoryStrategyResourceModel into an awstypes.ModifyMemoryStrategyInput
// via the TypedExpander.ExpandTo implementation.
func TestMemoryStrategyResourceModelExpand(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := []struct {
		name     string
		model    tfbedrockagentcore.MemoryStrategyResourceModel
		expected awstypes.ModifyMemoryStrategyInput
	}{
		{
			name: "basic_summarization_no_configuration",
			model: tfbedrockagentcore.MemoryStrategyResourceModel{
				MemoryStrategyID: types.StringValue("strat-001"),
				Description:     types.StringValue("summarization strategy"),
				NamespaceTemplates: fwtypes.NewSetValueOfMust[types.String](ctx, []attr.Value{
					types.StringValue("ns/{{actorId}}/summary"),
				}),
				Namespaces: fwtypes.NewSetValueOfMust[types.String](ctx, []attr.Value{
					types.StringValue("ns/{{actorId}}/summary"),
				}),
				Configuration: fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.CustomConfigurationModel](ctx),
			},
			expected: awstypes.ModifyMemoryStrategyInput{
				MemoryStrategyId:   aws.String("strat-001"),
				Description:        aws.String("summarization strategy"),
				Namespaces:         []string{"ns/{{actorId}}/summary"},
				NamespaceTemplates: []string{"ns/{{actorId}}/summary"},
				Configuration:      nil,
			},
		},
		{
			name: "custom_with_semantic_override_configuration",
			model: tfbedrockagentcore.MemoryStrategyResourceModel{
				MemoryStrategyID: types.StringValue("strat-002"),
				Description:     types.StringValue("custom semantic strategy"),
				NamespaceTemplates: fwtypes.NewSetValueOfMust[types.String](ctx, []attr.Value{
					types.StringValue("ns/{{actorId}}/semantic"),
				}),
				Namespaces: fwtypes.NewSetValueOfMust[types.String](ctx, []attr.Value{
					types.StringValue("ns/{{actorId}}/semantic"),
				}),
				Configuration: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.CustomConfigurationModel{
					Type: fwtypes.StringEnumValue(awstypes.OverrideTypeSemanticOverride),
					Consolidation: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.OverrideDetailsModel{
						AppendToPrompt: types.StringValue("append consolidation"),
						ModelID:        types.StringValue("amazon.titan-text-v1"),
					}),
					Extraction: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.OverrideDetailsModel{
						AppendToPrompt: types.StringValue("append extraction"),
						ModelID:        types.StringValue("amazon.titan-text-v1"),
					}),
				}),
			},
			expected: awstypes.ModifyMemoryStrategyInput{
				MemoryStrategyId:   aws.String("strat-002"),
				Description:        aws.String("custom semantic strategy"),
				Namespaces:         []string{"ns/{{actorId}}/semantic"},
				NamespaceTemplates: []string{"ns/{{actorId}}/semantic"},
				Configuration: &awstypes.ModifyStrategyConfiguration{
					Consolidation: &awstypes.ModifyConsolidationConfigurationMemberCustomConsolidationConfiguration{
						Value: &awstypes.CustomConsolidationConfigurationInputMemberSemanticConsolidationOverride{
							Value: awstypes.SemanticOverrideConsolidationConfigurationInput{
								AppendToPrompt: aws.String("append consolidation"),
								ModelId:        aws.String("amazon.titan-text-v1"),
							},
						},
					},
					Extraction: &awstypes.ModifyExtractionConfigurationMemberCustomExtractionConfiguration{
						Value: &awstypes.CustomExtractionConfigurationInputMemberSemanticExtractionOverride{
							Value: awstypes.SemanticOverrideExtractionConfigurationInput{
								AppendToPrompt: aws.String("append extraction"),
								ModelId:        aws.String("amazon.titan-text-v1"),
							},
						},
					},
				},
			},
		},
		{
			name: "null_configuration_is_cleared",
			// Even when AutoFlex might emit an empty ModifyStrategyConfiguration from a null
			// model field, expandToModifyMemoryStrategyInput explicitly nils it out.
			model: tfbedrockagentcore.MemoryStrategyResourceModel{
				MemoryStrategyID: types.StringValue("strat-003"),
				NamespaceTemplates: fwtypes.NewSetValueOfMust[types.String](ctx, []attr.Value{
					types.StringValue("ns/{{actorId}}/sem"),
				}),
				Namespaces: fwtypes.NewSetValueOfMust[types.String](ctx, []attr.Value{
					types.StringValue("ns/{{actorId}}/sem"),
				}),
				Configuration: fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.CustomConfigurationModel](ctx),
			},
			expected: awstypes.ModifyMemoryStrategyInput{
				MemoryStrategyId:   aws.String("strat-003"),
				Namespaces:         []string{"ns/{{actorId}}/sem"},
				NamespaceTemplates: []string{"ns/{{actorId}}/sem"},
				Configuration:      nil,
			},
		},
	}

	opts := cmp.Options{
		cmpopts.IgnoreUnexported(
			awstypes.ModifyMemoryStrategyInput{},
			awstypes.ModifyStrategyConfiguration{},
			awstypes.ModifyConsolidationConfigurationMemberCustomConsolidationConfiguration{},
			awstypes.ModifyExtractionConfigurationMemberCustomExtractionConfiguration{},
			awstypes.CustomConsolidationConfigurationInputMemberSemanticConsolidationOverride{},
			awstypes.CustomExtractionConfigurationInputMemberSemanticExtractionOverride{},
			awstypes.SemanticOverrideConsolidationConfigurationInput{},
			awstypes.SemanticOverrideExtractionConfigurationInput{},
		),
		cmpopts.SortSlices(func(a, b string) bool { return a < b }),
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var out awstypes.ModifyMemoryStrategyInput
			if diags := fwflex.Expand(ctx, tc.model, &out); diags.HasError() {
				t.Fatalf("Expand diagnostics: %v", diags)
			}

			if diff := cmp.Diff(tc.expected, out, opts...); diff != "" {
				t.Errorf("model -> ModifyMemoryStrategyInput mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestMemoryStrategyResourceModelFlatten verifies that fwflex.Flatten correctly
// converts an awstypes.MemoryStrategy into a memoryStrategyResourceModel via
// the Flattener.Flatten implementation.
func TestMemoryStrategyResourceModelFlatten(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := []struct {
		name       string
		input      awstypes.MemoryStrategy
		checkModel func(t *testing.T, model tfbedrockagentcore.MemoryStrategyResourceModel)
	}{
		{
			name: "summarization_no_configuration",
			input: awstypes.MemoryStrategy{
				StrategyId:  aws.String("strat-001"),
				Name:        aws.String("my-summary"),
				Description: aws.String("a summary strategy"),
				Type:        awstypes.MemoryStrategyTypeSummarization,
				Namespaces:  []string{"ns/{{actorId}}/summary"},
			},
			checkModel: func(t *testing.T, model tfbedrockagentcore.MemoryStrategyResourceModel) {
				t.Helper()
				if got, want := model.MemoryStrategyID.ValueString(), "strat-001"; got != want {
					t.Errorf("MemoryStrategyID = %q, want %q", got, want)
				}
				if got, want := model.Name.ValueString(), "my-summary"; got != want {
					t.Errorf("Name = %q, want %q", got, want)
				}
				if got, want := model.Description.ValueString(), "a summary strategy"; got != want {
					t.Errorf("Description = %q, want %q", got, want)
				}
				if got, want := model.Type.ValueEnum(), awstypes.MemoryStrategyTypeSummarization; got != want {
					t.Errorf("Type = %v, want %v", got, want)
				}
				// For non-CUSTOM types, Configuration must be null or empty.
				if !model.Configuration.IsNull() && len(model.Configuration.Elements()) > 0 {
					t.Errorf("Configuration must be null or empty for non-CUSTOM type, got %d element(s)", len(model.Configuration.Elements()))
				}
				// Namespaces is populated directly from MemoryStrategy.Namespaces.
				ns, diags := model.Namespaces.ToSetValue(ctx)
				if diags.HasError() {
					t.Fatalf("Namespaces.ToSetValue: %v", diags)
				}
				if got, want := len(ns.Elements()), 1; got != want {
					t.Errorf("Namespaces len = %d, want %d", got, want)
				}
			},
		},
		{
			name: "non_custom_type_with_api_configuration_is_cleared",
			// The Flatten implementation strips Configuration for non-CUSTOM types to
			// avoid invalid OverrideType values that the API may return for built-in
			// strategy types (e.g. "EPISODIC" is not a valid OverrideType enum value).
			input: awstypes.MemoryStrategy{
				StrategyId: aws.String("strat-002"),
				Name:       aws.String("my-semantic"),
				Type:       awstypes.MemoryStrategyTypeSemantic,
				Namespaces: []string{"ns/{{actorId}}/sem"},
				// The API returns a StrategyConfiguration even for non-CUSTOM types.
				Configuration: &awstypes.StrategyConfiguration{
					Type: awstypes.OverrideTypeSemanticOverride,
				},
			},
			checkModel: func(t *testing.T, model tfbedrockagentcore.MemoryStrategyResourceModel) {
				t.Helper()
				if got, want := model.MemoryStrategyID.ValueString(), "strat-002"; got != want {
					t.Errorf("MemoryStrategyID = %q, want %q", got, want)
				}
				if got, want := model.Type.ValueEnum(), awstypes.MemoryStrategyTypeSemantic; got != want {
					t.Errorf("Type = %v, want %v", got, want)
				}
				// Configuration must be cleared regardless of what the API returned.
				if !model.Configuration.IsNull() && len(model.Configuration.Elements()) > 0 {
					t.Errorf("Configuration must be null or empty for non-CUSTOM type, got %d element(s)", len(model.Configuration.Elements()))
				}
			},
		},
		{
			name: "custom_type_with_semantic_override_configuration",
			input: awstypes.MemoryStrategy{
				StrategyId:  aws.String("strat-003"),
				Name:        aws.String("my-custom"),
				Description: aws.String("a custom strategy"),
				Type:        awstypes.MemoryStrategyTypeCustom,
				Namespaces:  []string{"ns/{{actorId}}/custom"},
				Configuration: &awstypes.StrategyConfiguration{
					Type: awstypes.OverrideTypeSemanticOverride,
					Consolidation: &awstypes.ConsolidationConfigurationMemberCustomConsolidationConfiguration{
						Value: &awstypes.CustomConsolidationConfigurationMemberSemanticConsolidationOverride{
							Value: awstypes.SemanticConsolidationOverride{
								AppendToPrompt: aws.String("consolidation prompt"),
								ModelId:        aws.String("amazon.titan-text-v1"),
							},
						},
					},
					Extraction: &awstypes.ExtractionConfigurationMemberCustomExtractionConfiguration{
						Value: &awstypes.CustomExtractionConfigurationMemberSemanticExtractionOverride{
							Value: awstypes.SemanticExtractionOverride{
								AppendToPrompt: aws.String("extraction prompt"),
								ModelId:        aws.String("amazon.titan-text-v1"),
							},
						},
					},
				},
			},
			checkModel: func(t *testing.T, model tfbedrockagentcore.MemoryStrategyResourceModel) {
				t.Helper()
				if got, want := model.MemoryStrategyID.ValueString(), "strat-003"; got != want {
					t.Errorf("MemoryStrategyID = %q, want %q", got, want)
				}
				if got, want := model.Type.ValueEnum(), awstypes.MemoryStrategyTypeCustom; got != want {
					t.Errorf("Type = %v, want %v", got, want)
				}
				// Configuration must be populated for CUSTOM type.
				if model.Configuration.IsNull() {
					t.Fatal("Configuration must not be null for CUSTOM type")
				}
				cfgs, diags := model.Configuration.ToSlice(ctx)
				if diags.HasError() {
					t.Fatalf("Configuration.ToSlice: %v", diags)
				}
				if got, want := len(cfgs), 1; got != want {
					t.Fatalf("Configuration len = %d, want %d", got, want)
				}
				cfg := cfgs[0]
				if got, want := cfg.Type.ValueEnum(), awstypes.OverrideTypeSemanticOverride; got != want {
					t.Errorf("Configuration[0].Type = %v, want %v", got, want)
				}
				// Consolidation.
				cons, diags := cfg.Consolidation.ToSlice(ctx)
				if diags.HasError() {
					t.Fatalf("Consolidation.ToSlice: %v", diags)
				}
				if got, want := len(cons), 1; got != want {
					t.Fatalf("Consolidation len = %d, want %d", got, want)
				}
				if got, want := cons[0].AppendToPrompt.ValueString(), "consolidation prompt"; got != want {
					t.Errorf("Consolidation[0].AppendToPrompt = %q, want %q", got, want)
				}
				if got, want := cons[0].ModelID.ValueString(), "amazon.titan-text-v1"; got != want {
					t.Errorf("Consolidation[0].ModelID = %q, want %q", got, want)
				}
				// Extraction.
				exts, diags := cfg.Extraction.ToSlice(ctx)
				if diags.HasError() {
					t.Fatalf("Extraction.ToSlice: %v", diags)
				}
				if got, want := len(exts), 1; got != want {
					t.Fatalf("Extraction len = %d, want %d", got, want)
				}
				if got, want := exts[0].AppendToPrompt.ValueString(), "extraction prompt"; got != want {
					t.Errorf("Extraction[0].AppendToPrompt = %q, want %q", got, want)
				}
				if got, want := exts[0].ModelID.ValueString(), "amazon.titan-text-v1"; got != want {
					t.Errorf("Extraction[0].ModelID = %q, want %q", got, want)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var model tfbedrockagentcore.MemoryStrategyResourceModel
			if diags := fwflex.Flatten(ctx, tc.input, &model); diags.HasError() {
				t.Fatalf("Flatten diagnostics: %v", diags)
			}

			tc.checkModel(t, model)
		})
	}
}
