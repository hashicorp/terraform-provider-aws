package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsTrafficMirrorFilter() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsTrafficMirrorinFilterCreate,
		Read:   resourceAwsTrafficMirrorFilterRead,
		Delete: resourceAwsTrafficMirrorFilterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsTrafficMirrorinFilterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.CreateTrafficMirrorFilterInput{
		Description: aws.String("Traffic Mirror filter"),
	}

	if description, ok := d.GetOk("description"); ok {
		input.Description = aws.String(description.(string))
	}

	_, err := conn.CreateTrafficMirrorFilter(input)
	if err != nil {
		return fmt.Errorf("Error while creating traffic filter %s", err)
	}

	return resourceAwsTrafficMirrorFilterRead(d, meta)
}

func resourceAwsTrafficMirrorFilterRead(d *schema.ResourceData, meta interface{}) error {
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

	return nil
}

func resourceAwsTrafficMirrorFilterDelete(d *schema.ResourceData, meta interface{}) error {
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
