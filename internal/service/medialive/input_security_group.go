// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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

// @SDKResource("aws_medialive_input_security_group", name="Input Security Group")
// @Tags(identifierAttribute="arn")
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
			names.AttrARN: {
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameInputSecurityGroup = "Input Security Group"
)

func resourceInputSecurityGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).MediaLiveClient(ctx)

	in := &medialive.CreateInputSecurityGroupInput{
		Tags:           getTagsIn(ctx),
		WhitelistRules: expandWhitelistRules(d.Get("whitelist_rules").(*schema.Set).List()),
	}

	out, err := conn.CreateInputSecurityGroup(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.MediaLive, create.ErrActionCreating, ResNameInputSecurityGroup, "", err)
	}

	if out == nil || out.SecurityGroup == nil {
		return create.AppendDiagError(diags, names.MediaLive, create.ErrActionCreating, ResNameInputSecurityGroup, "", errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.SecurityGroup.Id))

	if _, err := waitInputSecurityGroupCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.MediaLive, create.ErrActionWaitingForCreation, ResNameInputSecurityGroup, d.Id(), err)
	}

	return append(diags, resourceInputSecurityGroupRead(ctx, d, meta)...)
}

func resourceInputSecurityGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).MediaLiveClient(ctx)

	out, err := FindInputSecurityGroupByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MediaLive InputSecurityGroup (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.MediaLive, create.ErrActionReading, ResNameInputSecurityGroup, d.Id(), err)
	}

	d.Set(names.AttrARN, out.Arn)
	d.Set("inputs", out.Inputs)
	d.Set("whitelist_rules", flattenInputWhitelistRules(out.WhitelistRules))

	return diags
}

func resourceInputSecurityGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).MediaLiveClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		in := &medialive.UpdateInputSecurityGroupInput{
			InputSecurityGroupId: aws.String(d.Id()),
		}

		if d.HasChange("whitelist_rules") {
			in.WhitelistRules = expandWhitelistRules(d.Get("whitelist_rules").(*schema.Set).List())
		}

		log.Printf("[DEBUG] Updating MediaLive InputSecurityGroup (%s): %#v", d.Id(), in)
		out, err := conn.UpdateInputSecurityGroup(ctx, in)
		if err != nil {
			return create.AppendDiagError(diags, names.MediaLive, create.ErrActionUpdating, ResNameInputSecurityGroup, d.Id(), err)
		}

		if _, err := waitInputSecurityGroupUpdated(ctx, conn, aws.ToString(out.SecurityGroup.Id), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return create.AppendDiagError(diags, names.MediaLive, create.ErrActionWaitingForUpdate, ResNameInputSecurityGroup, d.Id(), err)
		}
	}

	return append(diags, resourceInputSecurityGroupRead(ctx, d, meta)...)
}

func resourceInputSecurityGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).MediaLiveClient(ctx)

	log.Printf("[INFO] Deleting MediaLive InputSecurityGroup %s", d.Id())

	_, err := conn.DeleteInputSecurityGroup(ctx, &medialive.DeleteInputSecurityGroupInput{
		InputSecurityGroupId: aws.String(d.Id()),
	})

	if err != nil {
		var nfe *types.NotFoundException
		if errors.As(err, &nfe) {
			return diags
		}

		return create.AppendDiagError(diags, names.MediaLive, create.ErrActionDeleting, ResNameInputSecurityGroup, d.Id(), err)
	}

	if _, err := waitInputSecurityGroupDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.AppendDiagError(diags, names.MediaLive, create.ErrActionWaitingForDeletion, ResNameInputSecurityGroup, d.Id(), err)
	}

	return diags
}

func waitInputSecurityGroupCreated(ctx context.Context, conn *medialive.Client, id string, timeout time.Duration) (*medialive.DescribeInputSecurityGroupOutput, error) {
	stateConf := &retry.StateChangeConf{
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
	stateConf := &retry.StateChangeConf{
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
	stateConf := &retry.StateChangeConf{
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

func statusInputSecurityGroup(ctx context.Context, conn *medialive.Client, id string) retry.StateRefreshFunc {
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
			return nil, &retry.NotFoundError{
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
