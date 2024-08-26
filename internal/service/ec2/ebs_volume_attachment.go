// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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

func resourceVolumeAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)
	deviceName := d.Get(names.AttrDeviceName).(string)
	instanceID := d.Get(names.AttrInstanceID).(string)
	volumeID := d.Get("volume_id").(string)

	_, err := findVolumeAttachment(ctx, conn, volumeID, instanceID, deviceName)

	if tfresource.NotFound(err) {
		// This handles the situation where the instance is created by
		// a spot request and whilst the request has been fulfilled the
		// instance is not running yet.
		if _, err := waitVolumeAttachmentInstanceReady(ctx, conn, instanceID, instanceReadyTimeout); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EC2 Instance (%s) to be ready: %s", instanceID, err)
		}

		input := &ec2.AttachVolumeInput{
			Device:     aws.String(deviceName),
			InstanceId: aws.String(instanceID),
			VolumeId:   aws.String(volumeID),
		}

		_, err := conn.AttachVolume(ctx, input)

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

func resourceVolumeAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)
	deviceName := d.Get(names.AttrDeviceName).(string)
	instanceID := d.Get(names.AttrInstanceID).(string)
	volumeID := d.Get("volume_id").(string)

	_, err := findVolumeAttachment(ctx, conn, volumeID, instanceID, deviceName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EBS Volume Attachment %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EBS Volume (%s) Attachment (%s): %s", volumeID, instanceID, err)
	}

	return diags
}

func resourceVolumeAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	input := &ec2.DetachVolumeInput{
		Device:     aws.String(deviceName),
		Force:      aws.Bool(d.Get("force_detach").(bool)),
		InstanceId: aws.String(instanceID),
		VolumeId:   aws.String(volumeID),
	}

	log.Printf("[DEBUG] Deleting EBS Volume Attachment: %s", d.Id())
	_, err := conn.DetachVolume(ctx, input)

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
	buf.WriteString(fmt.Sprintf("%s-", name))
	buf.WriteString(fmt.Sprintf("%s-", instanceID))
	buf.WriteString(fmt.Sprintf("%s-", volumeID))

	return fmt.Sprintf("vai-%d", create.StringHashcode(buf.String()))
}

func findVolumeAttachment(ctx context.Context, conn *ec2.Client, volumeID, instanceID, deviceName string) (*awstypes.VolumeAttachment, error) {
	input := &ec2.DescribeVolumesInput{
		Filters: newAttributeFilterList(map[string]string{
			"attachment.device":      deviceName,
			"attachment.instance-id": instanceID,
		}),
		VolumeIds: []string{volumeID},
	}

	output, err := findEBSVolume(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.VolumeStateAvailable || state == awstypes.VolumeStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.VolumeId) != volumeID {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	for _, v := range output.Attachments {
		if v.State == awstypes.VolumeAttachmentStateDetached {
			continue
		}

		if aws.ToString(v.Device) == deviceName && aws.ToString(v.InstanceId) == instanceID {
			return &v, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

func stopVolumeAttachmentInstance(ctx context.Context, conn *ec2.Client, id string, force bool, timeout time.Duration) error {
	tflog.Info(ctx, "Stopping EC2 Instance", map[string]any{
		"ec2_instance_id": id,
		"force":           force,
	})
	_, err := conn.StopInstances(ctx, &ec2.StopInstancesInput{
		Force:       aws.Bool(force),
		InstanceIds: []string{id},
	})

	if err != nil {
		return fmt.Errorf("stopping EC2 Instance (%s): %w", id, err)
	}

	if _, err := waitVolumeAttachmentInstanceStopped(ctx, conn, id, timeout); err != nil {
		return fmt.Errorf("waiting for EC2 Instance (%s) stop: %w", id, err)
	}

	return nil
}

func waitVolumeAttachmentInstanceStopped(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.Instance, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.InstanceStateNamePending,
			awstypes.InstanceStateNameRunning,
			awstypes.InstanceStateNameShuttingDown,
			awstypes.InstanceStateNameStopping,
		),
		Target:     enum.Slice(awstypes.InstanceStateNameStopped),
		Refresh:    statusVolumeAttachmentInstanceState(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Instance); ok {
		if stateReason := output.StateReason; stateReason != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(stateReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitVolumeAttachmentInstanceReady(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.Instance, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.InstanceStateNamePending, awstypes.InstanceStateNameStopping),
		Target:     enum.Slice(awstypes.InstanceStateNameRunning, awstypes.InstanceStateNameStopped),
		Refresh:    statusVolumeAttachmentInstanceState(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Instance); ok {
		if stateReason := output.StateReason; stateReason != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(stateReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitVolumeAttachmentDeleted(ctx context.Context, conn *ec2.Client, volumeID, instanceID, deviceName string, timeout time.Duration) (*awstypes.VolumeAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.VolumeAttachmentStateDetaching),
		Target:     []string{},
		Refresh:    statusVolumeAttachment(ctx, conn, volumeID, instanceID, deviceName),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VolumeAttachment); ok {
		return output, err
	}

	return nil, err
}

func statusVolumeAttachmentInstanceState(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		// Don't call FindInstanceByID as it maps useful status codes to NotFoundError.
		output, err := findInstance(ctx, conn, &ec2.DescribeInstancesInput{
			InstanceIds: []string{id},
		})

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State.Name), nil
	}
}
