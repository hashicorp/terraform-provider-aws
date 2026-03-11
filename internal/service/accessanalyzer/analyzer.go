// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

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
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
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
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/accessanalyzer/types;types.AnalyzerSummary", serialize="true")
// @Testing(preCheck="testAccPreCheck")
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
						"internal_access": {
							Type:          schema.TypeList,
							Optional:      true,
							ForceNew:      true,
							MaxItems:      1,
							ConflictsWith: []string{"configuration.0.unused_access"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"analysis_rule": {
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"inclusion": {
													Type:     schema.TypeList,
													Optional: true,
													ForceNew: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"account_ids": {
																Type:     schema.TypeList,
																Optional: true,
																ForceNew: true,
																Elem: &schema.Schema{
																	Type:         schema.TypeString,
																	ValidateFunc: verify.ValidAccountID,
																},
															},
															"resource_arns": {
																Type:     schema.TypeList,
																Optional: true,
																ForceNew: true,
																Elem: &schema.Schema{
																	Type:         schema.TypeString,
																	ValidateFunc: verify.ValidARN,
																},
															},
															"resource_types": {
																Type:     schema.TypeList,
																Optional: true,
																ForceNew: true,
																Elem: &schema.Schema{
																	Type:             schema.TypeString,
																	ValidateDiagFunc: enum.Validate[types.ResourceType](),
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
						"unused_access": {
							Type:          schema.TypeList,
							Optional:      true,
							ForceNew:      true,
							MaxItems:      1,
							ConflictsWith: []string{"configuration.0.internal_access"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
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
																	ValidateFunc: verify.ValidAccountID,
																},
															},
															names.AttrResourceTags: {
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
									"unused_access_age": {
										Type:     schema.TypeInt,
										Optional: true,
										ForceNew: true,
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
		ClientToken:  aws.String(sdkid.UniqueId()),
		Tags:         getTagsIn(ctx),
		Type:         types.Type(d.Get(names.AttrType).(string)),
	}

	if v, ok := d.GetOk(names.AttrConfiguration); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.Configuration = expandAnalyzerConfiguration(v.([]any)[0].(map[string]any))
	}

	// Handle Organizations eventual consistency.
	_, err := tfresource.RetryWhenIsAErrorMessageContains[any, *types.ValidationException](ctx, organizationCreationTimeout,
		func(ctx context.Context) (any, error) {
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

	if !d.IsNewResource() && retry.NotFound(err) {
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
		if err := d.Set(names.AttrConfiguration, []any{flattenAnalyzerConfiguration(analyzer.Configuration)}); err != nil {
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
		ClientToken:  aws.String(sdkid.UniqueId()),
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
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Analyzer == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Analyzer, nil
}

func expandAnalyzerConfiguration(tfMap map[string]any) types.AnalyzerConfiguration {
	if tfMap == nil {
		return nil
	}

	var apiObject types.AnalyzerConfiguration

	if v, ok := tfMap["internal_access"].([]any); ok && len(v) > 0 && v[0] != nil {
		internalAccess := &types.AnalyzerConfigurationMemberInternalAccess{}
		internalAccess.Value = expandInternalAccessConfiguration(v[0].(map[string]any))
		apiObject = internalAccess
	}
	if v, ok := tfMap["unused_access"].([]any); ok && len(v) > 0 && v[0] != nil {
		unusedAccess := &types.AnalyzerConfigurationMemberUnusedAccess{}
		unusedAccess.Value = expandUnusedAccessConfiguration(v[0].(map[string]any))
		apiObject = unusedAccess
	}

	return apiObject
}

func expandInternalAccessConfiguration(tfMap map[string]any) types.InternalAccessConfiguration {
	apiObject := types.InternalAccessConfiguration{}

	if v, ok := tfMap["analysis_rule"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.AnalysisRule = expandInternalAccessAnalysisRule(v[0].(map[string]any))
	}

	return apiObject
}

func expandInternalAccessAnalysisRule(tfMap map[string]any) *types.InternalAccessAnalysisRule {
	apiObject := &types.InternalAccessAnalysisRule{}

	if v, ok := tfMap["inclusion"].([]any); ok && len(v) > 0 {
		apiObject.Inclusions = expandInternalAccessAnalysisRuleCriterias(v)
	}

	return apiObject
}

func expandInternalAccessAnalysisRuleCriterias(tfList []any) []types.InternalAccessAnalysisRuleCriteria {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.InternalAccessAnalysisRuleCriteria

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandInternalAccessAnalysisRuleCriteria(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandInternalAccessAnalysisRuleCriteria(tfMap map[string]any) *types.InternalAccessAnalysisRuleCriteria {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.InternalAccessAnalysisRuleCriteria{}

	if tfList, ok := tfMap["account_ids"].([]any); ok && len(tfList) > 0 {
		apiObject.AccountIds = flex.ExpandStringValueList(tfList)
	}

	if tfList, ok := tfMap["resource_arns"].([]any); ok && len(tfList) > 0 {
		apiObject.ResourceArns = flex.ExpandStringValueList(tfList)
	}

	if tfList, ok := tfMap["resource_types"].([]any); ok && len(tfList) > 0 {
		apiObject.ResourceTypes = flex.ExpandStringyValueList[types.ResourceType](tfList)
	}

	return apiObject
}

func expandUnusedAccessConfiguration(tfMap map[string]any) types.UnusedAccessConfiguration {
	apiObject := types.UnusedAccessConfiguration{}

	if v, ok := tfMap["analysis_rule"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.AnalysisRule = expandAnalysisRule(v[0].(map[string]any))
	}

	if v, ok := tfMap["unused_access_age"].(int); ok && v != 0 {
		apiObject.UnusedAccessAge = aws.Int32(int32(v))
	}

	return apiObject
}

func expandAnalysisRule(tfMap map[string]any) *types.AnalysisRule {
	apiObject := &types.AnalysisRule{}

	if v, ok := tfMap["exclusion"].([]any); ok && len(v) > 0 {
		apiObject.Exclusions = expandAnalysisRuleCriterias(v)
	}

	return apiObject
}

func expandAnalysisRuleCriterias(tfList []any) []types.AnalysisRuleCriteria {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.AnalysisRuleCriteria

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandAnalysisRuleCriteria(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandAnalysisRuleCriteria(tfMap map[string]any) *types.AnalysisRuleCriteria {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.AnalysisRuleCriteria{}

	if tfList, ok := tfMap["account_ids"].([]any); ok && len(tfList) > 0 {
		apiObject.AccountIds = flex.ExpandStringValueList(tfList)
	}

	if tfList, ok := tfMap[names.AttrResourceTags].([]any); ok && len(tfList) > 0 {
		for _, v := range tfList {
			if v == nil {
				continue
			}
			apiObject.ResourceTags = append(apiObject.ResourceTags, flex.ExpandStringValueMap(v.(map[string]any)))
		}
	}

	return apiObject
}

func flattenAnalyzerConfiguration(apiObject types.AnalyzerConfiguration) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	switch v := apiObject.(type) {
	case *types.AnalyzerConfigurationMemberInternalAccess:
		tfMap["internal_access"] = []any{flattenInternalAccessConfiguration(&v.Value)}
	case *types.AnalyzerConfigurationMemberUnusedAccess:
		tfMap["unused_access"] = []any{flattenUnusedAccessConfiguration(&v.Value)}
	}

	return tfMap
}

func flattenInternalAccessConfiguration(apiObject *types.InternalAccessConfiguration) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.AnalysisRule; v != nil {
		tfMap["analysis_rule"] = []any{flattenInternalAccessAnalysisRule(v)}
	}

	return tfMap
}

func flattenInternalAccessAnalysisRule(apiObject *types.InternalAccessAnalysisRule) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.Inclusions; v != nil {
		tfMap["inclusion"] = flattenInternalAccessAnalysisRuleCriterias(v)
	}

	return tfMap
}

func flattenInternalAccessAnalysisRuleCriterias(apiObjects []types.InternalAccessAnalysisRuleCriteria) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := flattenInternalAccessAnalysisRuleCriteria(&apiObject)
		if tfMap != nil {
			tfList = append(tfList, tfMap)
		}
	}

	return tfList
}

func flattenInternalAccessAnalysisRuleCriteria(apiObject *types.InternalAccessAnalysisRuleCriteria) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.AccountIds; len(v) > 0 {
		tfMap["account_ids"] = v
	}

	if v := apiObject.ResourceArns; len(v) > 0 {
		tfMap["resource_arns"] = v
	}

	if v := apiObject.ResourceTypes; len(v) > 0 {
		tfMap["resource_types"] = v
	}

	return tfMap
}

func flattenUnusedAccessConfiguration(apiObject *types.UnusedAccessConfiguration) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.AnalysisRule; v != nil {
		tfMap["analysis_rule"] = []any{flattenAnalysisRule(v)}
	}

	if v := apiObject.UnusedAccessAge; v != nil {
		tfMap["unused_access_age"] = aws.ToInt32(v)
	}

	return tfMap
}

func flattenAnalysisRule(apiObject *types.AnalysisRule) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.Exclusions; v != nil {
		tfMap["exclusion"] = flattenAnalysisRuleCriterias(v)
	}

	return tfMap
}

func flattenAnalysisRuleCriterias(apiObjects []types.AnalysisRuleCriteria) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := flattenAnalysisRuleCriteria(&apiObject)
		if tfMap != nil {
			tfList = append(tfList, tfMap)
		}
	}

	return tfList
}

func flattenAnalysisRuleCriteria(apiObject *types.AnalysisRuleCriteria) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.AccountIds; len(v) > 0 {
		tfMap["account_ids"] = v
	}

	if v := apiObject.ResourceTags; len(v) > 0 {
		tfMap[names.AttrResourceTags] = v
	}

	return tfMap
}
