// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package efs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/efs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_efs_file_system", name="File System")
// @Tags
func dataSourceFileSystem() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceFileSystemRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zone_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_token": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(0, 64),
			},
			names.AttrDNSName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEncrypted: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrFileSystemID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrKMSKeyID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"lifecycle_policy": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"transition_to_archive": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"transition_to_ia": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"transition_to_primary_storage_class": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"performance_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"protection": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"replication_overwrite": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"provisioned_throughput_in_mibps": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"size_in_bytes": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"throughput_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceFileSystemRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &efs.DescribeFileSystemsInput{}

	if v, ok := d.GetOk("creation_token"); ok {
		input.CreationToken = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrFileSystemID); ok {
		input.FileSystemId = aws.String(v.(string))
	}

	filter := tfslices.PredicateTrue[*awstypes.FileSystemDescription]()

	if tagsToMatch := tftags.New(ctx, d.Get(names.AttrTags).(map[string]interface{})).IgnoreAWS().IgnoreConfig(ignoreTagsConfig); len(tagsToMatch) > 0 {
		filter = func(v *awstypes.FileSystemDescription) bool {
			return KeyValueTags(ctx, v.Tags).ContainsAll(tagsToMatch)
		}
	}

	fs, err := findFileSystem(ctx, conn, input, filter)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("EFS File System", err))
	}

	fsID := aws.ToString(fs.FileSystemId)
	d.SetId(fsID)
	d.Set(names.AttrARN, fs.FileSystemArn)
	d.Set("availability_zone_id", fs.AvailabilityZoneId)
	d.Set("availability_zone_name", fs.AvailabilityZoneName)
	d.Set("creation_token", fs.CreationToken)
	d.Set(names.AttrDNSName, meta.(*conns.AWSClient).RegionalHostname(ctx, d.Id()+".efs"))
	d.Set(names.AttrFileSystemID, fsID)
	d.Set(names.AttrEncrypted, fs.Encrypted)
	d.Set(names.AttrKMSKeyID, fs.KmsKeyId)
	d.Set(names.AttrName, fs.Name)
	d.Set("performance_mode", fs.PerformanceMode)
	if err := d.Set("protection", flattenFileSystemProtectionDescription(fs.FileSystemProtection)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting protection: %s", err)
	}
	d.Set("provisioned_throughput_in_mibps", fs.ProvisionedThroughputInMibps)
	if fs.SizeInBytes != nil {
		d.Set("size_in_bytes", fs.SizeInBytes.Value)
	}
	d.Set("throughput_mode", fs.ThroughputMode)

	setTagsOut(ctx, fs.Tags)

	output, err := conn.DescribeLifecycleConfiguration(ctx, &efs.DescribeLifecycleConfigurationInput{
		FileSystemId: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EFS File System (%s) lifecycle configuration: %s", d.Id(), err)
	}

	if err := d.Set("lifecycle_policy", flattenLifecyclePolicies(output.LifecyclePolicies)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting lifecycle_policy: %s", err)
	}

	return diags
}
