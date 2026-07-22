// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrockagentcore "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagentcore"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestMemoryStrategyResourceModelExpandOnCreate(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	testCases := []struct {
		name     string
		model    tfbedrockagentcore.MemoryStrategyResourceModel
		expected awstypes.MemoryStrategyInput
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
			expected: &awstypes.MemoryStrategyInputMemberSummaryMemoryStrategy{
				Value: awstypes.SummaryMemoryStrategyInput{
					Name:               aws.String("summarization_builtin_001"),
					NamespaceTemplates: []string{"/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}/"},
				},
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
			expected: &awstypes.MemoryStrategyInputMemberSummaryMemoryStrategy{
				Value: awstypes.SummaryMemoryStrategyInput{
					Description:        aws.String("description_001"),
					Name:               aws.String("summarization_builtin_001"),
					NamespaceTemplates: []string{"/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}/"},
				},
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
			expected: &awstypes.MemoryStrategyInputMemberSemanticMemoryStrategy{
				Value: awstypes.SemanticMemoryStrategyInput{
					Name:               aws.String("semantic_builtin_001"),
					NamespaceTemplates: []string{"/strategies/{memoryStrategyId}/actors/{actorId}/"},
				},
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
			expected: &awstypes.MemoryStrategyInputMemberUserPreferenceMemoryStrategy{
				Value: awstypes.UserPreferenceMemoryStrategyInput{
					Name:               aws.String("user_preference_builtin_001"),
					NamespaceTemplates: []string{"/strategies/{memoryStrategyId}/actors/{actorId}/"},
				},
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
			expected: &awstypes.MemoryStrategyInputMemberEpisodicMemoryStrategy{
				Value: awstypes.EpisodicMemoryStrategyInput{
					Name:               aws.String("episodic_builtin_001"),
					NamespaceTemplates: []string{"/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}/"},
					ReflectionConfiguration: &awstypes.EpisodicReflectionConfigurationInput{
						NamespaceTemplates: []string{"/strategies/{memoryStrategyId}/actors/{actorId}/"},
					},
				},
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
					Reflection: fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.EpisodicReflectionOverrideDetailsModel](ctx),
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
			expected: &awstypes.MemoryStrategyInputMemberCustomMemoryStrategy{
				Value: awstypes.CustomMemoryStrategyInput{
					Configuration: &awstypes.CustomConfigurationInputMemberSummaryOverride{
						Value: awstypes.SummaryOverrideConfigurationInput{
							Consolidation: &awstypes.SummaryOverrideConsolidationConfigurationInput{
								AppendToPrompt: aws.String("<task>Consolidate</task>"),
								ModelId:        aws.String("us.amazon.nova-2-lite-v1:0"),
							},
						},
					},
					Name:               aws.String("summarization_override_001"),
					NamespaceTemplates: []string{"/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}/"},
				},
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
					Reflection: fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.EpisodicReflectionOverrideDetailsModel](ctx),
					Type:       fwtypes.StringEnumValue(awstypes.OverrideTypeSemanticOverride),
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
			expected: &awstypes.MemoryStrategyInputMemberCustomMemoryStrategy{
				Value: awstypes.CustomMemoryStrategyInput{
					Configuration: &awstypes.CustomConfigurationInputMemberSemanticOverride{
						Value: awstypes.SemanticOverrideConfigurationInput{
							Consolidation: &awstypes.SemanticOverrideConsolidationConfigurationInput{
								AppendToPrompt: aws.String("<task>Consolidate</task>"),
								ModelId:        aws.String("us.amazon.nova-2-lite-v1:0"),
							},
							Extraction: &awstypes.SemanticOverrideExtractionConfigurationInput{
								AppendToPrompt: aws.String("<task>Extract</task>"),
								ModelId:        aws.String("amazon.nova-lite-v1:0"),
							},
						},
					},
					Name:               aws.String("semantic_override_001"),
					NamespaceTemplates: []string{"/strategies/{memoryStrategyId}/actors/{actorId}/"},
				},
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
					Reflection: fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.EpisodicReflectionOverrideDetailsModel](ctx),
					Type:       fwtypes.StringEnumValue(awstypes.OverrideTypeUserPreferenceOverride),
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
			expected: &awstypes.MemoryStrategyInputMemberCustomMemoryStrategy{
				Value: awstypes.CustomMemoryStrategyInput{
					Configuration: &awstypes.CustomConfigurationInputMemberUserPreferenceOverride{
						Value: awstypes.UserPreferenceOverrideConfigurationInput{
							Consolidation: &awstypes.UserPreferenceOverrideConsolidationConfigurationInput{
								AppendToPrompt: aws.String("<task>Consolidate</task>"),
								ModelId:        aws.String("us.amazon.nova-2-lite-v1:0"),
							},
							Extraction: &awstypes.UserPreferenceOverrideExtractionConfigurationInput{
								AppendToPrompt: aws.String("<task>Extract</task>"),
								ModelId:        aws.String("amazon.nova-lite-v1:0"),
							},
						},
					},
					Name:               aws.String("user_preference_override_001"),
					NamespaceTemplates: []string{"/strategies/{memoryStrategyId}/actors/{actorId}/"},
				},
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
					Reflection: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.EpisodicReflectionOverrideDetailsModel{
						AppendToPrompt:     types.StringValue("<task>Reflect</task>"),
						ModelID:            types.StringValue("us.amazon.nova-micro-v1:0"),
						NamespaceTemplates: fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"/strategies/{memoryStrategyId}/actors/{actorId}/"}),
					}),
					Type: fwtypes.StringEnumValue(awstypes.OverrideTypeEpisodicOverride),
				}),
				Description:             types.StringNull(),
				MemoryExecutionRoleARN:  fwtypes.ARNNull(),
				MemoryID:                types.StringValue("memory_001"),
				Name:                    types.StringValue("episodic_override_001"),
				Namespaces:              fwtypes.NewSetValueOfNull[types.String](ctx),
				NamespaceTemplates:      fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}/"}),
				ReflectionConfiguration: fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.EpisodicReflectionConfigurationModel](ctx),
				Type:                    fwtypes.StringEnumValue(awstypes.MemoryStrategyTypeCustom),
			},
			expected: &awstypes.MemoryStrategyInputMemberCustomMemoryStrategy{
				Value: awstypes.CustomMemoryStrategyInput{
					Configuration: &awstypes.CustomConfigurationInputMemberEpisodicOverride{
						Value: awstypes.EpisodicOverrideConfigurationInput{
							Consolidation: &awstypes.EpisodicOverrideConsolidationConfigurationInput{
								AppendToPrompt: aws.String("<task>Consolidate</task>"),
								ModelId:        aws.String("us.amazon.nova-2-lite-v1:0"),
							},
							Extraction: &awstypes.EpisodicOverrideExtractionConfigurationInput{
								AppendToPrompt: aws.String("<task>Extract</task>"),
								ModelId:        aws.String("amazon.nova-lite-v1:0"),
							},
							Reflection: &awstypes.EpisodicOverrideReflectionConfigurationInput{
								AppendToPrompt:     aws.String("<task>Reflect</task>"),
								ModelId:            aws.String("us.amazon.nova-micro-v1:0"),
								NamespaceTemplates: []string{"/strategies/{memoryStrategyId}/actors/{actorId}/"},
							},
						},
					},
					Name:               aws.String("episodic_override_001"),
					NamespaceTemplates: []string{"/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}/"},
				},
			},
		},
	}

	opts := cmp.Options{
		cmpopts.IgnoreUnexported(
			awstypes.CustomConfigurationInputMemberEpisodicOverride{},
			awstypes.CustomConfigurationInputMemberSemanticOverride{},
			awstypes.CustomConfigurationInputMemberSummaryOverride{},
			awstypes.CustomConfigurationInputMemberUserPreferenceOverride{},
			awstypes.CustomMemoryStrategyInput{},
			awstypes.EpisodicMemoryStrategyInput{},
			awstypes.EpisodicOverrideConfigurationInput{},
			awstypes.EpisodicOverrideConsolidationConfigurationInput{},
			awstypes.EpisodicOverrideExtractionConfigurationInput{},
			awstypes.EpisodicOverrideReflectionConfigurationInput{},
			awstypes.EpisodicReflectionConfigurationInput{},
			awstypes.MemoryStrategyInputMemberCustomMemoryStrategy{},
			awstypes.MemoryStrategyInputMemberEpisodicMemoryStrategy{},
			awstypes.MemoryStrategyInputMemberSemanticMemoryStrategy{},
			awstypes.MemoryStrategyInputMemberSummaryMemoryStrategy{},
			awstypes.MemoryStrategyInputMemberUserPreferenceMemoryStrategy{},
			awstypes.SemanticMemoryStrategyInput{},
			awstypes.SummaryMemoryStrategyInput{},
			awstypes.SemanticOverrideConfigurationInput{},
			awstypes.SemanticOverrideConsolidationConfigurationInput{},
			awstypes.SemanticOverrideExtractionConfigurationInput{},
			awstypes.SummaryOverrideConfigurationInput{},
			awstypes.SummaryOverrideConsolidationConfigurationInput{},
			awstypes.UserPreferenceMemoryStrategyInput{},
			awstypes.UserPreferenceOverrideConfigurationInput{},
			awstypes.UserPreferenceOverrideConsolidationConfigurationInput{},
			awstypes.UserPreferenceOverrideExtractionConfigurationInput{},
		),
		cmpopts.SortSlices(func(a, b string) bool { return a < b }),
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var out awstypes.MemoryStrategyInput
			if diags := fwflex.Expand(ctx, tc.model, &out); diags.HasError() {
				t.Fatalf("Expand: %v", diags)
			}

			if diff := cmp.Diff(tc.expected, out, opts...); diff != "" {
				t.Errorf("model -> ModifyMemoryStrategyInput mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMemoryStrategyResourceModelExpandOnUpdate(t *testing.T) {
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
				Configuration: &awstypes.ModifyStrategyConfiguration{
					Reflection: &awstypes.ModifyReflectionConfigurationMemberEpisodicReflectionConfiguration{
						Value: awstypes.EpisodicReflectionConfigurationInput{
							NamespaceTemplates: []string{"/strategies/{memoryStrategyId}/actors/{actorId}/"},
						},
					},
				},
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
				Configuration: &awstypes.ModifyStrategyConfiguration{
					Reflection: &awstypes.ModifyReflectionConfigurationMemberEpisodicReflectionConfiguration{
						Value: awstypes.EpisodicReflectionConfigurationInput{
							NamespaceTemplates: []string{"/strategies/{memoryStrategyId}/actors/{actorId}/"},
						},
					},
				},
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
					Reflection: fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.EpisodicReflectionOverrideDetailsModel](ctx),
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
					Reflection: fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.EpisodicReflectionOverrideDetailsModel](ctx),
					Type:       fwtypes.StringEnumValue(awstypes.OverrideTypeSemanticOverride),
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
					Reflection: fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.EpisodicReflectionOverrideDetailsModel](ctx),
					Type:       fwtypes.StringEnumValue(awstypes.OverrideTypeUserPreferenceOverride),
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
					Reflection: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.EpisodicReflectionOverrideDetailsModel{
						AppendToPrompt:     types.StringValue("<task>Reflect</task>"),
						ModelID:            types.StringValue("us.amazon.nova-micro-v1:0"),
						NamespaceTemplates: fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"/strategies/{memoryStrategyId}/actors/{actorId}/"}),
					}),
					Type: fwtypes.StringEnumValue(awstypes.OverrideTypeEpisodicOverride),
				}),
				Description:             types.StringNull(),
				MemoryExecutionRoleARN:  fwtypes.ARNNull(),
				MemoryID:                types.StringValue("memory_001"),
				Name:                    types.StringValue("episodic_override_001"),
				Namespaces:              fwtypes.NewSetValueOfNull[types.String](ctx),
				NamespaceTemplates:      fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}/"}),
				ReflectionConfiguration: fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.EpisodicReflectionConfigurationModel](ctx),
				Type:                    fwtypes.StringEnumValue(awstypes.MemoryStrategyTypeCustom),
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
					Reflection: &awstypes.ModifyReflectionConfigurationMemberCustomReflectionConfiguration{
						Value: &awstypes.CustomReflectionConfigurationInputMemberEpisodicReflectionOverride{
							Value: awstypes.EpisodicOverrideReflectionConfigurationInput{
								AppendToPrompt:     aws.String("<task>Reflect</task>"),
								ModelId:            aws.String("us.amazon.nova-micro-v1:0"),
								NamespaceTemplates: []string{"/strategies/{memoryStrategyId}/actors/{actorId}/"},
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
			awstypes.CustomReflectionConfigurationInputMemberEpisodicReflectionOverride{},
			awstypes.EpisodicOverrideConsolidationConfigurationInput{},
			awstypes.EpisodicOverrideExtractionConfigurationInput{},
			awstypes.EpisodicReflectionConfigurationInput{},
			awstypes.EpisodicOverrideReflectionConfigurationInput{},
			awstypes.ModifyConsolidationConfigurationMemberCustomConsolidationConfiguration{},
			awstypes.ModifyExtractionConfigurationMemberCustomExtractionConfiguration{},
			awstypes.ModifyMemoryStrategyInput{},
			awstypes.ModifyReflectionConfigurationMemberCustomReflectionConfiguration{},
			awstypes.ModifyReflectionConfigurationMemberEpisodicReflectionConfiguration{},
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
				Configuration:          fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.CustomConfigurationModel](ctx),
				Description:            types.StringNull(),
				MemoryExecutionRoleARN: fwtypes.ARNNull(),
				MemoryID:               types.StringNull(),
				MemoryStrategyID:       types.StringValue("episodic_builtin_001-Qw47TlFGX5"),
				Name:                   types.StringValue("episodic_builtin_001"),
				Namespaces:             fwtypes.NewSetValueOfNull[types.String](ctx),
				NamespaceTemplates:     fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}/"}),
				ReflectionConfiguration: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.EpisodicReflectionConfigurationModel{
					NamespaceTemplates: fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"/strategies/{memoryStrategyId}/actors/{actorId}/"}),
				}),
				Type: fwtypes.StringEnumValue(awstypes.MemoryStrategyTypeEpisodic),
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
					Reflection: fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.EpisodicReflectionOverrideDetailsModel](ctx),
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
					Reflection: fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.EpisodicReflectionOverrideDetailsModel](ctx),
					Type:       fwtypes.StringEnumValue(awstypes.OverrideTypeSemanticOverride),
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
					Reflection: fwtypes.NewListNestedObjectValueOfNull[tfbedrockagentcore.EpisodicReflectionOverrideDetailsModel](ctx),
					Type:       fwtypes.StringEnumValue(awstypes.OverrideTypeUserPreferenceOverride),
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
								AppendToPrompt:     aws.String("<task>Reflect</task>"),
								ModelId:            aws.String("us.amazon.nova-micro-v1:0"),
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
					Reflection: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfbedrockagentcore.EpisodicReflectionOverrideDetailsModel{
						AppendToPrompt:     types.StringValue("<task>Reflect</task>"),
						ModelID:            types.StringValue("us.amazon.nova-micro-v1:0"),
						NamespaceTemplates: fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"/strategies/{memoryStrategyId}/actors/{actorId}/"}),
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

