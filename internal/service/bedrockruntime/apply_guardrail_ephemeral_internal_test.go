// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockruntime

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflogtest"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func TestApplyGuardrailExpand(t *testing.T) {
	t.Parallel()

	for _, source := range []awstypes.GuardrailContentSource{
		awstypes.GuardrailContentSourceInput,
		awstypes.GuardrailContentSourceOutput,
	} {
		t.Run(string(source), func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			model := applyGuardrailEphemeralResourceModel{
				GuardrailIdentifier: types.StringValue("guardrail-id"),
				GuardrailVersion:    types.StringValue("DRAFT"),
				Source:              fwtypes.StringEnumValue(source),
				Content: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []guardrailContentModel{
					{
						Text: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &guardrailTextModel{
							Text: types.StringValue("Reference text."),
							Qualifiers: fwtypes.NewListValueOfMust[fwtypes.StringEnum[awstypes.GuardrailContentQualifier]](ctx, []attr.Value{
								fwtypes.StringEnumValue(awstypes.GuardrailContentQualifierGroundingSource),
							}),
						}),
					},
					{
						Text: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &guardrailTextModel{
							Text:       types.StringValue("Text to evaluate."),
							Qualifiers: fwtypes.NewListValueOfNull[fwtypes.StringEnum[awstypes.GuardrailContentQualifier]](ctx),
						}),
					},
				}),
			}
			var input bedrockruntime.ApplyGuardrailInput
			if diags := flex.Expand(ctx, model, &input); diags.HasError() {
				t.Fatalf("expanding input: %v", diags)
			}

			want := bedrockruntime.ApplyGuardrailInput{
				GuardrailIdentifier: aws.String("guardrail-id"),
				GuardrailVersion:    aws.String("DRAFT"),
				Source:              source,
				Content: []awstypes.GuardrailContentBlock{
					&awstypes.GuardrailContentBlockMemberText{
						Value: awstypes.GuardrailTextBlock{
							Text:       aws.String("Reference text."),
							Qualifiers: []awstypes.GuardrailContentQualifier{awstypes.GuardrailContentQualifierGroundingSource},
						},
					},
					&awstypes.GuardrailContentBlockMemberText{
						Value: awstypes.GuardrailTextBlock{Text: aws.String("Text to evaluate.")},
					},
				},
			}
			if diff := cmp.Diff(want, input, cmpopts.IgnoreUnexported(
				bedrockruntime.ApplyGuardrailInput{},
				awstypes.GuardrailContentBlockMemberText{},
				awstypes.GuardrailTextBlock{},
			)); diff != "" {
				t.Errorf("unexpected input (-want +got):\n%s", diff)
			}
		})
	}
}

func TestApplyGuardrailExpandInvalidTextBlock(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	text := guardrailTextModel{
		Text:       types.StringValue("Text to evaluate."),
		Qualifiers: fwtypes.NewListValueOfNull[fwtypes.StringEnum[awstypes.GuardrailContentQualifier]](ctx),
	}
	testCases := map[string]fwtypes.ListNestedObjectValueOf[guardrailTextModel]{
		"null":     fwtypes.NewListNestedObjectValueOfNull[guardrailTextModel](ctx),
		"empty":    fwtypes.NewListNestedObjectValueOfEmpty[guardrailTextModel](ctx),
		"multiple": fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []guardrailTextModel{text, text}),
	}
	for name, value := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			model := guardrailContentModel{Text: value}
			result, diags := model.Expand(t.Context())
			if !diags.HasError() {
				t.Fatal("expected a diagnostic for an invalid text block count")
			}
			if result != nil {
				t.Errorf("expected no union member, got %v", result)
			}
		})
	}
}

