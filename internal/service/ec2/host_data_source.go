package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func DataSourceHost() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceHostRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_placement": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cores": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"filter": DataSourceFiltersSchema(),
			"host_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"host_recovery": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_family": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sockets": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"total_vcpus": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceHostRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	host, err := FindHostByIDAndFilters(conn, d.Get("host_id").(string), BuildFiltersDataSource(d.Get("filter").(*schema.Set)))

	if err != nil {
		return tfresource.SingularDataSourceFindError("EC2 Host", err)
	}

	d.SetId(aws.StringValue(host.HostId))

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: aws.StringValue(host.OwnerId),
		Resource:  fmt.Sprintf("dedicated-host/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("auto_placement", host.AutoPlacement)
	d.Set("availability_zone", host.AvailabilityZone)
	d.Set("cores", host.HostProperties.Cores)
	d.Set("host_id", host.HostId)
	d.Set("host_recovery", host.HostRecovery)
	d.Set("instance_family", host.HostProperties.InstanceFamily)
	d.Set("instance_type", host.HostProperties.InstanceType)
	d.Set("owner_id", host.OwnerId)
	d.Set("sockets", host.HostProperties.Sockets)
	d.Set("total_vcpus", host.HostProperties.TotalVCpus)

	if err := d.Set("tags", KeyValueTags(host.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}
