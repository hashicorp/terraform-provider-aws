// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ebs_volume", name="EBS Volume")
// @Tags
// @Testing(tagsTest=false)
func dataSourceEBSVolume() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceEBSVolumeRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAvailabilityZone: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEncrypted: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrFilter: customFiltersSchema(),
			names.AttrIOPS: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrKMSKeyID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrMostRecent: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"multi_attach_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"outpost_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrSize: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrSnapshotID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			names.AttrThroughput: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrVolumeType: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"volume_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceEBSVolumeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeVolumesInput{}

	input.Filters = append(input.Filters, newCustomFilterList(
		d.Get(names.AttrFilter).(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	output, err := findEBSVolumes(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EBS Volumes: %s", err)
	}

	if len(output) < 1 {
		return sdkdiag.AppendErrorf(diags, "Your query returned no results. Please change your search criteria and try again.")
	}

	var volume awstypes.Volume

	if len(output) > 1 {
		recent := d.Get(names.AttrMostRecent).(bool)

		if !recent {
			return sdkdiag.AppendErrorf(diags, "Your query returned more than one result. Please try a more "+
				"specific search criteria, or set `most_recent` attribute to true.")
		}

		volume = mostRecentVolume(output)
	} else {
		// Query returned single result.
		volume = output[0]
	}

	d.SetId(aws.ToString(volume.VolumeId))

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   names.EC2,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("volume/%s", d.Id()),
	}
	d.Set(names.AttrARN, arn.String())
	d.Set(names.AttrAvailabilityZone, volume.AvailabilityZone)
	d.Set(names.AttrEncrypted, volume.Encrypted)
	d.Set(names.AttrIOPS, volume.Iops)
	d.Set(names.AttrKMSKeyID, volume.KmsKeyId)
	d.Set("multi_attach_enabled", volume.MultiAttachEnabled)
	d.Set("outpost_arn", volume.OutpostArn)
	d.Set(names.AttrSize, volume.Size)
	d.Set(names.AttrSnapshotID, volume.SnapshotId)
	d.Set(names.AttrThroughput, volume.Throughput)
	d.Set("volume_id", volume.VolumeId)
	d.Set(names.AttrVolumeType, volume.VolumeType)

	setTagsOut(ctx, volume.Tags)

	return diags
}

type volumeSort []awstypes.Volume

func (a volumeSort) Len() int      { return len(a) }
func (a volumeSort) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a volumeSort) Less(i, j int) bool {
	itime := aws.ToTime(a[i].CreateTime)
	jtime := aws.ToTime(a[j].CreateTime)
	return itime.Unix() < jtime.Unix()
}

func mostRecentVolume(volumes []awstypes.Volume) awstypes.Volume {
	sortedVolumes := volumes
	sort.Sort(volumeSort(sortedVolumes))
	return sortedVolumes[len(sortedVolumes)-1]
}