func TestAccBedrockAgentCoreMemoryStrategy_basic(t *testing.T) {
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
				Config: testAccMemoryStrategyConfig_basicNamespaceTemplates(rName, awstypes.MemoryStrategyTypeSemantic, "default"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_execution_role_arn"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_strategy_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespaces"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("default"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespace_templates"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("default"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(awstypes.MemoryStrategyTypeSemantic)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccMemoryStrategyImportStateIDFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "memory_strategy_id",
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
				Config: testAccMemoryStrategyConfig_basicNamespaceTemplates(rName, awstypes.MemoryStrategyTypeSemantic, "default"),
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

func TestAccBedrockAgentCoreMemoryStrategy_multipleStrategies(t *testing.T) {
	ctx := acctest.Context(t)
	var m1, m2 awstypes.MemoryStrategy
	rName := randomMemoryName(t)
	resourceName1 := "aws_bedrockagentcore_memory_strategy.test"
	resourceName2 := "aws_bedrockagentcore_memory_strategy.test2"

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
				// Two strategies of different types on the same memory: creating the
				// second must not fail with "too many results" against the first.
				Config: testAccMemoryStrategyConfig_multipleTypes(rName, "Summarization strategy"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName1, &m1),
					testAccCheckMemoryStrategyExists(ctx, t, resourceName2, &m2),
					resource.TestCheckResourceAttr(resourceName1, names.AttrType, string(awstypes.MemoryStrategyTypeUserPreference)),
					resource.TestCheckResourceAttr(resourceName2, names.AttrType, string(awstypes.MemoryStrategyTypeSummarization)),
				),
			},
			{
				// Updating one of the two strategies must not fail either, for the
				// same reason.
				Config: testAccMemoryStrategyConfig_multipleTypes(rName, "Summarization strategy updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName1, &m1),
					testAccCheckMemoryStrategyExists(ctx, t, resourceName2, &m2),
					resource.TestCheckResourceAttr(resourceName2, names.AttrDescription, "Summarization strategy updated"),
				),
			},
		},
	})
}

