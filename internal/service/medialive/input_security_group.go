package medialive

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/medialive"
	"github.com/aws/aws-sdk-go-v2/service/medialive/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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

func ResourceInputSecurityGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInputSecurityGroupCreate,
		ReadWithoutTimeout:   resourceInputSecurityGroupRead,
		UpdateWithoutTimeout: resourceInputSecurityGroupUpdate,
		DeleteWithoutTimeout: resourceInputSecurityGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"inputs": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"whitelist_rules": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cidr": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validation.ToDiagFunc(verify.ValidCIDRNetworkAddress),
						},
					},
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameInputSecurityGroup = "Input Security Group"
)

func resourceInputSecurityGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MediaLiveClient()

	in := &medialive.CreateInputSecurityGroupInput{
		WhitelistRules: expandWhitelistRules(d.Get("whitelist_rules").(*schema.Set).List()),
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	if len(tags) > 0 {
		in.Tags = Tags(tags.IgnoreAWS())
	}

	out, err := conn.CreateInputSecurityGroup(ctx, in)
	if err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionCreating, ResNameInputSecurityGroup, "", err)
	}

	if out == nil || out.SecurityGroup == nil {
		return create.DiagError(names.MediaLive, create.ErrActionCreating, ResNameInputSecurityGroup, "", errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.SecurityGroup.Id))

	if _, err := waitInputSecurityGroupCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionWaitingForCreation, ResNameInputSecurityGroup, d.Id(), err)
	}

	return resourceInputSecurityGroupRead(ctx, d, meta)
}

func resourceInputSecurityGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MediaLiveClient()

	out, err := FindInputSecurityGroupByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MediaLive InputSecurityGroup (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionReading, ResNameInputSecurityGroup, d.Id(), err)
	}

	d.Set("arn", out.Arn)
	d.Set("inputs", out.Inputs)
	d.Set("whitelist_rules", flattenInputWhitelistRules(out.WhitelistRules))

	tags, err := ListTags(ctx, conn, aws.ToString(out.Arn))
	if err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionReading, ResNameInputSecurityGroup, d.Id(), err)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionSetting, ResNameInputSecurityGroup, d.Id(), err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionSetting, ResNameInputSecurityGroup, d.Id(), err)
	}

	return nil
}

func resourceInputSecurityGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MediaLiveClient()

	if d.HasChangesExcept("tags", "tags_all") {
		in := &medialive.UpdateInputSecurityGroupInput{
			InputSecurityGroupId: aws.String(d.Id()),
		}

		if d.HasChange("whitelist_rules") {
			in.WhitelistRules = expandWhitelistRules(d.Get("whitelist_rules").(*schema.Set).List())
		}

		log.Printf("[DEBUG] Updating MediaLive InputSecurityGroup (%s): %#v", d.Id(), in)
		out, err := conn.UpdateInputSecurityGroup(ctx, in)
		if err != nil {
			return create.DiagError(names.MediaLive, create.ErrActionUpdating, ResNameInputSecurityGroup, d.Id(), err)
		}

		if _, err := waitInputSecurityGroupUpdated(ctx, conn, aws.ToString(out.SecurityGroup.Id), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return create.DiagError(names.MediaLive, create.ErrActionWaitingForUpdate, ResNameInputSecurityGroup, d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return create.DiagError(names.MediaLive, create.ErrActionUpdating, ResNameInputSecurityGroup, d.Id(), err)
		}
	}

	return resourceInputSecurityGroupRead(ctx, d, meta)
}

func resourceInputSecurityGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MediaLiveClient()

	log.Printf("[INFO] Deleting MediaLive InputSecurityGroup %s", d.Id())

	_, err := conn.DeleteInputSecurityGroup(ctx, &medialive.DeleteInputSecurityGroupInput{
		InputSecurityGroupId: aws.String(d.Id()),
	})

	if err != nil {
		var nfe *types.NotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.DiagError(names.MediaLive, create.ErrActionDeleting, ResNameInputSecurityGroup, d.Id(), err)
	}

	if _, err := waitInputSecurityGroupDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionWaitingForDeletion, ResNameInputSecurityGroup, d.Id(), err)
	}

	return nil
}

func waitInputSecurityGroupCreated(ctx context.Context, conn *medialive.Client, id string, timeout time.Duration) (*medialive.DescribeInputSecurityGroupOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{},
		Target:                    enum.Slice(types.InputSecurityGroupStateIdle, types.InputSecurityGroupStateInUse),
		Refresh:                   statusInputSecurityGroup(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*medialive.DescribeInputSecurityGroupOutput); ok {
		return out, err
	}

	return nil, err
}

func waitInputSecurityGroupUpdated(ctx context.Context, conn *medialive.Client, id string, timeout time.Duration) (*medialive.DescribeInputSecurityGroupOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   enum.Slice(types.InputSecurityGroupStateUpdating),
		Target:                    enum.Slice(types.InputSecurityGroupStateIdle, types.InputSecurityGroupStateInUse),
		Refresh:                   statusInputSecurityGroup(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*medialive.DescribeInputSecurityGroupOutput); ok {
		return out, err
	}

	return nil, err
}

func waitInputSecurityGroupDeleted(ctx context.Context, conn *medialive.Client, id string, timeout time.Duration) (*medialive.DescribeInputSecurityGroupOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{},
		Target:  enum.Slice(types.InputSecurityGroupStateDeleted),
		Refresh: statusInputSecurityGroup(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*medialive.DescribeInputSecurityGroupOutput); ok {
		return out, err
	}

	return nil, err
}

func statusInputSecurityGroup(ctx context.Context, conn *medialive.Client, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindInputSecurityGroupByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.State), nil
	}
}

func FindInputSecurityGroupByID(ctx context.Context, conn *medialive.Client, id string) (*medialive.DescribeInputSecurityGroupOutput, error) {
	in := &medialive.DescribeInputSecurityGroupInput{
		InputSecurityGroupId: aws.String(id),
	}
	out, err := conn.DescribeInputSecurityGroup(ctx, in)
	if err != nil {
		var nfe *types.NotFoundException
		if errors.As(err, &nfe) {
			return nil, &resource.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func flattenInputWhitelistRule(apiObject types.InputWhitelistRule) map[string]interface{} {
	if apiObject == (types.InputWhitelistRule{}) {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.Cidr; v != nil {
		m["cidr"] = aws.ToString(v)
	}

	return m
}

func flattenInputWhitelistRules(apiObjects []types.InputWhitelistRule) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var l []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == (types.InputWhitelistRule{}) {
			continue
		}

		l = append(l, flattenInputWhitelistRule(apiObject))
	}

	return l
}

func expandWhitelistRules(tfList []interface{}) []types.InputWhitelistRuleCidr {
	if len(tfList) == 0 {
		return nil
	}

	var s []types.InputWhitelistRuleCidr

	for _, v := range tfList {
		m, ok := v.(map[string]interface{})

		if !ok {
			continue
		}

		var id types.InputWhitelistRuleCidr
		if val, ok := m["cidr"]; ok {
			id.Cidr = aws.String(val.(string))
			s = append(s, id)
		}
	}
	return s
}
