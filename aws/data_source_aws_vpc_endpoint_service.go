package aws

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsVpcEndpointService() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsVpcEndpointServiceRead,

		Schema: map[string]*schema.Schema{
			"service": {
				Type:     schema.TypeString,
				Required: true,
			},
			"service_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_endpoint_policy_supported": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"acceptance_required": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"availability_zones": &schema.Schema{
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				Set:      schema.HashString,
			},
			"private_dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"base_endpoint_dns_names": &schema.Schema{
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				Set:      schema.HashString,
			},
		},
	}
}

func dataSourceAwsVpcEndpointServiceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	service := d.Get("service").(string)

	log.Printf("[DEBUG] Reading VPC Endpoint Services.")

	request := &ec2.DescribeVpcEndpointServicesInput{}

	resp, err := conn.DescribeVpcEndpointServices(request)
	if err != nil {
		return fmt.Errorf("Error fetching VPC Endpoint Services: %s", err)
	}

	for _, sd := range resp.ServiceDetails {
		serviceName := aws.StringValue(sd.ServiceName)
		if strings.HasSuffix(serviceName, "."+service) {
			d.SetId(strconv.Itoa(hashcode.String(serviceName)))
			d.Set("service_name", serviceName)
			d.Set("service_type", sd.ServiceType[0].ServiceType)
			d.Set("owner", sd.Owner)
			d.Set("vpc_endpoint_policy_supported", sd.VpcEndpointPolicySupported)
			d.Set("acceptance_required", sd.AcceptanceRequired)
			d.Set("availability_zones", flattenStringList(sd.AvailabilityZones))
			d.Set("private_dns_name", sd.PrivateDnsName)
			d.Set("base_endpoint_dns_names", flattenStringList(sd.BaseEndpointDnsNames))

			return nil
		}
	}

	return fmt.Errorf("VPC Endpoint Service (%s) not found", service)
}
