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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_iot_role_alias", name="Role Alias")
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
			names.AttrAlias: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"credential_duration": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      3600,
				ValidateFunc: validation.IntBetween(900, 43200),
			},
			names.AttrRoleARN: {
				Type:     schema.TypeString,
				Required: true,
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
	input := &iot.CreateRoleAliasInput{
		RoleAlias:                 aws.String(roleAlias),
		RoleArn:                   aws.String(d.Get(names.AttrRoleARN).(string)),
		CredentialDurationSeconds: aws.Int32(int32(d.Get("credential_duration").(int))),
		Tags:                      getTagsIn(ctx),
	}

	_, err := conn.CreateRoleAlias(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IoT Role Alias (%s): %s", roleAlias, err)
	}

	d.SetId(roleAlias)

	return append(diags, resourceRoleAliasRead(ctx, d, meta)...)
}

func resourceRoleAliasRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	output, err := findRoleAliasByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IoT Role Alias (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IoT Role Alias (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrAlias, output.RoleAlias)
	d.Set(names.AttrARN, output.RoleAliasArn)
	d.Set("credential_duration", output.CredentialDurationSeconds)
	d.Set(names.AttrRoleARN, output.RoleArn)

	return diags
}

func resourceRoleAliasUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	if d.HasChange("credential_duration") {
		input := &iot.UpdateRoleAliasInput{
			CredentialDurationSeconds: aws.Int32(int32(d.Get("credential_duration").(int))),
			RoleAlias:                 aws.String(d.Id()),
		}

		_, err := conn.UpdateRoleAlias(ctx, input)

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

func resourceRoleAliasDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	log.Printf("[INFO] Deleting IoT Role Alias: %s", d.Id())
	_, err := conn.DeleteRoleAlias(ctx, &iot.DeleteRoleAliasInput{
		RoleAlias: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IoT Role Alias (%s): %s", d.Id(), err)
	}

	return diags
}

func findRoleAliasByID(ctx context.Context, conn *iot.Client, alias string) (*awstypes.RoleAliasDescription, error) {
	input := &iot.DescribeRoleAliasInput{
		RoleAlias: aws.String(alias),
	}

	output, err := conn.DescribeRoleAlias(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.RoleAliasDescription, nil
}
