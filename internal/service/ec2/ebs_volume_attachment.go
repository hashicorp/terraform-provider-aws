package ec2

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceVolumeAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceVolumeAttachmentCreate,
		Read:   resourceVolumeAttachmentRead,
		Update: schema.Noop,
		Delete: resourceVolumeAttachmentDelete,

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), ":")

				if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected DEVICE_NAME:VOLUME_ID:INSTANCE_ID", d.Id())
				}

				deviceName := idParts[0]
				volumeID := idParts[1]
				instanceID := idParts[2]
				d.SetId(volumeAttachmentID(deviceName, volumeID, instanceID))
				d.Set("device_name", deviceName)
				d.Set("instance_id", instanceID)
				d.Set("volume_id", volumeID)

				return []*schema.ResourceData{d}, nil
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"device_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"force_detach": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"skip_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"stop_instance_before_detaching": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"volume_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceVolumeAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	deviceName := d.Get("device_name").(string)
	instanceID := d.Get("instance_id").(string)
	volumeID := d.Get("volume_id").(string)

	_, err := FindEBSVolumeAttachment(conn, volumeID, instanceID, deviceName)

	if tfresource.NotFound(err) {
		// This handles the situation where the instance is created by
		// a spot request and whilst the request has been fulfilled the
		// instance is not running yet.
		if _, err := WaitInstanceReady(conn, instanceID, InstanceReadyTimeout); err != nil {
			return fmt.Errorf("waiting for EC2 Instance (%s) to be ready: %w", instanceID, err)
		}

		input := &ec2.AttachVolumeInput{
			Device:     aws.String(deviceName),
			InstanceId: aws.String(instanceID),
			VolumeId:   aws.String(volumeID),
		}

		log.Printf("[DEBUG] Create EBS Volume Attachment: %s", input)
		_, err := conn.AttachVolume(input)

		if err != nil {
			return fmt.Errorf("attaching EBS Volume (%s) to EC2 Instance (%s): %w", volumeID, instanceID, err)
		}
	} else if err != nil {
		return fmt.Errorf("reading EBS Volume (%s) Attachment (%s): %w", volumeID, instanceID, err)
	}

	if _, err := WaitVolumeAttachmentCreated(conn, volumeID, instanceID, deviceName, d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("waiting for EBS Volume (%s) Attachment (%s) create: %w", volumeID, instanceID, err)
	}

	d.SetId(volumeAttachmentID(deviceName, volumeID, instanceID))

	return resourceVolumeAttachmentRead(d, meta)
}

func resourceVolumeAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	deviceName := d.Get("device_name").(string)
	instanceID := d.Get("instance_id").(string)
	volumeID := d.Get("volume_id").(string)

	_, err := FindEBSVolumeAttachment(conn, volumeID, instanceID, deviceName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EBS Volume Attachment %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading EBS Volume (%s) Attachment (%s): %w", volumeID, instanceID, err)
	}

	return nil
}

func resourceVolumeAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if _, ok := d.GetOk("skip_destroy"); ok {
		return nil
	}

	deviceName := d.Get("device_name").(string)
	instanceID := d.Get("instance_id").(string)
	volumeID := d.Get("volume_id").(string)

	if _, ok := d.GetOk("stop_instance_before_detaching"); ok {
		if err := StopInstance(conn, instanceID, InstanceStopTimeout); err != nil {
			return err
		}
	}

	input := &ec2.DetachVolumeInput{
		Device:     aws.String(deviceName),
		Force:      aws.Bool(d.Get("force_detach").(bool)),
		InstanceId: aws.String(instanceID),
		VolumeId:   aws.String(volumeID),
	}

	log.Printf("[DEBUG] Deleting EBS Volume Attachment: %s", d.Id())
	_, err := conn.DetachVolume(input)

	if err != nil {
		return fmt.Errorf("deleting EBS Volume (%s) Attachment (%s): %w", volumeID, instanceID, err)
	}

	if _, err := WaitVolumeAttachmentDeleted(conn, volumeID, instanceID, deviceName, d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("waiting for EBS Volume (%s) Attachment (%s) delete: %w", volumeID, instanceID, err)
	}

	return nil
}

func volumeAttachmentID(name, volumeID, instanceID string) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%s-", name))
	buf.WriteString(fmt.Sprintf("%s-", instanceID))
	buf.WriteString(fmt.Sprintf("%s-", volumeID))

	return fmt.Sprintf("vai-%d", create.StringHashcode(buf.String()))
}
