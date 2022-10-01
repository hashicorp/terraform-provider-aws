package ec2

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
			// computed
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

	input := &ec2.GetIpamPoolCidrsInput{}

	if v, ok := d.GetOk("ipam_pool_id"); ok {
		input.IpamPoolId = aws.String(v.(string))
	}

	filters, filtersOk := d.GetOk("filter")
	if filtersOk {
		input.Filters = BuildFiltersDataSource(filters.(*schema.Set))
	}

	output, err := conn.GetIpamPoolCidrs(input)
	var cidrs []*ec2.IpamPoolCidr

	if err != nil {
		return err
	}

	if len(output.IpamPoolCidrs) == 0 || output.IpamPoolCidrs[0] == nil {
		return tfresource.SingularDataSourceFindError("EC2 VPC IPAM POOL CIDRS", tfresource.NewEmptyResultError(input))
	}

	cidrs = output.IpamPoolCidrs

	d.SetId(d.Get("ipam_pool_id").(string))
	d.Set("ipam_pool_cidrs", flattenIPAMPoolCIDRs(cidrs))

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
