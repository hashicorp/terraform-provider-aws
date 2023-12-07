// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_securityhub_automation_rule")
// @Tags(identifierAttribute="arn")
func ResourceAutomationRule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAutomationRuleCreate,
		ReadWithoutTimeout:   resourceAutomationRuleRead,
		UpdateWithoutTimeout: resourceAutomationRuleUpdate,
		DeleteWithoutTimeout: resourceAutomationRuleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"actions": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"finding_fields_update": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"confidence": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
									"criticality": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
									"note": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"text": {
													Type:     schema.TypeString,
													Required: true,
												},
												"updated_by": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
									"related_findings": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"id": {
													Type:     schema.TypeString,
													Required: true,
												},
												"product_arn": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: verify.ValidARN,
												},
											},
										},
									},
									"severity": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"label": {
													Type:             schema.TypeString,
													Optional:         true,
													Computed:         true,
													ValidateDiagFunc: enum.Validate[types.SeverityLabel](),
												},
												"product": {
													Type:     schema.TypeFloat,
													Optional: true,
												},
											},
										},
									},
									"types": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"user_defined_fields": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"verification_state": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.VerificationState](),
									},
									"workflow": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"status": {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[types.WorkflowStatus](),
												},
											},
										},
									},
								},
							},
						},
						"type": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[types.AutomationRulesActionType](),
						},
					},
				},
			},
			"criteria": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"aws_account_id":                     stringFilterSchema(),
						"aws_account_name":                   stringFilterSchema(),
						"company_name":                       stringFilterSchema(),
						"compliance_associated_standards_id": stringFilterSchema(),
						"compliance_security_control_id":     stringFilterSchema(),
						"compliance_status":                  stringFilterSchema(),
						"confidence":                         numberFilterSchema(),
						"created_at":                         dateFilterSchema(),
						"criticality":                        numberFilterSchema(),
						"description":                        stringFilterSchema(),
						"first_observed_at":                  dateFilterSchema(),
						"generator_id":                       stringFilterSchema(),
						"id":                                 stringFilterSchema(),
						"last_observed_at":                   dateFilterSchema(),
						"note_text":                          stringFilterSchema(),
						"note_updated_at":                    dateFilterSchema(),
						"note_updated_by":                    stringFilterSchema(),
						"product_arn":                        stringFilterSchema(),
						"product_name":                       stringFilterSchema(),
						"record_state":                       stringFilterSchema(),
						"related_findings_id":                stringFilterSchema(),
						"related_findings_product_arn":       stringFilterSchema(),
						"resource_application_arn":           stringFilterSchema(),
						"resource_application_name":          stringFilterSchema(),
						"resource_details_other":             mapFilterSchema(),
						"resource_id":                        stringFilterSchema(),
						"resource_partition":                 stringFilterSchema(),
						"resource_region":                    stringFilterSchema(),
						"resource_tags":                      mapFilterSchema(),
						"resource_type":                      stringFilterSchema(),
						"severity_label":                     stringFilterSchema(),
						"source_url":                         stringFilterSchema(),
						"title":                              stringFilterSchema(),
						"type":                               stringFilterSchema(),
						"updated_at":                         dateFilterSchema(),
						"user_defined_fields":                mapFilterSchema(),
						"verification_state":                 stringFilterSchema(),
						"workflow_status":                    stringFilterSchema(),
					},
				},
			},
			"description": {
				Type:     schema.TypeString,
				Required: true,
			},
			"is_terminal": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"rule_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"rule_order": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"rule_status": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[types.RuleStatus](),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAutomationRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	ruleName := d.Get("rule_name").(string)
	input := &securityhub.CreateAutomationRuleInput{
		Criteria:    expandCriteria(d.Get("criteria").([]interface{})),
		Description: aws.String(d.Get("description").(string)),
		IsTerminal:  aws.Bool(d.Get("is_terminal").(bool)),
		RuleName:    aws.String(ruleName),
		RuleOrder:   aws.Int32(int32(d.Get("rule_order").(int))),
		Tags:        getTagsIn(ctx),
	}

	if v, ok := d.GetOk("actions"); ok {
		input.Actions = expandActions(v.([]interface{}))
	}

	if v, ok := d.GetOk("rule_status"); ok {
		input.RuleStatus = types.RuleStatus(v.(string))
	}

	output, err := conn.CreateAutomationRule(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Security Hub Automation Rule (%s): %s", ruleName, err)
	}

	d.SetId(aws.ToString(output.RuleArn))

	return append(diags, resourceAutomationRuleRead(ctx, d, meta)...)
}

func resourceAutomationRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	output, err := FindAutomationRuleByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Security Hub Product Automation Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Security Hub Automation Rule (%s): %s", d.Id(), err)
	}

	d.Set("arn", output.RuleArn)
	d.Set("actions", flattenActions(output.Actions))
	d.Set("criteria", flattenCriteria(output.Criteria))
	d.Set("description", output.Description)
	d.Set("is_terminal", output.IsTerminal)
	d.Set("rule_name", output.RuleName)
	d.Set("rule_order", output.RuleOrder)
	d.Set("rule_status", output.RuleStatus)

	return diags
}

func resourceAutomationRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &securityhub.BatchUpdateAutomationRulesInput{}
		automationRuleItem := types.UpdateAutomationRulesRequestItem{
			RuleArn: aws.String(d.Id()),
		}

		if d.HasChange("actions") {
			automationRuleItem.Actions = expandActions(d.Get("actions").([]interface{}))
		}

		if d.HasChange("criteria") {
			automationRuleItem.Criteria = expandCriteria(d.Get("criteria").([]interface{}))
		}

		if d.HasChange("description") {
			automationRuleItem.Description = aws.String(d.Get("description").(string))
		}

		if d.HasChange("is_terminal") {
			automationRuleItem.IsTerminal = aws.Bool(d.Get("is_terminal").(bool))
		}

		if d.HasChange("rule_name") {
			automationRuleItem.RuleName = aws.String(d.Get("rule_name").(string))
		}

		if d.HasChange("rule_order") {
			automationRuleItem.RuleOrder = aws.Int32(int32(d.Get("rule_order").(int)))
		}

		if d.HasChange("rule_status") {
			automationRuleItem.RuleStatus = types.RuleStatus(d.Get("rule_status").(string))
		}

		input.UpdateAutomationRulesRequestItems = append(input.UpdateAutomationRulesRequestItems, automationRuleItem)

		_, err := conn.BatchUpdateAutomationRules(ctx, input)

		if err != nil {
			return diag.Errorf("updating Security Hub Automation Rule (%s): %s", d.Id(), err)
		}
	}

	return resourceAutomationRuleRead(ctx, d, meta)
}

func resourceAutomationRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	log.Printf("[DEBUG] Deleting Security Hub Automation Rule: %s", d.Id())
	_, err := conn.BatchDeleteAutomationRules(ctx, &securityhub.BatchDeleteAutomationRulesInput{
		AutomationRulesArns: []string{d.Id()},
	})

	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Security Hub Automatuon Rule (%s): %s", d.Id(), err)
	}

	return diags
}

