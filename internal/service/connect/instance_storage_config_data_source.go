// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"

	awstypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_connect_instance_storage_config", name="Instance Storage Config")
func dataSourceInstanceStorageConfig() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceInstanceStorageConfigRead,

		Schema: map[string]*schema.Schema{
			names.AttrAssociationID: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			names.AttrInstanceID: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			names.AttrResourceType: {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.InstanceStorageResourceType](),
			},
			"storage_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kinesis_firehose_config": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"firehose_arn": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"kinesis_stream_config": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrStreamARN: {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"kinesis_video_stream_config": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"encryption_config": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"encryption_type": {
													Type:     schema.TypeString,
													Computed: true,
												},
												names.AttrKeyID: {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
									names.AttrPrefix: {
										Type:     schema.TypeString,
										Computed: true,
									},
									"retention_period_hours": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
						"s3_config": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrBucketName: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrBucketPrefix: {
										Type:     schema.TypeString,
										Computed: true,
									},
									"encryption_config": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"encryption_type": {
													Type:     schema.TypeString,
													Computed: true,
												},
												names.AttrKeyID: {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
						names.AttrStorageType: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceInstanceStorageConfigRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	associationID := d.Get(names.AttrAssociationID).(string)
	instanceID := d.Get(names.AttrInstanceID).(string)
	resourceType := awstypes.InstanceStorageResourceType(d.Get(names.AttrResourceType).(string))
	id := instanceStorageConfigCreateResourceID(instanceID, associationID, resourceType)
	storageConfig, err := findInstanceStorageConfigByThreePartKey(ctx, conn, instanceID, associationID, resourceType)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Instance Storage Config (%s): %s", id, err)
	}

	d.SetId(id)
	if err := d.Set("storage_config", flattenStorageConfig(storageConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting storage_config: %s", err)
	}

	return diags
}
