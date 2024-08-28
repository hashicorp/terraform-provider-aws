// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_lambda_code_signing_config", name="Code Signing Config")
func dataSourceCodeSigningConfig() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceCodeSigningConfigRead,

		Schema: map[string]*schema.Schema{
			"allowed_publishers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"signing_profile_version_arns": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			names.AttrARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"config_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"policies": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"untrusted_artifact_on_deployment": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceCodeSigningConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	arn := d.Get(names.AttrARN).(string)
	output, err := findCodeSigningConfigByARN(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Lambda Code Signing Config (%s): %s", arn, err)
	}

	d.SetId(aws.ToString(output.CodeSigningConfigArn))
	if err := d.Set("allowed_publishers", flattenAllowedPublishers(output.AllowedPublishers)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting allowed_publishers: %s", err)
	}
	d.Set("config_id", output.CodeSigningConfigId)
	d.Set(names.AttrDescription, output.Description)
	d.Set("last_modified", output.LastModified)
	if err := d.Set("policies", []interface{}{map[string]interface{}{
		"untrusted_artifact_on_deployment": output.CodeSigningPolicies.UntrustedArtifactOnDeployment,
	}}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting policies: %s", err)
	}

	return diags
}
