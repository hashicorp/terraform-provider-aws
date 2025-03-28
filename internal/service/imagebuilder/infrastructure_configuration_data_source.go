// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_imagebuilder_infrastructure_configuration", name="Infrastructure Configuration")
// @Tags
func dataSourceInfrastructureConfiguration() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceInfrastructureConfigurationRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"date_created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"date_updated": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_metadata_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"http_put_response_hop_limit": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"http_tokens": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"instance_profile_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_types": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"key_pair": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"logging": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"s3_logs": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrS3BucketName: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrS3KeyPrefix: {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrResourceTags: tftags.TagsSchemaComputed(),
			names.AttrSecurityGroupIDs: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrSNSTopicARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrSubnetID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"terminate_instance_on_failure": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceInfrastructureConfigurationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	arn := d.Get(names.AttrARN).(string)
	infrastructureConfiguration, err := findInfrastructureConfigurationByARN(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Image Builder Infrastructure Configuration (%s): %s", arn, err)
	}

	d.SetId(aws.ToString(infrastructureConfiguration.Arn))
	d.Set(names.AttrARN, infrastructureConfiguration.Arn)
	d.Set("date_created", infrastructureConfiguration.DateCreated)
	d.Set("date_updated", infrastructureConfiguration.DateUpdated)
	d.Set(names.AttrDescription, infrastructureConfiguration.Description)
	if infrastructureConfiguration.InstanceMetadataOptions != nil {
		if err := d.Set("instance_metadata_options", []any{flattenInstanceMetadataOptions(infrastructureConfiguration.InstanceMetadataOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting instance_metadata_options: %s", err)
		}
	} else {
		d.Set("instance_metadata_options", nil)
	}
	d.Set("instance_profile_name", infrastructureConfiguration.InstanceProfileName)
	d.Set("instance_types", infrastructureConfiguration.InstanceTypes)
	d.Set("key_pair", infrastructureConfiguration.KeyPair)
	if infrastructureConfiguration.Logging != nil {
		if err := d.Set("logging", []any{flattenLogging(infrastructureConfiguration.Logging)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting logging: %s", err)
		}
	} else {
		d.Set("logging", nil)
	}
	d.Set(names.AttrName, infrastructureConfiguration.Name)
	d.Set(names.AttrResourceTags, keyValueTags(ctx, infrastructureConfiguration.ResourceTags).Map())
	d.Set(names.AttrSecurityGroupIDs, infrastructureConfiguration.SecurityGroupIds)
	d.Set(names.AttrSNSTopicARN, infrastructureConfiguration.SnsTopicArn)
	d.Set(names.AttrSubnetID, infrastructureConfiguration.SubnetId)
	d.Set("terminate_instance_on_failure", infrastructureConfiguration.TerminateInstanceOnFailure)

	setTagsOut(ctx, infrastructureConfiguration.Tags)

	return diags
}
