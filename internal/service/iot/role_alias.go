// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iot/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_iot_role_alias")
// @Tags(identifierAttribute="arn")
func ResourceRoleAlias() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRoleAliasCreate,
		ReadWithoutTimeout:   resourceRoleAliasRead,
		UpdateWithoutTimeout: resourceRoleAliasUpdate,
		DeleteWithoutTimeout: resourceRoleAliasDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAlias: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrRoleARN: {
				Type:     schema.TypeString,
				Required: true,
			},
			"credential_duration": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      3600,
				ValidateFunc: validation.IntBetween(900, 43200),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceRoleAliasCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	roleAlias := d.Get(names.AttrAlias).(string)
	roleArn := d.Get(names.AttrRoleARN).(string)
	credentialDuration := d.Get("credential_duration").(int)

	_, err := conn.CreateRoleAlias(ctx, &iot.CreateRoleAliasInput{
		RoleAlias:                 aws.String(roleAlias),
		RoleArn:                   aws.String(roleArn),
		CredentialDurationSeconds: aws.Int32(int32(credentialDuration)),
		Tags:                      getTagsIn(ctx),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating role alias %s for role %s: %s", roleAlias, roleArn, err)
	}

	d.SetId(roleAlias)
	return append(diags, resourceRoleAliasRead(ctx, d, meta)...)
}

func GetRoleAliasDescription(ctx context.Context, conn *iot.Client, alias string) (*awstypes.RoleAliasDescription, error) {
	roleAliasDescriptionOutput, err := conn.DescribeRoleAlias(ctx, &iot.DescribeRoleAliasInput{
		RoleAlias: aws.String(alias),
	})

	if err != nil {
		return nil, err
	}

	if roleAliasDescriptionOutput == nil {
		return nil, nil
	}

	return roleAliasDescriptionOutput.RoleAliasDescription, nil
}

func resourceRoleAliasRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	var roleAliasDescription *awstypes.RoleAliasDescription

	roleAliasDescription, err := GetRoleAliasDescription(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing role alias %s: %s", d.Id(), err)
	}

	if roleAliasDescription == nil {
		log.Printf("[WARN] Role alias (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	d.Set(names.AttrARN, roleAliasDescription.RoleAliasArn)
	d.Set(names.AttrAlias, roleAliasDescription.RoleAlias)
	d.Set(names.AttrRoleARN, roleAliasDescription.RoleArn)
	d.Set("credential_duration", roleAliasDescription.CredentialDurationSeconds)

	return diags
}

func resourceRoleAliasDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	alias := d.Get(names.AttrAlias).(string)

	_, err := conn.DeleteRoleAlias(ctx, &iot.DeleteRoleAliasInput{
		RoleAlias: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting role alias %s: %s", alias, err)
	}

	return diags
}

func resourceRoleAliasUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	if d.HasChange("credential_duration") {
		roleAliasInput := &iot.UpdateRoleAliasInput{
			RoleAlias:                 aws.String(d.Id()),
			CredentialDurationSeconds: aws.Int32(int32(d.Get("credential_duration").(int))),
		}
		_, err := conn.UpdateRoleAlias(ctx, roleAliasInput)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating role alias %s: %s", d.Id(), err)
		}
	}

	if d.HasChange(names.AttrRoleARN) {
		roleAliasInput := &iot.UpdateRoleAliasInput{
			RoleAlias: aws.String(d.Id()),
			RoleArn:   aws.String(d.Get(names.AttrRoleARN).(string)),
		}
		_, err := conn.UpdateRoleAlias(ctx, roleAliasInput)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating role alias %s: %s", d.Id(), err)
		}
	}

	return append(diags, resourceRoleAliasRead(ctx, d, meta)...)
}
