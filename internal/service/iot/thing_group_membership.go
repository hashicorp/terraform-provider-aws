package iot

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceThingGroupMembership() *schema.Resource {
	return &schema.Resource{
		Create: resourceThingGroupMembershipCreate,
		Read:   resourceThingGroupMembershipRead,
		Delete: resourceThingGroupMembershipDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"override_dynamic_group": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"thing_group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"thing_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceThingGroupMembershipCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn

	thingGroupName := d.Get("thing_group_name").(string)
	thingName := d.Get("thing_name").(string)
	input := &iot.AddThingToThingGroupInput{
		ThingGroupName: aws.String(thingGroupName),
		ThingName:      aws.String(thingName),
	}

	if v, ok := d.GetOk("override_dynamic_group"); ok {
		input.OverrideDynamicGroups = aws.Bool(v.(bool))
	}

	log.Printf("[DEBUG] Creating IoT Thing Group Membership: %s", input)
	_, err := conn.AddThingToThingGroup(input)

	if err != nil {
		return fmt.Errorf("error adding IoT Thing (%s) to IoT Thing Group (%s): %w", thingName, thingGroupName, err)
	}

	d.SetId(ThingGroupMembershipCreateResourceID(thingGroupName, thingName))

	return resourceThingGroupMembershipRead(d, meta)
}

func resourceThingGroupMembershipRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn

	thingGroupName, thingName, err := ThingGroupMembershipParseResourceID(d.Id())

	if err != nil {
		return err
	}

	err = FindThingGroupMembership(conn, thingGroupName, thingName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IoT Thing Group Membership (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading IoT Thing Group Membership (%s): %w", d.Id(), err)
	}

	d.Set("thing_group_name", thingGroupName)
	d.Set("thing_name", thingName)

	return nil
}

func resourceThingGroupMembershipDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn

	thingGroupName, thingName, err := ThingGroupMembershipParseResourceID(d.Id())

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Deleting IoT Thing Group Membership: %s", d.Id())
	_, err = conn.RemoveThingFromThingGroup(&iot.RemoveThingFromThingGroupInput{
		ThingGroupName: aws.String(thingGroupName),
		ThingName:      aws.String(thingName),
	})

	if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error removing IoT Thing (%s) from IoT Thing Group (%s): %w", thingName, thingGroupName, err)
	}

	return nil
}

const thingGroupMembershipResourceIDSeparator = "/"

func ThingGroupMembershipCreateResourceID(thingGroupName, thingName string) string {
	parts := []string{thingGroupName, thingName}
	id := strings.Join(parts, thingGroupMembershipResourceIDSeparator)

	return id
}

func ThingGroupMembershipParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, thingGroupMembershipResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected thing-group-name%[2]sthing-name", id, thingGroupMembershipResourceIDSeparator)
}
