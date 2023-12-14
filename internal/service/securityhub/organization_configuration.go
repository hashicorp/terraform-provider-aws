// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_securityhub_organization_configuration")
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
			"auto_enable_standards": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[types.AutoEnableStandards](),
			},
		},
	}
}

func resourceOrganizationConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	input := &securityhub.UpdateOrganizationConfigurationInput{
		AutoEnable: aws.Bool(d.Get("auto_enable").(bool)),
	}

	if v, ok := d.GetOk("auto_enable_standards"); ok {
		input.AutoEnableStandards = types.AutoEnableStandards(v.(string))
	}

	_, err := conn.UpdateOrganizationConfiguration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Security Hub Organization Configuration (%s): %s", d.Id(), err)
	}

	if d.IsNewResource() {
		d.SetId(meta.(*conns.AWSClient).AccountID)
	}

	return append(diags, resourceOrganizationConfigurationRead(ctx, d, meta)...)
}

func resourceOrganizationConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	output, err := FindOrganizationConfiguration(ctx, conn)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Security Hub Organization Configuration %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Security Hub Organization Configuration (%s): %s", d.Id(), err)
	}

	d.Set("auto_enable", output.AutoEnable)
	d.Set("auto_enable_standards", output.AutoEnableStandards)

	return diags
}

func FindOrganizationConfiguration(ctx context.Context, conn *securityhub.Client) (*securityhub.DescribeOrganizationConfigurationOutput, error) {
	input := &securityhub.DescribeOrganizationConfigurationInput{}

	output, err := conn.DescribeOrganizationConfiguration(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) || tfawserr.ErrMessageContains(err, errCodeInvalidAccessException, "not subscribed to AWS Security Hub") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.OrganizationConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
