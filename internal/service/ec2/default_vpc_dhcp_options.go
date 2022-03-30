package ec2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceDefaultVPCDHCPOptions() *schema.Resource {
	// reuse aws_vpc_dhcp_options schema, and methods for READ, UPDATE
	dvpc := ResourceVPCDHCPOptions()
	dvpc.Create = resourceDefaultVPCDHCPOptionsCreate
	dvpc.Delete = resourceDefaultVPCDHCPOptionsDelete

	// domain_name is a computed value for Default Default DHCP Options Sets
	dvpc.Schema["domain_name"] = &schema.Schema{
		Type:     schema.TypeString,
		Computed: true,
	}
	// domain_name_servers is a computed value for Default Default DHCP Options Sets
	dvpc.Schema["domain_name_servers"] = &schema.Schema{
		Type:     schema.TypeString,
		Computed: true,
	}
	// ntp_servers is a computed value for Default Default DHCP Options Sets
	dvpc.Schema["ntp_servers"] = &schema.Schema{
		Type:     schema.TypeString,
		Computed: true,
	}

	dvpc.Schema["owner_id"] = &schema.Schema{
		Type:     schema.TypeString,
		Computed: true,
		Optional: true,
	}

	return dvpc
}

func resourceDefaultVPCDHCPOptionsCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	filters := []*ec2.Filter{
		{
			Name:   aws.String("key"),
			Values: aws.StringSlice([]string{"domain-name"}),
		},
		{
			Name:   aws.String("value"),
			Values: aws.StringSlice([]string{RegionalPrivateDNSSuffix(meta.(*conns.AWSClient).Region)}),
		},
		{
			Name:   aws.String("key"),
			Values: aws.StringSlice([]string{"domain-name-servers"}),
		},
		{
			Name:   aws.String("value"),
			Values: aws.StringSlice([]string{"AmazonProvidedDNS"}),
		},
	}

	if v, ok := d.GetOk("owner_id"); ok {
		filter := &ec2.Filter{
			Name:   aws.String("owner-id"),
			Values: aws.StringSlice([]string{v.(string)}),
		}

		filters = append(filters, filter)
	}

	req := &ec2.DescribeDhcpOptionsInput{
		Filters: filters,
	}

	var dhcpOptions []*ec2.DhcpOptions
	err := conn.DescribeDhcpOptionsPages(req, func(page *ec2.DescribeDhcpOptionsOutput, lastPage bool) bool {
		dhcpOptions = append(dhcpOptions, page.DhcpOptions...)
		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("Error describing DHCP options: %s", err)
	}

	if len(dhcpOptions) == 0 {
		return fmt.Errorf("Default DHCP Options Set not found")
	}

	if len(dhcpOptions) > 1 {
		return fmt.Errorf("Multiple default DHCP Options Sets found")
	}

	if dhcpOptions[0] == nil {
		return fmt.Errorf("Default DHCP Options Set is empty")
	}
	d.SetId(aws.StringValue(dhcpOptions[0].DhcpOptionsId))

	return resourceVPCDHCPOptionsUpdate(d, meta)
}

func resourceDefaultVPCDHCPOptionsDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[WARN] Cannot destroy Default DHCP Options Set. Terraform will remove this resource from the state file, however resources may remain.")
	return nil
}

func RegionalPrivateDNSSuffix(region string) string {
	if region == endpoints.UsEast1RegionID {
		return "ec2.internal"
	}

	return fmt.Sprintf("%s.compute.internal", region)
}

func RegionalPublicDNSSuffix(region string) string {
	if region == endpoints.UsEast1RegionID {
		return "compute-1"
	}

	return fmt.Sprintf("%s.compute", region)
}
