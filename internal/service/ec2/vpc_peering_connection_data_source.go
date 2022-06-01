package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func DataSourceVPCPeeringConnection() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVPCPeeringConnectionRead,

		Schema: map[string]*schema.Schema{
			"accepter": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeBool},
			},
			"cidr_block": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"cidr_block_set": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cidr_block": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"filter": CustomFiltersSchema(),
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"peer_cidr_block": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"peer_cidr_block_set": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cidr_block": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"peer_owner_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"peer_region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"peer_vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"requester": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeBool},
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceVPCPeeringConnectionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeVpcPeeringConnectionsInput{}

	if v, ok := d.GetOk("id"); ok {
		input.VpcPeeringConnectionIds = aws.StringSlice([]string{v.(string)})
	}

	input.Filters = BuildAttributeFilterList(
		map[string]string{
			"status-code":                   d.Get("status").(string),
			"requester-vpc-info.vpc-id":     d.Get("vpc_id").(string),
			"requester-vpc-info.owner-id":   d.Get("owner_id").(string),
			"requester-vpc-info.cidr-block": d.Get("cidr_block").(string),
			"accepter-vpc-info.vpc-id":      d.Get("peer_vpc_id").(string),
			"accepter-vpc-info.owner-id":    d.Get("peer_owner_id").(string),
			"accepter-vpc-info.cidr-block":  d.Get("peer_cidr_block").(string),
		},
	)

	if tags, tagsOk := d.GetOk("tags"); tagsOk {
		input.Filters = append(input.Filters, BuildTagFilterList(
			Tags(tftags.New(tags.(map[string]interface{}))),
		)...)
	}

	input.Filters = append(input.Filters, BuildCustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	vpcPeeringConnection, err := FindVPCPeeringConnection(conn, input)

	if err != nil {
		return tfresource.SingularDataSourceFindError("EC2 VPC Peering Connection", err)
	}

	d.SetId(aws.StringValue(vpcPeeringConnection.VpcPeeringConnectionId))
	d.Set("status", vpcPeeringConnection.Status.Code)
	d.Set("vpc_id", vpcPeeringConnection.RequesterVpcInfo.VpcId)
	d.Set("owner_id", vpcPeeringConnection.RequesterVpcInfo.OwnerId)
	d.Set("cidr_block", vpcPeeringConnection.RequesterVpcInfo.CidrBlock)

	cidrBlockSet := []interface{}{}
	for _, v := range vpcPeeringConnection.RequesterVpcInfo.CidrBlockSet {
		cidrBlockSet = append(cidrBlockSet, map[string]interface{}{
			"cidr_block": aws.StringValue(v.CidrBlock),
		})
	}
	if err := d.Set("cidr_block_set", cidrBlockSet); err != nil {
		return fmt.Errorf("error setting cidr_block_set: %w", err)
	}

	d.Set("region", vpcPeeringConnection.RequesterVpcInfo.Region)
	d.Set("peer_vpc_id", vpcPeeringConnection.AccepterVpcInfo.VpcId)
	d.Set("peer_owner_id", vpcPeeringConnection.AccepterVpcInfo.OwnerId)
	d.Set("peer_cidr_block", vpcPeeringConnection.AccepterVpcInfo.CidrBlock)

	peerCidrBlockSet := []interface{}{}
	for _, v := range vpcPeeringConnection.AccepterVpcInfo.CidrBlockSet {
		peerCidrBlockSet = append(peerCidrBlockSet, map[string]interface{}{
			"cidr_block": aws.StringValue(v.CidrBlock),
		})
	}
	if err := d.Set("peer_cidr_block_set", peerCidrBlockSet); err != nil {
		return fmt.Errorf("error setting peer_cidr_block_set: %w", err)
	}

	d.Set("peer_region", vpcPeeringConnection.AccepterVpcInfo.Region)

	if err := d.Set("tags", KeyValueTags(vpcPeeringConnection.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if vpcPeeringConnection.AccepterVpcInfo.PeeringOptions != nil {
		if err := d.Set("accepter", flattenVPCPeeringConnectionOptionsDescription(vpcPeeringConnection.AccepterVpcInfo.PeeringOptions)); err != nil {
			return fmt.Errorf("error setting accepter: %w", err)
		}
	}

	if vpcPeeringConnection.RequesterVpcInfo.PeeringOptions != nil {
		if err := d.Set("requester", flattenVPCPeeringConnectionOptionsDescription(vpcPeeringConnection.RequesterVpcInfo.PeeringOptions)); err != nil {
			return fmt.Errorf("error setting requester: %w", err)
		}
	}

	return nil
}