func FindAutomationRuleByARN(ctx context.Context, conn *securityhub.Client, arn string) (*types.AutomationRulesConfig, error) {
	input := &securityhub.BatchGetAutomationRulesInput{
		AutomationRulesArns: []string{arn},
	}

	output, err := conn.BatchGetAutomationRules(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) || tfawserr.ErrMessageContains(err, errCodeInvalidAccessException, "not subscribed to AWS Security Hub") {
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

	return tfresource.AssertSingleValueResult(output.Rules)
}

func expandActions(l []interface{}) []types.AutomationRulesAction {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	var output []types.AutomationRulesAction

	for _, object := range l {
		if object == nil {
			continue
		}

		action := object.(map[string]interface{})
		apiObject := types.AutomationRulesAction{}

		if v, ok := action["finding_fields_update"]; ok {
			apiObject.FindingFieldsUpdate = expandFindingFieldsUpdate(v.([]interface{}))
		}

		if v, ok := action["type"]; ok {
			apiObject.Type = types.AutomationRulesActionType(v.(string))
		}
		output = append(output, apiObject)
	}

	return output
}

func expandCriteria(l []interface{}) *types.AutomationRulesFindingFilters {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	output := types.AutomationRulesFindingFilters{}

	if v, ok := tfMap["aws_account_id"].(*schema.Set); ok && v.Len() > 0 {
		output.AwsAccountId = expandStringFilters(v.List())
	}

	if v, ok := tfMap["aws_account_name"].(*schema.Set); ok && v.Len() > 0 {
		output.AwsAccountName = expandStringFilters(v.List())
	}

	if v, ok := tfMap["company_name"].(*schema.Set); ok && v.Len() > 0 {
		output.CompanyName = expandStringFilters(v.List())
	}

	if v, ok := tfMap["compliance_associated_standards_id"].(*schema.Set); ok && v.Len() > 0 {
		output.ComplianceAssociatedStandardsId = expandStringFilters(v.List())
	}

	if v, ok := tfMap["compliance_security_control_id"].(*schema.Set); ok && v.Len() > 0 {
		output.ComplianceSecurityControlId = expandStringFilters(v.List())
	}

	if v, ok := tfMap["compliance_status"].(*schema.Set); ok && v.Len() > 0 {
		output.ComplianceStatus = expandStringFilters(v.List())
	}

	if v, ok := tfMap["confidence"].(*schema.Set); ok && v.Len() > 0 {
		output.Confidence = expandNumberFilters(v.List())
	}

	if v, ok := tfMap["created_at"].(*schema.Set); ok && v.Len() > 0 {
		output.CreatedAt = expandDateFilters(v.List())
	}

	if v, ok := tfMap["criticality"].(*schema.Set); ok && v.Len() > 0 {
		output.Criticality = expandNumberFilters(v.List())
	}

	if v, ok := tfMap["description"].(*schema.Set); ok && v.Len() > 0 {
		output.Description = expandStringFilters(v.List())
	}

	if v, ok := tfMap["first_observed_at"].(*schema.Set); ok && v.Len() > 0 {
		output.FirstObservedAt = expandDateFilters(v.List())
	}

	if v, ok := tfMap["generator_id"].(*schema.Set); ok && v.Len() > 0 {
		output.GeneratorId = expandStringFilters(v.List())
	}

	if v, ok := tfMap["id"].(*schema.Set); ok && v.Len() > 0 {
		output.Id = expandStringFilters(v.List())
	}

	if v, ok := tfMap["last_observed_at"].(*schema.Set); ok && v.Len() > 0 {
		output.LastObservedAt = expandDateFilters(v.List())
	}

	if v, ok := tfMap["note_text"].(*schema.Set); ok && v.Len() > 0 {
		output.NoteText = expandStringFilters(v.List())
	}

	if v, ok := tfMap["note_updated_at"].(*schema.Set); ok && v.Len() > 0 {
		output.NoteUpdatedAt = expandDateFilters(v.List())
	}

	if v, ok := tfMap["note_updated_by"].(*schema.Set); ok && v.Len() > 0 {
		output.NoteUpdatedBy = expandStringFilters(v.List())
	}

	if v, ok := tfMap["product_arn"].(*schema.Set); ok && v.Len() > 0 {
		output.ProductArn = expandStringFilters(v.List())
	}

	if v, ok := tfMap["product_name"].(*schema.Set); ok && v.Len() > 0 {
		output.ProductName = expandStringFilters(v.List())
	}

	if v, ok := tfMap["record_state"].(*schema.Set); ok && v.Len() > 0 {
		output.RecordState = expandStringFilters(v.List())
	}

	if v, ok := tfMap["related_findings_id"].(*schema.Set); ok && v.Len() > 0 {
		output.RelatedFindingsId = expandStringFilters(v.List())
	}

	if v, ok := tfMap["related_findings_product_arn"].(*schema.Set); ok && v.Len() > 0 {
		output.RelatedFindingsProductArn = expandStringFilters(v.List())
	}

	if v, ok := tfMap["resource_application_arn"].(*schema.Set); ok && v.Len() > 0 {
		output.ResourceApplicationArn = expandStringFilters(v.List())
	}

	if v, ok := tfMap["resource_application_name"].(*schema.Set); ok && v.Len() > 0 {
		output.ResourceApplicationName = expandStringFilters(v.List())
	}

	if v, ok := tfMap["resource_details_other"].(*schema.Set); ok && v.Len() > 0 {
		output.ResourceDetailsOther = expandMapFilters(v.List())
	}

	if v, ok := tfMap["resource_id"].(*schema.Set); ok && v.Len() > 0 {
		output.ResourceId = expandStringFilters(v.List())
	}

	if v, ok := tfMap["resource_partition"].(*schema.Set); ok && v.Len() > 0 {
		output.ResourcePartition = expandStringFilters(v.List())
	}

	if v, ok := tfMap["resource_region"].(*schema.Set); ok && v.Len() > 0 {
		output.ResourceRegion = expandStringFilters(v.List())
	}

	if v, ok := tfMap["resource_tags"].(*schema.Set); ok && v.Len() > 0 {
		output.ResourceTags = expandMapFilters(v.List())
	}

	if v, ok := tfMap["resource_type"].(*schema.Set); ok && v.Len() > 0 {
		output.ResourceType = expandStringFilters(v.List())
	}

	if v, ok := tfMap["severity_label"].(*schema.Set); ok && v.Len() > 0 {
		output.SeverityLabel = expandStringFilters(v.List())
	}

	if v, ok := tfMap["source_url"].(*schema.Set); ok && v.Len() > 0 {
		output.SourceUrl = expandStringFilters(v.List())
	}

	if v, ok := tfMap["title"].(*schema.Set); ok && v.Len() > 0 {
		output.Title = expandStringFilters(v.List())
	}

	if v, ok := tfMap["type"].(*schema.Set); ok && v.Len() > 0 {
		output.Type = expandStringFilters(v.List())
	}

	if v, ok := tfMap["updated_at"].(*schema.Set); ok && v.Len() > 0 {
		output.UpdatedAt = expandDateFilters(v.List())
	}

	if v, ok := tfMap["user_defined_fields"].(*schema.Set); ok && v.Len() > 0 {
		output.UserDefinedFields = expandMapFilters(v.List())
	}

	if v, ok := tfMap["verification_state"].(*schema.Set); ok && v.Len() > 0 {
		output.VerificationState = expandStringFilters(v.List())
	}

	if v, ok := tfMap["workflow_status"].(*schema.Set); ok && v.Len() > 0 {
		output.WorkflowStatus = expandStringFilters(v.List())
	}

	return &output
}

func expandFindingFieldsUpdate(l []interface{}) *types.AutomationRulesFindingFieldsUpdate {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	output := &types.AutomationRulesFindingFieldsUpdate{}

	if v, ok := tfMap["confidence"]; ok {
		output.Confidence = aws.Int32(int32(v.(int)))
	}

	if v, ok := tfMap["criticality"]; ok {
		output.Criticality = aws.Int32(int32(v.(int)))
	}

	if v, ok := tfMap["note"]; ok {
		output.Note = expandNote(v.([]interface{}))
	}

	if v, ok := tfMap["related_findings"].(*schema.Set); ok && v.Len() > 0 {
		output.RelatedFindings = expandRelatedFindings(v.List())
	}

	if v, ok := tfMap["severity"]; ok {
		output.Severity = expandSeverity(v.([]interface{}))
	}

	if v, ok := tfMap["types"]; ok {
		output.Types = flex.ExpandStringValueList(v.([]interface{}))
	}

	if v, ok := tfMap["user_defined_fields"]; ok {
		output.UserDefinedFields = flex.ExpandStringValueMap(v.(map[string]interface{}))
	}

	if v, ok := tfMap["verification_state"]; ok {
		output.VerificationState = types.VerificationState(v.(string))
	}

	if v, ok := tfMap["workflow"]; ok {
		output.Workflow = expandWorkflow(v.([]interface{}))
	}

	return output
}

func expandNote(l []interface{}) *types.NoteUpdate {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	output := &types.NoteUpdate{}

	if v, ok := tfMap["text"]; ok {
		output.Text = aws.String(v.(string))
	}

	if v, ok := tfMap["updated_by"]; ok {
		output.UpdatedBy = aws.String(v.(string))
	}

	return output
}

func expandRelatedFindings(l []interface{}) []types.RelatedFinding {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	var output []types.RelatedFinding

	for _, item := range l {
		tfMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		relatedFinding := types.RelatedFinding{}

		if v, ok := tfMap["id"]; ok {
			relatedFinding.Id = aws.String(v.(string))
		}

		if v, ok := tfMap["product_arn"]; ok {
			relatedFinding.ProductArn = aws.String(v.(string))
		}

		output = append(output, relatedFinding)
	}
	return output
}

func expandSeverity(l []interface{}) *types.SeverityUpdate {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	output := &types.SeverityUpdate{}

	if v, ok := tfMap["label"]; ok {
		output.Label = types.SeverityLabel(v.(string))
	}

	if v, ok := tfMap["product"]; ok {
		output.Product = aws.Float64(v.(float64))
	}

	return output
}

func expandWorkflow(l []interface{}) *types.WorkflowUpdate {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	output := &types.WorkflowUpdate{}

	if v, ok := tfMap["status"]; ok {
		output.Status = types.WorkflowStatus(v.(string))
	}

	return output
}

func flattenActions(apiObject []types.AutomationRulesAction) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var output []interface{}

	for _, action := range apiObject {
		tfMap := map[string]interface{}{
			"type": string(action.Type),
		}

		if v := action.FindingFieldsUpdate; v != nil {
			tfMap["finding_fields_update"] = flattenFindingFieldsUpdate(v)
		}

		output = append(output, tfMap)
	}

	return output
}

