// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ecr_account_setting", name="Account Settings")
func resourceAccountSetting() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAccountSettingPut,
		ReadWithoutTimeout:   resourceAccountSettingRead,
		UpdateWithoutTimeout: resourceAccountSettingPut,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"BASIC_SCAN_TYPE_VERSION"}, false),
			},
			names.AttrValue: {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"AWS_NATIVE",
					"CLAIR",
				}, false),
			},
		},
	}
}

func resourceAccountSettingPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	name := aws.String(d.Get(names.AttrName).(string))
	value := aws.String(d.Get(names.AttrValue).(string))
	input := &ecr.PutAccountSettingInput{
		Name:  name,
		Value: value,
	}

	output, err := conn.PutAccountSetting(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ECR Account Setting (%s): %s", *name, err)
	}

	d.SetId(aws.ToString(output.Name))

	return append(diags, resourceAccountSettingRead(ctx, d, meta)...)
}

func resourceAccountSettingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	output, err := findAccountSettingByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ECR Account Setting (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECR Account Setting (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrName, output.Name)
	d.Set(names.AttrValue, output.Value)

	return diags
}

func findAccountSettingByName(ctx context.Context, conn *ecr.Client, name string) (*ecr.GetAccountSettingOutput, error) {
	input := &ecr.GetAccountSettingInput{
		Name: aws.String(name),
	}

	output, err := conn.GetAccountSetting(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
