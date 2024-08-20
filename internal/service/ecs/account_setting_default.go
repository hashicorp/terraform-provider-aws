// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ecs_account_setting_default", name="Account Setting Default")
func resourceAccountSettingDefault() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAccountSettingDefaultPut,
		ReadWithoutTimeout:   resourceAccountSettingDefaultRead,
		UpdateWithoutTimeout: resourceAccountSettingDefaultPut,
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

func resourceAccountSettingDefaultPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	settingName := awstypes.SettingName(d.Get(names.AttrName).(string))
	input := &ecs.PutAccountSettingDefaultInput{
		Name:  settingName,
		Value: aws.String(d.Get(names.AttrValue).(string)),
	}

	output, err := conn.PutAccountSettingDefault(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting ECS Account Setting Default (%s): %s", settingName, err)
	}

	if d.IsNewResource() {
		// Huh?
		d.SetId(aws.ToString(output.Setting.Value))
	}

	return append(diags, resourceAccountSettingDefaultRead(ctx, d, meta)...)
}

func resourceAccountSettingDefaultRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	settingName := awstypes.SettingName(d.Get(names.AttrName).(string))
	setting, err := findEffectiveAccountSettingByName(ctx, conn, settingName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ECS Account Setting Default (%s) not found, removing from state", settingName)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECS Account Setting Default (%s): %s", settingName, err)
	}

	principalARN := aws.ToString(setting.PrincipalArn)
	d.SetId(principalARN)
	d.Set(names.AttrName, setting.Name)
	d.Set("principal_arn", principalARN)
	d.Set(names.AttrValue, setting.Value)

	return diags
}

func resourceAccountSettingDefaultDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	settingName := awstypes.SettingName(d.Get(names.AttrName).(string))
	settingValue := "disabled"

	// Default value: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-maintenance.html#task-retirement-change.
	if settingName == awstypes.SettingNameFargateTaskRetirementWaitPeriod {
		const (
			fargateTaskRetirementWaitPeriodValue = "7"
		)
		settingValue = fargateTaskRetirementWaitPeriodValue
	}

	log.Printf("[WARN] Deleting ECS Account Setting Default: %s", settingName)
	input := &ecs.PutAccountSettingDefaultInput{
		Name:  settingName,
		Value: aws.String(settingValue),
	}

	_, err := conn.PutAccountSettingDefault(ctx, input)

	if errs.IsAErrorMessageContains[*awstypes.InvalidParameterException](err, "You can no longer disable") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling ECS Account Setting Default (%s): %s", settingName, err)
	}

	return diags
}

func resourceAccountSettingDefaultImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.Set(names.AttrName, d.Id())
	d.SetId(arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Service:   names.ECSEndpointID,
		Resource:  "cluster/" + d.Id(),
	}.String())

	return []*schema.ResourceData{d}, nil
}

func findSetting(ctx context.Context, conn *ecs.Client, input *ecs.ListAccountSettingsInput) (*awstypes.Setting, error) {
	output, err := findSettings(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findSettings(ctx context.Context, conn *ecs.Client, input *ecs.ListAccountSettingsInput) ([]awstypes.Setting, error) {
	var output []awstypes.Setting

	pages := ecs.NewListAccountSettingsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.Settings...)
	}

	return output, nil
}

func findEffectiveAccountSettingByName(ctx context.Context, conn *ecs.Client, name awstypes.SettingName) (*awstypes.Setting, error) {
	input := &ecs.ListAccountSettingsInput{
		EffectiveSettings: true,
		Name:              name,
	}

	return findSetting(ctx, conn, input)
}