func flattenCriteria(apiObject *types.AutomationRulesFindingFilters) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AwsAccountId; v != nil {
		tfMap["aws_account_id"] = flattenStringFilters(apiObject.AwsAccountId)
	}

	if v := apiObject.AwsAccountName; v != nil {
		tfMap["aws_account_name"] = flattenStringFilters(apiObject.AwsAccountName)
	}

	if v := apiObject.CompanyName; v != nil {
		tfMap["company_name"] = flattenStringFilters(apiObject.CompanyName)
	}

	if v := apiObject.ComplianceAssociatedStandardsId; v != nil {
		tfMap["compliance_associated_standards_id"] = flattenStringFilters(apiObject.ComplianceAssociatedStandardsId)
	}

	if v := apiObject.ComplianceSecurityControlId; v != nil {
		tfMap["compliance_security_control_id"] = flattenStringFilters(apiObject.ComplianceSecurityControlId)
	}

	if v := apiObject.ComplianceStatus; v != nil {
		tfMap["compliance_status"] = flattenStringFilters(apiObject.ComplianceStatus)
	}

	if v := apiObject.Confidence; v != nil {
		tfMap["confidence"] = flattenNumberFilters(apiObject.Confidence)
	}

	if v := apiObject.CreatedAt; v != nil {
		tfMap["created_at"] = flattenDateFilters(apiObject.CreatedAt)
	}

	if v := apiObject.Criticality; v != nil {
		tfMap["criticality"] = flattenNumberFilters(apiObject.Criticality)
	}

	if v := apiObject.Description; v != nil {
		tfMap["description"] = flattenStringFilters(apiObject.Description)
	}

	if v := apiObject.FirstObservedAt; v != nil {
		tfMap["first_observed_at"] = flattenDateFilters(apiObject.FirstObservedAt)
	}

	if v := apiObject.GeneratorId; v != nil {
		tfMap["generator_id"] = flattenStringFilters(apiObject.GeneratorId)
	}

	if v := apiObject.Id; v != nil {
		tfMap["id"] = flattenStringFilters(apiObject.Id)
	}

	if v := apiObject.LastObservedAt; v != nil {
		tfMap["last_observed_at"] = flattenDateFilters(apiObject.LastObservedAt)
	}

	if v := apiObject.NoteText; v != nil {
		tfMap["note_text"] = flattenStringFilters(apiObject.NoteText)
	}

	if v := apiObject.NoteUpdatedAt; v != nil {
		tfMap["note_updated_at"] = flattenDateFilters(apiObject.NoteUpdatedAt)
	}

	if v := apiObject.NoteUpdatedBy; v != nil {
		tfMap["note_updated_by"] = flattenStringFilters(apiObject.NoteUpdatedBy)
	}

	if v := apiObject.ProductArn; v != nil {
		tfMap["product_arn"] = flattenStringFilters(apiObject.ProductArn)
	}

	if v := apiObject.ProductName; v != nil {
		tfMap["product_name"] = flattenStringFilters(apiObject.ProductName)
	}

	if v := apiObject.RecordState; v != nil {
		tfMap["record_state"] = flattenStringFilters(apiObject.RecordState)
	}

	if v := apiObject.RelatedFindingsId; v != nil {
		tfMap["related_findings_id"] = flattenStringFilters(apiObject.RelatedFindingsId)
	}

	if v := apiObject.RelatedFindingsProductArn; v != nil {
		tfMap["related_findings_product_arn"] = flattenStringFilters(apiObject.RelatedFindingsProductArn)
	}

	if v := apiObject.ResourceApplicationArn; v != nil {
		tfMap["resource_application_arn"] = flattenStringFilters(apiObject.ResourceApplicationArn)
	}

	if v := apiObject.ResourceApplicationName; v != nil {
		tfMap["resource_application_name"] = flattenStringFilters(apiObject.ResourceApplicationName)
	}

	if v := apiObject.ResourceDetailsOther; v != nil {
		tfMap["resource_details_other"] = flattenMapFilters(apiObject.ResourceDetailsOther)
	}

	if v := apiObject.ResourceId; v != nil {
		tfMap["resource_id"] = flattenStringFilters(apiObject.ResourceId)
	}

	if v := apiObject.ResourcePartition; v != nil {
		tfMap["resource_partition"] = flattenStringFilters(apiObject.ResourcePartition)
	}

	if v := apiObject.ResourceRegion; v != nil {
		tfMap["resource_region"] = flattenStringFilters(apiObject.ResourceRegion)
	}

	if v := apiObject.ResourceTags; v != nil {
		tfMap["resource_tags"] = flattenMapFilters(apiObject.ResourceTags)
	}

	if v := apiObject.ResourceType; v != nil {
		tfMap["resource_type"] = flattenStringFilters(apiObject.ResourceType)
	}

	if v := apiObject.SeverityLabel; v != nil {
		tfMap["severity_label"] = flattenStringFilters(apiObject.SeverityLabel)
	}

	if v := apiObject.SourceUrl; v != nil {
		tfMap["source_url"] = flattenStringFilters(apiObject.SourceUrl)
	}

	if v := apiObject.Title; v != nil {
		tfMap["title"] = flattenStringFilters(apiObject.Title)
	}

	if v := apiObject.Type; v != nil {
		tfMap["type"] = flattenStringFilters(apiObject.Type)
	}

	if v := apiObject.UpdatedAt; v != nil {
		tfMap["updated_at"] = flattenDateFilters(apiObject.UpdatedAt)
	}

	if v := apiObject.UserDefinedFields; v != nil {
		tfMap["user_defined_fields"] = flattenMapFilters(apiObject.UserDefinedFields)
	}

	if v := apiObject.VerificationState; v != nil {
		tfMap["verification_state"] = flattenStringFilters(apiObject.VerificationState)
	}

	if v := apiObject.WorkflowStatus; v != nil {
		tfMap["workflow_status"] = flattenStringFilters(apiObject.WorkflowStatus)
	}

	return []interface{}{tfMap}
}