func TestApplyGuardrailFlatten(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		output bedrockruntime.ApplyGuardrailOutput
		text   string
	}{
		"allowed": {
			output: bedrockruntime.ApplyGuardrailOutput{Action: awstypes.GuardrailActionNone},
		},
		"empty assessments": {
			output: bedrockruntime.ApplyGuardrailOutput{
				Action:      awstypes.GuardrailActionNone,
				Assessments: []awstypes.GuardrailAssessment{},
			},
		},
		"intervened": {
			output: bedrockruntime.ApplyGuardrailOutput{
				Action:       awstypes.GuardrailActionGuardrailIntervened,
				ActionReason: aws.String("Word policy intervened."),
				Assessments: []awstypes.GuardrailAssessment{
					{
						WordPolicy: &awstypes.GuardrailWordPolicyAssessment{
							CustomWords: []awstypes.GuardrailCustomWord{
								{
									Action:   awstypes.GuardrailWordPolicyActionBlocked,
									Detected: aws.Bool(true),
									Match:    aws.String("exampleblockedword"),
								},
							},
						},
					},
				},
				GuardrailCoverage: &awstypes.GuardrailCoverage{
					TextCharacters: &awstypes.GuardrailTextCharactersCoverage{
						Guarded: aws.Int32(18),
						Total:   aws.Int32(18),
					},
				},
				Outputs: []awstypes.GuardrailOutputContent{{Text: aws.String("Blocked.")}},
				Usage: &awstypes.GuardrailUsage{
					WordPolicyUnits: aws.Int32(1),
				},
			},
			text: "Blocked.",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			var logs bytes.Buffer
			ctx = flex.RegisterLogger(tflogtest.RootLogger(ctx, &logs))
			content := fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []guardrailContentModel{
				{
					Text: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &guardrailTextModel{
						Text:       types.StringValue("Original content."),
						Qualifiers: fwtypes.NewListValueOfNull[fwtypes.StringEnum[awstypes.GuardrailContentQualifier]](ctx),
					}),
				},
			})
			model := applyGuardrailEphemeralResourceModel{
				Content:             content,
				GuardrailIdentifier: types.StringValue("guardrail-id"),
				GuardrailVersion:    types.StringValue("DRAFT"),
				Source:              fwtypes.StringEnumValue(awstypes.GuardrailContentSourceInput),
			}
			if diags := flex.Flatten(ctx, &tc.output, &model); diags.HasError() {
				t.Fatalf("flattening output: %v", diags)
			}
			entries, err := tflogtest.MultilineJSONDecode(&logs)
			if err != nil {
				t.Fatalf("decoding logs: %v", err)
			}
			for _, entry := range entries {
				if entry["@level"] == "error" {
					t.Errorf("unexpected error log: %v", entry)
				}
			}
			if model.Action.ValueString() != string(tc.output.Action) {
				t.Errorf("unexpected action: %s", model.Action)
			}
			if !model.ActionReason.Equal(flex.StringToFramework(ctx, tc.output.ActionReason)) {
				t.Errorf("unexpected action reason: %s", model.ActionReason)
			}
			var assessments []awstypes.GuardrailAssessment
			if diags := flex.Expand(ctx, model.Assessments, &assessments); diags.HasError() {
				t.Fatalf("expanding assessments: %v", diags)
			}
			if diff := cmp.Diff(tc.output.Assessments, assessments, cmpopts.EquateEmpty(), cmpopts.IgnoreUnexported(
				awstypes.GuardrailAssessment{}, awstypes.GuardrailWordPolicyAssessment{}, awstypes.GuardrailCustomWord{},
			)); diff != "" {
				t.Errorf("unexpected assessments (-want +got):\n%s", diff)
			}
			if len(model.Assessments.Elements()) != len(tc.output.Assessments) {
				t.Errorf("unexpected assessment count: %s", model.Assessments)
			}
			if tc.output.Assessments == nil && !model.Assessments.IsNull() {
				t.Errorf("expected null assessments, got %s", model.Assessments)
			}
			if tc.output.Assessments != nil && model.Assessments.IsNull() {
				t.Error("non-nil assessments must remain a list")
			}
			if !model.Content.Equal(content) || model.GuardrailIdentifier.ValueString() != "guardrail-id" ||
				model.GuardrailVersion.ValueString() != "DRAFT" || model.Source.ValueString() != "INPUT" {
				t.Error("flattening overwrote configured input")
			}
			if tc.output.GuardrailCoverage != nil {
				coverage, diags := model.GuardrailCoverage.ToPtr(ctx)
				if diags.HasError() {
					t.Fatalf("converting guardrail coverage: %v", diags)
				}
				if coverage == nil || coverage.TextCharacters.IsNull() {
					t.Errorf("unexpected guardrail coverage: %s", model.GuardrailCoverage)
				}
			}
			if tc.output.Usage != nil {
				usage, diags := model.Usage.ToPtr(ctx)
				if diags.HasError() {
					t.Fatalf("converting usage: %v", diags)
				}
				if usage == nil || usage.WordPolicyUnits.ValueInt64() != int64(aws.ToInt32(tc.output.Usage.WordPolicyUnits)) {
					t.Errorf("unexpected usage: %s", model.Usage)
				}
			}
			if tc.text == "" {
				if !model.Output.IsNull() && len(model.Output.Elements()) != 0 {
					t.Errorf("expected no output, got %s", model.Output)
				}
			} else {
				outputs, diags := model.Output.ToSlice(ctx)
				if diags.HasError() {
					t.Fatalf("converting output: %v", diags)
				}
				if len(outputs) != 1 || outputs[0].Text.ValueString() != tc.text {
					t.Errorf("unexpected output: %s", model.Output)
				}
			}
		})
	}
}

