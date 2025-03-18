// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package storagegateway

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/storagegateway"
	awstypes "github.com/aws/aws-sdk-go-v2/service/storagegateway/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_storagegateway_stored_iscsi_volume", name="Stored iSCSI Volume")
// @Tags(identifierAttribute="arn")
func resourceStorediSCSIVolume() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStorediSCSIVolumeCreate,
		ReadWithoutTimeout:   resourceStorediSCSIVolumeRead,
		UpdateWithoutTimeout: resourceStorediSCSIVolumeUpdate,
		DeleteWithoutTimeout: resourceStorediSCSIVolumeDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"chap_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"disk_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"gateway_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"kms_encrypted": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			names.AttrKMSKey: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
				RequiredWith: []string{"kms_encrypted"},
			},
			"lun_number": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			// Poor API naming: this accepts the IP address of the network interface.
			names.AttrNetworkInterfaceID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"network_interface_port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"preserve_existing_data": {
				Type:     schema.TypeBool,
				Required: true,
				ForceNew: true,
			},
			names.AttrSnapshotID: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrTargetARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"target_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"volume_attachment_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"volume_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"volume_size_in_bytes": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"volume_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrVolumeType: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceStorediSCSIVolumeCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayClient(ctx)

	input := &storagegateway.CreateStorediSCSIVolumeInput{
		DiskId:               aws.String(d.Get("disk_id").(string)),
		GatewayARN:           aws.String(d.Get("gateway_arn").(string)),
		NetworkInterfaceId:   aws.String(d.Get(names.AttrNetworkInterfaceID).(string)),
		PreserveExistingData: d.Get("preserve_existing_data").(bool),
		Tags:                 getTagsIn(ctx),
		TargetName:           aws.String(d.Get("target_name").(string)),
	}

	if v, ok := d.GetOk("kms_encrypted"); ok {
		input.KMSEncrypted = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk(names.AttrKMSKey); ok {
		input.KMSKey = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrSnapshotID); ok {
		input.SnapshotId = aws.String(v.(string))
	}

	output, err := conn.CreateStorediSCSIVolume(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Storage Gateway Stored iSCSI Volume: %s", err)
	}

	d.SetId(aws.ToString(output.VolumeARN))

	if _, err := waitStorediSCSIVolumeAvailable(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Storage Gateway Stored iSCSI Volume (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceStorediSCSIVolumeRead(ctx, d, meta)...)
}

func resourceStorediSCSIVolumeRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayClient(ctx)

	volume, err := findStorediSCSIVolumeByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Storage Gateway Stored iSCSI Volume (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Storage Gateway Stored iSCSI Volume (%s): %s", d.Id(), err)
	}

	arn := aws.ToString(volume.VolumeARN)
	d.Set(names.AttrARN, arn)
	d.Set("disk_id", volume.VolumeDiskId)
	d.Set("kms_encrypted", volume.KMSKey != nil)
	d.Set(names.AttrKMSKey, volume.KMSKey)
	d.Set("preserve_existing_data", volume.PreservedExistingData)
	d.Set(names.AttrSnapshotID, volume.SourceSnapshotId)
	d.Set("volume_attachment_status", volume.VolumeAttachmentStatus)
	d.Set("volume_id", volume.VolumeId)
	d.Set("volume_size_in_bytes", volume.VolumeSizeInBytes)
	d.Set("volume_status", volume.VolumeStatus)
	d.Set(names.AttrVolumeType, volume.VolumeType)

	if attr := volume.VolumeiSCSIAttributes; attr != nil {
		d.Set("chap_enabled", attr.ChapEnabled)
		d.Set("lun_number", attr.LunNumber)
		d.Set(names.AttrNetworkInterfaceID, attr.NetworkInterfaceId)
		d.Set("network_interface_port", attr.NetworkInterfacePort)

		targetARN := aws.ToString(attr.TargetARN)
		d.Set(names.AttrTargetARN, targetARN)

		gatewayARN, targetName, err := parseVolumeGatewayARNAndTargetNameFromARN(targetARN)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
		d.Set("gateway_arn", gatewayARN)
		d.Set("target_name", targetName)
	}

	return diags
}

func resourceStorediSCSIVolumeUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceStorediSCSIVolumeRead(ctx, d, meta)...)
}

func resourceStorediSCSIVolumeDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayClient(ctx)

	log.Printf("[DEBUG] Deleting Storage Gateway Stored iSCSI Volume: %s", d.Id())
	const (
		timeout = 2 * time.Minute
	)
	_, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.InvalidGatewayRequestException](ctx, timeout, func() (any, error) {
		return conn.DeleteVolume(ctx, &storagegateway.DeleteVolumeInput{
			VolumeARN: aws.String(d.Id()),
		})
	}, "The specified gateway is not connected")

	if isVolumeNotFoundErr(err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Storage Gateway Stored iSCSI Volume (%s): %s", d.Id(), err)
	}

	return diags
}

func findStorediSCSIVolumeByARN(ctx context.Context, conn *storagegateway.Client, arn string) (*awstypes.StorediSCSIVolume, error) {
	input := &storagegateway.DescribeStorediSCSIVolumesInput{
		VolumeARNs: []string{arn},
	}
	output, err := findStorediSCSIVolume(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.VolumeARN) != arn {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findStorediSCSIVolume(ctx context.Context, conn *storagegateway.Client, input *storagegateway.DescribeStorediSCSIVolumesInput) (*awstypes.StorediSCSIVolume, error) {
	output, err := findStorediSCSIVolumes(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findStorediSCSIVolumes(ctx context.Context, conn *storagegateway.Client, input *storagegateway.DescribeStorediSCSIVolumesInput) ([]awstypes.StorediSCSIVolume, error) {
	output, err := conn.DescribeStorediSCSIVolumes(ctx, input)

	if isVolumeNotFoundErr(err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.StorediSCSIVolumes, nil
}

func statusStorediSCSIVolume(ctx context.Context, conn *storagegateway.Client, volumeARN string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findStorediSCSIVolumeByARN(ctx, conn, volumeARN)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.VolumeStatus), nil
	}
}

func waitStorediSCSIVolumeAvailable(ctx context.Context, conn *storagegateway.Client, volumeARN string) (*awstypes.StorediSCSIVolume, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: []string{"BOOTSTRAPPING", "CREATING", "RESTORING"},
		Target:  []string{"AVAILABLE"},
		Refresh: statusStorediSCSIVolume(ctx, conn, volumeARN),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.StorediSCSIVolume); ok {
		return output, err
	}

	return nil, err
}