func TestAccBedrockAgentCoreMemoryStrategy_namespacesToNamespaceTemplates(t *testing.T) {
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
				Config: testAccMemoryStrategyConfig_basicNamespaces(rName, awstypes.MemoryStrategyTypeSemantic, "{sessionId}"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespaces"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("{sessionId}"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespace_templates"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("{sessionId}"),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccMemoryStrategyImportStateIDFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "memory_strategy_id",
			},
			{
				Config: testAccMemoryStrategyConfig_basicNamespaceTemplates(rName, awstypes.MemoryStrategyTypeSemantic, "{sessionId}"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
			{
				Config: testAccMemoryStrategyConfig_basicNamespaces(rName, awstypes.MemoryStrategyTypeSemantic, "default"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespaces"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("default"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespace_templates"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("default"),
					})),
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreMemoryStrategy_namespaceTemplatesToNamespaces(t *testing.T) {
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
				Config: testAccMemoryStrategyConfig_basicNamespaceTemplates(rName, awstypes.MemoryStrategyTypeSemantic, "{sessionId}"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespaces"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("{sessionId}"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespace_templates"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("{sessionId}"),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccMemoryStrategyImportStateIDFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "memory_strategy_id",
			},
			{
				Config: testAccMemoryStrategyConfig_basicNamespaces(rName, awstypes.MemoryStrategyTypeSemantic, "{sessionId}"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
			{
				Config: testAccMemoryStrategyConfig_basicNamespaceTemplates(rName, awstypes.MemoryStrategyTypeSemantic, "default"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespaces"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("default"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespace_templates"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("default"),
					})),
				},
			},
		},
	})
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
				Config: testAccMemoryStrategyConfig_withExecutionRole(rName, awstypes.MemoryStrategyTypeEpisodic, "Episodic strategy", "/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("Episodic strategy")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_execution_role_arn"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_strategy_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespace_templates"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(awstypes.MemoryStrategyTypeEpisodic)),
				},
			},
			// Step 2: Update episodic description (in-place)
			{
				Config: testAccMemoryStrategyConfig_withExecutionRole(rName, awstypes.MemoryStrategyTypeEpisodic, "Updated episodic strategy", "/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("Updated episodic strategy")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_execution_role_arn"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_strategy_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespace_templates"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(awstypes.MemoryStrategyTypeEpisodic)),
				},
			},
			// Step 3: Change type episodic→semantic (replacement)
			{
				Config: testAccMemoryStrategyConfig_withExecutionRole(rName, awstypes.MemoryStrategyTypeSemantic, "desc1", "default"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("desc1")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_execution_role_arn"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_strategy_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespace_templates"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("default"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(awstypes.MemoryStrategyTypeSemantic)),
				},
			},
			// Step 4: Update description + namespace (in-place)
			{
				Config: testAccMemoryStrategyConfig_withExecutionRole(rName, awstypes.MemoryStrategyTypeSemantic, "desc2", "custom"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("desc2")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_execution_role_arn"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_strategy_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespace_templates"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("custom"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(awstypes.MemoryStrategyTypeSemantic)),
				},
			},
			// Step 5: Change type semantic→user_preference (replacement)
			{
				Config: testAccMemoryStrategyConfig_withExecutionRole(rName, awstypes.MemoryStrategyTypeUserPreference, "User preference strategy", "preferences"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("User preference strategy")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_execution_role_arn"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_strategy_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespace_templates"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("preferences"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(awstypes.MemoryStrategyTypeUserPreference)),
				},
			},
			// Step 6: Try to create ANOTHER user_preference strategy → should ERROR
			{
				Config:      testAccMemoryStrategyConfig_duplicateType(rName, awstypes.MemoryStrategyTypeUserPreference),
				ExpectError: regexache.MustCompile("Found multiple strategies of type"),
			},
			// Step 7: Change type user_preference→summarization (replacement)
			{
				Config: testAccMemoryStrategyConfig_withExecutionRole(rName, awstypes.MemoryStrategyTypeSummarization, "Summarization strategy", "{sessionId}"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("Summarization strategy")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_execution_role_arn"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_strategy_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespace_templates"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("{sessionId}"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(awstypes.MemoryStrategyTypeSummarization)),
				},
			},
			// Step 8: Import test - verify composite ID import works
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
				ExpectError: regexache.MustCompile(`Attribute "configuration" must be configured`),
			},
			// Step 2: Create CUSTOM strategy with consolidation block
			{
				Config: testAccMemoryStrategyConfig_customConsolidationOnly(rName, awstypes.OverrideTypeSemanticOverride, "Focus on semantic relationships", "us.amazon.nova-2-lite-v1:0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"consolidation": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
							"append_to_prompt": knownvalue.StringExact("Focus on semantic relationships"),
							"model_id":         knownvalue.StringExact("us.amazon.nova-2-lite-v1:0"),
						})}),
						"extraction":   knownvalue.ListSizeExact(0),
						"reflection":   knownvalue.ListSizeExact(0),
						names.AttrType: tfknownvalue.StringExact(awstypes.OverrideTypeSemanticOverride),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(awstypes.MemoryStrategyTypeCustom)),
				},
			},
			// Step 3: Add extraction block and update consolidation properties (same override type)
			{
				Config: testAccMemoryStrategyConfig_custom(rName, awstypes.OverrideTypeSemanticOverride, "Updated semantic consolidation", "amazon.nova-lite-v1:0", "Extract semantic meaning", "us.amazon.nova-2-lite-v1:0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"consolidation": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
							"append_to_prompt": knownvalue.StringExact("Updated semantic consolidation"),
							"model_id":         knownvalue.StringExact("amazon.nova-lite-v1:0"),
						})}),
						"extraction": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
							"append_to_prompt": knownvalue.StringExact("Extract semantic meaning"),
							"model_id":         knownvalue.StringExact("us.amazon.nova-2-lite-v1:0"),
						})}),
						"reflection":   knownvalue.ListSizeExact(0),
						names.AttrType: tfknownvalue.StringExact(awstypes.OverrideTypeSemanticOverride),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(awstypes.MemoryStrategyTypeCustom)),
				},
			},
			// Step 4: Remove consolidation block → should replace resource
			{
				Config: testAccMemoryStrategyConfig_customExtractionOnly(rName, awstypes.OverrideTypeSemanticOverride, "Extract semantic meaning", "us.amazon.nova-2-lite-v1:0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"consolidation": knownvalue.ListSizeExact(0),
						"extraction": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
							"append_to_prompt": knownvalue.StringExact("Extract semantic meaning"),
							"model_id":         knownvalue.StringExact("us.amazon.nova-2-lite-v1:0"),
						})}),
						"reflection":   knownvalue.ListSizeExact(0),
						names.AttrType: tfknownvalue.StringExact(awstypes.OverrideTypeSemanticOverride),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(awstypes.MemoryStrategyTypeCustom)),
				},
			},
			//// Step 5: Change override type → should replace resource
			{
				Config: testAccMemoryStrategyConfig_custom(rName, awstypes.OverrideTypeUserPreferenceOverride, "Store user preferences", "amazon.nova-lite-v1:0", "Extract user preferences", "us.amazon.nova-2-lite-v1:0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"consolidation": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
							"append_to_prompt": knownvalue.StringExact("Store user preferences"),
							"model_id":         knownvalue.StringExact("amazon.nova-lite-v1:0"),
						})}),
						"extraction": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
							"append_to_prompt": knownvalue.StringExact("Extract user preferences"),
							"model_id":         knownvalue.StringExact("us.amazon.nova-2-lite-v1:0"),
						})}),
						"reflection":   knownvalue.ListSizeExact(0),
						names.AttrType: tfknownvalue.StringExact(awstypes.OverrideTypeUserPreferenceOverride),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(awstypes.MemoryStrategyTypeCustom)),
				},
			},
			//// Step 6: SUMMARY_OVERRIDE with extraction block → ValidateConfig error
			{
				Config:      testAccMemoryStrategyConfig_custom(rName, awstypes.OverrideTypeSummaryOverride, "Summary consolidation", "amazon.nova-lite-v1:0", "Summary extraction", "us.amazon.nova-2-lite-v1:0"),
				ExpectError: regexache.MustCompile(`Attribute "configuration\[0\].extraction" must not be configured`),
			},
			//// Step 7: SUMMARY_OVERRIDE with no extraction block → should succeed
			{
				Config: testAccMemoryStrategyConfig_customConsolidationOnly(rName, awstypes.OverrideTypeSummaryOverride, "Summary consolidation only", "amazon.nova-lite-v1:0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"consolidation": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
							"append_to_prompt": knownvalue.StringExact("Summary consolidation only"),
							"model_id":         knownvalue.StringExact("amazon.nova-lite-v1:0"),
						})}),
						"extraction":   knownvalue.ListSizeExact(0),
						"reflection":   knownvalue.ListSizeExact(0),
						names.AttrType: tfknownvalue.StringExact(awstypes.OverrideTypeSummaryOverride),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(awstypes.MemoryStrategyTypeCustom)),
				},
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
					resource.TestCheckResourceAttr(resourceName, "configuration.0.self_managed.0.trigger_conditions.message_based_trigger.message_count", "6"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.self_managed.0.trigger_conditions.token_based_trigger.token_count", "5000"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.self_managed.0.trigger_conditions.time_based_trigger.idle_session_timeout", "20"),
					resource.TestCheckResourceAttrSet(resourceName, "memory_strategy_id"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			// Step 2: Partially configure only the message threshold. The other
			// service-normalized defaults remain represented as computed children.
			{
				Config: testAccMemoryStrategyConfig_selfManaged(rName, 25, testAccMemoryStrategyTriggerConditionsMessageOnly(12)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.self_managed.0.historical_context_window_size", "25"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.self_managed.0.trigger_conditions.message_based_trigger.message_count", "12"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.self_managed.0.trigger_conditions.token_based_trigger.token_count", "5000"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.self_managed.0.trigger_conditions.time_based_trigger.idle_session_timeout", "20"),
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
					resource.TestCheckResourceAttr(resourceName, "configuration.0.self_managed.0.trigger_conditions.message_based_trigger.message_count", "18"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.self_managed.0.trigger_conditions.token_based_trigger.token_count", "6000"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.self_managed.0.trigger_conditions.time_based_trigger.idle_session_timeout", "30"),
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

func TestAccBedrockAgentCoreMemoryStrategy_episodicBuiltin(t *testing.T) {
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
				ConfigDirectory: config.StaticDirectory("testdata/MemoryStrategy/episodic_builtin/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:      config.StringVariable(rName),
					"namespace_template": config.StringVariable("/strategy/{memoryStrategyId}/"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_execution_role_arn"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_strategy_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName+"_s")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespaces"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("/strategy/{memoryStrategyId}/"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespace_templates"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("/strategy/{memoryStrategyId}/"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("reflection_configuration"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(awstypes.MemoryStrategyTypeEpisodic)),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/MemoryStrategy/episodic_builtin/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:      config.StringVariable(rName),
					"namespace_template": config.StringVariable("/strategy/{memoryStrategyId}/"),
				},
				ImportStateIdFunc:                    testAccMemoryStrategyImportStateIDFunc(resourceName),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "memory_strategy_id",
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/MemoryStrategy/episodic_builtin/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:      config.StringVariable(rName),
					"namespace_template": config.StringVariable("/strategy/{memoryStrategyId}/actor/{actorId}/"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_execution_role_arn"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_strategy_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName+"_s")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespaces"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("/strategy/{memoryStrategyId}/actor/{actorId}/"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespace_templates"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("/strategy/{memoryStrategyId}/actor/{actorId}/"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("reflection_configuration"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(awstypes.MemoryStrategyTypeEpisodic)),
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreMemoryStrategy_episodicBuiltinReflectionConfiguration(t *testing.T) {
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
				ConfigDirectory: config.StaticDirectory("testdata/MemoryStrategy/episodic_builtin_reflection_configuration/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:                 config.StringVariable(rName),
					"namespace_template":            config.StringVariable("/strategy/{memoryStrategyId}/actor/{actorId}/session/{sessionId}/"),
					"reflection_namespace_template": config.StringVariable("/strategy/{memoryStrategyId}/actor/{actorId}/"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_execution_role_arn"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_strategy_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName+"_s")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespaces"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("/strategy/{memoryStrategyId}/actor/{actorId}/session/{sessionId}/"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespace_templates"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("/strategy/{memoryStrategyId}/actor/{actorId}/session/{sessionId}/"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("reflection_configuration"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"namespace_templates": knownvalue.SetExact([]knownvalue.Check{
							knownvalue.StringExact("/strategy/{memoryStrategyId}/actor/{actorId}/"),
						}),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(awstypes.MemoryStrategyTypeEpisodic)),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/MemoryStrategy/episodic_builtin_reflection_configuration/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:                 config.StringVariable(rName),
					"namespace_template":            config.StringVariable("/strategy/{memoryStrategyId}/"),
					"reflection_namespace_template": config.StringVariable("/strategy/{memoryStrategyId}/actor/{actorId}/"),
				},
				ImportStateIdFunc:                    testAccMemoryStrategyImportStateIDFunc(resourceName),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "memory_strategy_id",
				ImportStateVerifyIgnore:              []string{"reflection_configuration"},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/MemoryStrategy/episodic_builtin_reflection_configuration/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:                 config.StringVariable(rName),
					"namespace_template":            config.StringVariable("/strategy/{memoryStrategyId}/actor/{actorId}/session/{sessionId}/"),
					"reflection_namespace_template": config.StringVariable("/strategy/{memoryStrategyId}/"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_execution_role_arn"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_strategy_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName+"_s")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespaces"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("/strategy/{memoryStrategyId}/actor/{actorId}/session/{sessionId}/"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespace_templates"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("/strategy/{memoryStrategyId}/actor/{actorId}/session/{sessionId}/"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("reflection_configuration"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"namespace_templates": knownvalue.SetExact([]knownvalue.Check{
							knownvalue.StringExact("/strategy/{memoryStrategyId}/"),
						}),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(awstypes.MemoryStrategyTypeEpisodic)),
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreMemoryStrategy_episodicOverride(t *testing.T) {
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
				ConfigDirectory: config.StaticDirectory("testdata/MemoryStrategy/episodic_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:                 config.StringVariable(rName),
					"namespace_template":            config.StringVariable("/strategy/{memoryStrategyId}/actor/{actorId}/session/{sessionId}/"),
					"reflection_namespace_template": config.StringVariable("/strategy/{memoryStrategyId}/actor/{actorId}/"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"consolidation": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
							"append_to_prompt": knownvalue.StringExact("<task>Consolidate</task>"),
							"model_id":         knownvalue.StringExact("us.amazon.nova-2-lite-v1:0"),
						})}),
						"extraction": knownvalue.ListSizeExact(0),
						"reflection": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
							"append_to_prompt": knownvalue.StringExact("<task>Reflect</task>"),
							"model_id":         knownvalue.StringExact("amazon.nova-lite-v1:0"),
							"namespace_templates": knownvalue.SetExact([]knownvalue.Check{
								knownvalue.StringExact("/strategy/{memoryStrategyId}/actor/{actorId}/"),
							}),
						})}),
						names.AttrType: tfknownvalue.StringExact(awstypes.OverrideTypeEpisodicOverride),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_execution_role_arn"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_strategy_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName+"_s")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespaces"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("/strategy/{memoryStrategyId}/actor/{actorId}/session/{sessionId}/"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespace_templates"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("/strategy/{memoryStrategyId}/actor/{actorId}/session/{sessionId}/"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("reflection_configuration"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(awstypes.MemoryStrategyTypeCustom)),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/MemoryStrategy/episodic_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:                 config.StringVariable(rName),
					"namespace_template":            config.StringVariable("/strategy/{memoryStrategyId}/actor/{actorId}/session/{sessionId}/"),
					"reflection_namespace_template": config.StringVariable("/strategy/{memoryStrategyId}/actor/{actorId}/"),
				},
				ImportStateIdFunc:                    testAccMemoryStrategyImportStateIDFunc(resourceName),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "memory_strategy_id",
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/MemoryStrategy/episodic_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:                 config.StringVariable(rName),
					"namespace_template":            config.StringVariable("/strategy/{memoryStrategyId}/actor/{actorId}/session/{sessionId}/"),
					"reflection_namespace_template": config.StringVariable("/strategy/{memoryStrategyId}/"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"consolidation": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
							"append_to_prompt": knownvalue.StringExact("<task>Consolidate</task>"),
							"model_id":         knownvalue.StringExact("us.amazon.nova-2-lite-v1:0"),
						})}),
						"extraction": knownvalue.ListSizeExact(0),
						"reflection": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
							"append_to_prompt": knownvalue.StringExact("<task>Reflect</task>"),
							"model_id":         knownvalue.StringExact("amazon.nova-lite-v1:0"),
							"namespace_templates": knownvalue.SetExact([]knownvalue.Check{
								knownvalue.StringExact("/strategy/{memoryStrategyId}/"),
							}),
						})}),
						names.AttrType: tfknownvalue.StringExact(awstypes.OverrideTypeEpisodicOverride),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_execution_role_arn"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_strategy_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName+"_s")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespaces"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("/strategy/{memoryStrategyId}/actor/{actorId}/session/{sessionId}/"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespace_templates"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("/strategy/{memoryStrategyId}/actor/{actorId}/session/{sessionId}/"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("reflection_configuration"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(awstypes.MemoryStrategyTypeCustom)),
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

func testAccMemoryStrategyConfig_basicNamespaces(rName string, strategyType awstypes.MemoryStrategyType, nss ...string) string {
	return acctest.ConfigCompose(testAccMemoryConfig_basic(rName), fmt.Sprintf(`	
resource "aws_bedrockagentcore_memory_strategy" "test" {
  name       = %[1]q
  memory_id  = aws_bedrockagentcore_memory.test.id
  type       = %[2]q
  namespaces = [%[3]s]
}
`, rName, strategyType, acctest.ListOfStrings(nss...)))
}

func testAccMemoryStrategyConfig_basicNamespaceTemplates(rName string, strategyType awstypes.MemoryStrategyType, nss ...string) string {
	return acctest.ConfigCompose(testAccMemoryConfig_basic(rName), fmt.Sprintf(`	
resource "aws_bedrockagentcore_memory_strategy" "test" {
  name                = %[1]q
  memory_id           = aws_bedrockagentcore_memory.test.id
  type                = %[2]q
  namespace_templates = [%[3]s]
}
`, rName, strategyType, acctest.ListOfStrings(nss...)))
}

func testAccMemoryStrategyConfig_noDescription(rName string, strategyType awstypes.MemoryStrategyType, nss ...string) string {
	return acctest.ConfigCompose(testAccMemoryConfig_basic(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_memory_strategy" "test" {
  name                = %[1]q
  memory_id           = aws_bedrockagentcore_memory.test.id
  type                = %[2]q
  namespace_templates = [%[3]s]
}
`, rName, strategyType, acctest.ListOfStrings(nss...)))
}

func testAccMemoryStrategyConfig_withExecutionRole(rName string, strategyType awstypes.MemoryStrategyType, description string, nss ...string) string {
	return acctest.ConfigCompose(testAccMemoryConfig_memoryExecutionRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_memory_strategy" "test" {
  name                      = %[1]q
  memory_id                 = aws_bedrockagentcore_memory.test.id
  memory_execution_role_arn = aws_bedrockagentcore_memory.test.memory_execution_role_arn
  type                      = %[2]q
  description               = %[3]q
  namespace_templates       = [%[4]s]
}
`, rName, strategyType, description, acctest.ListOfStrings(nss...)))
}

