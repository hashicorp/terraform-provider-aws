package ec2

import (
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceEBSSnapshot() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceEBSSnapshotRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"data_encryption_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encrypted": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"filter": DataSourceFiltersSchema(),
			"kms_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"most_recent": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"outpost_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_alias": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owners": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"restorable_by_user_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"snapshot_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"snapshot_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"storage_tier": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"volume_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"volume_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceEBSSnapshotRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeSnapshotsInput{}

	if v, ok := d.GetOk("owners"); ok && len(v.([]interface{})) > 0 {
		input.OwnerIds = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("restorable_by_user_ids"); ok && len(v.([]interface{})) > 0 {
		input.RestorableByUserIds = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("snapshot_ids"); ok && len(v.([]interface{})) > 0 {
		input.SnapshotIds = flex.ExpandStringList(v.([]interface{}))
	}

	input.Filters = append(input.Filters, BuildFiltersDataSource(
		d.Get("filter").(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	snapshots, err := FindSnapshots(conn, input)

	if err != nil {
		return fmt.Errorf("reading EBS Snapshots: %w", err)
	}

	if len(snapshots) < 1 {
		return fmt.Errorf("Your query returned no results. Please change your search criteria and try again.")
	}

	if len(snapshots) > 1 {
		if !d.Get("most_recent").(bool) {
			return fmt.Errorf("Your query returned more than one result. Please try a more " +
				"specific search criteria, or set `most_recent` attribute to true.")
		}

		sort.Slice(snapshots, func(i, j int) bool {
			return aws.TimeValue(snapshots[i].StartTime).Unix() > aws.TimeValue(snapshots[j].StartTime).Unix()
		})
	}

	snapshot := snapshots[0]

	d.SetId(aws.StringValue(snapshot.SnapshotId))
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("snapshot/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("data_encryption_key_id", snapshot.DataEncryptionKeyId)
	d.Set("description", snapshot.Description)
	d.Set("encrypted", snapshot.Encrypted)
	d.Set("kms_key_id", snapshot.KmsKeyId)
	d.Set("outpost_arn", snapshot.OutpostArn)
	d.Set("owner_alias", snapshot.OwnerAlias)
	d.Set("owner_id", snapshot.OwnerId)
	d.Set("snapshot_id", snapshot.SnapshotId)
	d.Set("state", snapshot.State)
	d.Set("storage_tier", snapshot.StorageTier)
	d.Set("volume_id", snapshot.VolumeId)
	d.Set("volume_size", snapshot.VolumeSize)

	if err := d.Set("tags", KeyValueTags(snapshot.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	return nil
}
