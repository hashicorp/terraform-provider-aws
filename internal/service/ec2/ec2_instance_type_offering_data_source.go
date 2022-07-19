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
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(ec2.LocationType_Values(), false),
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

	instanceTypeOfferings, err := FindInstanceTypeOfferings(conn, input)

	if err != nil {
		return fmt.Errorf("reading EC2 Instance Type Offerings: %w", err)
	}

	if len(instanceTypeOfferings) == 0 {
		return fmt.Errorf("no EC2 Instance Type Offerings found matching criteria; try different search")
	}

	var foundInstanceTypes []string

	for _, instanceTypeOffering := range instanceTypeOfferings {
		foundInstanceTypes = append(foundInstanceTypes, aws.StringValue(instanceTypeOffering.InstanceType))
	}

	var resultInstanceType string

	// Search preferred instance types in their given order and set result
	// instance type for first match found
	if v, ok := d.GetOk("preferred_instance_types"); ok {
		for _, v := range v.([]interface{}) {
			if v, ok := v.(string); ok {
				for _, foundInstanceType := range foundInstanceTypes {
					if foundInstanceType == v {
						resultInstanceType = v
						break
					}
				}

				if resultInstanceType != "" {
					break
				}
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

	d.SetId(resultInstanceType)
	d.Set("instance_type", resultInstanceType)

	return nil
}
