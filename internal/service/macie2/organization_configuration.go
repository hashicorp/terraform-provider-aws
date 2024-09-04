// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package macie2

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/macie2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_macie2_organization_configuration", name="Organization Configuration")
func resourceOrganizationConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOrganizationConfigurationUpdate,
		UpdateWithoutTimeout: resourceOrganizationConfigurationUpdate,
		DeleteWithoutTimeout: schema.NoopContext,
		ReadWithoutTimeout:   resourceOrganizationConfigurationRead,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"auto_enable": {
				Type:     schema.TypeBool,
				Required: true,
			},
		},
	}
}

func resourceOrganizationConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).Macie2Client(ctx)

	input := macie2.UpdateOrganizationConfigurationInput{
		AutoEnable: aws.Bool(d.Get("auto_enable").(bool)),
	}

	log.Printf("[DEBUG] Updating Macie organization configuration: %+v", input)

	_, err := conn.UpdateOrganizationConfiguration(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Macie organization configuration failed: %s", err)
	}

	d.SetId("macie2-org-config")

	return append(diags, resourceOrganizationConfigurationRead(ctx, d, meta)...)
}

func resourceOrganizationConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).Macie2Client(ctx)

	input := macie2.DescribeOrganizationConfigurationInput{}
	output, err := conn.DescribeOrganizationConfiguration(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Macie organization configuration failed: %s", err)
	}

	d.Set("auto_enable", output.AutoEnable)

	return diags
}
