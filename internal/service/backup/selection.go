// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_backup_selection")
func ResourceSelection() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSelectionCreate,
		ReadWithoutTimeout:   resourceSelectionRead,
		DeleteWithoutTimeout: resourceSelectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceSelectionImportState,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 50),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9\-\_\.]+$`), "must contain only alphanumeric, hyphen, underscore, and period characters"),
				),
			},
			"plan_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"condition": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"string_equals": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									"value": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
								},
							},
						},
						"string_like": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									"value": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
								},
							},
						},
						"string_not_equals": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									"value": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
								},
							},
						},
						"string_not_like": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									"value": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
								},
							},
						},
					},
				},
			},
			"iam_role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"selection_tag": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								backup.ConditionTypeStringequals,
							}, false),
						},
						"key": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			"not_resources": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"resources": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceSelectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupConn(ctx)

	selection := &backup.Selection{
		Conditions:    expandConditions(d.Get("condition").(*schema.Set).List()),
		IamRoleArn:    aws.String(d.Get("iam_role_arn").(string)),
		ListOfTags:    expandConditionTags(d.Get("selection_tag").(*schema.Set).List()),
		NotResources:  flex.ExpandStringSet(d.Get("not_resources").(*schema.Set)),
		Resources:     flex.ExpandStringSet(d.Get("resources").(*schema.Set)),
		SelectionName: aws.String(d.Get("name").(string)),
	}

	input := &backup.CreateBackupSelectionInput{
		BackupPlanId:    aws.String(d.Get("plan_id").(string)),
		BackupSelection: selection,
	}

	// Retry for IAM eventual consistency
	var output *backup.CreateBackupSelectionOutput
	err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		var err error
		output, err = conn.CreateBackupSelectionWithContext(ctx, input)

		// Retry on the following error:
		// InvalidParameterValueException: IAM Role arn:aws:iam::123456789012:role/XXX cannot be assumed by AWS Backup
		if tfawserr.ErrMessageContains(err, backup.ErrCodeInvalidParameterValueException, "cannot be assumed") {
			log.Printf("[DEBUG] Received %s, retrying create backup selection.", err)
			return retry.RetryableError(err)
		}

		// Retry on the following error:
		// InvalidParameterValueException: IAM Role arn:aws:iam::123456789012:role/XXX is not authorized to call tag:GetResources
		if tfawserr.ErrMessageContains(err, backup.ErrCodeInvalidParameterValueException, "is not authorized to call") {
			log.Printf("[DEBUG] Received %s, retrying create backup selection.", err)
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateBackupSelectionWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Backup Selection: %s", err)
	}

	d.SetId(aws.StringValue(output.SelectionId))

	return append(diags, resourceSelectionRead(ctx, d, meta)...)
}

func resourceSelectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupConn(ctx)

	input := &backup.GetBackupSelectionInput{
		BackupPlanId: aws.String(d.Get("plan_id").(string)),
		SelectionId:  aws.String(d.Id()),
	}

	var resp *backup.GetBackupSelectionOutput

	err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		var err error

		resp, err = conn.GetBackupSelectionWithContext(ctx, input)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, backup.ErrCodeResourceNotFoundException) {
			return retry.RetryableError(err)
		}

		if d.IsNewResource() && tfawserr.ErrMessageContains(err, backup.ErrCodeInvalidParameterValueException, "Cannot find Backup plan") {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		resp, err = conn.GetBackupSelectionWithContext(ctx, input)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, backup.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Backup Selection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if !d.IsNewResource() && tfawserr.ErrMessageContains(err, backup.ErrCodeInvalidParameterValueException, "Cannot find Backup plan") {
		log.Printf("[WARN] Backup Selection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Backup Selection (%s): %s", d.Id(), err)
	}

	if resp == nil {
		return sdkdiag.AppendErrorf(diags, "reading Backup Selection (%s): empty response", d.Id())
	}

	d.Set("plan_id", resp.BackupPlanId)
	d.Set("name", resp.BackupSelection.SelectionName)
	d.Set("iam_role_arn", resp.BackupSelection.IamRoleArn)

	if conditions := resp.BackupSelection.Conditions; conditions != nil {
		if err := d.Set("condition", flattenConditions(conditions)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting conditions: %s", err)
		}
	}

	if resp.BackupSelection.ListOfTags != nil {
		tags := make([]map[string]interface{}, 0)

		for _, r := range resp.BackupSelection.ListOfTags {
			m := make(map[string]interface{})

			m["type"] = aws.StringValue(r.ConditionType)
			m["key"] = aws.StringValue(r.ConditionKey)
			m["value"] = aws.StringValue(r.ConditionValue)

			tags = append(tags, m)
		}

		if err := d.Set("selection_tag", tags); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting selection tag: %s", err)
		}
	}

	if resp.BackupSelection.Resources != nil {
		if err := d.Set("resources", aws.StringValueSlice(resp.BackupSelection.Resources)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting resources: %s", err)
		}
	}

	if resp.BackupSelection.NotResources != nil {
		if err := d.Set("not_resources", aws.StringValueSlice(resp.BackupSelection.NotResources)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting not resources: %s", err)
		}
	}

	return diags
}

func resourceSelectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupConn(ctx)

	input := &backup.DeleteBackupSelectionInput{
		BackupPlanId: aws.String(d.Get("plan_id").(string)),
		SelectionId:  aws.String(d.Id()),
	}

	_, err := conn.DeleteBackupSelectionWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Backup Selection: %s", err)
	}

	return diags
}

func resourceSelectionImportState(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.Split(d.Id(), "|")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%q), expected <plan-id>|<selection-id>", d.Id())
	}

	planID := idParts[0]
	selectionID := idParts[1]

	d.Set("plan_id", planID)
	d.SetId(selectionID)

	return []*schema.ResourceData{d}, nil
}

func expandConditionTags(tagList []interface{}) []*backup.Condition {
	conditions := []*backup.Condition{}

	for _, i := range tagList {
		item := i.(map[string]interface{})
		tag := &backup.Condition{}

		tag.ConditionType = aws.String(item["type"].(string))
		tag.ConditionKey = aws.String(item["key"].(string))
		tag.ConditionValue = aws.String(item["value"].(string))

		conditions = append(conditions, tag)
	}

	return conditions
}

func expandConditions(conditionsList []interface{}) *backup.Conditions {
	conditions := &backup.Conditions{}

	for _, condition := range conditionsList {
		mCondition := condition.(map[string]interface{})

		if vStringEquals := expandConditionParameters(mCondition["string_equals"].(*schema.Set).List()); len(vStringEquals) > 0 {
			conditions.StringEquals = vStringEquals
		}
		if vStringNotEquals := expandConditionParameters(mCondition["string_not_equals"].(*schema.Set).List()); len(vStringNotEquals) > 0 {
			conditions.StringNotEquals = vStringNotEquals
		}
		if vStringLike := expandConditionParameters(mCondition["string_like"].(*schema.Set).List()); len(vStringLike) > 0 {
			conditions.StringLike = vStringLike
		}
		if vStringNotLike := expandConditionParameters(mCondition["string_not_like"].(*schema.Set).List()); len(vStringNotLike) > 0 {
			conditions.StringNotLike = vStringNotLike
		}
	}

	return conditions
}

func expandConditionParameters(conditionParametersList []interface{}) []*backup.ConditionParameter {
	conditionParameters := []*backup.ConditionParameter{}

	for _, i := range conditionParametersList {
		item := i.(map[string]interface{})
		conditionParameter := &backup.ConditionParameter{}

		conditionParameter.ConditionKey = aws.String(item["key"].(string))
		conditionParameter.ConditionValue = aws.String(item["value"].(string))

		conditionParameters = append(conditionParameters, conditionParameter)
	}

	return conditionParameters
}

func flattenConditions(conditions *backup.Conditions) *schema.Set {
	var vConditions []interface{}

	mCondition := map[string]interface{}{}

	mCondition["string_equals"] = flattenConditionParameters(conditions.StringEquals)
	mCondition["string_not_equals"] = flattenConditionParameters(conditions.StringNotEquals)
	mCondition["string_like"] = flattenConditionParameters(conditions.StringLike)
	mCondition["string_not_like"] = flattenConditionParameters(conditions.StringNotLike)

	vConditions = append(vConditions, mCondition)

	return schema.NewSet(conditionsHash, vConditions)
}

func conditionsHash(vCondition interface{}) int {
	var buf bytes.Buffer

	mCondition := vCondition.(map[string]interface{})

	if v, ok := mCondition["string_equals"].(string); ok {
		buf.WriteString(fmt.Sprintf("%s-", v))
	}

	if v, ok := mCondition["string_not_equals"].(string); ok {
		buf.WriteString(fmt.Sprintf("%s-", v))
	}

	if v, ok := mCondition["string_like"].(string); ok {
		buf.WriteString(fmt.Sprintf("%s-", v))
	}

	if v, ok := mCondition["string_not_like"].(string); ok {
		buf.WriteString(fmt.Sprintf("%s-", v))
	}

	return create.StringHashcode(buf.String())
}

func flattenConditionParameters(conditionParameters []*backup.ConditionParameter) []interface{} {
	if len(conditionParameters) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, conditionParameter := range conditionParameters {
		if conditionParameter == nil {
			continue
		}

		tfMap := map[string]interface{}{
			"key":   aws.StringValue(conditionParameter.ConditionKey),
			"value": aws.StringValue(conditionParameter.ConditionValue),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
