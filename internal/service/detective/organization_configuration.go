// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package detective

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/detective"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_detective_organization_configuration")
func ResourceOrganizationConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOrganizationConfigurationUpdate,
		ReadWithoutTimeout:   resourceOrganizationConfigurationRead,
		UpdateWithoutTimeout: resourceOrganizationConfigurationUpdate,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"auto_enable": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"graph_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceOrganizationConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).DetectiveClient(ctx)

	graphARN := d.Get("graph_arn").(string)
	input := &detective.UpdateOrganizationConfigurationInput{
		AutoEnable: d.Get("auto_enable").(bool),
		GraphArn:   aws.String(graphARN),
	}

	_, err := conn.UpdateOrganizationConfiguration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Detective Organization Configuration (%s): %s", graphARN, err)
	}

	if d.IsNewResource() {
		d.SetId(graphARN)
	}

	return append(diags, resourceOrganizationConfigurationRead(ctx, d, meta)...)
}

func resourceOrganizationConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).DetectiveClient(ctx)

	input := &detective.DescribeOrganizationConfigurationInput{
		GraphArn: aws.String(d.Id()),
	}

	output, err := conn.DescribeOrganizationConfiguration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Detective Organization Configuration (%s): %s", d.Id(), err)
	}

	d.Set("auto_enable", output.AutoEnable)
	d.Set("graph_arn", d.Id())

	return diags
}
