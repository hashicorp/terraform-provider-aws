// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_volume_attachment", name="EBS Volume Attachment")
func resourceVolumeAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVolumeAttachmentCreate,
		ReadWithoutTimeout:   resourceVolumeAttachmentRead,
		UpdateWithoutTimeout: schema.NoopContext,
		DeleteWithoutTimeout: resourceVolumeAttachmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), ":")

				if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected DEVICE_NAME:VOLUME_ID:INSTANCE_ID", d.Id())
				}

				deviceName := idParts[0]
				volumeID := idParts[1]
				instanceID := idParts[2]
				d.SetId(volumeAttachmentID(deviceName, volumeID, instanceID))
				d.Set(names.AttrDeviceName, deviceName)
				d.Set(names.AttrInstanceID, instanceID)
				d.Set("volume_id", volumeID)

				return []*schema.ResourceData{d}, nil
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrDeviceName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"force_detach": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			names.AttrInstanceID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrSkipDestroy: {
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

func resourceVolumeAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)
	deviceName := d.Get(names.AttrDeviceName).(string)
	instanceID := d.Get(names.AttrInstanceID).(string)
	volumeID := d.Get("volume_id").(string)

	_, err := findVolumeAttachment(ctx, conn, volumeID, instanceID, deviceName)

	if retry.NotFound(err) {
		// This handles the situation where the instance is created by
		// a spot request and whilst the request has been fulfilled the
		// instance is not running yet.
		if _, err := waitVolumeAttachmentInstanceReady(ctx, conn, instanceID, instanceReadyTimeout); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EC2 Instance (%s) to be ready: %s", instanceID, err)
		}

		input := ec2.AttachVolumeInput{
			Device:     aws.String(deviceName),
			InstanceId: aws.String(instanceID),
			VolumeId:   aws.String(volumeID),
		}

		_, err := conn.AttachVolume(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "attaching EBS Volume (%s) to EC2 Instance (%s): %s", volumeID, instanceID, err)
		}
	} else if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EBS Volume (%s) Attachment (%s): %s", volumeID, instanceID, err)
	}

	if _, err := waitVolumeAttachmentCreated(ctx, conn, volumeID, instanceID, deviceName, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EBS Volume (%s) Attachment (%s) create: %s", volumeID, instanceID, err)
	}

	d.SetId(volumeAttachmentID(deviceName, volumeID, instanceID))

	return append(diags, resourceVolumeAttachmentRead(ctx, d, meta)...)
}

func resourceVolumeAttachmentRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)
	deviceName := d.Get(names.AttrDeviceName).(string)
	instanceID := d.Get(names.AttrInstanceID).(string)
	volumeID := d.Get("volume_id").(string)

	_, err := findVolumeAttachment(ctx, conn, volumeID, instanceID, deviceName)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] EBS Volume Attachment %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EBS Volume (%s) Attachment (%s): %s", volumeID, instanceID, err)
	}

	return diags
}

func resourceVolumeAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if _, ok := d.GetOk(names.AttrSkipDestroy); ok {
		return diags
	}

	deviceName := d.Get(names.AttrDeviceName).(string)
	instanceID := d.Get(names.AttrInstanceID).(string)
	volumeID := d.Get("volume_id").(string)

	if _, ok := d.GetOk("stop_instance_before_detaching"); ok {
		if err := stopVolumeAttachmentInstance(ctx, conn, instanceID, false, instanceStopTimeout); err != nil {
			return sdkdiag.AppendErrorf(diags, "deleting EBS Volume (%s) Attachment (%s): %s", volumeID, instanceID, err)
		}
	}

	input := ec2.DetachVolumeInput{
		Device:     aws.String(deviceName),
		Force:      aws.Bool(d.Get("force_detach").(bool)),
		InstanceId: aws.String(instanceID),
		VolumeId:   aws.String(volumeID),
	}

	log.Printf("[DEBUG] Deleting EBS Volume Attachment: %s", d.Id())
	_, err := conn.DetachVolume(ctx, &input)

	if tfawserr.ErrMessageContains(err, errCodeIncorrectState, "available") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EBS Volume (%s) Attachment (%s): %s", volumeID, instanceID, err)
	}

	if _, err := waitVolumeAttachmentDeleted(ctx, conn, volumeID, instanceID, deviceName, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EBS Volume (%s) Attachment (%s) delete: %s", volumeID, instanceID, err)
	}

	return diags
}

func volumeAttachmentID(name, volumeID, instanceID string) string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%s-", name)
	fmt.Fprintf(&buf, "%s-", instanceID)
	fmt.Fprintf(&buf, "%s-", volumeID)

	return fmt.Sprintf("vai-%d", create.StringHashcode(buf.String()))
}

func stopVolumeAttachmentInstance(ctx context.Context, conn *ec2.Client, id string, force bool, timeout time.Duration) error {
	tflog.Info(ctx, "Stopping EC2 Instance", map[string]any{
		"ec2_instance_id": id,
		"force":           force,
	})
	input := ec2.StopInstancesInput{
		Force:       aws.Bool(force),
		InstanceIds: []string{id},
	}
	_, err := conn.StopInstances(ctx, &input)

	if err != nil {
		return fmt.Errorf("stopping EC2 Instance (%s): %w", id, err)
	}

	if _, err := waitVolumeAttachmentInstanceStopped(ctx, conn, id, timeout); err != nil {
		return fmt.Errorf("waiting for EC2 Instance (%s) stop: %w", id, err)
	}

	return nil
}
