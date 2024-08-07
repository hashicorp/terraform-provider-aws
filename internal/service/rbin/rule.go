// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rbin

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rbin"
	"github.com/aws/aws-sdk-go-v2/service/rbin/types"
	awsarn "github.com/aws/aws-sdk-go/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_rbin_rule", name="Rule")
// @Tags(identifierAttribute="arn")
func ResourceRule() *schema.Resource {
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
			names.AttrID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrResourceTags: {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
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
			"lock_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"lock_end_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameRule = "Rule"
)

func resourceRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RBinClient(ctx)

	in := &rbin.CreateRuleInput{
		ResourceType:    types.ResourceType(d.Get(names.AttrResourceType).(string)),
		RetentionPeriod: expandRetentionPeriod(d.Get(names.AttrRetentionPeriod).([]interface{})),
		Tags:            getTagsIn(ctx),
	}

	if _, ok := d.GetOk(names.AttrDescription); ok {
		in.Description = aws.String(d.Get(names.AttrDescription).(string))
	}

	if v, ok := d.GetOk(names.AttrResourceTags); ok && v.(*schema.Set).Len() > 0 {
		in.ResourceTags = expandResourceTags(v.(*schema.Set).List())
	}

	out, err := conn.CreateRule(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.RBin, create.ErrActionCreating, ResNameRule, d.Get(names.AttrResourceType).(string), err)
	}

	if out == nil || out.Identifier == nil {
		return create.AppendDiagError(diags, names.RBin, create.ErrActionCreating, ResNameRule, d.Get(names.AttrResourceType).(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.Identifier))

	if _, err := waitRuleCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.RBin, create.ErrActionWaitingForCreation, ResNameRule, d.Id(), err)
	}

	return append(diags, resourceRuleRead(ctx, d, meta)...)
}

func resourceRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RBinClient(ctx)

	out, err := findRuleByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RBin Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.RBin, create.ErrActionReading, ResNameRule, d.Id(), err)
	}

	ruleArn := awsarn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   rbin.ServiceID,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("rule/%s", aws.ToString(out.Identifier)),
	}.String()
	d.Set(names.AttrARN, ruleArn)

	d.Set(names.AttrDescription, out.Description)
	d.Set(names.AttrResourceType, string(out.ResourceType))
	d.Set(names.AttrStatus, string(out.Status))

	if err := d.Set(names.AttrResourceTags, flattenResourceTags(out.ResourceTags)); err != nil {
		return create.AppendDiagError(diags, names.RBin, create.ErrActionSetting, ResNameRule, d.Id(), err)
	}

	if err := d.Set(names.AttrRetentionPeriod, flattenRetentionPeriod(out.RetentionPeriod)); err != nil {
		return create.AppendDiagError(diags, names.RBin, create.ErrActionSetting, ResNameRule, d.Id(), err)
	}

	return diags
}

func resourceRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RBinClient(ctx)

	update := false

	in := &rbin.UpdateRuleInput{
		Identifier: aws.String(d.Id()),
	}

	if d.HasChanges(names.AttrDescription) {
		in.Description = aws.String(d.Get(names.AttrDescription).(string))
		update = true
	}

	if d.HasChanges(names.AttrResourceTags) {
		in.ResourceTags = expandResourceTags(d.Get(names.AttrResourceTags).(*schema.Set).List())
		update = true
	}

	if d.HasChanges(names.AttrRetentionPeriod) {
		in.RetentionPeriod = expandRetentionPeriod(d.Get(names.AttrRetentionPeriod).([]interface{}))
		update = true
	}

	if !update {
		return diags
	}

	log.Printf("[DEBUG] Updating RBin Rule (%s): %#v", d.Id(), in)
	out, err := conn.UpdateRule(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.RBin, create.ErrActionUpdating, ResNameRule, d.Id(), err)
	}

	if _, err := waitRuleUpdated(ctx, conn, aws.ToString(out.Identifier), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return create.AppendDiagError(diags, names.RBin, create.ErrActionWaitingForUpdate, ResNameRule, d.Id(), err)
	}

	return append(diags, resourceRuleRead(ctx, d, meta)...)
}

func resourceRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RBinClient(ctx)

	log.Printf("[INFO] Deleting RBin Rule: %s", d.Id())
	_, err := conn.DeleteRule(ctx, &rbin.DeleteRuleInput{
		Identifier: aws.String(d.Id()),
	})

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return diags
		}

		return create.AppendDiagError(diags, names.RBin, create.ErrActionDeleting, ResNameRule, d.Id(), err)
	}

	if _, err := waitRuleDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.AppendDiagError(diags, names.RBin, create.ErrActionWaitingForDeletion, ResNameRule, d.Id(), err)
	}

	return diags
}

func waitRuleCreated(ctx context.Context, conn *rbin.Client, id string, timeout time.Duration) (*rbin.GetRuleOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.RuleStatusPending),
		Target:                    enum.Slice(types.RuleStatusAvailable),
		Refresh:                   statusRule(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*rbin.GetRuleOutput); ok {
		return out, err
	}

	return nil, err
}

func waitRuleUpdated(ctx context.Context, conn *rbin.Client, id string, timeout time.Duration) (*rbin.GetRuleOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.RuleStatusPending),
		Target:                    enum.Slice(types.RuleStatusAvailable),
		Refresh:                   statusRule(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*rbin.GetRuleOutput); ok {
		return out, err
	}

	return nil, err
}

func waitRuleDeleted(ctx context.Context, conn *rbin.Client, id string, timeout time.Duration) (*rbin.GetRuleOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.RuleStatusPending, types.RuleStatusAvailable),
		Target:  []string{},
		Refresh: statusRule(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*rbin.GetRuleOutput); ok {
		return out, err
	}

	return nil, err
}

func statusRule(ctx context.Context, conn *rbin.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findRuleByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findRuleByID(ctx context.Context, conn *rbin.Client, id string) (*rbin.GetRuleOutput, error) {
	in := &rbin.GetRuleInput{
		Identifier: aws.String(id),
	}
	out, err := conn.GetRule(ctx, in)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.Identifier == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func flattenResourceTag(rTag types.ResourceTag) map[string]interface{} {
	m := map[string]interface{}{}

	if v := rTag.ResourceTagKey; v != nil {
		m["resource_tag_key"] = aws.ToString(v)
	}

	if v := rTag.ResourceTagValue; v != nil {
		m["resource_tag_value"] = aws.ToString(v)
	}

	return m
}

func flattenResourceTags(rTags []types.ResourceTag) []interface{} {
	if len(rTags) == 0 {
		return nil
	}

	var l []interface{}

	for _, rTag := range rTags {
		l = append(l, flattenResourceTag(rTag))
	}

	return l
}

func flattenRetentionPeriod(retPeriod *types.RetentionPeriod) []interface{} {
	m := map[string]interface{}{}

	if v := retPeriod.RetentionPeriodUnit; v != "" {
		m["retention_period_unit"] = string(v)
	}

	if v := retPeriod.RetentionPeriodValue; v != aws.Int32(0) {
		m["retention_period_value"] = v
	}

	return []interface{}{m}
}

func expandResourceTag(tfMap map[string]interface{}) *types.ResourceTag {
	if tfMap == nil {
		return nil
	}

	a := &types.ResourceTag{}

	if v, ok := tfMap["resource_tag_key"].(string); ok && v != "" {
		a.ResourceTagKey = aws.String(v)
	}

	if v, ok := tfMap["resource_tag_value"].(string); ok && v != "" {
		a.ResourceTagValue = aws.String(v)
	}

	return a
}

func expandResourceTags(tfList []interface{}) []types.ResourceTag {
	if len(tfList) == 0 {
		return nil
	}

	var s []types.ResourceTag

	for _, r := range tfList {
		m, ok := r.(map[string]interface{})

		if !ok {
			continue
		}

		a := expandResourceTag(m)

		if a == nil {
			continue
		}

		s = append(s, *a)
	}

	return s
}

func expandRetentionPeriod(tfList []interface{}) *types.RetentionPeriod {
	if tfList == nil {
		return nil
	}
	tfMap := tfList[0].(map[string]interface{})

	a := types.RetentionPeriod{}

	if v, ok := tfMap["retention_period_value"].(int); ok {
		a.RetentionPeriodValue = aws.Int32(int32(v))
	}

	if v, ok := tfMap["retention_period_unit"].(string); ok && v != "" {
		a.RetentionPeriodUnit = types.RetentionPeriodUnit(v)
	}

	return &a
}
