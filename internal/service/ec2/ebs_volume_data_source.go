package ec2

import (
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceEBSVolume() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceEBSVolumeRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encrypted": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"filter": DataSourceFiltersSchema(),
			"iops": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"most_recent": {
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
			"size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"snapshot_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"throughput": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"volume_type": {
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

func dataSourceEBSVolumeRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeVolumesInput{}

	input.Filters = append(input.Filters, BuildFiltersDataSource(
		d.Get("filter").(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	output, err := FindEBSVolumes(conn, input)

	if err != nil {
		return fmt.Errorf("reading EBS Volumes: %w", err)
	}

	if len(output) < 1 {
		return fmt.Errorf("Your query returned no results. Please change your search criteria and try again.")
	}

	var volume *ec2.Volume

	if len(output) > 1 {
		recent := d.Get("most_recent").(bool)

		if !recent {
			return fmt.Errorf("Your query returned more than one result. Please try a more " +
				"specific search criteria, or set `most_recent` attribute to true.")
		}

		volume = mostRecentVolume(output)
	} else {
		// Query returned single result.
		volume = output[0]
	}

	d.SetId(aws.StringValue(volume.VolumeId))

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("volume/%s", d.Id()),
	}
	d.Set("arn", arn.String())
	d.Set("availability_zone", volume.AvailabilityZone)
	d.Set("encrypted", volume.Encrypted)
	d.Set("iops", volume.Iops)
	d.Set("kms_key_id", volume.KmsKeyId)
	d.Set("multi_attach_enabled", volume.MultiAttachEnabled)
	d.Set("outpost_arn", volume.OutpostArn)
	d.Set("size", volume.Size)
	d.Set("snapshot_id", volume.SnapshotId)
	d.Set("throughput", volume.Throughput)
	d.Set("volume_id", volume.VolumeId)
	d.Set("volume_type", volume.VolumeType)

	if err := d.Set("tags", KeyValueTags(volume.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	return nil
}

type volumeSort []*ec2.Volume

func (a volumeSort) Len() int      { return len(a) }
func (a volumeSort) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a volumeSort) Less(i, j int) bool {
	itime := aws.TimeValue(a[i].CreateTime)
	jtime := aws.TimeValue(a[j].CreateTime)
	return itime.Unix() < jtime.Unix()
}

func mostRecentVolume(volumes []*ec2.Volume) *ec2.Volume {
	sortedVolumes := volumes
	sort.Sort(volumeSort(sortedVolumes))
	return sortedVolumes[len(sortedVolumes)-1]
}
