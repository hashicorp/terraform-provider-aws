// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rbin

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rbin"
	"github.com/aws/aws-sdk-go-v2/service/rbin/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_rbin_rule", name="Rule")
// @Tags(identifierAttribute="arn")
func resourceRule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRuleCreate,
		ReadWithoutTimeout:   resourceRuleRead,
		UpdateWithoutTimeout: resourceRuleUpdate,
		DeleteWithoutTimeout: resourceRuleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(0, 500),
			},
			"exclude_resource_tags": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 5,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_tag_key": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 127),
						},
						"resource_tag_value": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 256),
						},
					},
				},
				ConflictsWith: []string{names.AttrResourceTags},
			},
			names.AttrID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"lock_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"unlock_delay": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"unlock_delay_unit": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[types.UnlockDelayUnit](),
									},
									"unlock_delay_value": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntBetween(7, 30),
									},
								},
							},
						},
					},
				},
			},
			"lock_end_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"lock_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrResourceTags: {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 50,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_tag_key": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 127),
						},
						"resource_tag_value": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 256),
						},
					},
				},
				ConflictsWith: []string{"exclude_resource_tags"},
			},
			names.AttrResourceType: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.ResourceType](),
			},
			names.AttrRetentionPeriod: {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"retention_period_value": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 365),
						},
						"retention_period_unit": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.RetentionPeriodUnit](),
						},
					},
				},
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceRuleCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RBinClient(ctx)

	input := rbin.CreateRuleInput{
		ResourceType:    types.ResourceType(d.Get(names.AttrResourceType).(string)),
		RetentionPeriod: expandRetentionPeriod(d.Get(names.AttrRetentionPeriod).([]any)),
		Tags:            getTagsIn(ctx),
	}

	if _, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(d.Get(names.AttrDescription).(string))
	}

	if v, ok := d.GetOk("exclude_resource_tags"); ok && v.(*schema.Set).Len() > 0 {
		input.ExcludeResourceTags = expandResourceTags(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk(names.AttrResourceTags); ok && v.(*schema.Set).Len() > 0 {
		input.ResourceTags = expandResourceTags(v.(*schema.Set).List())
	}

	output, err := conn.CreateRule(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RBin Rule: %s", err)
	}

	d.SetId(aws.ToString(output.Identifier))

	if _, err := waitRuleCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RBin Rule (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceRuleRead(ctx, d, meta)...)
}

func resourceRuleRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.RBinClient(ctx)

	output, err := findRuleByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] RBin Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RBin Rule (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, c.RegionalARN(ctx, "rbin", "rule/"+d.Id()))
	d.Set(names.AttrDescription, output.Description)
	if err := d.Set("exclude_resource_tags", flattenResourceTags(output.ExcludeResourceTags)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting exclude_resource_tags: %s", err)
	}
	if err := d.Set(names.AttrResourceTags, flattenResourceTags(output.ResourceTags)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting resource_tags: %s", err)
	}
	d.Set(names.AttrResourceType, output.ResourceType)
	if err := d.Set(names.AttrRetentionPeriod, flattenRetentionPeriod(output.RetentionPeriod)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting retention_period: %s", err)
	}
	d.Set(names.AttrStatus, output.Status)

	return diags
}

func resourceRuleUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RBinClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := rbin.UpdateRuleInput{
			Identifier: aws.String(d.Id()),
		}

		if d.HasChanges(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChanges("exclude_resource_tags") {
			v := d.Get("exclude_resource_tags")
			if v == nil || v.(*schema.Set).Len() == 0 {
				input.ExcludeResourceTags = []types.ResourceTag{}
			} else {
				input.ExcludeResourceTags = expandResourceTags(v.(*schema.Set).List())
			}
		}

		if d.HasChanges(names.AttrResourceTags) {
			v := d.Get(names.AttrResourceTags)
			if v == nil || v.(*schema.Set).Len() == 0 {
				input.ResourceTags = []types.ResourceTag{}
			} else {
				input.ResourceTags = expandResourceTags(v.(*schema.Set).List())
			}
		}

		if d.HasChanges(names.AttrRetentionPeriod) {
			input.RetentionPeriod = expandRetentionPeriod(d.Get(names.AttrRetentionPeriod).([]any))
		}

		_, err := conn.UpdateRule(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating RBin Rule (%s): %s", d.Id(), err)
		}

		if _, err := waitRuleUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for RBin Rule (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceRuleRead(ctx, d, meta)...)
}

func resourceRuleDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RBinClient(ctx)

	log.Printf("[INFO] Deleting RBin Rule: %s", d.Id())
	input := rbin.DeleteRuleInput{
		Identifier: aws.String(d.Id()),
	}
	_, err := conn.DeleteRule(ctx, &input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RBin Rule (%s): %s", d.Id(), err)
	}

	if _, err := waitRuleDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RBin Rule (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findRuleByID(ctx context.Context, conn *rbin.Client, id string) (*rbin.GetRuleOutput, error) {
	input := rbin.GetRuleInput{
		Identifier: aws.String(id),
	}
	output, err := conn.GetRule(ctx, &input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Identifier == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func statusRule(conn *rbin.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findRuleByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitRuleCreated(ctx context.Context, conn *rbin.Client, id string, timeout time.Duration) (*rbin.GetRuleOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.RuleStatusPending),
		Target:                    enum.Slice(types.RuleStatusAvailable),
		Refresh:                   statusRule(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*rbin.GetRuleOutput); ok {
		return output, err
	}

	return nil, err
}

func waitRuleUpdated(ctx context.Context, conn *rbin.Client, id string, timeout time.Duration) (*rbin.GetRuleOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.RuleStatusPending),
		Target:                    enum.Slice(types.RuleStatusAvailable),
		Refresh:                   statusRule(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*rbin.GetRuleOutput); ok {
		return output, err
	}

	return nil, err
}

func waitRuleDeleted(ctx context.Context, conn *rbin.Client, id string, timeout time.Duration) (*rbin.GetRuleOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.RuleStatusPending, types.RuleStatusAvailable),
		Target:  []string{},
		Refresh: statusRule(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*rbin.GetRuleOutput); ok {
		return output, err
	}

	return nil, err
}

func flattenResourceTag(apiObject types.ResourceTag) map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.ResourceTagKey; v != nil {
		tfMap["resource_tag_key"] = aws.ToString(v)
	}

	if v := apiObject.ResourceTagValue; v != nil {
		tfMap["resource_tag_value"] = aws.ToString(v)
	}

	return tfMap
}

func flattenResourceTags(apiObjects []types.ResourceTag) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenResourceTag(apiObject))
	}

	return tfList
}

func flattenRetentionPeriod(apiObject *types.RetentionPeriod) []any {
	tfMap := map[string]any{
		"retention_period_unit": apiObject.RetentionPeriodUnit,
	}

	if v := apiObject.RetentionPeriodValue; v != aws.Int32(0) {
		tfMap["retention_period_value"] = v
	}

	return []any{tfMap}
}

func expandResourceTag(tfMap map[string]any) *types.ResourceTag {
	if tfMap == nil {
		return nil
	}

	apiObject := types.ResourceTag{}

	if v, ok := tfMap["resource_tag_key"].(string); ok && v != "" {
		apiObject.ResourceTagKey = aws.String(v)
	}

	if v, ok := tfMap["resource_tag_value"].(string); ok && v != "" {
		apiObject.ResourceTagValue = aws.String(v)
	}

	return &apiObject
}

func expandResourceTags(tfList []any) []types.ResourceTag {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.ResourceTag

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandResourceTag(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandRetentionPeriod(tfList []any) *types.RetentionPeriod {
	if tfList == nil {
		return nil
	}
	tfMap := tfList[0].(map[string]any)

	apiObject := types.RetentionPeriod{}

	if v, ok := tfMap["retention_period_value"].(int); ok {
		apiObject.RetentionPeriodValue = aws.Int32(int32(v))
	}

	if v, ok := tfMap["retention_period_unit"].(string); ok && v != "" {
		apiObject.RetentionPeriodUnit = types.RetentionPeriodUnit(v)
	}

	return &apiObject
}
