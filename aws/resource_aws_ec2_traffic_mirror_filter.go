package aws

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsEc2TrafficMirrorFilter() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEc2TrafficMirrorinFilterCreate,
		Read:   resourceAwsEc2TrafficMirrorFilterRead,
		Update: resourceAwsEc2TrafficMirrorFilterUpdate,
		Delete: resourceAwsEc2TrafficMirrorFilterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"network_services": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"amazon-dns",
					}, false),
				},
			},
		},
	}
}

func resourceAwsEc2TrafficMirrorinFilterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.CreateTrafficMirrorFilterInput{}

	if description, ok := d.GetOk("description"); ok {
		input.Description = aws.String(description.(string))
	}

	out, err := conn.CreateTrafficMirrorFilter(input)
	if err != nil {
		return fmt.Errorf("Error while creating traffic filter %s", err)
	}

	d.Partial(true)
	d.SetPartial("description")
	d.Partial(false)

	d.SetId(*out.TrafficMirrorFilter.TrafficMirrorFilterId)

	return resourceAwsEc2TrafficMirrorFilterUpdate(d, meta)
}

func resourceAwsEc2TrafficMirrorFilterUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	filterId := d.Id()

	d.Partial(true)
	if d.HasChange("network_services") {
		input := &ec2.ModifyTrafficMirrorFilterNetworkServicesInput{
			TrafficMirrorFilterId: aws.String(filterId),
		}

		o, n := d.GetChange("network_services")
		newServices := n.(*schema.Set).Difference(o.(*schema.Set)).List()
		if len(newServices) > 0 {
			input.SetAddNetworkServices(expandStringList(newServices))
		}

		removeServices := o.(*schema.Set).Difference(n.(*schema.Set)).List()
		if len(removeServices) > 0 {
			input.SetRemoveNetworkServices(expandStringList(removeServices))
		}

		_, err := conn.ModifyTrafficMirrorFilterNetworkServices(input)
		if err != nil {
			return fmt.Errorf("Error modifying network services for traffic mirror filter %v", filterId)
		}

		d.SetPartial("network_services")
	}
	d.Partial(false)

	return resourceAwsEc2TrafficMirrorFilterRead(d, meta)
}

func resourceAwsEc2TrafficMirrorFilterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	filterId := d.Id()

	var filterIds []*string
	filterIds = append(filterIds, &filterId)

	input := &ec2.DescribeTrafficMirrorFiltersInput{
		TrafficMirrorFilterIds: filterIds,
	}

	out, err := conn.DescribeTrafficMirrorFilters(input)
	if err != nil {
		return fmt.Errorf("Error describing traffic mirror filter %v: %v", filterId, err)
	}

	if len(out.TrafficMirrorFilters) == 0 {
		return fmt.Errorf("Error finding traffic mirror filter %v", filterId)
	}

	d.SetId(*out.TrafficMirrorFilters[0].TrafficMirrorFilterId)
	d.Set("description", out.TrafficMirrorFilters[0].Description)

	if len(out.TrafficMirrorFilters[0].NetworkServices) > 0 {
		d.Set("network_services",
			schema.NewSet(schema.HashString,
				flattenStringList(out.TrafficMirrorFilters[0].NetworkServices)))
	}

	return nil
}

func resourceAwsEc2TrafficMirrorFilterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	filterId := d.Id()

	input := &ec2.DeleteTrafficMirrorFilterInput{
		TrafficMirrorFilterId: &filterId,
	}

	_, err := conn.DeleteTrafficMirrorFilter(input)
	if err != nil {
		return fmt.Errorf("Error deleting traffic mirror filter %v: %v", filterId, err)
	}

	d.SetId("")
	d.Set("description", "")

	return nil
}