func flattenFindingFieldsUpdate(apiObject *types.AutomationRulesFindingFieldsUpdate) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"verification_state": string(apiObject.VerificationState),
	}

	if v := apiObject.Confidence; v != nil {
		tfMap["confidence"] = int(aws.ToInt32(v))
	}

	if v := apiObject.Criticality; v != nil {
		tfMap["criticality"] = int(aws.ToInt32(v))
	}

	if v := apiObject.Note; v != nil {
		tfMap["note"] = flattenNotes(v)
	}

	if v := apiObject.RelatedFindings; v != nil {
		tfMap["related_findings"] = flattenRelatedFindings(v)
	}

	if v := apiObject.Severity; v != nil {
		tfMap["severity"] = flattenSeverity(v)
	}

	if v := apiObject.Types; v != nil {
		tfMap["types"] = flex.FlattenStringValueList(v)
	}

	if v := apiObject.UserDefinedFields; v != nil {
		tfMap["user_defined_fields"] = v
	}

	if v := apiObject.Workflow; v != nil {
		tfMap["workflow"] = flattenWorkflow(v)
	}

	return []interface{}{tfMap}
}

func flattenNotes(apiObject *types.NoteUpdate) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Text; v != nil {
		tfMap["text"] = aws.ToString(v)
	}

	if v := apiObject.UpdatedBy; v != nil {
		tfMap["updated_by"] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

func flattenRelatedFindings(apiObject []types.RelatedFinding) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var output []interface{}

	for _, relatedFinding := range apiObject {
		tfMap := map[string]interface{}{}

		if v := relatedFinding.Id; v != nil {
			tfMap["id"] = aws.ToString(v)
		}

		if v := relatedFinding.ProductArn; v != nil {
			tfMap["product_arn"] = aws.ToString(v)
		}

		output = append(output, tfMap)
	}

	return output
}

func flattenSeverity(apiObject *types.SeverityUpdate) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"label": string(apiObject.Label),
	}

	if v := apiObject.Product; v != nil {
		tfMap["product"] = aws.ToFloat64(v)
	}

	return []interface{}{tfMap}
}

func flattenWorkflow(apiObject *types.WorkflowUpdate) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"status": string(apiObject.Status),
	}

	return []interface{}{tfMap}
}
