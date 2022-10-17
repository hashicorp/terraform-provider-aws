package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDefaultVPCDHCPOptions() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		Create: resourceDefaultVPCDHCPOptionsCreate,
		Read:   resourceVPCDHCPOptionsRead,
		Update: resourceVPCDHCPOptionsUpdate,
		Delete: schema.Noop,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: verify.SetTagsDiff,

		// Keep in sync with aws_vpc_dhcp_options' schema with the following changes:
		//   - domain_name is Computed-only
		//   - domain_name_servers is Computed-only and is TypeString
		//   - netbios_name_servers is Computed-only and is TypeString
		//   - netbios_node_type is Computed-only
		//   - ntp_servers is Computed-only and is TypeString
		//   - owner_id is Optional/Computed
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_name_servers": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"netbios_name_servers": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"netbios_node_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ntp_servers": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceDefaultVPCDHCPOptionsCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeDhcpOptionsInput{}

	input.Filters = append(input.Filters,
		NewFilter("key", []string{"domain-name"}),
		NewFilter("value", []string{RegionalPrivateDNSSuffix(meta.(*conns.AWSClient).Region)}),
		NewFilter("key", []string{"domain-name-servers"}),
		NewFilter("value", []string{"AmazonProvidedDNS"}),
	)

	if v, ok := d.GetOk("owner_id"); ok {
		input.Filters = append(input.Filters, BuildAttributeFilterList(map[string]string{
			"owner-id": v.(string),
		})...)
	}

	dhcpOptions, err := FindDHCPOptions(conn, input)

	if err != nil {
		return fmt.Errorf("reading EC2 Default DHCP Options Set: %w", err)
	}

	d.SetId(aws.StringValue(dhcpOptions.DhcpOptionsId))

	return resourceVPCDHCPOptionsUpdate(d, meta)
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
