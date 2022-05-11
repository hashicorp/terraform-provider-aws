package ec2

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
				d.Set("device_name", deviceName)
				d.Set("volume_id", volumeID)
				d.Set("instance_id", instanceID)
				d.SetId(volumeAttachmentID(deviceName, volumeID, instanceID))
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
	name := d.Get("device_name").(string)
	iID := d.Get("instance_id").(string)
	vID := d.Get("volume_id").(string)

	// Find out if the volume is already attached to the instance, in which case
	// we have nothing to do
	request := &ec2.DescribeVolumesInput{
		VolumeIds: []*string{aws.String(vID)},
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("attachment.instance-id"),
				Values: []*string{aws.String(iID)},
			},
			{
				Name:   aws.String("attachment.device"),
				Values: []*string{aws.String(name)},
			},
		},
	}

	vols, err := conn.DescribeVolumes(request)
	if (err != nil) || (len(vols.Volumes) == 0) {
		// This handles the situation where the instance is created by
		// a spot request and whilst the request has been fulfilled the
		// instance is not running yet
		if _, err := WaitInstanceReady(conn, iID, InstanceReadyTimeout); err != nil {
			return fmt.Errorf("waiting for EC2 Instance (%s) to be ready: %w", iID, err)
		}

		// not attached
		opts := &ec2.AttachVolumeInput{
			Device:     aws.String(name),
			InstanceId: aws.String(iID),
			VolumeId:   aws.String(vID),
		}

		log.Printf("[DEBUG] Attaching Volume (%s) to Instance (%s)", vID, iID)
		_, err := conn.AttachVolume(opts)

		if err != nil {
			return fmt.Errorf("attaching EBS Volume (%s) to EC2 Instance (%s): %w", vID, iID, err)
		}
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{ec2.VolumeAttachmentStateAttaching},
		Target:     []string{ec2.VolumeAttachmentStateAttached},
		Refresh:    volumeAttachmentStateRefreshFunc(conn, name, vID, iID),
		Timeout:    5 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf(
			"Error waiting for Volume (%s) to attach to Instance: %s, error: %s",
			vID, iID, err)
	}

	d.SetId(volumeAttachmentID(name, vID, iID))
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

func volumeAttachmentStateRefreshFunc(conn *ec2.EC2, name, volumeID, instanceID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		request := &ec2.DescribeVolumesInput{
			VolumeIds: []*string{aws.String(volumeID)},
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("attachment.device"),
					Values: []*string{aws.String(name)},
				},
				{
					Name:   aws.String("attachment.instance-id"),
					Values: []*string{aws.String(instanceID)},
				},
			},
		}

		resp, err := conn.DescribeVolumes(request)
		if err != nil {
			return nil, "failed", err
		}

		if len(resp.Volumes) > 0 {
			v := resp.Volumes[0]
			for _, a := range v.Attachments {
				if aws.StringValue(a.InstanceId) == instanceID {
					return a, aws.StringValue(a.State), nil
				}
			}
		}
		// assume detached if volume count is 0
		return 42, ec2.VolumeAttachmentStateDetached, nil
	}
}

func volumeAttachmentID(name, volumeID, instanceID string) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%s-", name))
	buf.WriteString(fmt.Sprintf("%s-", instanceID))
	buf.WriteString(fmt.Sprintf("%s-", volumeID))

	return fmt.Sprintf("vai-%d", create.StringHashcode(buf.String()))
}
