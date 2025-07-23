// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package storagegateway

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/storagegateway"
	awstypes "github.com/aws/aws-sdk-go-v2/service/storagegateway/types"
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
		},
	}
}

func resourceCachediSCSIVolumeCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayClient(ctx)

	input := &storagegateway.CreateCachediSCSIVolumeInput{
		ClientToken:        aws.String(id.UniqueId()),
		GatewayARN:         aws.String(d.Get("gateway_arn").(string)),
		NetworkInterfaceId: aws.String(d.Get(names.AttrNetworkInterfaceID).(string)),
		Tags:               getTagsIn(ctx),
		TargetName:         aws.String(d.Get("target_name").(string)),
		VolumeSizeInBytes:  int64(d.Get("volume_size_in_bytes").(int)),
	}

	if v, ok := d.GetOk(names.AttrKMSKey); ok {
		input.KMSKey = aws.String(v.(string))
	}

	if v, ok := d.GetOk("kms_encrypted"); ok {
		input.KMSEncrypted = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk(names.AttrSnapshotID); ok {
		input.SnapshotId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("source_volume_arn"); ok {
		input.SourceVolumeARN = aws.String(v.(string))
	}

	output, err := conn.CreateCachediSCSIVolume(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Storage Gateway Cached iSCSI Volume: %s", err)
	}

	d.SetId(aws.ToString(output.VolumeARN))

	return append(diags, resourceCachediSCSIVolumeRead(ctx, d, meta)...)
}

func resourceCachediSCSIVolumeRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayClient(ctx)

	volume, err := findCachediSCSIVolumeByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Storage Gateway Cached iSCSI Volume (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Storage Gateway Cached iSCSI Volume (%s): %s", d.Id(), err)
	}

	arn := aws.ToString(volume.VolumeARN)
	d.Set(names.AttrARN, arn)
	d.Set("kms_encrypted", volume.KMSKey != nil)
	d.Set(names.AttrKMSKey, volume.KMSKey)
	d.Set(names.AttrSnapshotID, volume.SourceSnapshotId)
	d.Set("volume_arn", arn)
	d.Set("volume_id", volume.VolumeId)
	d.Set("volume_size_in_bytes", volume.VolumeSizeInBytes)

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

func resourceCachediSCSIVolumeUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceCachediSCSIVolumeRead(ctx, d, meta)...)
}

func resourceCachediSCSIVolumeDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayClient(ctx)

	log.Printf("[DEBUG] Deleting Storage Gateway Cached iSCSI Volume: %s", d.Id())
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
		return sdkdiag.AppendErrorf(diags, "deleting Storage Gateway cached iSCSI Volume (%s): %s", d.Id(), err)
	}

	return diags
}

func findCachediSCSIVolumeByARN(ctx context.Context, conn *storagegateway.Client, arn string) (*awstypes.CachediSCSIVolume, error) {
	input := &storagegateway.DescribeCachediSCSIVolumesInput{
		VolumeARNs: []string{arn},
	}
	output, err := findCachediSCSIVolume(ctx, conn, input)

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

func findCachediSCSIVolume(ctx context.Context, conn *storagegateway.Client, input *storagegateway.DescribeCachediSCSIVolumesInput) (*awstypes.CachediSCSIVolume, error) {
	output, err := findCachediSCSIVolumes(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findCachediSCSIVolumes(ctx context.Context, conn *storagegateway.Client, input *storagegateway.DescribeCachediSCSIVolumesInput) ([]awstypes.CachediSCSIVolume, error) {
	output, err := conn.DescribeCachediSCSIVolumes(ctx, input)

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

	return output.CachediSCSIVolumes, nil
}

func parseVolumeGatewayARNAndTargetNameFromARN(inputARN string) (string, string, error) {
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
