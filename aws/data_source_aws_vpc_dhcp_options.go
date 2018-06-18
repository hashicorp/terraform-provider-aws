package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/validation"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsVpcDhcpOptions() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsVpcDhcpOptionsRead,

		Schema: map[string]*schema.Schema{
			"dhcp_options_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_name_servers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"netbios_name_servers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"netbios_node_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ntp_servers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsVpcDhcpOptionsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	dhcpOptionID := d.Get("dhcp_options_id").(string)

	input := &ec2.DescribeDhcpOptionsInput{
		DhcpOptionsIds: []*string{aws.String(dhcpOptionID)},
	}

	log.Printf("[DEBUG] Reading EC2 DHCP Options: %s", input)
	output, err := conn.DescribeDhcpOptions(input)
	if err != nil {
		if isNoSuchDhcpOptionIDErr(err) {
			return fmt.Errorf("EC2 DHCP Options %q not found", dhcpOptionID)
		}
		return fmt.Errorf("error reading EC2 DHCP Options: %s", err)
	}

	if len(output.DhcpOptions) == 0 {
		return fmt.Errorf("EC2 DHCP Options %q not found", dhcpOptionID)
	}

	d.SetId(dhcpOptionID)

	dhcpConfigurations := output.DhcpOptions[0].DhcpConfigurations

	for _, dhcpConfiguration := range dhcpConfigurations {
		key := aws.StringValue(dhcpConfiguration.Key)
		tfKey := strings.Replace(key, "-", "_", -1)

		if len(dhcpConfiguration.Values) == 0 {
			continue
		}

		switch key {
		case "domain-name":
			d.Set(tfKey, aws.StringValue(dhcpConfiguration.Values[0].Value))
		case "domain-name-servers":
			if err := d.Set(tfKey, flattenEc2AttributeValues(dhcpConfiguration.Values)); err != nil {
				return fmt.Errorf("error setting %s: %s", tfKey, err)
			}
		case "netbios-name-servers":
			if err := d.Set(tfKey, flattenEc2AttributeValues(dhcpConfiguration.Values)); err != nil {
				return fmt.Errorf("error setting %s: %s", tfKey, err)
			}
		case "netbios-node-type":
			d.Set(tfKey, aws.StringValue(dhcpConfiguration.Values[0].Value))
		case "ntp-servers":
			if err := d.Set(tfKey, flattenEc2AttributeValues(dhcpConfiguration.Values)); err != nil {
				return fmt.Errorf("error setting %s: %s", tfKey, err)
			}
		}
	}

	if err := d.Set("tags", d.Set("tags", tagsToMap(output.DhcpOptions[0].Tags))); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}
