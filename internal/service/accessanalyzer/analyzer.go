// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package accessanalyzer

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/accessanalyzer"
	"github.com/aws/aws-sdk-go-v2/service/accessanalyzer/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	// Maximum amount of time to wait for Organizations eventual consistency on creation
	// This timeout value is much higher than usual since the cross-service validation
	// appears to be consistently caching for 5 minutes:
	// --- PASS: TestAccAccessAnalyzer_serial/Analyzer/Type_Organization (315.86s)
	organizationCreationTimeout = 10 * time.Minute
)

// @SDKResource("aws_accessanalyzer_analyzer", name="Analyzer")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/accessanalyzer/types;types.AnalyzerSummary", serialize="true", preCheck="true")
func resourceAnalyzer() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAnalyzerCreate,
		ReadWithoutTimeout:   resourceAnalyzerRead,
		UpdateWithoutTimeout: resourceAnalyzerUpdate,
		DeleteWithoutTimeout: resourceAnalyzerDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"analyzer_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexache.MustCompile(`^[A-Za-z][0-9A-Za-z_.-]*$`), "must begin with a letter and contain only alphanumeric, underscore, period, or hyphen characters"),
				),
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrConfiguration: {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"unused_access": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"unused_access_age": {
										Type:     schema.TypeInt,
										Optional: true,
										ForceNew: true,
									},
									"analysis_rule": {
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"exclusion": {
													Type:     schema.TypeList,
													Optional: true,
													ForceNew: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"account_ids": {
																Type:     schema.TypeList,
																Optional: true,
																ForceNew: true,
																MaxItems: 2000,
																Elem: &schema.Schema{
																	Type:         schema.TypeString,
																	ValidateFunc: validation.StringMatch(regexache.MustCompile(`^\d{12}$`), "Must be a 12-digit account ID"),
																},
															},
															"resource_tags": {
																Type:     schema.TypeList,
																Optional: true,
																ForceNew: true,
																Elem: &schema.Schema{
																	Type: schema.TypeMap,
																	Elem: &schema.Schema{
																		Type:         schema.TypeString,
																		ValidateFunc: validation.StringLenBetween(0, 256),
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrType: {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          types.TypeAccount,
				ValidateDiagFunc: enum.Validate[types.Type](),
			},
		},
	}
}

func resourceAnalyzerCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AccessAnalyzerClient(ctx)

	analyzerName := d.Get("analyzer_name").(string)
	input := accessanalyzer.CreateAnalyzerInput{
		AnalyzerName: aws.String(analyzerName),
		ClientToken:  aws.String(id.UniqueId()),
		Tags:         getTagsIn(ctx),
		Type:         types.Type(d.Get(names.AttrType).(string)),
	}

	if v, ok := d.GetOk(names.AttrConfiguration); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.Configuration = expandAnalyzerConfiguration(v.([]any)[0].(map[string]any))
	}

	// Handle Organizations eventual consistency.
	_, err := tfresource.RetryWhenIsAErrorMessageContains[*types.ValidationException](ctx, organizationCreationTimeout,
		func() (any, error) {
			return conn.CreateAnalyzer(ctx, &input)
		},
		"You must create an organization",
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IAM Access Analyzer Analyzer (%s): %s", analyzerName, err)
	}

	d.SetId(analyzerName)

	return append(diags, resourceAnalyzerRead(ctx, d, meta)...)
}

func resourceAnalyzerRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AccessAnalyzerClient(ctx)

	analyzer, err := findAnalyzerByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IAM Access Analyzer Analyzer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Access Analyzer Analyzer (%s): %s", d.Id(), err)
	}

	d.Set("analyzer_name", analyzer.Name)
	d.Set(names.AttrARN, analyzer.Arn)
	if analyzer.Configuration != nil {
		if err := d.Set(names.AttrConfiguration, []any{flattenConfiguration(analyzer.Configuration)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting configuration: %s", err)
		}
	} else {
		d.Set(names.AttrConfiguration, nil)
	}
	d.Set(names.AttrType, analyzer.Type)

	setTagsOut(ctx, analyzer.Tags)

	return diags
}

func resourceAnalyzerUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceAnalyzerRead(ctx, d, meta)...)
}

func resourceAnalyzerDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AccessAnalyzerClient(ctx)

	log.Printf("[DEBUG] Deleting IAM Access Analyzer Analyzer: %s", d.Id())
	input := accessanalyzer.DeleteAnalyzerInput{
		AnalyzerName: aws.String(d.Id()),
		ClientToken:  aws.String(id.UniqueId()),
	}
	_, err := conn.DeleteAnalyzer(ctx, &input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM Access Analyzer Analyzer (%s): %s", d.Id(), err)
	}

	return diags
}

func findAnalyzerByName(ctx context.Context, conn *accessanalyzer.Client, name string) (*types.AnalyzerSummary, error) {
	input := accessanalyzer.GetAnalyzerInput{
		AnalyzerName: aws.String(name),
	}

	output, err := conn.GetAnalyzer(ctx, &input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Analyzer == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Analyzer, nil
}

func expandAnalyzerConfiguration(tfMap map[string]any) types.AnalyzerConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.AnalyzerConfigurationMemberUnusedAccess{}

	if v, ok := tfMap["unused_access"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.Value = expandUnusedAccess(v[0].(map[string]any))
	}

	return apiObject
}

func expandUnusedAccess(tfMap map[string]any) types.UnusedAccessConfiguration {
	apiObject := types.UnusedAccessConfiguration{}

	if v, ok := tfMap["unused_access_age"].(int); ok && v != 0 {
		apiObject.UnusedAccessAge = aws.Int32(int32(v))
	}

	if v, ok := tfMap["analysis_rule"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.AnalysisRule = expandAnalysisRule(v[0].(map[string]any))
	}

	return apiObject
}

func expandAnalysisRule(tfMap map[string]any) *types.AnalysisRule {
	apiObject := &types.AnalysisRule{}

	if v, ok := tfMap["exclusion"].([]any); ok && len(v) > 0 {
		apiObject.Exclusions = expandExclusions(v)
	}

	return apiObject
}

func expandExclusions(tfList []any) []types.AnalysisRuleCriteria {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.AnalysisRuleCriteria

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)

		if !ok {
			continue
		}

		apiObject := expandExclusion(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandExclusion(tfMap map[string]any) *types.AnalysisRuleCriteria {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.AnalysisRuleCriteria{}

	if v, ok := tfMap["account_ids"].([]any); ok && len(v) > 0 {
		for _, accountId := range v {
			accountId, ok := accountId.(string)
			if !ok {
				continue
			}
			apiObject.AccountIds = append(apiObject.AccountIds, accountId)
		}
	}

	if v, ok := tfMap["resource_tags"].([]any); ok && len(v) > 0 {
		for _, resourceTag := range v {
			resourceTagMap := flex.ExpandStringValueMap(resourceTag.(map[string]any))
			apiObject.ResourceTags = append(apiObject.ResourceTags, resourceTagMap)
		}
	}

	return apiObject
}

func flattenConfiguration(apiObject types.AnalyzerConfiguration) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	switch v := apiObject.(type) {
	case *types.AnalyzerConfigurationMemberUnusedAccess:
		tfMap["unused_access"] = []any{flattenUnusedAccessConfiguration(&v.Value)}
	}

	return tfMap
}

func flattenUnusedAccessConfiguration(apiObject *types.UnusedAccessConfiguration) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.UnusedAccessAge; v != nil {
		tfMap["unused_access_age"] = aws.ToInt32(v)
	}

	if v := apiObject.AnalysisRule; v != nil {
		tfMap["analysis_rule"] = []any{flattenAnalysisRule(v)}
	}
	return tfMap
}

func flattenAnalysisRule(apiObject *types.AnalysisRule) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.Exclusions; v != nil {
		tfMap["exclusion"] = flattenExclusions(v)
	}

	return tfMap
}

func flattenExclusions(apiObjects []types.AnalysisRuleCriteria) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := flattenExclusion(&apiObject)
		if tfMap != nil {
			tfList = append(tfList, tfMap)
		}
	}

	return tfList
}
func flattenExclusion(apiObject *types.AnalysisRuleCriteria) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.AccountIds; len(v) > 0 {
		tfMap["account_ids"] = v
	}

	if v := apiObject.ResourceTags; len(v) > 0 {
		tfMap["resource_tags"] = v
	}

	return tfMap
}
