// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfbedrockagentcore "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagentcore"
)

func TestMemoryStrategyResourceModelExpand(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	testCases := []struct {
		name     string
		model    tfbedrockagentcore.MemoryStrategyResourceModel
		expected awstypes.ModifyMemoryStrategyInput
	}{
		{
			name: "basic summarization",
			model: tfbedrockagentcore.MemoryStrategyResourceModel{
				Configuration:           fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.CustomConfigurationModel](ctx),
				Description:             types.StringNull(),
				MemoryExecutionRoleARN:  fwtypes.ARNNull(),
				MemoryID:                types.StringValue("memory_001"),
				Name:                    types.StringValue("summarization_builtin_001"),
				Namespaces:              fwtypes.NewSetValueOfNull[types.String](ctx),
				NamespaceTemplates:      fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}/"}),
				ReflectionConfiguration: fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.EpisodicReflectionConfigurationModel](ctx),
				Type:                    fwtypes.StringEnumValue(awstypes.MemoryStrategyTypeSummarization),
			},
			expected: awstypes.ModifyMemoryStrategyInput{
				NamespaceTemplates: []string{"/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}/"},
			},
		},
		{
			name: "basic summarization with description",
			model: tfbedrockagentcore.MemoryStrategyResourceModel{
				Configuration:           fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.CustomConfigurationModel](ctx),
				Description:             types.StringValue("description_001"),
				MemoryExecutionRoleARN:  fwtypes.ARNNull(),
				MemoryID:                types.StringValue("memory_001"),
				Name:                    types.StringValue("summarization_builtin_001"),
				Namespaces:              fwtypes.NewSetValueOfNull[types.String](ctx),
				NamespaceTemplates:      fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}/"}),
				ReflectionConfiguration: fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.EpisodicReflectionConfigurationModel](ctx),
				Type:                    fwtypes.StringEnumValue(awstypes.MemoryStrategyTypeSummarization),
			},
			expected: awstypes.ModifyMemoryStrategyInput{
				Description:        aws.String("description_001"),
				NamespaceTemplates: []string{"/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}/"},
			},
		},
		{
			name: "basic semantic",
			model: tfbedrockagentcore.MemoryStrategyResourceModel{
				Configuration:           fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.CustomConfigurationModel](ctx),
				Description:             types.StringNull(),
				MemoryExecutionRoleARN:  fwtypes.ARNNull(),
				MemoryID:                types.StringValue("memory_001"),
				Name:                    types.StringValue("semantic_builtin_001"),
				Namespaces:              fwtypes.NewSetValueOfNull[types.String](ctx),
				NamespaceTemplates:      fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"/strategies/{memoryStrategyId}/actors/{actorId}/"}),
				ReflectionConfiguration: fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.EpisodicReflectionConfigurationModel](ctx),
				Type:                    fwtypes.StringEnumValue(awstypes.MemoryStrategyTypeSemantic),
			},
			expected: awstypes.ModifyMemoryStrategyInput{
				NamespaceTemplates: []string{"/strategies/{memoryStrategyId}/actors/{actorId}/"},
			},
		},
		{
			name: "basic semantic with description",
			model: tfbedrockagentcore.MemoryStrategyResourceModel{
				Configuration:           fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.CustomConfigurationModel](ctx),
				Description:             types.StringValue("description_001"),
				MemoryExecutionRoleARN:  fwtypes.ARNNull(),
				MemoryID:                types.StringValue("memory_001"),
				Name:                    types.StringValue("semantic_builtin_001"),
				Namespaces:              fwtypes.NewSetValueOfNull[types.String](ctx),
				NamespaceTemplates:      fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"/strategies/{memoryStrategyId}/actors/{actorId}/"}),
				ReflectionConfiguration: fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.EpisodicReflectionConfigurationModel](ctx),
				Type:                    fwtypes.StringEnumValue(awstypes.MemoryStrategyTypeSemantic),
			},
			expected: awstypes.ModifyMemoryStrategyInput{
				Description:        aws.String("description_001"),
				NamespaceTemplates: []string{"/strategies/{memoryStrategyId}/actors/{actorId}/"},
			},
		},
		{
			name: "basic user preference",
			model: tfbedrockagentcore.MemoryStrategyResourceModel{
				Configuration:           fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.CustomConfigurationModel](ctx),
				Description:             types.StringNull(),
				MemoryExecutionRoleARN:  fwtypes.ARNNull(),
				MemoryID:                types.StringValue("memory_001"),
				Name:                    types.StringValue("user_preference_builtin_001"),
				Namespaces:              fwtypes.NewSetValueOfNull[types.String](ctx),
				NamespaceTemplates:      fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"/strategies/{memoryStrategyId}/actors/{actorId}/"}),
				ReflectionConfiguration: fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.EpisodicReflectionConfigurationModel](ctx),
				Type:                    fwtypes.StringEnumValue(awstypes.MemoryStrategyTypeUserPreference),
			},
			expected: awstypes.ModifyMemoryStrategyInput{
				NamespaceTemplates: []string{"/strategies/{memoryStrategyId}/actors/{actorId}/"},
			},
		},
		{
			name: "basic user preference with description",
			model: tfbedrockagentcore.MemoryStrategyResourceModel{
				Configuration:           fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.CustomConfigurationModel](ctx),
				Description:             types.StringValue("description_001"),
				MemoryExecutionRoleARN:  fwtypes.ARNNull(),
				MemoryID:                types.StringValue("memory_001"),
				Name:                    types.StringValue("user_preference_builtin_001"),
				Namespaces:              fwtypes.NewSetValueOfNull[types.String](ctx),
				NamespaceTemplates:      fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"/strategies/{memoryStrategyId}/actors/{actorId}/"}),
				ReflectionConfiguration: fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.EpisodicReflectionConfigurationModel](ctx),
				Type:                    fwtypes.StringEnumValue(awstypes.MemoryStrategyTypeUserPreference),
			},
			expected: awstypes.ModifyMemoryStrategyInput{
				Description:        aws.String("description_001"),
				NamespaceTemplates: []string{"/strategies/{memoryStrategyId}/actors/{actorId}/"},
			},
		},
		{
			name: "basic episodic",
			model: tfbedrockagentcore.MemoryStrategyResourceModel{
				Configuration:          fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.CustomConfigurationModel](ctx),
				Description:            types.StringNull(),
				MemoryExecutionRoleARN: fwtypes.ARNNull(),
				MemoryID:               types.StringValue("memory_001"),
				Name:                   types.StringValue("episodic_builtin_001"),
				Namespaces:             fwtypes.NewSetValueOfNull[types.String](ctx),
				NamespaceTemplates:     fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}/"}),
				ReflectionConfiguration: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.EpisodicReflectionConfigurationModel{
					NamespaceTemplates: fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"/strategies/{memoryStrategyId}/actors/{actorId}/"}),
				}),
				Type: fwtypes.StringEnumValue(awstypes.MemoryStrategyTypeEpisodic),
			},
			expected: awstypes.ModifyMemoryStrategyInput{
				NamespaceTemplates: []string{"/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}/"},
			},
		},
		{
			name: "basic episodic with description",
			model: tfbedrockagentcore.MemoryStrategyResourceModel{
				Configuration:          fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.CustomConfigurationModel](ctx),
				Description:            types.StringValue("description_001"),
				MemoryExecutionRoleARN: fwtypes.ARNNull(),
				MemoryID:               types.StringValue("memory_001"),
				Name:                   types.StringValue("episodic_builtin_001"),
				Namespaces:             fwtypes.NewSetValueOfNull[types.String](ctx),
				NamespaceTemplates:     fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}/"}),
				ReflectionConfiguration: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.EpisodicReflectionConfigurationModel{
					NamespaceTemplates: fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"/strategies/{memoryStrategyId}/actors/{actorId}/"}),
				}),
				Type: fwtypes.StringEnumValue(awstypes.MemoryStrategyTypeEpisodic),
			},
			expected: awstypes.ModifyMemoryStrategyInput{
				Description:        aws.String("description_001"),
				NamespaceTemplates: []string{"/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}/"},
			},
		},
		{
			name: "summarization override",
			model: tfbedrockagentcore.MemoryStrategyResourceModel{
				Configuration: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.CustomConfigurationModel{
					Consolidation: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.OverrideDetailsModel{
						AppendToPrompt: types.StringValue("<task>Consolidate</task>"),
						ModelID:        types.StringValue("us.amazon.nova-2-lite-v1:0"),
					}),
					Extraction: fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.OverrideDetailsModel](ctx),
					Type:       fwtypes.StringEnumValue(awstypes.OverrideTypeSummaryOverride),
				}),
				Description:             types.StringNull(),
				MemoryExecutionRoleARN:  fwtypes.ARNNull(),
				MemoryID:                types.StringValue("memory_001"),
				Name:                    types.StringValue("summarization_override_001"),
				Namespaces:              fwtypes.NewSetValueOfNull[types.String](ctx),
				NamespaceTemplates:      fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}/"}),
				ReflectionConfiguration: fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.EpisodicReflectionConfigurationModel](ctx),
				Type:                    fwtypes.StringEnumValue(awstypes.MemoryStrategyTypeCustom),
			},
			expected: awstypes.ModifyMemoryStrategyInput{
				Configuration: &awstypes.ModifyStrategyConfiguration{
					Consolidation: &awstypes.ModifyConsolidationConfigurationMemberCustomConsolidationConfiguration{
						Value: &awstypes.CustomConsolidationConfigurationInputMemberSummaryConsolidationOverride{
							Value: awstypes.SummaryOverrideConsolidationConfigurationInput{
								AppendToPrompt: aws.String("<task>Consolidate</task>"),
								ModelId:        aws.String("us.amazon.nova-2-lite-v1:0"),
							},
						},
					},
				},
				NamespaceTemplates: []string{"/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}/"},
			},
		},
		{
			name: "semantic override",
			model: tfbedrockagentcore.MemoryStrategyResourceModel{
				Configuration: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.CustomConfigurationModel{
					Consolidation: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.OverrideDetailsModel{
						AppendToPrompt: types.StringValue("<task>Consolidate</task>"),
						ModelID:        types.StringValue("us.amazon.nova-2-lite-v1:0"),
					}),
					Extraction: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.OverrideDetailsModel{
						AppendToPrompt: types.StringValue("<task>Extract</task>"),
						ModelID:        types.StringValue("amazon.nova-lite-v1:0"),
					}),
					Type: fwtypes.StringEnumValue(awstypes.OverrideTypeSemanticOverride),
				}),
				Description:             types.StringNull(),
				MemoryExecutionRoleARN:  fwtypes.ARNNull(),
				MemoryID:                types.StringValue("memory_001"),
				Name:                    types.StringValue("semantic_override_001"),
				Namespaces:              fwtypes.NewSetValueOfNull[types.String](ctx),
				NamespaceTemplates:      fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"/strategies/{memoryStrategyId}/actors/{actorId}/"}),
				ReflectionConfiguration: fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.EpisodicReflectionConfigurationModel](ctx),
				Type:                    fwtypes.StringEnumValue(awstypes.MemoryStrategyTypeCustom),
			},
			expected: awstypes.ModifyMemoryStrategyInput{
				Configuration: &awstypes.ModifyStrategyConfiguration{
					Consolidation: &awstypes.ModifyConsolidationConfigurationMemberCustomConsolidationConfiguration{
						Value: &awstypes.CustomConsolidationConfigurationInputMemberSemanticConsolidationOverride{
							Value: awstypes.SemanticOverrideConsolidationConfigurationInput{
								AppendToPrompt: aws.String("<task>Consolidate</task>"),
								ModelId:        aws.String("us.amazon.nova-2-lite-v1:0"),
							},
						},
					},
					Extraction: &awstypes.ModifyExtractionConfigurationMemberCustomExtractionConfiguration{
						Value: &awstypes.CustomExtractionConfigurationInputMemberSemanticExtractionOverride{
							Value: awstypes.SemanticOverrideExtractionConfigurationInput{
								AppendToPrompt: aws.String("<task>Extract</task>"),
								ModelId:        aws.String("amazon.nova-lite-v1:0"),
							},
						},
					},
				},
				NamespaceTemplates: []string{"/strategies/{memoryStrategyId}/actors/{actorId}/"},
			},
		},
		{
			name: "user preference override",
			model: tfbedrockagentcore.MemoryStrategyResourceModel{
				Configuration: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.CustomConfigurationModel{
					Consolidation: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.OverrideDetailsModel{
						AppendToPrompt: types.StringValue("<task>Consolidate</task>"),
						ModelID:        types.StringValue("us.amazon.nova-2-lite-v1:0"),
					}),
					Extraction: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.OverrideDetailsModel{
						AppendToPrompt: types.StringValue("<task>Extract</task>"),
						ModelID:        types.StringValue("amazon.nova-lite-v1:0"),
					}),
					Type: fwtypes.StringEnumValue(awstypes.OverrideTypeUserPreferenceOverride),
				}),
				Description:             types.StringNull(),
				MemoryExecutionRoleARN:  fwtypes.ARNNull(),
				MemoryID:                types.StringValue("memory_001"),
				Name:                    types.StringValue("user_preference_override_001"),
				Namespaces:              fwtypes.NewSetValueOfNull[types.String](ctx),
				NamespaceTemplates:      fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"/strategies/{memoryStrategyId}/actors/{actorId}/"}),
				ReflectionConfiguration: fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.EpisodicReflectionConfigurationModel](ctx),
				Type:                    fwtypes.StringEnumValue(awstypes.MemoryStrategyTypeCustom),
			},
			expected: awstypes.ModifyMemoryStrategyInput{
				Configuration: &awstypes.ModifyStrategyConfiguration{
					Consolidation: &awstypes.ModifyConsolidationConfigurationMemberCustomConsolidationConfiguration{
						Value: &awstypes.CustomConsolidationConfigurationInputMemberUserPreferenceConsolidationOverride{
							Value: awstypes.UserPreferenceOverrideConsolidationConfigurationInput{
								AppendToPrompt: aws.String("<task>Consolidate</task>"),
								ModelId:        aws.String("us.amazon.nova-2-lite-v1:0"),
							},
						},
					},
					Extraction: &awstypes.ModifyExtractionConfigurationMemberCustomExtractionConfiguration{
						Value: &awstypes.CustomExtractionConfigurationInputMemberUserPreferenceExtractionOverride{
							Value: awstypes.UserPreferenceOverrideExtractionConfigurationInput{
								AppendToPrompt: aws.String("<task>Extract</task>"),
								ModelId:        aws.String("amazon.nova-lite-v1:0"),
							},
						},
					},
				},
				NamespaceTemplates: []string{"/strategies/{memoryStrategyId}/actors/{actorId}/"},
			},
		},
		{
			name: "episodic override",
			model: tfbedrockagentcore.MemoryStrategyResourceModel{
				Configuration: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.CustomConfigurationModel{
					Consolidation: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.OverrideDetailsModel{
						AppendToPrompt: types.StringValue("<task>Consolidate</task>"),
						ModelID:        types.StringValue("us.amazon.nova-2-lite-v1:0"),
					}),
					Extraction: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.OverrideDetailsModel{
						AppendToPrompt: types.StringValue("<task>Extract</task>"),
						ModelID:        types.StringValue("amazon.nova-lite-v1:0"),
					}),
					Type: fwtypes.StringEnumValue(awstypes.OverrideTypeEpisodicOverride),
				}),
				Description:            types.StringNull(),
				MemoryExecutionRoleARN: fwtypes.ARNNull(),
				MemoryID:               types.StringValue("memory_001"),
				Name:                   types.StringValue("episodic_override_001"),
				Namespaces:             fwtypes.NewSetValueOfNull[types.String](ctx),
				NamespaceTemplates:     fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}/"}),
				ReflectionConfiguration: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.EpisodicReflectionConfigurationModel{
					NamespaceTemplates: fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"/strategies/{memoryStrategyId}/actors/{actorId}/"}),
				}),
				Type: fwtypes.StringEnumValue(awstypes.MemoryStrategyTypeCustom),
			},
			expected: awstypes.ModifyMemoryStrategyInput{
				Configuration: &awstypes.ModifyStrategyConfiguration{
					Consolidation: &awstypes.ModifyConsolidationConfigurationMemberCustomConsolidationConfiguration{
						Value: &awstypes.CustomConsolidationConfigurationInputMemberEpisodicConsolidationOverride{
							Value: awstypes.EpisodicOverrideConsolidationConfigurationInput{
								AppendToPrompt: aws.String("<task>Consolidate</task>"),
								ModelId:        aws.String("us.amazon.nova-2-lite-v1:0"),
							},
						},
					},
					Extraction: &awstypes.ModifyExtractionConfigurationMemberCustomExtractionConfiguration{
						Value: &awstypes.CustomExtractionConfigurationInputMemberEpisodicExtractionOverride{
							Value: awstypes.EpisodicOverrideExtractionConfigurationInput{
								AppendToPrompt: aws.String("<task>Extract</task>"),
								ModelId:        aws.String("amazon.nova-lite-v1:0"),
							},
						},
					},
				},
				NamespaceTemplates: []string{"/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}/"},
			},
		},
	}

	opts := cmp.Options{
		cmpopts.IgnoreUnexported(
			awstypes.CustomConsolidationConfigurationInputMemberEpisodicConsolidationOverride{},
			awstypes.CustomConsolidationConfigurationInputMemberSemanticConsolidationOverride{},
			awstypes.CustomConsolidationConfigurationInputMemberSummaryConsolidationOverride{},
			awstypes.CustomConsolidationConfigurationInputMemberUserPreferenceConsolidationOverride{},
			awstypes.CustomExtractionConfigurationInputMemberEpisodicExtractionOverride{},
			awstypes.CustomExtractionConfigurationInputMemberSemanticExtractionOverride{},
			awstypes.CustomExtractionConfigurationInputMemberUserPreferenceExtractionOverride{},
			awstypes.EpisodicOverrideConsolidationConfigurationInput{},
			awstypes.EpisodicOverrideExtractionConfigurationInput{},
			awstypes.ModifyConsolidationConfigurationMemberCustomConsolidationConfiguration{},
			awstypes.ModifyExtractionConfigurationMemberCustomExtractionConfiguration{},
			awstypes.ModifyMemoryStrategyInput{},
			awstypes.ModifyStrategyConfiguration{},
			awstypes.SemanticOverrideConsolidationConfigurationInput{},
			awstypes.SemanticOverrideExtractionConfigurationInput{},
			awstypes.SummaryOverrideConsolidationConfigurationInput{},
			awstypes.UserPreferenceOverrideConsolidationConfigurationInput{},
			awstypes.UserPreferenceOverrideExtractionConfigurationInput{},
		),
		cmpopts.SortSlices(func(a, b string) bool { return a < b }),
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var out awstypes.ModifyMemoryStrategyInput
			if diags := fwflex.Expand(ctx, tc.model, &out); diags.HasError() {
				t.Fatalf("Expand: %v", diags)
			}

			if diff := cmp.Diff(tc.expected, out, opts...); diff != "" {
				t.Errorf("model -> ModifyMemoryStrategyInput mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMemoryStrategyResourceModelFlatten(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	testCases := []struct {
		name     string
		input    awstypes.MemoryStrategy
		expected tfbedrockagentcore.MemoryStrategyResourceModel
	}{
		{
			name: "basic summarization",
			input: awstypes.MemoryStrategy{
				Name:               aws.String("summarization_builtin_001"),
				NamespaceTemplates: []string{"/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}/"},
				StrategyId:         aws.String("summarization_builtin_001-YJnphM84r5"),
				Type:               awstypes.MemoryStrategyTypeSummarization,
			},
			expected: tfbedrockagentcore.MemoryStrategyResourceModel{
				Configuration:           fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.CustomConfigurationModel](ctx),
				Description:             types.StringNull(),
				MemoryExecutionRoleARN:  fwtypes.ARNNull(),
				MemoryID:                types.StringNull(),
				MemoryStrategyID:        types.StringValue("summarization_builtin_001-YJnphM84r5"),
				Name:                    types.StringValue("summarization_builtin_001"),
				Namespaces:              fwtypes.NewSetValueOfNull[types.String](ctx),
				NamespaceTemplates:      fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}/"}),
				ReflectionConfiguration: fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.EpisodicReflectionConfigurationModel](ctx),
				Type:                    fwtypes.StringEnumValue(awstypes.MemoryStrategyTypeSummarization),
			},
		},
		{
			name: "basic summarization with description",
			input: awstypes.MemoryStrategy{
				Description:        aws.String("description_001"),
				Name:               aws.String("summarization_builtin_001"),
				NamespaceTemplates: []string{"/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}/"},
				StrategyId:         aws.String("summarization_builtin_001-YJnphM84r5"),
				Type:               awstypes.MemoryStrategyTypeSummarization,
			},
			expected: tfbedrockagentcore.MemoryStrategyResourceModel{
				Configuration:           fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.CustomConfigurationModel](ctx),
				Description:             types.StringValue("description_001"),
				MemoryExecutionRoleARN:  fwtypes.ARNNull(),
				MemoryID:                types.StringNull(),
				MemoryStrategyID:        types.StringValue("summarization_builtin_001-YJnphM84r5"),
				Name:                    types.StringValue("summarization_builtin_001"),
				Namespaces:              fwtypes.NewSetValueOfNull[types.String](ctx),
				NamespaceTemplates:      fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}/"}),
				ReflectionConfiguration: fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.EpisodicReflectionConfigurationModel](ctx),
				Type:                    fwtypes.StringEnumValue(awstypes.MemoryStrategyTypeSummarization),
			},
		},
		{
			name: "basic semantic",
			input: awstypes.MemoryStrategy{
				Name:               aws.String("semantic_builtin_001"),
				NamespaceTemplates: []string{"/strategies/{memoryStrategyId}/actors/{actorId}/"},
				StrategyId:         aws.String("semantic_builtin_001-YJnphM84r5"),
				Type:               awstypes.MemoryStrategyTypeSemantic,
			},
			expected: tfbedrockagentcore.MemoryStrategyResourceModel{
				Configuration:           fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.CustomConfigurationModel](ctx),
				Description:             types.StringNull(),
				MemoryExecutionRoleARN:  fwtypes.ARNNull(),
				MemoryID:                types.StringNull(),
				MemoryStrategyID:        types.StringValue("semantic_builtin_001-YJnphM84r5"),
				Name:                    types.StringValue("semantic_builtin_001"),
				Namespaces:              fwtypes.NewSetValueOfNull[types.String](ctx),
				NamespaceTemplates:      fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"/strategies/{memoryStrategyId}/actors/{actorId}/"}),
				ReflectionConfiguration: fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.EpisodicReflectionConfigurationModel](ctx),
				Type:                    fwtypes.StringEnumValue(awstypes.MemoryStrategyTypeSemantic),
			},
		},
		{
			name: "basic user preference",
			input: awstypes.MemoryStrategy{
				Name:               aws.String("user_preference_builtin_001"),
				NamespaceTemplates: []string{"/strategies/{memoryStrategyId}/actors/{actorId}/"},
				StrategyId:         aws.String("user_preference_builtin_001-YJnphM84r5"),
				Type:               awstypes.MemoryStrategyTypeUserPreference,
			},
			expected: tfbedrockagentcore.MemoryStrategyResourceModel{
				Configuration:           fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.CustomConfigurationModel](ctx),
				Description:             types.StringNull(),
				MemoryExecutionRoleARN:  fwtypes.ARNNull(),
				MemoryID:                types.StringNull(),
				MemoryStrategyID:        types.StringValue("user_preference_builtin_001-YJnphM84r5"),
				Name:                    types.StringValue("user_preference_builtin_001"),
				Namespaces:              fwtypes.NewSetValueOfNull[types.String](ctx),
				NamespaceTemplates:      fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"/strategies/{memoryStrategyId}/actors/{actorId}/"}),
				ReflectionConfiguration: fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.EpisodicReflectionConfigurationModel](ctx),
				Type:                    fwtypes.StringEnumValue(awstypes.MemoryStrategyTypeUserPreference),
			},
		},
		{
			name: "basic user episodic",
			input: awstypes.MemoryStrategy{
				Configuration: &awstypes.StrategyConfiguration{
					Type: "EPISODIC", // Yes, really.
					Reflection: &awstypes.ReflectionConfigurationMemberEpisodicReflectionConfiguration{
						Value: awstypes.EpisodicReflectionConfiguration{
							NamespaceTemplates: []string{"/strategies/{memoryStrategyId}/actors/{actorId}/"},
						},
					},
				},
				Name:               aws.String("episodic_builtin_001"),
				NamespaceTemplates: []string{"/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}/"},
				StrategyId:         aws.String("episodic_builtin_001-Qw47TlFGX5"),
				Type:               awstypes.MemoryStrategyTypeEpisodic,
			},
			expected: tfbedrockagentcore.MemoryStrategyResourceModel{
				Configuration:           fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.CustomConfigurationModel](ctx),
				Description:             types.StringNull(),
				MemoryExecutionRoleARN:  fwtypes.ARNNull(),
				MemoryID:                types.StringNull(),
				MemoryStrategyID:        types.StringValue("episodic_builtin_001-Qw47TlFGX5"),
				Name:                    types.StringValue("episodic_builtin_001"),
				Namespaces:              fwtypes.NewSetValueOfNull[types.String](ctx),
				NamespaceTemplates:      fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}/"}),
				ReflectionConfiguration: fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.EpisodicReflectionConfigurationModel](ctx), // TODO
				Type:                    fwtypes.StringEnumValue(awstypes.MemoryStrategyTypeEpisodic),
			},
		},
		{
			name: "summarization override",
			input: awstypes.MemoryStrategy{
				Configuration: &awstypes.StrategyConfiguration{
					Consolidation: &awstypes.ConsolidationConfigurationMemberCustomConsolidationConfiguration{
						Value: &awstypes.CustomConsolidationConfigurationMemberSummaryConsolidationOverride{
							Value: awstypes.SummaryConsolidationOverride{
								AppendToPrompt: aws.String("<task>Consolidate</task>"),
								ModelId:        aws.String("us.amazon.nova-2-lite-v1:0"),
							},
						},
					},
					Type: awstypes.OverrideTypeSummaryOverride,
				},
				Name:               aws.String("summarization_override_001"),
				NamespaceTemplates: []string{"/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}/"},
				StrategyId:         aws.String("summarization_override_001-XJf3fg7IP1"),
				Type:               awstypes.MemoryStrategyTypeCustom,
			},
			expected: tfbedrockagentcore.MemoryStrategyResourceModel{
				Configuration: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.CustomConfigurationModel{
					Consolidation: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.OverrideDetailsModel{
						AppendToPrompt: types.StringValue("<task>Consolidate</task>"),
						ModelID:        types.StringValue("us.amazon.nova-2-lite-v1:0"),
					}),
					Extraction: fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.OverrideDetailsModel](ctx),
					Type:       fwtypes.StringEnumValue(awstypes.OverrideTypeSummaryOverride),
				}),
				Description:             types.StringNull(),
				MemoryExecutionRoleARN:  fwtypes.ARNNull(),
				MemoryID:                types.StringNull(),
				MemoryStrategyID:        types.StringValue("summarization_override_001-XJf3fg7IP1"),
				Name:                    types.StringValue("summarization_override_001"),
				Namespaces:              fwtypes.NewSetValueOfNull[types.String](ctx),
				NamespaceTemplates:      fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}/"}),
				ReflectionConfiguration: fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.EpisodicReflectionConfigurationModel](ctx),
				Type:                    fwtypes.StringEnumValue(awstypes.MemoryStrategyTypeCustom),
			},
		},
		{
			name: "semantic override",
			input: awstypes.MemoryStrategy{
				Configuration: &awstypes.StrategyConfiguration{
					Consolidation: &awstypes.ConsolidationConfigurationMemberCustomConsolidationConfiguration{
						Value: &awstypes.CustomConsolidationConfigurationMemberSemanticConsolidationOverride{
							Value: awstypes.SemanticConsolidationOverride{
								AppendToPrompt: aws.String("<task>Consolidate</task>"),
								ModelId:        aws.String("us.amazon.nova-2-lite-v1:0"),
							},
						},
					},
					Extraction: &awstypes.ExtractionConfigurationMemberCustomExtractionConfiguration{
						Value: &awstypes.CustomExtractionConfigurationMemberSemanticExtractionOverride{
							Value: awstypes.SemanticExtractionOverride{
								AppendToPrompt: aws.String("<task>Extract</task>"),
								ModelId:        aws.String("amazon.nova-lite-v1:0"),
							},
						},
					},
					Type: awstypes.OverrideTypeSemanticOverride,
				},
				Name:               aws.String("semantic_override_001"),
				NamespaceTemplates: []string{"/strategies/{memoryStrategyId}/actors/{actorId}/"},
				StrategyId:         aws.String("semantic_override_001-XJf3fg7IP1"),
				Type:               awstypes.MemoryStrategyTypeCustom,
			},
			expected: tfbedrockagentcore.MemoryStrategyResourceModel{
				Configuration: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.CustomConfigurationModel{
					Consolidation: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.OverrideDetailsModel{
						AppendToPrompt: types.StringValue("<task>Consolidate</task>"),
						ModelID:        types.StringValue("us.amazon.nova-2-lite-v1:0"),
					}),
					Extraction: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.OverrideDetailsModel{
						AppendToPrompt: types.StringValue("<task>Extract</task>"),
						ModelID:        types.StringValue("amazon.nova-lite-v1:0"),
					}),
					Type: fwtypes.StringEnumValue(awstypes.OverrideTypeSemanticOverride),
				}),
				Description:             types.StringNull(),
				MemoryExecutionRoleARN:  fwtypes.ARNNull(),
				MemoryID:                types.StringNull(),
				MemoryStrategyID:        types.StringValue("semantic_override_001-XJf3fg7IP1"),
				Name:                    types.StringValue("semantic_override_001"),
				Namespaces:              fwtypes.NewSetValueOfNull[types.String](ctx),
				NamespaceTemplates:      fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"/strategies/{memoryStrategyId}/actors/{actorId}/"}),
				ReflectionConfiguration: fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.EpisodicReflectionConfigurationModel](ctx),
				Type:                    fwtypes.StringEnumValue(awstypes.MemoryStrategyTypeCustom),
			},
		},
		{
			name: "user preference override",
			input: awstypes.MemoryStrategy{
				Configuration: &awstypes.StrategyConfiguration{
					Consolidation: &awstypes.ConsolidationConfigurationMemberCustomConsolidationConfiguration{
						Value: &awstypes.CustomConsolidationConfigurationMemberUserPreferenceConsolidationOverride{
							Value: awstypes.UserPreferenceConsolidationOverride{
								AppendToPrompt: aws.String("<task>Consolidate</task>"),
								ModelId:        aws.String("us.amazon.nova-2-lite-v1:0"),
							},
						},
					},
					Extraction: &awstypes.ExtractionConfigurationMemberCustomExtractionConfiguration{
						Value: &awstypes.CustomExtractionConfigurationMemberUserPreferenceExtractionOverride{
							Value: awstypes.UserPreferenceExtractionOverride{
								AppendToPrompt: aws.String("<task>Extract</task>"),
								ModelId:        aws.String("amazon.nova-lite-v1:0"),
							},
						},
					},
					Type: awstypes.OverrideTypeUserPreferenceOverride,
				},
				Name:               aws.String("user_preference_override_001"),
				NamespaceTemplates: []string{"/strategies/{memoryStrategyId}/actors/{actorId}/"},
				StrategyId:         aws.String("user_preference_override_001-XJf3fg7IP1"),
				Type:               awstypes.MemoryStrategyTypeCustom,
			},
			expected: tfbedrockagentcore.MemoryStrategyResourceModel{
				Configuration: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.CustomConfigurationModel{
					Consolidation: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.OverrideDetailsModel{
						AppendToPrompt: types.StringValue("<task>Consolidate</task>"),
						ModelID:        types.StringValue("us.amazon.nova-2-lite-v1:0"),
					}),
					Extraction: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.OverrideDetailsModel{
						AppendToPrompt: types.StringValue("<task>Extract</task>"),
						ModelID:        types.StringValue("amazon.nova-lite-v1:0"),
					}),
					Type: fwtypes.StringEnumValue(awstypes.OverrideTypeUserPreferenceOverride),
				}),
				Description:             types.StringNull(),
				MemoryExecutionRoleARN:  fwtypes.ARNNull(),
				MemoryID:                types.StringNull(),
				MemoryStrategyID:        types.StringValue("user_preference_override_001-XJf3fg7IP1"),
				Name:                    types.StringValue("user_preference_override_001"),
				Namespaces:              fwtypes.NewSetValueOfNull[types.String](ctx),
				NamespaceTemplates:      fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"/strategies/{memoryStrategyId}/actors/{actorId}/"}),
				ReflectionConfiguration: fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.EpisodicReflectionConfigurationModel](ctx),
				Type:                    fwtypes.StringEnumValue(awstypes.MemoryStrategyTypeCustom),
			},
		},
		{
			name: "episodic override",
			input: awstypes.MemoryStrategy{
				Configuration: &awstypes.StrategyConfiguration{
					Consolidation: &awstypes.ConsolidationConfigurationMemberCustomConsolidationConfiguration{
						Value: &awstypes.CustomConsolidationConfigurationMemberEpisodicConsolidationOverride{
							Value: awstypes.EpisodicConsolidationOverride{
								AppendToPrompt: aws.String("<task>Consolidate</task>"),
								ModelId:        aws.String("us.amazon.nova-2-lite-v1:0"),
							},
						},
					},
					Extraction: &awstypes.ExtractionConfigurationMemberCustomExtractionConfiguration{
						Value: &awstypes.CustomExtractionConfigurationMemberEpisodicExtractionOverride{
							Value: awstypes.EpisodicExtractionOverride{
								AppendToPrompt: aws.String("<task>Extract</task>"),
								ModelId:        aws.String("amazon.nova-lite-v1:0"),
							},
						},
					},
					Reflection: &awstypes.ReflectionConfigurationMemberCustomReflectionConfiguration{
						Value: &awstypes.CustomReflectionConfigurationMemberEpisodicReflectionOverride{
							Value: awstypes.EpisodicReflectionOverride{
								NamespaceTemplates: []string{"/strategies/{memoryStrategyId}/actors/{actorId}/"},
							},
						},
					},
					Type: awstypes.OverrideTypeEpisodicOverride,
				},
				Name:               aws.String("episodic_override_001"),
				NamespaceTemplates: []string{"/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}/"},
				StrategyId:         aws.String("episodic_override_001-XJf3fg7IP1"),
				Type:               awstypes.MemoryStrategyTypeCustom,
			},
			expected: tfbedrockagentcore.MemoryStrategyResourceModel{
				Configuration: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.CustomConfigurationModel{
					Consolidation: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.OverrideDetailsModel{
						AppendToPrompt: types.StringValue("<task>Consolidate</task>"),
						ModelID:        types.StringValue("us.amazon.nova-2-lite-v1:0"),
					}),
					Extraction: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.OverrideDetailsModel{
						AppendToPrompt: types.StringValue("<task>Extract</task>"),
						ModelID:        types.StringValue("amazon.nova-lite-v1:0"),
					}),
					Type: fwtypes.StringEnumValue(awstypes.OverrideTypeEpisodicOverride),
				}),
				Description:             types.StringNull(),
				MemoryExecutionRoleARN:  fwtypes.ARNNull(),
				MemoryID:                types.StringNull(),
				MemoryStrategyID:        types.StringValue("episodic_override_001-XJf3fg7IP1"),
				Name:                    types.StringValue("episodic_override_001"),
				Namespaces:              fwtypes.NewSetValueOfNull[types.String](ctx),
				NamespaceTemplates:      fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}/"}),
				ReflectionConfiguration: fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.EpisodicReflectionConfigurationModel](ctx),
				Type:                    fwtypes.StringEnumValue(awstypes.MemoryStrategyTypeCustom),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var model tfbedrockagentcore.MemoryStrategyResourceModel
			if diags := fwflex.Flatten(ctx, tc.input, &model, fwflex.WithFieldNamePrefix("Memory")); diags.HasError() {
				t.Fatalf("Flatten: %v", diags)
			}

			if diff := cmp.Diff(tc.expected, model); diff != "" {
				t.Errorf("MemoryStrategy -> model mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