func TestApplyGuardrailFlattenAssessments(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	want := []awstypes.GuardrailAssessment{
		{
			AppliedGuardrailDetails: &awstypes.AppliedGuardrailDetails{
				GuardrailArn:       aws.String("arn:aws:bedrock:us-east-1:123456789012:guardrail/example"),
				GuardrailId:        aws.String("example"),
				GuardrailOrigin:    []awstypes.GuardrailOrigin{awstypes.GuardrailOriginRequest, awstypes.GuardrailOriginAccountEnforced},
				GuardrailOwnership: awstypes.GuardrailOwnershipCrossAccount,
				GuardrailVersion:   aws.String("1"),
			},
			ContentPolicy: &awstypes.GuardrailContentPolicyAssessment{
				Filters: []awstypes.GuardrailContentFilter{
					{Action: "BLOCKED", Confidence: "HIGH", Detected: aws.Bool(true), FilterStrength: "MEDIUM", Type: "HATE"},
					{Action: "NONE", Confidence: "LOW", Detected: aws.Bool(false), FilterStrength: "LOW", Type: "VIOLENCE"},
				},
			},
			ContextualGroundingPolicy: &awstypes.GuardrailContextualGroundingPolicyAssessment{
				Filters: []awstypes.GuardrailContextualGroundingFilter{
					{Action: "BLOCKED", Detected: aws.Bool(true), Score: aws.Float64(0.25), Threshold: aws.Float64(0.75), Type: "GROUNDING"},
					{Action: "NONE", Detected: aws.Bool(false), Score: aws.Float64(1), Threshold: aws.Float64(0), Type: "RELEVANCE"},
				},
			},
			InvocationMetrics: &awstypes.GuardrailInvocationMetrics{
				GuardrailCoverage: &awstypes.GuardrailCoverage{
					Images:         &awstypes.GuardrailImageCoverage{Guarded: aws.Int32(1), Total: aws.Int32(2)},
					TextCharacters: &awstypes.GuardrailTextCharactersCoverage{Guarded: aws.Int32(12), Total: aws.Int32(34)},
				},
				GuardrailProcessingLatency: aws.Int64(42),
				Usage: &awstypes.GuardrailUsage{
					AutomatedReasoningPolicies: aws.Int32(1), AutomatedReasoningPolicyUnits: aws.Int32(2),
					ContentPolicyImageUnits: aws.Int32(3), ContentPolicyUnits: aws.Int32(4),
					ContextualGroundingPolicyUnits: aws.Int32(5), SensitiveInformationPolicyFreeUnits: aws.Int32(6),
					SensitiveInformationPolicyUnits: aws.Int32(7), TopicPolicyUnits: aws.Int32(8), WordPolicyUnits: aws.Int32(9),
				},
			},
			SensitiveInformationPolicy: &awstypes.GuardrailSensitiveInformationPolicyAssessment{
				PiiEntities: []awstypes.GuardrailPiiEntityFilter{
					{Action: "ANONYMIZED", Detected: aws.Bool(true), Match: aws.String("user@example.com"), Type: "EMAIL"},
					{Action: "NONE", Detected: aws.Bool(false), Match: aws.String("123"), Type: "PIN"},
				},
				Regexes: []awstypes.GuardrailRegexFilter{
					{Action: "BLOCKED", Detected: aws.Bool(true), Match: aws.String("secret-123"), Name: aws.String("secret"), Regex: aws.String("secret-[0-9]+")},
				},
			},
			TopicPolicy: &awstypes.GuardrailTopicPolicyAssessment{
				Topics: []awstypes.GuardrailTopic{
					{Action: "BLOCKED", Detected: aws.Bool(true), Name: aws.String("restricted"), Type: "DENY"},
				},
			},
			WordPolicy: &awstypes.GuardrailWordPolicyAssessment{
				CustomWords: []awstypes.GuardrailCustomWord{
					{Action: "BLOCKED", Detected: aws.Bool(true), Match: aws.String("blocked")},
					{Action: "NONE", Detected: aws.Bool(false), Match: aws.String("allowed")},
				},
				ManagedWordLists: []awstypes.GuardrailManagedWord{
					{Action: "BLOCKED", Detected: aws.Bool(true), Match: aws.String("profanity"), Type: "PROFANITY"},
				},
			},
		},
		{
			WordPolicy: &awstypes.GuardrailWordPolicyAssessment{
				CustomWords: []awstypes.GuardrailCustomWord{{Action: "NONE", Match: aws.String("not detected")}},
			},
		},
	}
	var model applyGuardrailEphemeralResourceModel
	if diags := flex.Flatten(ctx, &bedrockruntime.ApplyGuardrailOutput{Assessments: want}, &model); diags.HasError() {
		t.Fatalf("flattening assessments: %v", diags)
	}
	var got []awstypes.GuardrailAssessment
	if diags := flex.Expand(ctx, model.Assessments, &got); diags.HasError() {
		t.Fatalf("expanding assessments: %v", diags)
	}
	if diff := cmp.Diff(want, got, cmp.Exporter(func(t reflect.Type) bool {
		return t.PkgPath() == reflect.TypeFor[awstypes.GuardrailAssessment]().PkgPath()
	})); diff != "" {
		t.Errorf("assessment fields lost during conversion (-want +got):\n%s", diff)
	}
	assessments, diags := model.Assessments.ToSlice(ctx)
	if diags.HasError() || len(assessments) != 2 {
		t.Fatalf("unexpected assessments: %s (%v)", model.Assessments, diags)
	}
	if !assessments[1].ContentPolicy.IsNull() || !assessments[1].AutomatedReasoningPolicy.IsNull() {
		t.Error("absent policies must remain null")
	}

	var resp ephemeral.SchemaResponse
	e := &applyGuardrailEphemeralResource{}
	e.Schema(ctx, ephemeral.SchemaRequest{}, &resp)
	if diags := resp.Schema.ValidateImplementation(ctx); diags.HasError() {
		t.Fatalf("invalid schema: %v", diags)
	}
	if a := resp.Schema.Attributes["assessments"]; !a.IsComputed() || !a.IsSensitive() {
		t.Error("assessments must be computed and sensitive")
	}
	// The framework's region interceptor adds this attribute to the resource schema.
	resp.Schema.Attributes["region"] = schema.StringAttribute{Computed: true}
	model.Content = fwtypes.NewListNestedObjectValueOfNull[guardrailContentModel](ctx)
	result := tfsdk.EphemeralResultData{Schema: resp.Schema}
	if diags := result.Set(ctx, &model); diags.HasError() {
		t.Fatalf("serializing ephemeral result: %v", diags)
	}
	var restored applyGuardrailEphemeralResourceModel
	if diags := result.Get(ctx, &restored); diags.HasError() {
		t.Fatalf("reading ephemeral result: %v", diags)
	}
	if !restored.Assessments.Equal(model.Assessments) {
		t.Error("schema serialization changed assessments")
	}
}

