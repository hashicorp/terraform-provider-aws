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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_rbin_rbin_rule")
func ResourceRBinRule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRBinRuleCreate,
		ReadWithoutTimeout:   resourceRBinRuleRead,
		UpdateWithoutTimeout: resourceRBinRuleUpdate,
		DeleteWithoutTimeout: resourceRBinRuleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(0, 500),
			},
			"identifier": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_tags": {
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
			"resource_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"EBS_SNAPSHOT", "EC2_IMAGE"}, false),
			},
			"retention_period": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"retention_period_value": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 365),
						},
						"retention_period_unit": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"DAYS"}, false),
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameRBinRule  = "Recycle Bin Rule"
	ResourceNameRBin = "Recycle Bin"
)

func resourceRBinRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).RBinClient()

	in := &rbin.CreateRuleInput{
		Description:     aws.String(d.Get("description").(string)),
		ResourceType:    types.ResourceType(d.Get("resource_type").(string)),
		RetentionPeriod: expandRetentionPeriod(d.Get("retention_period").(*schema.Set).List()),
	}

	if v, ok := d.GetOk("resource_tags"); ok && v.(*schema.Set).Len() > 0 {
		in.ResourceTags = expandResourceTags(v.(*schema.Set).List())
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(ctx, d.Get("tags").(map[string]interface{})))

	if len(tags) > 0 {
		in.Tags = Tags(tags.IgnoreAWS())
	}

	out, err := conn.CreateRule(ctx, in)
	if err != nil {
		return create.DiagError(names.RBin, create.ErrActionCreating, ResNameRBinRule, d.Get("identifier").(string), err)
	}

	if out == nil || out.Identifier == nil {
		return create.DiagError(names.RBin, create.ErrActionCreating, ResNameRBinRule, d.Get("identifier").(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.Identifier))

	if _, err := waitRBinRuleCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.DiagError(names.RBin, create.ErrActionWaitingForCreation, ResNameRBinRule, d.Id(), err)
	}

	return resourceRBinRuleRead(ctx, d, meta)
}

func resourceRBinRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).RBinClient()

	out, err := findRBinRuleByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RBin RBinRule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.RBin, create.ErrActionReading, ResNameRBinRule, d.Id(), err)
	}

	d.Set("description", out.Description)
	d.Set("identifier", out.Identifier)
	d.Set("resource_type", string(out.ResourceType))
	d.Set("status", string(out.Status))

	if err := d.Set("resource_tags", flattenResourceTags(out.ResourceTags)); err != nil {
		return create.DiagError(names.RBin, create.ErrActionSetting, ResNameRBinRule, d.Id(), err)
	}

	if err := d.Set("retention_period", flattenRetentionPeriod(out.RetentionPeriod)); err != nil {
		return create.DiagError(names.RBin, create.ErrActionSetting, ResNameRBinRule, d.Id(), err)
	}

	c := meta.(*conns.AWSClient)
	ARN := awsarn.ARN{
		Partition: c.Partition,
		Service:   rbin.ServiceID,
		Region:    c.Region,
		AccountID: c.AccountID,
		Resource:  fmt.Sprintf("rule/%s", d.Id()),
	}
	tags, err := ListTags(ctx, conn, ARN.String())
	if err != nil {
		return create.DiagError(names.RBin, create.ErrActionReading, ResNameRBinRule, d.Id(), err)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return create.DiagError(names.RBin, create.ErrActionSetting, ResNameRBinRule, d.Id(), err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return create.DiagError(names.RBin, create.ErrActionSetting, ResNameRBinRule, d.Id(), err)
	}

	return nil
}

func resourceRBinRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).RBinClient()

	update := false

	in := &rbin.UpdateRuleInput{
		Identifier: aws.String(d.Id()),
	}

	if d.HasChanges("description") {
		in.Description = aws.String(d.Get("description").(string))
		update = true
	}

	if d.HasChanges("resource_tags") {
		in.ResourceTags = expandResourceTags(d.Get("resource_tags").([]interface{}))
		update = true
	}

	if d.HasChanges("retention_period") {
		tfMap := d.Get("retention_period").(*schema.Set).List()
		in.RetentionPeriod = expandRetentionPeriod(tfMap)
		update = true
	}

	if !update {
		return nil
	}

	log.Printf("[DEBUG] Updating RBin RBinRule (%s): %#v", d.Id(), in)
	out, err := conn.UpdateRule(ctx, in)
	if err != nil {
		return create.DiagError(names.RBin, create.ErrActionUpdating, ResNameRBinRule, d.Id(), err)
	}

	if _, err := waitRBinRuleUpdated(ctx, conn, aws.ToString(out.Identifier), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return create.DiagError(names.RBin, create.ErrActionWaitingForUpdate, ResNameRBinRule, d.Id(), err)
	}

	return resourceRBinRuleRead(ctx, d, meta)
}

func resourceRBinRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[INFO] Deleting RBin RBinRule %s", d.Id())

	conn := meta.(*conns.AWSClient).RBinClient()

	_, err := conn.DeleteRule(ctx, &rbin.DeleteRuleInput{
		Identifier: aws.String(d.Id()),
	})

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.DiagError(names.Comprehend, create.ErrActionDeleting, ResourceNameRBin, d.Id(), err)
	}

	if _, err := waitRBinRuleDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.DiagError(names.RBin, create.ErrActionWaitingForDeletion, ResNameRBinRule, d.Id(), err)
	}

	return nil
}

func waitRBinRuleCreated(ctx context.Context, conn *rbin.Client, id string, timeout time.Duration) (*rbin.GetRuleOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{string(types.RuleStatusPending)},
		Target:                    []string{string(types.RuleStatusAvailable)},
		Refresh:                   statusRBinRule(ctx, conn, id),
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

func waitRBinRuleUpdated(ctx context.Context, conn *rbin.Client, id string, timeout time.Duration) (*rbin.GetRuleOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{string(types.RuleStatusPending)},
		Target:                    []string{string(types.RuleStatusAvailable)},
		Refresh:                   statusRBinRule(ctx, conn, id),
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

func waitRBinRuleDeleted(ctx context.Context, conn *rbin.Client, id string, timeout time.Duration) (*rbin.GetRuleOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{string(types.RuleStatusPending), string(types.RuleStatusAvailable)},
		Target:  []string{},
		Refresh: statusRBinRule(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*rbin.GetRuleOutput); ok {
		return out, err
	}

	return nil, err
}

func statusRBinRule(ctx context.Context, conn *rbin.Client, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findRBinRuleByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findRBinRuleByID(ctx context.Context, conn *rbin.Client, id string) (*rbin.GetRuleOutput, error) {
	in := &rbin.GetRuleInput{
		Identifier: aws.String(id),
	}
	out, err := conn.GetRule(ctx, in)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &resource.NotFoundError{
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
