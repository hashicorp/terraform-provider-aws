package ec2

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceIPAMPoolCIDRs() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceIPAMPoolCIDRsRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(1 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"filter": DataSourceFiltersSchema(),
			"ipam_pool_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ipam_pool_cidrs": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cidr": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"state": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceIPAMPoolCIDRsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	poolID := d.Get("ipam_pool_id").(string)
	input := &ec2.GetIpamPoolCidrsInput{
		IpamPoolId: aws.String(poolID),
	}

	input.Filters = append(input.Filters, BuildFiltersDataSource(
		d.Get("filter").(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	output, err := FindIPAMPoolCIDRs(conn, input)

	if err != nil {
		return fmt.Errorf("reading IPAM Pool CIDRs: %w", err)
	}

	d.SetId(poolID)
	d.Set("ipam_pool_cidrs", flattenIPAMPoolCIDRs(output))

	return nil
}

func flattenIPAMPoolCIDRs(c []*ec2.IpamPoolCidr) []interface{} {
	cidrs := []interface{}{}
	for _, cidr := range c {
		cidrs = append(cidrs, flattenIPAMPoolCIDR(cidr))
	}
	return cidrs
}

func flattenIPAMPoolCIDR(c *ec2.IpamPoolCidr) map[string]interface{} {
	cidr := make(map[string]interface{})
	cidr["cidr"] = aws.StringValue(c.Cidr)
	cidr["state"] = aws.StringValue(c.State)
	return cidr
}
