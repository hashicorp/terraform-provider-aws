// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package detective

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/detective"
	awstypes "github.com/aws/aws-sdk-go-v2/service/detective/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_detective_organization_configuration", name="Organization Configuration")
// @ArnIdentity("graph_arn")
// @Testing(preIdentityVersion="v6.50.0")
// @Testing(serialize=true)
// @Testing(generator=false)
// @Testing(preCheck="github.com/hashicorp/terraform-provider-aws/internal/acctest;acctest.PreCheckOrganizationManagementAccount")
func resourceOrganizationConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOrganizationConfigurationUpdate,
		ReadWithoutTimeout:   resourceOrganizationConfigurationRead,
		UpdateWithoutTimeout: resourceOrganizationConfigurationUpdate,
		DeleteWithoutTimeout: schema.NoopContext,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
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
			}
		},
	}
}

func resourceOrganizationConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DetectiveClient(ctx)

	graphARN := d.Get("graph_arn").(string)
	input := detective.UpdateOrganizationConfigurationInput{
		AutoEnable: d.Get("auto_enable").(bool),
		GraphArn:   aws.String(graphARN),
	}

	_, err := conn.UpdateOrganizationConfiguration(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Detective Organization Configuration (%s): %s", graphARN, err)
	}

	if d.IsNewResource() {
		d.SetId(graphARN)
	}

	return append(diags, resourceOrganizationConfigurationRead(ctx, d, meta)...)
}

func resourceOrganizationConfigurationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DetectiveClient(ctx)

	output, err := findOrganizationConfigurationByGraphARN(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Detective Organization Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Detective Organization Configuration (%s): %s", d.Id(), err)
	}

	d.Set("auto_enable", output.AutoEnable)
	d.Set("graph_arn", d.Id())

	return diags
}

func findOrganizationConfigurationByGraphARN(ctx context.Context, conn *detective.Client, graphARN string) (*detective.DescribeOrganizationConfigurationOutput, error) {
	input := detective.DescribeOrganizationConfigurationInput{
		GraphArn: aws.String(graphARN),
	}

	return findOrganizationConfiguration(ctx, conn, &input)
}

func findOrganizationConfiguration(ctx context.Context, conn *detective.Client, input *detective.DescribeOrganizationConfigurationInput) (*detective.DescribeOrganizationConfigurationOutput, error) {
	output, err := conn.DescribeOrganizationConfiguration(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "a delegated administrator account has not been enabled") {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}
