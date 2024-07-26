// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudformation

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_cloudformation_type", name="Type")
func dataSourceType() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceTypeRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"default_version_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deprecated_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"documentation_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrExecutionRoleARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_default_version": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"logging_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrLogGroupName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"log_role_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"provisioning_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrSchema: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"source_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrType: {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[awstypes.RegistryType](),
			},
			"type_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(10, 204),
					validation.StringMatch(regexache.MustCompile(`[0-9A-Za-z]{2,64}::[0-9A-Za-z]{2,64}::[0-9A-Za-z]{2,64}(::MODULE){0,1}`), "three alphanumeric character sections separated by double colons (::)"),
				),
			},
			"version_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"visibility": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceTypeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFormationClient(ctx)

	input := &cloudformation.DescribeTypeInput{}

	if v, ok := d.GetOk(names.AttrARN); ok {
		input.Arn = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrType); ok {
		input.Type = awstypes.RegistryType(v.(string))
	}

	if v, ok := d.GetOk("type_name"); ok {
		input.TypeName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("version_id"); ok {
		input.VersionId = aws.String(v.(string))
	}

	output, err := findType(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFormation Type: %s", err)
	}

	d.SetId(aws.ToString(output.Arn))
	d.Set(names.AttrARN, output.Arn)
	d.Set("default_version_id", output.DefaultVersionId)
	d.Set("deprecated_status", output.DeprecatedStatus)
	d.Set(names.AttrDescription, output.Description)
	d.Set("documentation_url", output.DocumentationUrl)
	d.Set(names.AttrExecutionRoleARN, output.ExecutionRoleArn)
	d.Set("is_default_version", output.IsDefaultVersion)
	if output.LoggingConfig != nil {
		if err := d.Set("logging_config", []interface{}{flattenLoggingConfig(output.LoggingConfig)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting logging_config: %s", err)
		}
	} else {
		d.Set("logging_config", nil)
	}
	d.Set("provisioning_type", output.ProvisioningType)
	d.Set(names.AttrSchema, output.Schema)
	d.Set("source_url", output.SourceUrl)
	d.Set(names.AttrType, output.Type)
	d.Set("type_name", output.TypeName)
	d.Set("visibility", output.Visibility)

	return diags
}
