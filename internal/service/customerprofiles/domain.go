// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package customerprofiles

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/customerprofiles"
	"github.com/aws/aws-sdk-go-v2/service/customerprofiles/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_customerprofiles_domain")
// @Tags(identifierAttribute="arn")
func ResourceDomain() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainCreate,
		ReadWithoutTimeout:   resourceDomainRead,
		UpdateWithoutTimeout: resourceDomainUpdate,
		DeleteWithoutTimeout: resourceDomainDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDomainName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"dead_letter_queue_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"default_encryption_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"default_expiration_days": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"matching": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auto_merging": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"conflict_resolution": conflictResolutionSchema(),
									"consolidation": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"matching_attributes_list": {
													Type:     schema.TypeList,
													Required: true,
													Elem: &schema.Schema{
														Type: schema.TypeList,
														Elem: &schema.Schema{Type: schema.TypeString},
													},
												},
											},
										},
									},
									names.AttrEnabled: {
										Type:     schema.TypeBool,
										Required: true,
									},
									"min_allowed_confidence_score_for_merging": {
										Type:     schema.TypeFloat,
										Optional: true,
									},
								},
							},
						},
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Required: true,
						},
						"exporting_config": exportingConfigSchema(),
						"job_schedule": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"day_of_the_week": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[types.JobScheduleDayOfTheWeek](),
									},
									"time": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"rule_based_matching": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attribute_types_selector": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrAddress: {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"attribute_matching_model": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[types.AttributeMatchingModel](),
									},
									"email_address": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"phone_number": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"conflict_resolution": conflictResolutionSchema(),
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Required: true,
						},
						"exporting_config": exportingConfigSchema(),
						"matching_rules": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrRule: {
										Type:     schema.TypeList,
										Required: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"max_allowed_rule_level_for_matching": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"max_allowed_rule_level_for_merging": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						names.AttrStatus: {
							Type:             schema.TypeString,
							Computed:         true,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[types.RuleBasedMatchingStatus](),
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func conflictResolutionSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"conflict_resolving_model": {
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: enum.Validate[types.ConflictResolvingModel](),
				},
				"source_name": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		},
	}
}

func exportingConfigSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"s3_exporting": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrS3BucketName: {
								Type:     schema.TypeString,
								Required: true,
							},
							"s3_key_name": {
								Type:     schema.TypeString,
								Optional: true,
							},
						},
					},
				},
			},
		},
	}
}

func resourceDomainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CustomerProfilesClient(ctx)

	name := d.Get(names.AttrDomainName).(string)
	input := &customerprofiles.CreateDomainInput{
		DomainName: aws.String(name),
		Tags:       getTagsIn(ctx),
	}

	if v, ok := d.GetOk("dead_letter_queue_url"); ok {
		input.DeadLetterQueueUrl = aws.String(v.(string))
	}

	if v, ok := d.GetOk("default_encryption_key"); ok {
		input.DefaultEncryptionKey = aws.String(v.(string))
	}

	if v, ok := d.GetOk("default_expiration_days"); ok {
		input.DefaultExpirationDays = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("matching"); ok {
		input.Matching = expandMatching(v.([]interface{}))
	}

	if v, ok := d.GetOk("rule_based_matching"); ok {
		input.RuleBasedMatching = expandRuleBasedMatching(v.([]interface{}))
	}

	output, err := conn.CreateDomain(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Customer Profiles Domain (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.DomainName))

	return append(diags, resourceDomainRead(ctx, d, meta)...)
}

func resourceDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CustomerProfilesClient(ctx)

	output, err := FindDomainByDomainName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Customer Profiles Domain with DomainName (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Customer Profiles Domain: (%s) %s", d.Id(), err)
	}

	d.Set(names.AttrARN, buildDomainARN(meta.(*conns.AWSClient), d.Id()))
	d.Set(names.AttrDomainName, output.DomainName)
	d.Set("dead_letter_queue_url", output.DeadLetterQueueUrl)
	d.Set("default_encryption_key", output.DefaultEncryptionKey)
	d.Set("default_expiration_days", output.DefaultExpirationDays)
	d.Set("matching", flattenMatching(output.Matching))
	d.Set("rule_based_matching", flattenRuleBasedMatching(output.RuleBasedMatching))

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceDomainUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CustomerProfilesClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &customerprofiles.UpdateDomainInput{
			DomainName: aws.String(d.Get(names.AttrDomainName).(string)),
		}

		if d.HasChange("dead_letter_queue_url") {
			input.DeadLetterQueueUrl = aws.String(d.Get("dead_letter_queue_url").(string))
		}

		if d.HasChange("default_encryption_key") {
			input.DefaultEncryptionKey = aws.String(d.Get("default_encryption_key").(string))
		}

		if d.HasChange("default_expiration_days") {
			input.DefaultExpirationDays = aws.Int32(int32(d.Get("default_expiration_days").(int)))
		}

		if d.HasChange("matching") {
			input.Matching = expandMatching(d.Get("matching").([]interface{}))
		}

		if d.HasChange("rule_based_matching") {
			input.RuleBasedMatching = expandRuleBasedMatching(d.Get("rule_based_matching").([]interface{}))
		}

		_, err := conn.UpdateDomain(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Customer Profiles Domain (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceDomainRead(ctx, d, meta)...)
}

func resourceDomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CustomerProfilesClient(ctx)

	log.Printf("[DEBUG] Deleting Customer Profiles Profile: %s", d.Id())
	_, err := conn.DeleteDomain(ctx, &customerprofiles.DeleteDomainInput{
		DomainName: aws.String(d.Id()),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Customer Profiles Domain (%s): %s", d.Id(), err)
	}

	return diags
}

func FindDomainByDomainName(ctx context.Context, conn *customerprofiles.Client, domainName string) (*customerprofiles.GetDomainOutput, error) {
	input := &customerprofiles.GetDomainInput{
		DomainName: aws.String(domainName),
	}

	output, err := conn.GetDomain(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func expandMatching(tfMap []interface{}) *types.MatchingRequest {
	if len(tfMap) == 0 {
		return nil
	}

	tfList, ok := tfMap[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &types.MatchingRequest{}

	if v, ok := tfList[names.AttrEnabled]; ok {
		apiObject.Enabled = aws.Bool(v.(bool))
	}

	if v, ok := tfList["auto_merging"]; ok {
		apiObject.AutoMerging = expandAutoMerging(v.([]interface{}))
	}

	if v, ok := tfList["exporting_config"]; ok {
		apiObject.ExportingConfig = expandExportingConfig(v.([]interface{}))
	}

	if v, ok := tfList["job_schedule"]; ok {
		apiObject.JobSchedule = expandJobSchedule(v.([]interface{}))
	}

	return apiObject
}

func expandAutoMerging(tfMap []interface{}) *types.AutoMerging {
	if len(tfMap) == 0 {
		return nil
	}

	tfList, ok := tfMap[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &types.AutoMerging{}

	if v, ok := tfList[names.AttrEnabled]; ok {
		apiObject.Enabled = aws.Bool(v.(bool))
	}

	if v, ok := tfList["conflict_resolution"]; ok {
		apiObject.ConflictResolution = expandConflictResolution(v.([]interface{}))
	}

	if v, ok := tfList["consolidation"]; ok {
		apiObject.Consolidation = expandConsolidation(v.([]interface{}))
	}

	if v, ok := tfList["min_allowed_confidence_score_for_merging"]; ok {
		apiObject.MinAllowedConfidenceScoreForMerging = aws.Float64(v.(float64))
	}

	return apiObject
}

func expandConflictResolution(tfMap []interface{}) *types.ConflictResolution {
	if len(tfMap) == 0 {
		return nil
	}

	tfList, ok := tfMap[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &types.ConflictResolution{}

	if v, ok := tfList["conflict_resolving_model"]; ok {
		apiObject.ConflictResolvingModel = types.ConflictResolvingModel(v.(string))
	}

	if v, ok := tfList["source_name"]; ok && v != "" {
		apiObject.SourceName = aws.String(v.(string))
	}

	return apiObject
}

func expandConsolidation(tfMap []interface{}) *types.Consolidation {
	if len(tfMap) == 0 {
		return nil
	}

	tfList, ok := tfMap[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &types.Consolidation{}

	if v, ok := tfList["matching_attributes_list"]; ok {
		apiObject.MatchingAttributesList = expandMatchingAttributesList(v.([]interface{}))
	}

	return apiObject
}

func expandExportingConfig(tfMap []interface{}) *types.ExportingConfig {
	if len(tfMap) == 0 {
		return nil
	}

	tfList, ok := tfMap[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &types.ExportingConfig{}

	if v, ok := tfList["s3_exporting"]; ok {
		apiObject.S3Exporting = expandS3ExportingConfig(v.([]interface{}))
	}

	return apiObject
}

func expandS3ExportingConfig(tfMap []interface{}) *types.S3ExportingConfig {
	if len(tfMap) == 0 {
		return nil
	}

	tfList, ok := tfMap[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &types.S3ExportingConfig{}

	if v, ok := tfList[names.AttrS3BucketName]; ok {
		apiObject.S3BucketName = aws.String(v.(string))
	}

	if v, ok := tfList["s3_key_name"]; ok {
		apiObject.S3KeyName = aws.String(v.(string))
	}

	return apiObject
}

func expandJobSchedule(tfMap []interface{}) *types.JobSchedule {
	if len(tfMap) == 0 {
		return nil
	}

	tfList, ok := tfMap[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &types.JobSchedule{}

	if v, ok := tfList["day_of_the_week"]; ok {
		apiObject.DayOfTheWeek = types.JobScheduleDayOfTheWeek(v.(string))
	}

	if v, ok := tfList["time"]; ok {
		apiObject.Time = aws.String(v.(string))
	}

	return apiObject
}

func expandRuleBasedMatching(tfMap []interface{}) *types.RuleBasedMatchingRequest {
	if len(tfMap) == 0 {
		return nil
	}

	tfList, ok := tfMap[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &types.RuleBasedMatchingRequest{}

	if v, ok := tfList[names.AttrEnabled]; ok {
		apiObject.Enabled = aws.Bool(v.(bool))
	}

	if v, ok := tfList["attribute_types_selector"]; ok {
		apiObject.AttributeTypesSelector = expandAttributesTypesSelector(v.([]interface{}))
	}

	if v, ok := tfList["conflict_resolution"]; ok {
		apiObject.ConflictResolution = expandConflictResolution(v.([]interface{}))
	}

	if v, ok := tfList["exporting_config"]; ok {
		apiObject.ExportingConfig = expandExportingConfig(v.([]interface{}))
	}

	if v, ok := tfList["matching_rules"]; ok {
		apiObject.MatchingRules = expandMatchingRules(v.(*schema.Set).List())
	}

	if v, ok := tfList["max_allowed_rule_level_for_matching"]; ok {
		apiObject.MaxAllowedRuleLevelForMatching = aws.Int32(int32(v.(int)))
	}

	if v, ok := tfList["max_allowed_rule_level_for_merging"]; ok {
		apiObject.MaxAllowedRuleLevelForMerging = aws.Int32(int32(v.(int)))
	}

	return apiObject
}

func expandAttributesTypesSelector(tfMap []interface{}) *types.AttributeTypesSelector {
	if len(tfMap) == 0 {
		return nil
	}

	tfList, ok := tfMap[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &types.AttributeTypesSelector{}

	if v, ok := tfList["attribute_matching_model"]; ok {
		apiObject.AttributeMatchingModel = types.AttributeMatchingModel(v.(string))
	}

	if v, ok := tfList[names.AttrAddress]; ok {
		apiObject.Address = flex.ExpandStringValueList(v.([]interface{}))
	}

	if v, ok := tfList["email_address"]; ok {
		apiObject.EmailAddress = flex.ExpandStringValueList(v.([]interface{}))
	}

	if v, ok := tfList["phone_number"]; ok {
		apiObject.PhoneNumber = flex.ExpandStringValueList(v.([]interface{}))
	}

	return apiObject
}

func expandMatchingAttributesList(tfMap []interface{}) [][]string {
	if len(tfMap) == 0 {
		return nil
	}

	result := make([][]string, 0, len(tfMap))

	for _, row := range tfMap {
		if row == nil {
			continue
		}
		result = append(result, flex.ExpandStringValueList(row.([]interface{})))
	}

	return result
}

func expandMatchingRules(tfMap []interface{}) []types.MatchingRule {
	if len(tfMap) == 0 {
		return nil
	}

	apiArray := make([]types.MatchingRule, 0, len(tfMap))

	for _, object := range tfMap {
		if object == nil {
			continue
		}

		matchingRule := object.(map[string]interface{})

		apiObject := types.MatchingRule{}

		if v, ok := matchingRule[names.AttrRule]; ok {
			apiObject.Rule = flex.ExpandStringValueList(v.([]interface{}))
		}

		apiArray = append(apiArray, apiObject)
	}

	return apiArray
}

func flattenMatching(apiObject *types.MatchingResponse) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AutoMerging; v != nil {
		tfMap["auto_merging"] = flattenAutoMerging(v)
	}

	if v := apiObject.Enabled; v != nil {
		tfMap[names.AttrEnabled] = aws.ToBool(v)
	}

	if v := apiObject.ExportingConfig; v != nil {
		tfMap["exporting_config"] = flattenExportingConfig(v)
	}

	if v := apiObject.JobSchedule; v != nil {
		tfMap["job_schedule"] = flattenJobSchedule(v)
	}

	return []interface{}{tfMap}
}

func flattenRuleBasedMatching(apiObject *types.RuleBasedMatchingResponse) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AttributeTypesSelector; v != nil {
		tfMap["attribute_types_selector"] = flattenAttributeTypesSelector(v)
	}

	if v := apiObject.ConflictResolution; v != nil {
		tfMap["conflict_resolution"] = flattenConflictResolution(v)
	}

	if v := apiObject.Enabled; v != nil {
		tfMap[names.AttrEnabled] = aws.ToBool(v)
	}

	if v := apiObject.ExportingConfig; v != nil {
		tfMap["exporting_config"] = flattenExportingConfig(v)
	}

	if v := apiObject.MatchingRules; v != nil {
		tfMap["matching_rules"] = flattenMatchingRules(v)
	}

	if v := apiObject.MaxAllowedRuleLevelForMatching; v != nil {
		tfMap["max_allowed_rule_level_for_matching"] = aws.ToInt32(v)
	}

	if v := apiObject.MaxAllowedRuleLevelForMerging; v != nil {
		tfMap["max_allowed_rule_level_for_merging"] = aws.ToInt32(v)
	}

	tfMap[names.AttrStatus] = types.IdentityResolutionJobStatus(apiObject.Status)

	return []interface{}{tfMap}
}

func flattenAutoMerging(apiObject *types.AutoMerging) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ConflictResolution; v != nil {
		tfMap["conflict_resolution"] = flattenConflictResolution(v)
	}

	if v := apiObject.Consolidation; v != nil {
		tfMap["consolidation"] = flattenConsolidation(v)
	}

	if v := apiObject.Enabled; v != nil {
		tfMap[names.AttrEnabled] = aws.ToBool(v)
	}

	if v := apiObject.MinAllowedConfidenceScoreForMerging; v != nil {
		tfMap["min_allowed_confidence_score_for_merging"] = aws.ToFloat64(v)
	}

	return []interface{}{tfMap}
}

func flattenConflictResolution(apiObject *types.ConflictResolution) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["conflict_resolving_model"] = apiObject.ConflictResolvingModel

	if v := apiObject.SourceName; v != nil {
		tfMap["source_name"] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

func flattenConsolidation(apiObject *types.Consolidation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.MatchingAttributesList; v != nil {
		tfMap["matching_attributes_list"] = v
	}

	return []interface{}{tfMap}
}

func flattenExportingConfig(apiObject *types.ExportingConfig) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.S3Exporting; v != nil {
		tfMap["s3_exporting"] = flattenS3Exporting(v)
	}

	return []interface{}{tfMap}
}

func flattenS3Exporting(apiObject *types.S3ExportingConfig) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.S3BucketName; v != nil {
		tfMap[names.AttrS3BucketName] = aws.ToString(v)
	}

	if v := apiObject.S3KeyName; v != nil {
		tfMap["s3_key_name"] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

func flattenJobSchedule(apiObject *types.JobSchedule) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["day_of_the_week"] = apiObject.DayOfTheWeek

	if v := apiObject.Time; v != nil {
		tfMap["time"] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

func flattenAttributeTypesSelector(apiObject *types.AttributeTypesSelector) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Address; v != nil {
		tfMap[names.AttrAddress] = flex.FlattenStringValueList(v)
	}

	tfMap["attribute_matching_model"] = apiObject.AttributeMatchingModel

	if v := apiObject.EmailAddress; v != nil {
		tfMap["email_address"] = flex.FlattenStringValueList(v)
	}

	if v := apiObject.PhoneNumber; v != nil {
		tfMap["phone_number"] = flex.FlattenStringValueList(v)
	}

	return []interface{}{tfMap}
}

func flattenMatchingRules(apiObject []types.MatchingRule) []interface{} {
	if apiObject == nil {
		return nil
	}

	var tfList []interface{}

	for _, matchingRule := range apiObject {
		if v := matchingRule.Rule; v != nil {
			tfMap := map[string]interface{}{}
			tfMap[names.AttrRule] = flex.FlattenStringValueList(v)
			tfList = append(tfList, tfMap)
		}
	}

	return tfList
}

// CreateDomainOutput does not have an ARN attribute which is needed for Tagging, therefore we construct it.
// https://docs.aws.amazon.com/service-authorization/latest/reference/list_amazonconnectcustomerprofiles.html#amazonconnectcustomerprofiles-resources-for-iam-policies
func buildDomainARN(conn *conns.AWSClient, domainName string) string {
	return fmt.Sprintf("arn:%s:profile:%s:%s:domains/%s", conn.Partition, conn.Region, conn.AccountID, domainName)
}
