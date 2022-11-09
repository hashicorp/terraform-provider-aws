package ec2

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceNetworkInterfaceAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetworkInterfaceAttachmentCreate,
		Read:   resourceNetworkInterfaceAttachmentRead,
		Delete: resourceNetworkInterfaceAttachmentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"attachment_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"device_index": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"network_interface_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceNetworkInterfaceAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	attachmentID, err := attachNetworkInterface(
		conn,
		d.Get("network_interface_id").(string),
		d.Get("instance_id").(string),
		d.Get("device_index").(int),
		networkInterfaceAttachedTimeout,
	)

	if attachmentID != "" {
		d.SetId(attachmentID)
	}

	if err != nil {
		return err
	}

	return resourceNetworkInterfaceAttachmentRead(d, meta)
}

func resourceNetworkInterfaceAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	network_interface, err := FindNetworkInterfaceByAttachmentID(context.TODO(), conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Network Interface Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Network Interface Attachment (%s): %w", d.Id(), err)
	}

	d.Set("network_interface_id", network_interface.NetworkInterfaceId)
	d.Set("attachment_id", network_interface.Attachment.AttachmentId)
	d.Set("device_index", network_interface.Attachment.DeviceIndex)
	d.Set("instance_id", network_interface.Attachment.InstanceId)
	d.Set("status", network_interface.Attachment.Status)

	return nil
}

func resourceNetworkInterfaceAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	return DetachNetworkInterface(conn, d.Get("network_interface_id").(string), d.Id(), NetworkInterfaceDetachedTimeout)
}
