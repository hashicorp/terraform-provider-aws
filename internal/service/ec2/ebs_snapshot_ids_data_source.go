package ec2

import (
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func DataSourceEBSSnapshotIDs() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceEBSSnapshotIDsRead,

		Schema: map[string]*schema.Schema{
			"filter": DataSourceFiltersSchema(),
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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
		},
	}
}

func dataSourceEBSSnapshotIDsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeSnapshotsInput{}

	if v, ok := d.GetOk("owners"); ok && len(v.([]interface{})) > 0 {
		input.OwnerIds = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("restorable_by_user_ids"); ok && len(v.([]interface{})) > 0 {
		input.RestorableByUserIds = flex.ExpandStringList(v.([]interface{}))
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

	sort.Slice(snapshots, func(i, j int) bool {
		return aws.TimeValue(snapshots[i].StartTime).Unix() > aws.TimeValue(snapshots[j].StartTime).Unix()
	})

	var snapshotIDs []string

	for _, v := range snapshots {
		snapshotIDs = append(snapshotIDs, aws.StringValue(v.SnapshotId))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("ids", snapshotIDs)

	return nil
}
