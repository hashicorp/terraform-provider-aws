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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/datafy"
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

		CustomizeDiff: customdiff.Sequence(
			func(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
				vId, _ := diff.GetChange("volume_id")

				// once the volume is managed, datafy has control on the volume, and it can't be updated via terraform.
				if changes := diff.GetChangedKeysPrefix(""); len(changes) > 0 {
					dc := meta.(*conns.AWSClient).DatafyClient(ctx)
					if datafyVolume, datafyErr := dc.GetVolume(vId.(string)); datafyErr == nil {
						if datafyVolume.IsManaged {
							return fmt.Errorf("can't modify EBS Volume Attachment (%s) of a datafid EBS Volume (%s). Changed keys: (%s)", diff.Id(), vId.(string), strings.Join(changes, ","))
						}
					}
				}
				return nil
			},
		),

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

		dc := meta.(*conns.AWSClient).DatafyClient(ctx)
		datafyVolume, err := dc.GetVolume(volumeID)
		// If an error occurs, and it's not a "not found" error, return the error.
		// If the volume exists and is managed by Datafy, proceed to attach it.
		// If the error is "not found", the volume hasn't been discovered yet and can't be managed.
		// When creating a volume from a Datafy snapshot (dsnap-), the volume is immediately marked as managed in the database and should be found.
		if err != nil && !datafy.NotFound(err) {
			return sdkdiag.AppendErrorf(diags, "attaching EBS Volume (%s) to EC2 Instance (%s): %s", volumeID, instanceID, err)
		} else if datafyVolume != nil && datafyVolume.IsManaged {
			err := dc.AttachVolume(instanceID, volumeID, deviceName)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "attaching datafy managed EBS Volume (%s) to EC2 Instance (%s): %s", volumeID, instanceID, err)
			}

			dvo, err := conn.DescribeVolumes(ctx, datafy.DescribeDatafiedVolumesInput(volumeID))
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "can't find datafy volumes of EBS volume (%s) Attachement (%s): %s", volumeID, instanceID, err)
			} else if len(dvo.Volumes) == 0 {
				return sdkdiag.AppendErrorf(diags, "can't find datafy volumes of EBS volume (%s) Attachement (%s)", volumeID, instanceID)
			}

			for _, volume := range dvo.Volumes {
				if _, err := waitDatafyVolumeAttachmentCreated(ctx, conn, aws.ToString(volume.VolumeId), instanceID, d.Timeout(schema.TimeoutCreate)); err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for EBS Volume (%s) Attachment (%s) create: %s", volumeID, instanceID, err)
				}
			}

			d.SetId(volumeAttachmentID(deviceName, volumeID, instanceID))
			return diags
		} else {
			input := &ec2.AttachVolumeInput{
				Device:     aws.String(deviceName),
				InstanceId: aws.String(instanceID),
				VolumeId:   aws.String(volumeID),
			}

			_, err := conn.AttachVolume(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "attaching EBS Volume (%s) to EC2 Instance (%s): %s", volumeID, instanceID, err)
			}

			if _, err := waitVolumeAttachmentCreated(ctx, conn, volumeID, instanceID, deviceName, d.Timeout(schema.TimeoutCreate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for EBS Volume (%s) Attachment (%s) create: %s", volumeID, instanceID, err)
			}

			d.SetId(volumeAttachmentID(deviceName, volumeID, instanceID))
			return append(diags, resourceVolumeAttachmentRead(ctx, d, meta)...)
		}
	} else if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EBS Volume (%s) Attachment (%s): %s", volumeID, instanceID, err)
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
		// if not found on aws, it may mean we datafied it and deleted the volume
		dc := meta.(*conns.AWSClient).DatafyClient(ctx)
		if datafyVolume, datafyErr := dc.GetVolume(volumeID); datafyErr == nil {
			// if we are managing this volume, just return the state as is
			if datafyVolume.IsManaged {
				return diags
			}

			// if the volume was replaced (new source due to undatafy), it means the new
			// volume is now the source volume, and we need to set the "new" values from aws
			if datafyVolume.ReplacedBy != "" {
				d.SetId(volumeAttachmentID(deviceName, datafyVolume.ReplacedBy, instanceID))
				d.Set("volume_id", datafyVolume.ReplacedBy)
				return append(
					sdkdiag.AppendWarningf(diags, "new EBS Volume (%s) has been created to replace the undatafied EBS Volume (%s)", datafyVolume.ReplacedBy, volumeID),
					resourceVolumeAttachmentRead(ctx, d, meta)...,
				)
			}
		} else if datafy.NotFound(datafyErr) {
			log.Printf("[WARN] EBS Volume Attachment %s not found, removing from state", d.Id())
			d.SetId("")
			return diags
		} else {
			err = datafyErr
		}
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

	dc := meta.(*conns.AWSClient).DatafyClient(ctx)
	datafyVolume, err := dc.GetVolume(volumeID)
	if err != nil && !datafy.NotFound(err) {
		return sdkdiag.AppendErrorf(diags, "deleting EBS Volume (%s) Attachment (%s): %s", volumeID, d.Id(), err)
	}

	if datafyVolume != nil && datafyVolume.IsManaged {
		dvo, err := conn.DescribeVolumes(ctx, datafy.DescribeDatafiedVolumesInput(volumeID))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "can't find datafy volumes of EBS volume (%s) Attachement (%s): %s", volumeID, d.Id(), err)
		} else if len(dvo.Volumes) == 0 {
			return sdkdiag.AppendErrorf(diags, "can't find datafy volumes of EBS volume (%s) Attachement (%s)", volumeID, d.Id())
		}

		volumesToDelete := make(map[string]string)
		if datafyVolume.HasSource {
			volumesToDelete[volumeID] = deviceName
		}
		for _, volume := range dvo.Volumes {
			if len(volume.Attachments) == 0 {
				return sdkdiag.AppendErrorf(diags, "can't find device name of datafy volume (%s) for EBS volume (%s) Attachement (%s)", aws.ToString(volume.VolumeId), volumeID, d.Id())
			}
			volumesToDelete[aws.ToString(volume.VolumeId)] = aws.ToString(volume.Attachments[0].Device)
		}

		err = dc.DetachVolume(instanceID, volumeID)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "detaching datafy EBS volume (%s) from EC2 Instance (%s): %s", volumeID, instanceID, err)
		}

		for id, dn := range volumesToDelete {
			if _, err := waitVolumeAttachmentDeleted(ctx, conn, id, instanceID, dn, d.Timeout(schema.TimeoutDelete)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for datafy EBS Volume (%s) for EBS volume (%s) Attachment (%s) delete: %s", id, volumeID, d.Id(), err)
			}
		}
		return diags
	}

	if datafyVolume != nil && datafyVolume.ReplacedBy != "" {
		d.SetId(volumeAttachmentID(deviceName, datafyVolume.ReplacedBy, instanceID))
		d.Set("volume_id", datafyVolume.ReplacedBy)
		return append(
			sdkdiag.AppendWarningf(diags, "new EBS Volume (%s) has been created to replace the undatafied EBS Volume (%s)", datafyVolume.ReplacedBy, volumeID),
			resourceVolumeAttachmentDelete(ctx, d, meta)...,
		)
	}

	if _, ok := d.GetOk("stop_instance_before_detaching"); ok {
		if err := stopVolumeAttachmentInstance(ctx, conn, instanceID, false, instanceStopTimeout); err != nil {
			return sdkdiag.AppendErrorf(diags, "deleting EBS Volume (%s) Attachment (%s): %s", volumeID, instanceID, err)
		}
	}

	// fallback if not managed / not discovered
	input := &ec2.DetachVolumeInput{
		Device:     aws.String(deviceName),
		Force:      aws.Bool(d.Get("force_detach").(bool)),
		InstanceId: aws.String(instanceID),
		VolumeId:   aws.String(volumeID),
	}

	log.Printf("[DEBUG] Deleting EBS Volume Attachment: %s", d.Id())
	_, err = conn.DetachVolume(ctx, input)

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

func findDatafyVolumeAttachment(ctx context.Context, conn *ec2.Client, volumeID, instanceID string) (*awstypes.VolumeAttachment, error) {
	input := &ec2.DescribeVolumesInput{
		Filters: newAttributeFilterList(map[string]string{
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

		if aws.ToString(v.InstanceId) == instanceID {
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