func testAccMemoryStrategyConfig_duplicateType(rName string, strategyType awstypes.MemoryStrategyType) string {
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
  namespace_templates       = [%[3]q]
}
`, rName, strategyType, duplicateNamespace))
}

func testAccMemoryStrategyConfig_multipleTypes(rName, description string) string {
	return acctest.ConfigCompose(testAccMemoryStrategyConfig_withExecutionRole(rName, awstypes.MemoryStrategyTypeUserPreference, "User preference strategy", "preferences"), fmt.Sprintf(`
resource "aws_bedrockagentcore_memory_strategy" "test2" {
  name                      = "%[1]s_summary"
  memory_id                 = aws_bedrockagentcore_memory.test.id
  memory_execution_role_arn = aws_bedrockagentcore_memory.test.memory_execution_role_arn
  type                      = %[2]q
  description               = %[3]q
  namespace_templates       = ["{sessionId}"]
}
`, rName, awstypes.MemoryStrategyTypeSummarization, description))
}

func testAccMemoryStrategyConfig_custom(rName string, overrideType awstypes.OverrideType, consolidationPrompt, consolidationModel, extractionPrompt, extractionModel string) string {
	return acctest.ConfigCompose(testAccMemoryConfig_memoryExecutionRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_memory_strategy" "test" {
  name                = %[1]q
  memory_id           = aws_bedrockagentcore_memory.test.id
  type                = "CUSTOM"
  description         = "Test custom strategy"
  namespace_templates = ["{sessionId}"]

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

func testAccMemoryStrategyConfig_customConsolidationOnly(rName string, overrideType awstypes.OverrideType, consolidationPrompt, consolidationModel string) string {
	return acctest.ConfigCompose(testAccMemoryConfig_memoryExecutionRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_memory_strategy" "test" {
  name                = %[1]q
  memory_id           = aws_bedrockagentcore_memory.test.id
  type                = "CUSTOM"
  description         = "Test custom strategy"
  namespace_templates = ["{sessionId}"]

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

func testAccMemoryStrategyConfig_customExtractionOnly(rName string, overrideType awstypes.OverrideType, extractionPrompt, extractionModel string) string {
	return acctest.ConfigCompose(testAccMemoryConfig_memoryExecutionRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_memory_strategy" "test" {
  name                = %[1]q
  memory_id           = aws_bedrockagentcore_memory.test.id
  type                = "CUSTOM"
  description         = "Test custom strategy"
  namespace_templates = ["{sessionId}"]

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

func testAccMemoryStrategyConfig_customInvalid(rName string) string {
	return acctest.ConfigCompose(testAccMemoryConfig_memoryExecutionRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_memory_strategy" "test" {
  name                = %[1]q
  memory_id           = aws_bedrockagentcore_memory.test.id
  type                = "CUSTOM"
  description         = "Test custom strategy"
  namespace_templates = ["default"]
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
trigger_conditions = {
  message_based_trigger = {
    message_count = %d
  }
}
`, messageCount)
}

func testAccMemoryStrategyTriggerConditions(messageCount, tokenCount, idleSessionTimeout int) string {
	return fmt.Sprintf(`
trigger_conditions = {
  message_based_trigger = {
    message_count = %d
  }
  token_based_trigger = {
    token_count = %d
  }
  time_based_trigger = {
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
