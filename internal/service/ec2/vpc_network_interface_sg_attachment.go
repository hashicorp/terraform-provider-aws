package ec2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
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
			"network_interface_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"security_group_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceNetworkInterfaceSGAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	networkInterfaceID := d.Get("network_interface_id").(string)
	sgID := d.Get("security_group_id").(string)
	mutexKey := "network_interface_sg_attachment_" + networkInterfaceID
	conns.GlobalMutexKV.Lock(mutexKey)
	defer conns.GlobalMutexKV.Unlock(mutexKey)

	eni, err := FindNetworkInterfaceByID(conn, networkInterfaceID)

	if err != nil {
		return fmt.Errorf("error reading EC2 Network Interface (%s): %w", networkInterfaceID, err)
	}

	groupIDs := []string{sgID}

	for _, group := range eni.Groups {
		if group == nil {
			continue
		}

		groupID := aws.StringValue(group.GroupId)

		if groupID == sgID {
			return fmt.Errorf("EC2 Security Group (%s) already attached to EC2 Network Interface (%s)", sgID, networkInterfaceID)
		}

		groupIDs = append(groupIDs, groupID)
	}

	input := &ec2.ModifyNetworkInterfaceAttributeInput{
		NetworkInterfaceId: aws.String(networkInterfaceID),
		Groups:             aws.StringSlice(groupIDs),
	}

	log.Printf("[INFO] Modifying EC2 Network Interface: %s", input)
	_, err = conn.ModifyNetworkInterfaceAttribute(input)

	if err != nil {
		return fmt.Errorf("error modifying EC2 Network Interface (%s): %w", networkInterfaceID, err)
	}

	d.SetId(fmt.Sprintf("%s_%s", sgID, networkInterfaceID))

	return resourceNetworkInterfaceSGAttachmentRead(d, meta)
}

func resourceNetworkInterfaceSGAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	networkInterfaceID := d.Get("network_interface_id").(string)
	sgID := d.Get("security_group_id").(string)
	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(propagationTimeout, func() (interface{}, error) {
		return FindNetworkInterfaceSecurityGroup(conn, networkInterfaceID, sgID)
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Network Interface (%s) Security Group (%s) Attachment not found, removing from state", networkInterfaceID, sgID)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Network Interface (%s) Security Group (%s) Attachment: %w", networkInterfaceID, sgID, err)
	}

	groupIdentifier := outputRaw.(*ec2.GroupIdentifier)

	d.Set("network_interface_id", networkInterfaceID)
	d.Set("security_group_id", groupIdentifier.GroupId)

	return nil
}

func resourceNetworkInterfaceSGAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	networkInterfaceID := d.Get("network_interface_id").(string)
	sgID := d.Get("security_group_id").(string)
	mutexKey := "network_interface_sg_attachment_" + networkInterfaceID
	conns.GlobalMutexKV.Lock(mutexKey)
	defer conns.GlobalMutexKV.Unlock(mutexKey)

	eni, err := FindNetworkInterfaceByID(conn, networkInterfaceID)

	if tfresource.NotFound(err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Network Interface (%s): %w", networkInterfaceID, err)
	}

	groupIDs := []string{}

	for _, group := range eni.Groups {
		if group == nil {
			continue
		}

		groupID := aws.StringValue(group.GroupId)

		if groupID == sgID {
			continue
		}

		groupIDs = append(groupIDs, groupID)
	}

	input := &ec2.ModifyNetworkInterfaceAttributeInput{
		NetworkInterfaceId: aws.String(networkInterfaceID),
		Groups:             aws.StringSlice(groupIDs),
	}

	log.Printf("[INFO] Modifying EC2 Network Interface: %s", input)
	_, err = conn.ModifyNetworkInterfaceAttribute(input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidNetworkInterfaceIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error modifying EC2 Network Interface (%s): %w", networkInterfaceID, err)
	}

	return nil
}
