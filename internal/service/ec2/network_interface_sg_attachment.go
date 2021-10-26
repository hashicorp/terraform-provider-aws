package ec2

import (
	"fmt"
	"log"
	"reflect"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceNetworkInterfaceSGAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetworkInterfaceSGAttachmentCreate,
		Read:   resourceNetworkInterfaceSGAttachmentRead,
		Delete: resourceNetworkInterfaceSGAttachmentDelete,
		Schema: map[string]*schema.Schema{
			"security_group_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"network_interface_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceNetworkInterfaceSGAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	mk := "network_interface_sg_attachment_" + d.Get("network_interface_id").(string)
	conns.GlobalMutexKV.Lock(mk)
	defer conns.GlobalMutexKV.Unlock(mk)

	sgID := d.Get("security_group_id").(string)
	interfaceID := d.Get("network_interface_id").(string)

	conn := meta.(*conns.AWSClient).EC2Conn

	iface, err := FindNetworkInterfaceByID(conn, interfaceID)

	if err != nil {
		return fmt.Errorf("error reading EC2 Network Interface (%s): %w", interfaceID, err)
	}

	groupIDs := []string{sgID}

	for _, group := range iface.Groups {
		if group == nil {
			continue
		}

		if aws.StringValue(group.GroupId) == sgID {
			return fmt.Errorf("EC2 Security Group (%s) already attached to EC2 Network Interface ID: %s", sgID, interfaceID)
		}

		groupIDs = append(groupIDs, aws.StringValue(group.GroupId))
	}

	params := &ec2.ModifyNetworkInterfaceAttributeInput{
		NetworkInterfaceId: iface.NetworkInterfaceId,
		Groups:             aws.StringSlice(groupIDs),
	}

	_, err = conn.ModifyNetworkInterfaceAttribute(params)

	if err != nil {
		return fmt.Errorf("error modifying EC2 Network Interface (%s): %w", interfaceID, err)
	}

	d.SetId(fmt.Sprintf("%s_%s", sgID, interfaceID))

	return resourceNetworkInterfaceSGAttachmentRead(d, meta)
}

func resourceNetworkInterfaceSGAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	sgID := d.Get("security_group_id").(string)
	interfaceID := d.Get("network_interface_id").(string)

	log.Printf("[DEBUG] Checking association of security group %s to network interface ID %s", sgID, interfaceID)

	conn := meta.(*conns.AWSClient).EC2Conn

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(PropagationTimeout, func() (interface{}, error) {
		return FindNetworkInterfaceSecurityGroup(conn, interfaceID, sgID)
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Network Interface Security Group Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Network Interface Security Group Attachment (%s): %w", d.Id(), err)
	}

	groupIdentifier := outputRaw.(*ec2.GroupIdentifier)

	d.Set("network_interface_id", interfaceID)
	d.Set("security_group_id", groupIdentifier.GroupId)

	return nil
}

func resourceNetworkInterfaceSGAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	mk := "network_interface_sg_attachment_" + d.Get("network_interface_id").(string)
	conns.GlobalMutexKV.Lock(mk)
	defer conns.GlobalMutexKV.Unlock(mk)

	sgID := d.Get("security_group_id").(string)
	interfaceID := d.Get("network_interface_id").(string)

	log.Printf("[DEBUG] Removing security group %s from interface ID %s", sgID, interfaceID)

	conn := meta.(*conns.AWSClient).EC2Conn

	iface, err := FindNetworkInterfaceByID(conn, interfaceID)

	if tfresource.NotFound(err) {
		return nil
	}

	if err != nil {
		return err
	}

	return delSGFromENI(conn, sgID, iface)
}

func delSGFromENI(conn *ec2.EC2, sgID string, iface *ec2.NetworkInterface) error {
	old := iface.Groups
	var new []*string
	for _, v := range iface.Groups {
		if *v.GroupId == sgID {
			continue
		}
		new = append(new, v.GroupId)
	}
	if reflect.DeepEqual(old, new) {
		// The interface already didn't have the security group, nothing to do
		return nil
	}

	params := &ec2.ModifyNetworkInterfaceAttributeInput{
		NetworkInterfaceId: iface.NetworkInterfaceId,
		Groups:             new,
	}

	_, err := conn.ModifyNetworkInterfaceAttribute(params)

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidNetworkInterfaceIDNotFound) {
		return nil
	}

	return err
}
