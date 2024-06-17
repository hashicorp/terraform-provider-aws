// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ecs_account_setting_default", name="Account Setting Defauilt")
func ResourceAccountSettingDefault() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAccountSettingDefaultCreate,
		ReadWithoutTimeout:   resourceAccountSettingDefaultRead,
		UpdateWithoutTimeout: resourceAccountSettingDefaultUpdate,
		DeleteWithoutTimeout: resourceAccountSettingDefaultDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceAccountSettingDefaultImport,
		},

		Schema: map[string]*schema.Schema{
			names.AttrName: {
				Type:             schema.TypeString,
				ForceNew:         true,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.SettingName](),
			},
			"principal_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrValue: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceAccountSettingDefaultImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.Set(names.AttrName, d.Id())
	d.SetId(arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Service:   names.ECSEndpointID,
		Resource:  fmt.Sprintf("cluster/%s", d.Id()),
	}.String())
	return []*schema.ResourceData{d}, nil
}

func resourceAccountSettingDefaultCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	settingName := d.Get(names.AttrName).(string)
	settingValue := d.Get(names.AttrValue).(string)
	log.Printf("[DEBUG] Setting Account Default %s", settingName)

	input := ecs.PutAccountSettingDefaultInput{
		Name:  awstypes.SettingName(settingName),
		Value: aws.String(settingValue),
	}

	out, err := conn.PutAccountSettingDefault(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ECS Account Setting Defauilt (%s): %s", settingName, err)
	}
	log.Printf("[DEBUG] Account Setting Default %s set", aws.ToString(out.Setting.Value))

	d.SetId(aws.ToString(out.Setting.Value))
	d.Set("principal_arn", out.Setting.PrincipalArn)

	return append(diags, resourceAccountSettingDefaultRead(ctx, d, meta)...)
}

func resourceAccountSettingDefaultRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	input := &ecs.ListAccountSettingsInput{
		Name:              awstypes.SettingName(d.Get(names.AttrName).(string)),
		EffectiveSettings: true,
	}

	log.Printf("[DEBUG] Reading Default Account Settings: %+v", input)
	resp, err := conn.ListAccountSettings(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECS Account Setting Defauilt (%s): %s", d.Get(names.AttrName).(string), err)
	}

	if len(resp.Settings) == 0 {
		log.Printf("[WARN] Account Setting Default not set. Removing from state")
		d.SetId("")
		return diags
	}

	for _, r := range resp.Settings {
		d.SetId(aws.ToString(r.PrincipalArn))
		d.Set(names.AttrName, r.Name)
		d.Set("principal_arn", r.PrincipalArn)
		d.Set(names.AttrValue, r.Value)
	}

	return diags
}

func resourceAccountSettingDefaultUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	settingName := d.Get(names.AttrName).(string)
	settingValue := d.Get(names.AttrValue).(string)

	if d.HasChange(names.AttrValue) {
		input := ecs.PutAccountSettingDefaultInput{
			Name:  awstypes.SettingName(settingName),
			Value: aws.String(settingValue),
		}

		_, err := conn.PutAccountSettingDefault(ctx, &input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating ECS Account Setting Default (%s): %s", settingName, err)
		}
	}

	return diags
}

func resourceAccountSettingDefaultDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	settingName := d.Get(names.AttrName).(string)
	settingValue := "disabled"

	//Default value: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-maintenance.html#task-retirement-change
	if settingName == string(awstypes.SettingNameFargateTaskRetirementWaitPeriod) {
		settingValue = fargateTaskRetirementWaitPeriodValue
	}

	log.Printf("[WARN] Disabling ECS Account Setting Default %s", settingName)
	input := ecs.PutAccountSettingDefaultInput{
		Name:  awstypes.SettingName(settingName),
		Value: aws.String(settingValue),
	}

	_, err := conn.PutAccountSettingDefault(ctx, &input)

	if errs.IsAErrorMessageContains[*awstypes.InvalidParameterException](err, "You can no longer disable") {
		log.Printf("[DEBUG] ECS Account Setting Default (%q) could not be disabled: %s", settingName, err)
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling ECS Account Setting Default: %s", err)
	}

	log.Printf("[DEBUG] ECS Account Setting Default (%q) disabled", settingName)
	return diags
}
