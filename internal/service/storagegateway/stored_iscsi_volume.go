package storagegateway

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceStorediSCSIVolume() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStorediSCSIVolumeCreate,
		ReadWithoutTimeout:   resourceStorediSCSIVolumeRead,
		UpdateWithoutTimeout: resourceStorediSCSIVolumeUpdate,
		DeleteWithoutTimeout: resourceStorediSCSIVolumeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
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
			"target_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"preserve_existing_data": {
				Type:     schema.TypeBool,
				Required: true,
				ForceNew: true,
			},
			"kms_encrypted": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"kms_key": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
				RequiredWith: []string{"kms_encrypted"},
			},
			// Poor API naming: this accepts the IP address of the network interface
			"network_interface_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"network_interface_port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"snapshot_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"chap_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"lun_number": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"target_arn": {
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
			"volume_attachment_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"volume_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceStorediSCSIVolumeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &storagegateway.CreateStorediSCSIVolumeInput{
		DiskId:               aws.String(d.Get("disk_id").(string)),
		GatewayARN:           aws.String(d.Get("gateway_arn").(string)),
		NetworkInterfaceId:   aws.String(d.Get("network_interface_id").(string)),
		TargetName:           aws.String(d.Get("target_name").(string)),
		PreserveExistingData: aws.Bool(d.Get("preserve_existing_data").(bool)),
		Tags:                 Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("snapshot_id"); ok {
		input.SnapshotId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("kms_key"); ok {
		input.KMSKey = aws.String(v.(string))
	}

	if v, ok := d.GetOk("kms_encrypted"); ok {
		input.KMSEncrypted = aws.Bool(v.(bool))
	}

	log.Printf("[DEBUG] Creating Storage Gateway Stored iSCSI volume: %s", input)
	output, err := conn.CreateStorediSCSIVolumeWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Storage Gateway Stored iSCSI volume: %s", err)
	}

	d.SetId(aws.StringValue(output.VolumeARN))

	_, err = waitStorediSCSIVolumeAvailable(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Stored Iscsi Volume %q to be Available: %s", d.Id(), err)
	}

	return append(diags, resourceStorediSCSIVolumeRead(ctx, d, meta)...)
}

func resourceStorediSCSIVolumeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayConn()

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating tags: %s", err)
		}
	}

	return append(diags, resourceStorediSCSIVolumeRead(ctx, d, meta)...)
}

func resourceStorediSCSIVolumeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &storagegateway.DescribeStorediSCSIVolumesInput{
		VolumeARNs: []*string{aws.String(d.Id())},
	}

	log.Printf("[DEBUG] Reading Storage Gateway Stored iSCSI volume: %s", input)
	output, err := conn.DescribeStorediSCSIVolumesWithContext(ctx, input)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, storagegateway.ErrorCodeVolumeNotFound) || tfawserr.ErrMessageContains(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified volume was not found") {
			log.Printf("[WARN] Storage Gateway Stored iSCSI volume %q not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading Storage Gateway Stored iSCSI volume %q: %s", d.Id(), err)
	}

	if output == nil || len(output.StorediSCSIVolumes) == 0 || output.StorediSCSIVolumes[0] == nil || aws.StringValue(output.StorediSCSIVolumes[0].VolumeARN) != d.Id() {
		log.Printf("[WARN] Storage Gateway Stored iSCSI volume %q not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	volume := output.StorediSCSIVolumes[0]

	arn := aws.StringValue(volume.VolumeARN)
	d.Set("arn", arn)
	d.Set("disk_id", volume.VolumeDiskId)
	d.Set("snapshot_id", volume.SourceSnapshotId)
	d.Set("volume_id", volume.VolumeId)
	d.Set("volume_type", volume.VolumeType)
	d.Set("volume_size_in_bytes", volume.VolumeSizeInBytes)
	d.Set("volume_status", volume.VolumeStatus)
	d.Set("volume_attachment_status", volume.VolumeAttachmentStatus)
	d.Set("preserve_existing_data", volume.PreservedExistingData)
	d.Set("kms_key", volume.KMSKey)
	d.Set("kms_encrypted", volume.KMSKey != nil)

	tags, err := ListTags(ctx, conn, arn)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for resource (%s): %s", arn, err)
	}
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	attr := volume.VolumeiSCSIAttributes
	d.Set("chap_enabled", attr.ChapEnabled)
	d.Set("lun_number", attr.LunNumber)
	d.Set("network_interface_id", attr.NetworkInterfaceId)
	d.Set("network_interface_port", attr.NetworkInterfacePort)

	targetARN := aws.StringValue(attr.TargetARN)
	d.Set("target_arn", targetARN)

	gatewayARN, targetName, err := ParseVolumeGatewayARNAndTargetNameFromARN(targetARN)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing Storage Gateway volume gateway ARN and target name from target ARN %q: %s", targetARN, err)
	}
	d.Set("gateway_arn", gatewayARN)
	d.Set("target_name", targetName)

	return diags
}

func resourceStorediSCSIVolumeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayConn()

	input := &storagegateway.DeleteVolumeInput{
		VolumeARN: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Storage Gateway Stored iSCSI volume: %s", input)
	err := resource.RetryContext(ctx, 2*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteVolumeWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, storagegateway.ErrorCodeVolumeNotFound) {
				return nil
			}
			// InvalidGatewayRequestException: The specified gateway is not connected.
			// Can occur during concurrent DeleteVolume operations
			if tfawserr.ErrMessageContains(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified gateway is not connected") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
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
		return sdkdiag.AppendErrorf(diags, "deleting Storage Gateway Stored iSCSI volume %q: %s", d.Id(), err)
	}

	return diags
}
