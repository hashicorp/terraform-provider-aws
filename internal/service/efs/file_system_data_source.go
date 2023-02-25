package efs

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceFileSystem() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceFileSystemRead,

		Schema: map[string]*schema.Schema{
			"arn": {
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
			"encrypted": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"file_system_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"performance_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"throughput_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"provisioned_throughput_in_mibps": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"size_in_bytes": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"lifecycle_policy": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
		},
	}
}

func dataSourceFileSystemRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSConn()
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	tagsToMatch := tftags.New(d.Get("tags").(map[string]interface{})).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	describeEfsOpts := &efs.DescribeFileSystemsInput{}

	if v, ok := d.GetOk("creation_token"); ok {
		describeEfsOpts.CreationToken = aws.String(v.(string))
	}

	if v, ok := d.GetOk("file_system_id"); ok {
		describeEfsOpts.FileSystemId = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Reading EFS File System: %s", describeEfsOpts)
	describeResp, err := conn.DescribeFileSystemsWithContext(ctx, describeEfsOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EFS FileSystem: %s", err)
	}

	if describeResp == nil || len(describeResp.FileSystems) == 0 {
		return sdkdiag.AppendErrorf(diags, "reading EFS FileSystem: empty output")
	}

	results := describeResp.FileSystems

	if len(tagsToMatch) > 0 {
		var fileSystems []*efs.FileSystemDescription

		for _, fileSystem := range describeResp.FileSystems {
			tags := KeyValueTags(fileSystem.Tags)

			if !tags.ContainsAll(tagsToMatch) {
				continue
			}

			fileSystems = append(fileSystems, fileSystem)
		}

		results = fileSystems
	}

	if count := len(results); count != 1 {
		return sdkdiag.AppendErrorf(diags, "Search returned %d results, please revise so only one is returned", count)
	}

	fs := results[0]

	d.SetId(aws.StringValue(fs.FileSystemId))
	d.Set("availability_zone_id", fs.AvailabilityZoneId)
	d.Set("availability_zone_name", fs.AvailabilityZoneName)
	d.Set("creation_token", fs.CreationToken)
	d.Set("performance_mode", fs.PerformanceMode)
	d.Set("arn", fs.FileSystemArn)
	d.Set("file_system_id", fs.FileSystemId)
	d.Set("encrypted", fs.Encrypted)
	d.Set("kms_key_id", fs.KmsKeyId)
	d.Set("provisioned_throughput_in_mibps", fs.ProvisionedThroughputInMibps)
	d.Set("throughput_mode", fs.ThroughputMode)
	if fs.SizeInBytes != nil {
		d.Set("size_in_bytes", fs.SizeInBytes.Value)
	}

	if err := d.Set("tags", KeyValueTags(fs.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	res, err := conn.DescribeLifecycleConfigurationWithContext(ctx, &efs.DescribeLifecycleConfigurationInput{
		FileSystemId: fs.FileSystemId,
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing lifecycle configuration for EFS file system (%s): %s",
			aws.StringValue(fs.FileSystemId), err)
	}

	if err := d.Set("lifecycle_policy", flattenFileSystemLifecyclePolicies(res.LifecyclePolicies)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting lifecycle_policy: %s", err)
	}

	d.Set("dns_name", meta.(*conns.AWSClient).RegionalHostname(fmt.Sprintf("%s.efs", aws.StringValue(fs.FileSystemId))))

	return diags
}
