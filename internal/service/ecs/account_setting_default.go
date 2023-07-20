// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_ecs_account_setting_default")
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
			"name": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringInSlice(ecs.SettingName_Values(), false),
			},
			"principal_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"value": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"enabled", "disabled"}, false),
			},
		},
	}
}

func resourceAccountSettingDefaultImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.Set("name", d.Id())
	d.SetId(arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Service:   ecs.ServiceName,
		Resource:  fmt.Sprintf("cluster/%s", d.Id()),
	}.String())
	return []*schema.ResourceData{d}, nil
}

func resourceAccountSettingDefaultCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSConn(ctx)

	settingName := d.Get("name").(string)
	settingValue := d.Get("value").(string)
	log.Printf("[DEBUG] Setting Account Default %s", settingName)

	input := ecs.PutAccountSettingDefaultInput{
		Name:  aws.String(settingName),
		Value: aws.String(settingValue),
	}

	out, err := conn.PutAccountSettingDefaultWithContext(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ECS Account Setting Defauilt (%s): %s", settingName, err)
	}
	log.Printf("[DEBUG] Account Setting Default %s set", aws.StringValue(out.Setting.Value))

	d.SetId(aws.StringValue(out.Setting.Value))
	d.Set("principal_arn", out.Setting.PrincipalArn)

	return append(diags, resourceAccountSettingDefaultRead(ctx, d, meta)...)
}

func resourceAccountSettingDefaultRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSConn(ctx)

	input := &ecs.ListAccountSettingsInput{
		Name:              aws.String(d.Get("name").(string)),
		EffectiveSettings: aws.Bool(true),
	}

	log.Printf("[DEBUG] Reading Default Account Settings: %s", input)
	resp, err := conn.ListAccountSettingsWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECS Account Setting Defauilt (%s): %s", d.Get("name").(string), err)
	}

	if len(resp.Settings) == 0 {
		log.Printf("[WARN] Account Setting Default not set. Removing from state")
		d.SetId("")
		return diags
	}

	for _, r := range resp.Settings {
		d.SetId(aws.StringValue(r.PrincipalArn))
		d.Set("name", r.Name)
		d.Set("principal_arn", r.PrincipalArn)
		d.Set("value", r.Value)
	}

	return diags
}

func resourceAccountSettingDefaultUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSConn(ctx)

	settingName := d.Get("name").(string)
	settingValue := d.Get("value").(string)

	if d.HasChange("value") {
		input := ecs.PutAccountSettingDefaultInput{
			Name:  aws.String(settingName),
			Value: aws.String(settingValue),
		}

		_, err := conn.PutAccountSettingDefaultWithContext(ctx, &input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating ECS Account Setting Default (%s): %s", settingName, err)
		}
	}

	return diags
}

func resourceAccountSettingDefaultDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSConn(ctx)

	settingName := d.Get("name").(string)

	log.Printf("[WARN] Disabling ECS Account Setting Default %s", settingName)
	input := ecs.PutAccountSettingDefaultInput{
		Name:  aws.String(settingName),
		Value: aws.String("disabled"),
	}

	_, err := conn.PutAccountSettingDefaultWithContext(ctx, &input)

	if tfawserr.ErrMessageContains(err, ecs.ErrCodeInvalidParameterException, "You can no longer disable") {
		log.Printf("[DEBUG] ECS Account Setting Default (%q) could not be disabled: %s", settingName, err)
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling ECS Account Setting Default: %s", err)
	}

	log.Printf("[DEBUG] ECS Account Setting Default (%q) disabled", settingName)
	return diags
}
