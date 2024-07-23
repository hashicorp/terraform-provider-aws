// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package storagegateway

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_storagegateway_cached_iscsi_volume", name="Cached iSCSI Volume")
// @Tags(identifierAttribute="arn")
func resourceCachediSCSIVolume() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCachediSCSIVolumeCreate,
		ReadWithoutTimeout:   resourceCachediSCSIVolumeRead,
		UpdateWithoutTimeout: resourceCachediSCSIVolumeUpdate,
		DeleteWithoutTimeout: resourceCachediSCSIVolumeDelete,

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
			"gateway_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"lun_number": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			// Poor API naming: this accepts the IP address of the network interface
			names.AttrNetworkInterfaceID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"network_interface_port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrSnapshotID: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"source_volume_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrTargetARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"target_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"volume_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"volume_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"volume_size_in_bytes": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
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
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceCachediSCSIVolumeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayConn(ctx)

	input := &storagegateway.CreateCachediSCSIVolumeInput{
		ClientToken:        aws.String(id.UniqueId()),
		GatewayARN:         aws.String(d.Get("gateway_arn").(string)),
		NetworkInterfaceId: aws.String(d.Get(names.AttrNetworkInterfaceID).(string)),
		TargetName:         aws.String(d.Get("target_name").(string)),
		VolumeSizeInBytes:  aws.Int64(int64(d.Get("volume_size_in_bytes").(int))),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrSnapshotID); ok {
		input.SnapshotId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("source_volume_arn"); ok {
		input.SourceVolumeARN = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrKMSKey); ok {
		input.KMSKey = aws.String(v.(string))
	}

	if v, ok := d.GetOk("kms_encrypted"); ok {
		input.KMSEncrypted = aws.Bool(v.(bool))
	}

	log.Printf("[DEBUG] Creating Storage Gateway cached iSCSI volume: %s", input)
	output, err := conn.CreateCachediSCSIVolumeWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Storage Gateway cached iSCSI volume: %s", err)
	}

	d.SetId(aws.StringValue(output.VolumeARN))

	return append(diags, resourceCachediSCSIVolumeRead(ctx, d, meta)...)
}

func resourceCachediSCSIVolumeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayConn(ctx)

	input := &storagegateway.DescribeCachediSCSIVolumesInput{
		VolumeARNs: []*string{aws.String(d.Id())},
	}

	log.Printf("[DEBUG] Reading Storage Gateway cached iSCSI volume: %s", input)
	output, err := conn.DescribeCachediSCSIVolumesWithContext(ctx, input)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, storagegateway.ErrorCodeVolumeNotFound) || tfawserr.ErrMessageContains(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified volume was not found") {
			log.Printf("[WARN] Storage Gateway cached iSCSI volume %q not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading Storage Gateway cached iSCSI volume %q: %s", d.Id(), err)
	}

	if output == nil || len(output.CachediSCSIVolumes) == 0 || output.CachediSCSIVolumes[0] == nil || aws.StringValue(output.CachediSCSIVolumes[0].VolumeARN) != d.Id() {
		log.Printf("[WARN] Storage Gateway cached iSCSI volume %q not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	volume := output.CachediSCSIVolumes[0]

	arn := aws.StringValue(volume.VolumeARN)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrSnapshotID, volume.SourceSnapshotId)
	d.Set("volume_arn", arn)
	d.Set("volume_id", volume.VolumeId)
	d.Set("volume_size_in_bytes", volume.VolumeSizeInBytes)
	d.Set(names.AttrKMSKey, volume.KMSKey)
	if volume.KMSKey != nil {
		d.Set("kms_encrypted", true)
	} else {
		d.Set("kms_encrypted", false)
	}

	if volume.VolumeiSCSIAttributes != nil {
		d.Set("chap_enabled", volume.VolumeiSCSIAttributes.ChapEnabled)
		d.Set("lun_number", volume.VolumeiSCSIAttributes.LunNumber)
		d.Set(names.AttrNetworkInterfaceID, volume.VolumeiSCSIAttributes.NetworkInterfaceId)
		d.Set("network_interface_port", volume.VolumeiSCSIAttributes.NetworkInterfacePort)

		targetARN := aws.StringValue(volume.VolumeiSCSIAttributes.TargetARN)
		d.Set(names.AttrTargetARN, targetARN)

		gatewayARN, targetName, err := ParseVolumeGatewayARNAndTargetNameFromARN(targetARN)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "parsing Storage Gateway volume gateway ARN and target name from target ARN %q: %s", targetARN, err)
		}
		d.Set("gateway_arn", gatewayARN)
		d.Set("target_name", targetName)
	}

	return diags
}

func resourceCachediSCSIVolumeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceCachediSCSIVolumeRead(ctx, d, meta)...)
}

func resourceCachediSCSIVolumeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayConn(ctx)

	input := &storagegateway.DeleteVolumeInput{
		VolumeARN: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Storage Gateway cached iSCSI volume: %s", input)
	err := retry.RetryContext(ctx, 2*time.Minute, func() *retry.RetryError {
		_, err := conn.DeleteVolumeWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, storagegateway.ErrorCodeVolumeNotFound) {
				return nil
			}
			// InvalidGatewayRequestException: The specified gateway is not connected.
			// Can occur during concurrent DeleteVolume operations
			if tfawserr.ErrMessageContains(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified gateway is not connected") {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DeleteVolumeWithContext(ctx, input)
	}
	if tfawserr.ErrMessageContains(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified volume was not found") {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Storage Gateway cached iSCSI volume %q: %s", d.Id(), err)
	}

	return diags
}

func ParseVolumeGatewayARNAndTargetNameFromARN(inputARN string) (string, string, error) {
	// inputARN = arn:aws:storagegateway:us-east-2:111122223333:gateway/sgw-12A3456B/target/iqn.1997-05.com.amazon:TargetName
	targetARN, err := arn.Parse(inputARN)
	if err != nil {
		return "", "", err
	}
	// We need to get:
	//  * The Gateway ARN portion of the target ARN resource (gateway/sgw-12A3456B)
	//  * The target name portion of the target ARN resource (TargetName)
	// First, let's split up the resource of the target ARN
	// targetARN.Resource = gateway/sgw-12A3456B/target/iqn.1997-05.com.amazon:TargetName
	expectedFormatErr := fmt.Errorf("expected resource format gateway/sgw-12A3456B/target/iqn.1997-05.com.amazon:TargetName, received: %s", targetARN.Resource)
	resourceParts := strings.SplitN(targetARN.Resource, "/", 4)
	if len(resourceParts) != 4 {
		return "", "", expectedFormatErr
	}
	gatewayARN := arn.ARN{
		AccountID: targetARN.AccountID,
		Partition: targetARN.Partition,
		Region:    targetARN.Region,
		Resource:  fmt.Sprintf("%s/%s", resourceParts[0], resourceParts[1]),
		Service:   targetARN.Service,
	}.String()
	// Second, let's split off the target name from the initiator name
	// resourceParts[3] = iqn.1997-05.com.amazon:TargetName
	initiatorParts := strings.SplitN(resourceParts[3], ":", 2)
	if len(initiatorParts) != 2 {
		return "", "", expectedFormatErr
	}
	targetName := initiatorParts[1]
	return gatewayARN, targetName, nil
}
