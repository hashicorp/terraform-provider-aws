package ec2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func DataSourceIPAMPoolCIDRs() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceIPAMPoolCIDRsRead,

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

func dataSourceIPAMPoolCIDRsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

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

	output, err := FindIPAMPoolCIDRs(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IPAM Pool CIDRs: %s", err)
	}

	d.SetId(poolID)
	d.Set("ipam_pool_cidrs", flattenIPAMPoolCIDRs(output))

	return diags
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
