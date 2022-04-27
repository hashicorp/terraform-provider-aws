package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceInstanceTypeOffering() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceInstanceTypeOfferingRead,

		Schema: map[string]*schema.Schema{
			"filter": DataSourceFiltersSchema(),
			"instance_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"location_type": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					ec2.LocationTypeAvailabilityZone,
					ec2.LocationTypeAvailabilityZoneId,
					ec2.LocationTypeRegion,
				}, false),
			},
			"preferred_instance_types": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceInstanceTypeOfferingRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeInstanceTypeOfferingsInput{}

	if v, ok := d.GetOk("filter"); ok {
		input.Filters = BuildFiltersDataSource(v.(*schema.Set))
	}

	if v, ok := d.GetOk("location_type"); ok {
		input.LocationType = aws.String(v.(string))
	}

	var foundInstanceTypes []string

	for {
		output, err := conn.DescribeInstanceTypeOfferings(input)

		if err != nil {
			return fmt.Errorf("error reading EC2 Instance Type Offerings: %w", err)
		}

		if output == nil {
			break
		}

		for _, instanceTypeOffering := range output.InstanceTypeOfferings {
			if instanceTypeOffering == nil {
				continue
			}

			foundInstanceTypes = append(foundInstanceTypes, aws.StringValue(instanceTypeOffering.InstanceType))
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	if len(foundInstanceTypes) == 0 {
		return fmt.Errorf("no EC2 Instance Type Offerings found matching criteria; try different search")
	}

	var resultInstanceType string

	// Search preferred instance types in their given order and set result
	// instance type for first match found
	if l := d.Get("preferred_instance_types").([]interface{}); len(l) > 0 {
		for _, elem := range l {
			preferredInstanceType, ok := elem.(string)

			if !ok {
				continue
			}

			for _, foundInstanceType := range foundInstanceTypes {
				if foundInstanceType == preferredInstanceType {
					resultInstanceType = preferredInstanceType
					break
				}
			}

			if resultInstanceType != "" {
				break
			}
		}
	}

	if resultInstanceType == "" && len(foundInstanceTypes) > 1 {
		return fmt.Errorf("multiple EC2 Instance Offerings found matching criteria; try different search")
	}

	if resultInstanceType == "" && len(foundInstanceTypes) == 1 {
		resultInstanceType = foundInstanceTypes[0]
	}

	if resultInstanceType == "" {
		return fmt.Errorf("no EC2 Instance Type Offerings found matching criteria; try different search")
	}

	d.Set("instance_type", resultInstanceType)

	d.SetId(resultInstanceType)

	return nil
}
