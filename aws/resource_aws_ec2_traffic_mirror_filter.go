package aws

import (
	"fmt"
	"log"

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

	d.SetId(*out.TrafficMirrorFilter.TrafficMirrorFilterId)

	return resourceAwsEc2TrafficMirrorFilterUpdate(d, meta)
}

func resourceAwsEc2TrafficMirrorFilterUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	if d.HasChange("network_services") {
		input := &ec2.ModifyTrafficMirrorFilterNetworkServicesInput{
			TrafficMirrorFilterId: aws.String(d.Id()),
		}

		o, n := d.GetChange("network_services")
		newServices := n.(*schema.Set).Difference(o.(*schema.Set)).List()
		if len(newServices) > 0 {
			input.AddNetworkServices = expandStringList(newServices)
		}

		removeServices := o.(*schema.Set).Difference(n.(*schema.Set)).List()
		if len(removeServices) > 0 {
			input.RemoveNetworkServices = expandStringList(removeServices)
		}

		_, err := conn.ModifyTrafficMirrorFilterNetworkServices(input)
		if err != nil {
			return fmt.Errorf("Error modifying network services for traffic mirror filter %v", d.Id())
		}
	}

	return resourceAwsEc2TrafficMirrorFilterRead(d, meta)
}

func resourceAwsEc2TrafficMirrorFilterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.DescribeTrafficMirrorFiltersInput{
		TrafficMirrorFilterIds: aws.StringSlice([]string{d.Id()}),
	}

	out, err := conn.DescribeTrafficMirrorFilters(input)
	if err != nil {
		return fmt.Errorf("Error describing traffic mirror filter %v: %v", d.Id(), err)
	}

	if len(out.TrafficMirrorFilters) == 0 {
		log.Printf("[WARN] EC2 Traffic Mirror Filter (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.SetId(*out.TrafficMirrorFilters[0].TrafficMirrorFilterId)
	d.Set("description", out.TrafficMirrorFilters[0].Description)

	if len(out.TrafficMirrorFilters[0].NetworkServices) > 0 {
		if err := d.Set("network_services", aws.StringValueSlice(out.TrafficMirrorFilters[0].NetworkServices)); err != nil {
			return fmt.Errorf("error setting network_services for filter %v: %s", d.Id(), err)
		}
	}

	return nil
}

func resourceAwsEc2TrafficMirrorFilterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.DeleteTrafficMirrorFilterInput{
		TrafficMirrorFilterId: aws.String(d.Id()),
	}

	_, err := conn.DeleteTrafficMirrorFilter(input)
	if err != nil {
		return fmt.Errorf("Error deleting traffic mirror filter %v: %v", d.Id(), err)
	}

	return nil
}