func TestApplyGuardrailFlattenAutomatedReasoning(t *testing.T) {
	t.Parallel()

	claims := []awstypes.GuardrailAutomatedReasoningStatement{
		{Logic: aws.String("eligible(person)"), NaturalLanguage: aws.String("The person is eligible.")},
		{Logic: aws.String("adult(person)"), NaturalLanguage: aws.String("The person is an adult.")},
	}
	premises := []awstypes.GuardrailAutomatedReasoningStatement{
		{Logic: aws.String("age(person) >= 18"), NaturalLanguage: aws.String("The person is at least 18.")},
	}
	translation := &awstypes.GuardrailAutomatedReasoningTranslation{
		Claims: claims, Confidence: aws.Float64(0.875), Premises: premises,
		UntranslatedClaims: []awstypes.GuardrailAutomatedReasoningInputTextReference{
			{Text: aws.String("Untranslated claim one.")}, {Text: aws.String("Untranslated claim two.")},
		},
		UntranslatedPremises: []awstypes.GuardrailAutomatedReasoningInputTextReference{{Text: aws.String("Untranslated premise.")}},
	}
	warning := &awstypes.GuardrailAutomatedReasoningLogicWarning{
		Claims: claims, Premises: premises, Type: awstypes.GuardrailAutomatedReasoningLogicWarningTypeAlwaysTrue,
	}
	rules := []awstypes.GuardrailAutomatedReasoningRule{
		{Identifier: aws.String("rule-1"), PolicyVersionArn: aws.String("arn:aws:bedrock:us-east-1:123456789012:automated-reasoning-policy/example:1")},
		{Identifier: aws.String("rule-2"), PolicyVersionArn: aws.String("arn:aws:bedrock:us-east-1:123456789012:automated-reasoning-policy/example:2")},
	}
	scenario := &awstypes.GuardrailAutomatedReasoningScenario{Statements: claims}
	impossible := awstypes.GuardrailAutomatedReasoningImpossibleFinding{ContradictingRules: rules, LogicWarning: warning, Translation: translation}
	invalid := awstypes.GuardrailAutomatedReasoningInvalidFinding{ContradictingRules: rules, LogicWarning: warning, Translation: translation}
	satisfiable := awstypes.GuardrailAutomatedReasoningSatisfiableFinding{
		ClaimsFalseScenario: &awstypes.GuardrailAutomatedReasoningScenario{Statements: premises},
		ClaimsTrueScenario:  scenario, LogicWarning: warning, Translation: translation,
	}
	ambiguous := awstypes.GuardrailAutomatedReasoningTranslationAmbiguousFinding{
		DifferenceScenarios: []awstypes.GuardrailAutomatedReasoningScenario{*scenario, {Statements: premises}},
		Options: []awstypes.GuardrailAutomatedReasoningTranslationOption{
			{Translations: []awstypes.GuardrailAutomatedReasoningTranslation{*translation, {Claims: premises, Confidence: aws.Float64(0)}}},
			{Translations: []awstypes.GuardrailAutomatedReasoningTranslation{{Premises: claims, Confidence: aws.Float64(1)}}},
		},
	}
	valid := awstypes.GuardrailAutomatedReasoningValidFinding{
		ClaimsTrueScenario: scenario, LogicWarning: warning, SupportingRules: rules, Translation: translation,
	}
	testCases := map[string]struct {
		finding awstypes.GuardrailAutomatedReasoningFinding
		want    any
	}{
		"impossible":            {&awstypes.GuardrailAutomatedReasoningFindingMemberImpossible{Value: impossible}, impossible},
		"invalid":               {&awstypes.GuardrailAutomatedReasoningFindingMemberInvalid{Value: invalid}, invalid},
		"no_translations":       {&awstypes.GuardrailAutomatedReasoningFindingMemberNoTranslations{}, awstypes.GuardrailAutomatedReasoningNoTranslationsFinding{}},
		"satisfiable":           {&awstypes.GuardrailAutomatedReasoningFindingMemberSatisfiable{Value: satisfiable}, satisfiable},
		"too_complex":           {&awstypes.GuardrailAutomatedReasoningFindingMemberTooComplex{}, awstypes.GuardrailAutomatedReasoningTooComplexFinding{}},
		"translation_ambiguous": {&awstypes.GuardrailAutomatedReasoningFindingMemberTranslationAmbiguous{Value: ambiguous}, ambiguous},
		"valid":                 {&awstypes.GuardrailAutomatedReasoningFindingMemberValid{Value: valid}, valid},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			var logs bytes.Buffer
			ctx = flex.RegisterLogger(tflogtest.RootLogger(ctx, &logs))
			output := bedrockruntime.ApplyGuardrailOutput{
				Assessments: []awstypes.GuardrailAssessment{
					{AutomatedReasoningPolicy: &awstypes.GuardrailAutomatedReasoningPolicyAssessment{
						Findings: []awstypes.GuardrailAutomatedReasoningFinding{
							tc.finding,
							&awstypes.GuardrailAutomatedReasoningFindingMemberNoTranslations{},
						},
					}},
				},
			}
			var model applyGuardrailEphemeralResourceModel
			if diags := flex.Flatten(ctx, &output, &model); diags.HasError() {
				t.Fatalf("flattening automated reasoning: %v", diags)
			}
			assessments, diags := model.Assessments.ToSlice(ctx)
			if diags.HasError() || len(assessments) != 1 {
				t.Fatalf("unexpected assessments: %s (%v)", model.Assessments, diags)
			}
			policy, diags := assessments[0].AutomatedReasoningPolicy.ToPtr(ctx)
			if diags.HasError() || policy == nil {
				t.Fatalf("unexpected automated reasoning policy: %v", diags)
			}
			findings, diags := policy.Findings.ToSlice(ctx)
			if diags.HasError() || len(findings) != 2 {
				t.Fatalf("unexpected findings: %s (%v)", policy.Findings, diags)
			}
			member := policy.Findings.Elements()[0].(fwtypes.ObjectValueOf[guardrailAutomatedReasoningFindingModel]).Attributes()[name]
			for key, value := range policy.Findings.Elements()[0].(fwtypes.ObjectValueOf[guardrailAutomatedReasoningFindingModel]).Attributes() {
				if key == name {
					if value.IsNull() || value.IsUnknown() {
						t.Fatalf("active union member %s is not known", key)
					}
				} else if !value.IsNull() {
					t.Errorf("inactive union member %s is not null", key)
				}
			}
			got := reflect.New(reflect.TypeOf(tc.want)).Interface()
			if diags := flex.Expand(ctx, member, got); diags.HasError() {
				t.Fatalf("expanding finding: %v", diags)
			}
			if diff := cmp.Diff(tc.want, reflect.ValueOf(got).Elem().Interface(), cmp.Exporter(func(t reflect.Type) bool {
				return t.PkgPath() == reflect.TypeFor[awstypes.GuardrailAssessment]().PkgPath()
			})); diff != "" {
				t.Errorf("finding fields lost during conversion (-want +got):\n%s", diff)
			}
			if len(findings[1].NoTranslations.Elements()) != 1 {
				t.Error("empty union variant must be a singleton list containing an empty object")
			}
			entries, err := tflogtest.MultilineJSONDecode(&logs)
			if err != nil {
				t.Fatalf("decoding logs: %v", err)
			}
			for _, entry := range entries {
				if entry["@level"] == "error" {
					t.Errorf("unexpected error log: %v", entry)
				}
			}
		})
	}
}

func TestApplyGuardrailFlattenUnknownFinding(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	var model applyGuardrailEphemeralResourceModel
	output := bedrockruntime.ApplyGuardrailOutput{
		Assessments: []awstypes.GuardrailAssessment{
			{AutomatedReasoningPolicy: &awstypes.GuardrailAutomatedReasoningPolicyAssessment{
				Findings: []awstypes.GuardrailAutomatedReasoningFinding{
					&awstypes.UnknownUnionMember{Tag: "future_finding"},
				},
			}},
		},
	}
	if diags := flex.Flatten(ctx, &output, &model); !diags.HasError() {
		t.Fatal("unknown findings must return a diagnostic rather than silently losing assessment data")
	}
}
