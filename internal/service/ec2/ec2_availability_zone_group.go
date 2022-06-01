package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceAvailabilityZoneGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAvailabilityZoneGroupCreate,
		Read:   resourceAvailabilityZoneGroupRead,
		Update: resourceAvailabilityZoneGroupUpdate,
		Delete: schema.Noop,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"opt_in_status": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					ec2.AvailabilityZoneOptInStatusOptedIn,
					ec2.AvailabilityZoneOptInStatusNotOptedIn,
				}, false),
			},
		},
	}
}

func resourceAvailabilityZoneGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	groupName := d.Get("group_name").(string)
	availabilityZone, err := FindAvailabilityZoneGroupByName(conn, groupName)

	if err != nil {
		return fmt.Errorf("reading EC2 Availability Zone Group (%s): %w", groupName, err)
	}

	if v := d.Get("opt_in_status").(string); v != aws.StringValue(availabilityZone.OptInStatus) {
		if err := modifyAvailabilityZoneOptInStatus(conn, groupName, v); err != nil {
			return err
		}
	}

	d.SetId(groupName)

	return resourceAvailabilityZoneGroupRead(d, meta)
}

func resourceAvailabilityZoneGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	availabilityZone, err := FindAvailabilityZoneGroupByName(conn, d.Id())

	if err != nil {
		return fmt.Errorf("reading EC2 Availability Zone Group (%s): %w", d.Id(), err)
	}

	if aws.StringValue(availabilityZone.OptInStatus) == ec2.AvailabilityZoneOptInStatusOptInNotRequired {
		return fmt.Errorf("unnecessary handling of EC2 Availability Zone Group (%s), status: %s", d.Id(), ec2.AvailabilityZoneOptInStatusOptInNotRequired)
	}

	d.Set("group_name", availabilityZone.GroupName)
	d.Set("opt_in_status", availabilityZone.OptInStatus)

	return nil
}

func resourceAvailabilityZoneGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if err := modifyAvailabilityZoneOptInStatus(conn, d.Id(), d.Get("opt_in_status").(string)); err != nil {
		return err
	}

	return resourceAvailabilityZoneGroupRead(d, meta)
}

func modifyAvailabilityZoneOptInStatus(conn *ec2.EC2, groupName, optInStatus string) error {
	input := &ec2.ModifyAvailabilityZoneGroupInput{
		GroupName:   aws.String(groupName),
		OptInStatus: aws.String(optInStatus),
	}

	if _, err := conn.ModifyAvailabilityZoneGroup(input); err != nil {
		return fmt.Errorf("modifying EC2 Availability Zone Group (%s): %w", groupName, err)
	}

	waiter := WaitAvailabilityZoneGroupOptedIn
	if optInStatus == ec2.AvailabilityZoneOptInStatusNotOptedIn {
		waiter = WaitAvailabilityZoneGroupNotOptedIn
	}

	if _, err := waiter(conn, groupName); err != nil {
		return fmt.Errorf("waiting for EC2 Availability Zone Group (%s) opt-in status update: %w", groupName, err)
	}

	return nil
}
