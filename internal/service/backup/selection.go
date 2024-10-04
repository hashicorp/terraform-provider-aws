// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	awstypes "github.com/aws/aws-sdk-go-v2/service/backup/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_backup_selection", name="Selection")
func resourceSelection() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSelectionCreate,
		ReadWithoutTimeout:   resourceSelectionRead,
		DeleteWithoutTimeout: resourceSelectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceSelectionImportState,
		},

		Schema: map[string]*schema.Schema{
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 50),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+$`), "must contain only alphanumeric, hyphen, underscore, and period characters"),
				),
			},
			names.AttrCondition: {
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
									names.AttrKey: {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									names.AttrValue: {
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
									names.AttrKey: {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									names.AttrValue: {
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
									names.AttrKey: {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									names.AttrValue: {
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
									names.AttrKey: {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									names.AttrValue: {
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
			names.AttrIAMRoleARN: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"not_resources": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"plan_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"selection_tag": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrKey: {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						names.AttrType: {
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ConditionType](),
						},
						names.AttrValue: {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			names.AttrResources: {
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
	conn := meta.(*conns.AWSClient).BackupClient(ctx)

	name := d.Get(names.AttrName).(string)
	planID := d.Get("plan_id").(string)
	input := &backup.CreateBackupSelectionInput{
		BackupPlanId: aws.String(planID),
		BackupSelection: &awstypes.BackupSelection{
			Conditions:    expandConditions(d.Get(names.AttrCondition).(*schema.Set).List()),
			IamRoleArn:    aws.String(d.Get(names.AttrIAMRoleARN).(string)),
			ListOfTags:    expandConditionTags(d.Get("selection_tag").(*schema.Set).List()),
			NotResources:  flex.ExpandStringValueSet(d.Get("not_resources").(*schema.Set)),
			Resources:     flex.ExpandStringValueSet(d.Get(names.AttrResources).(*schema.Set)),
			SelectionName: aws.String(name),
		},
	}

	// Retry for IAM eventual consistency.
	outputRaw, err := tfresource.RetryWhen(ctx, propagationTimeout,
		func() (interface{}, error) {
			return conn.CreateBackupSelection(ctx, input)
		},
		func(err error) (bool, error) {
			// InvalidParameterValueException: IAM Role arn:aws:iam::123456789012:role/XXX cannot be assumed by AWS Backup
			if errs.IsAErrorMessageContains[*awstypes.InvalidParameterValueException](err, "cannot be assumed") {
				return true, err
			}

			// InvalidParameterValueException: IAM Role arn:aws:iam::123456789012:role/XXX is not authorized to call tag:GetResources
			if errs.IsAErrorMessageContains[*awstypes.InvalidParameterValueException](err, "is not authorized to call") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Backup Selection (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*backup.CreateBackupSelectionOutput).SelectionId))

	const (
		// Maximum amount of time to wait for Backup changes to propagate.
		timeout = 2 * time.Minute
	)
	_, err = tfresource.RetryWhenNotFound(ctx, timeout, func() (interface{}, error) {
		return findSelectionByTwoPartKey(ctx, conn, planID, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Backup Selection (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceSelectionRead(ctx, d, meta)...)
}

func resourceSelectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupClient(ctx)

	planID := d.Get("plan_id").(string)
	output, err := findSelectionByTwoPartKey(ctx, conn, planID, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Backup Selection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Backup Selection (%s): %s", d.Id(), err)
	}

	if v := output.Conditions; v != nil {
		if err := d.Set(names.AttrCondition, flattenConditions(v)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting condition: %s", err)
		}
	}
	d.Set(names.AttrIAMRoleARN, output.IamRoleArn)
	d.Set(names.AttrName, output.SelectionName)
	if err := d.Set("not_resources", output.NotResources); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting not resources: %s", err)
	}
	d.Set("plan_id", planID)
	if err := d.Set(names.AttrResources, output.Resources); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting resources: %s", err)
	}
	if v := output.ListOfTags; v != nil {
		tfList := make([]interface{}, 0)

		for _, v := range v {
			tfMap := make(map[string]interface{})

			tfMap[names.AttrKey] = aws.ToString(v.ConditionKey)
			tfMap[names.AttrType] = v.ConditionType
			tfMap[names.AttrValue] = aws.ToString(v.ConditionValue)

			tfList = append(tfList, tfMap)
		}

		if err := d.Set("selection_tag", tfList); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting selection tag: %s", err)
		}
	}

	return diags
}

func resourceSelectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupClient(ctx)

	log.Printf("[DEBUG] Deleting Backup Selection: %s", d.Id())
	_, err := conn.DeleteBackupSelection(ctx, &backup.DeleteBackupSelectionInput{
		BackupPlanId: aws.String(d.Get("plan_id").(string)),
		SelectionId:  aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.InvalidParameterValueException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Backup Selection (%s): %s", d.Id(), err)
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

func findSelectionByTwoPartKey(ctx context.Context, conn *backup.Client, planID, selectionID string) (*awstypes.BackupSelection, error) {
	input := &backup.GetBackupSelectionInput{
		BackupPlanId: aws.String(planID),
		SelectionId:  aws.String(selectionID),
	}

	return findSelection(ctx, conn, input)
}

func findSelection(ctx context.Context, conn *backup.Client, input *backup.GetBackupSelectionInput) (*awstypes.BackupSelection, error) {
	output, err := conn.GetBackupSelection(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsAErrorMessageContains[*awstypes.InvalidParameterValueException](err, "Cannot find Backup plan") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if output == nil || output.BackupSelection == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.BackupSelection, nil
}

func expandConditionTags(tfList []interface{}) []awstypes.Condition {
	apiObjects := []awstypes.Condition{}

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]interface{})
		apiObject := awstypes.Condition{}

		apiObject.ConditionKey = aws.String(tfMap[names.AttrKey].(string))
		apiObject.ConditionType = awstypes.ConditionType(tfMap[names.AttrType].(string))
		apiObject.ConditionValue = aws.String(tfMap[names.AttrValue].(string))

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandConditions(tfList []interface{}) *awstypes.Conditions {
	apiObject := &awstypes.Conditions{}

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]interface{})

		if v := expandConditionParameters(tfMap["string_equals"].(*schema.Set).List()); len(v) > 0 {
			apiObject.StringEquals = v
		}
		if v := expandConditionParameters(tfMap["string_not_equals"].(*schema.Set).List()); len(v) > 0 {
			apiObject.StringNotEquals = v
		}
		if v := expandConditionParameters(tfMap["string_like"].(*schema.Set).List()); len(v) > 0 {
			apiObject.StringLike = v
		}
		if v := expandConditionParameters(tfMap["string_not_like"].(*schema.Set).List()); len(v) > 0 {
			apiObject.StringNotLike = v
		}
	}

	return apiObject
}

func expandConditionParameters(tfList []interface{}) []awstypes.ConditionParameter {
	apiObjects := []awstypes.ConditionParameter{}

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]interface{})
		apiObject := awstypes.ConditionParameter{}

		apiObject.ConditionKey = aws.String(tfMap[names.AttrKey].(string))
		apiObject.ConditionValue = aws.String(tfMap[names.AttrValue].(string))

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenConditions(apiObject *awstypes.Conditions) []interface{} {
	tfMap := map[string]interface{}{}

	tfMap["string_equals"] = flattenConditionParameters(apiObject.StringEquals)
	tfMap["string_not_equals"] = flattenConditionParameters(apiObject.StringNotEquals)
	tfMap["string_like"] = flattenConditionParameters(apiObject.StringLike)
	tfMap["string_not_like"] = flattenConditionParameters(apiObject.StringNotLike)

	return []interface{}{tfMap}
}

func flattenConditionParameters(apiObjects []awstypes.ConditionParameter) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			names.AttrKey:   aws.ToString(apiObject.ConditionKey),
			names.AttrValue: aws.ToString(apiObject.ConditionValue),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
