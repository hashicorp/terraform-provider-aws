package storagegateway

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceCachediSCSIVolume() *schema.Resource {
	return &schema.Resource{
		Create: resourceCachediSCSIVolumeCreate,
		Read:   resourceCachediSCSIVolumeRead,
		Update: resourceCachediSCSIVolumeUpdate,
		Delete: resourceCachediSCSIVolumeDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
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
			"source_volume_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"target_arn": {
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
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
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceCachediSCSIVolumeCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).StorageGatewayConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &storagegateway.CreateCachediSCSIVolumeInput{
		ClientToken:        aws.String(resource.UniqueId()),
		GatewayARN:         aws.String(d.Get("gateway_arn").(string)),
		NetworkInterfaceId: aws.String(d.Get("network_interface_id").(string)),
		TargetName:         aws.String(d.Get("target_name").(string)),
		VolumeSizeInBytes:  aws.Int64(int64(d.Get("volume_size_in_bytes").(int))),
		Tags:               tags.IgnoreAws().StoragegatewayTags(),
	}

	if v, ok := d.GetOk("snapshot_id"); ok {
		input.SnapshotId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("source_volume_arn"); ok {
		input.SourceVolumeARN = aws.String(v.(string))
	}

	if v, ok := d.GetOk("kms_key"); ok {
		input.KMSKey = aws.String(v.(string))
	}

	if v, ok := d.GetOk("kms_encrypted"); ok {
		input.KMSEncrypted = aws.Bool(v.(bool))
	}

	log.Printf("[DEBUG] Creating Storage Gateway cached iSCSI volume: %s", input)
	output, err := conn.CreateCachediSCSIVolume(input)
	if err != nil {
		return fmt.Errorf("error creating Storage Gateway cached iSCSI volume: %s", err)
	}

	d.SetId(aws.StringValue(output.VolumeARN))

	return resourceCachediSCSIVolumeRead(d, meta)
}

func resourceCachediSCSIVolumeUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).StorageGatewayConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := tftags.StoragegatewayUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceCachediSCSIVolumeRead(d, meta)
}

func resourceCachediSCSIVolumeRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).StorageGatewayConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &storagegateway.DescribeCachediSCSIVolumesInput{
		VolumeARNs: []*string{aws.String(d.Id())},
	}

	log.Printf("[DEBUG] Reading Storage Gateway cached iSCSI volume: %s", input)
	output, err := conn.DescribeCachediSCSIVolumes(input)

	if err != nil {
		if tfawserr.ErrMessageContains(err, storagegateway.ErrorCodeVolumeNotFound, "") || tfawserr.ErrMessageContains(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified volume was not found") {
			log.Printf("[WARN] Storage Gateway cached iSCSI volume %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading Storage Gateway cached iSCSI volume %q: %s", d.Id(), err)
	}

	if output == nil || len(output.CachediSCSIVolumes) == 0 || output.CachediSCSIVolumes[0] == nil || aws.StringValue(output.CachediSCSIVolumes[0].VolumeARN) != d.Id() {
		log.Printf("[WARN] Storage Gateway cached iSCSI volume %q not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	volume := output.CachediSCSIVolumes[0]

	arn := aws.StringValue(volume.VolumeARN)
	d.Set("arn", arn)
	d.Set("snapshot_id", volume.SourceSnapshotId)
	d.Set("volume_arn", arn)
	d.Set("volume_id", volume.VolumeId)
	d.Set("volume_size_in_bytes", volume.VolumeSizeInBytes)
	d.Set("kms_key", volume.KMSKey)
	if volume.KMSKey != nil {
		d.Set("kms_encrypted", true)
	} else {
		d.Set("kms_encrypted", false)
	}

	tags, err := tftags.StoragegatewayListTags(conn, arn)
	if err != nil {
		return fmt.Errorf("error listing tags for resource (%s): %s", arn, err)
	}
	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	if volume.VolumeiSCSIAttributes != nil {
		d.Set("chap_enabled", volume.VolumeiSCSIAttributes.ChapEnabled)
		d.Set("lun_number", volume.VolumeiSCSIAttributes.LunNumber)
		d.Set("network_interface_id", volume.VolumeiSCSIAttributes.NetworkInterfaceId)
		d.Set("network_interface_port", volume.VolumeiSCSIAttributes.NetworkInterfacePort)

		targetARN := aws.StringValue(volume.VolumeiSCSIAttributes.TargetARN)
		d.Set("target_arn", targetARN)

		gatewayARN, targetName, err := ParseVolumeGatewayARNAndTargetNameFromARN(targetARN)
		if err != nil {
			return fmt.Errorf("error parsing Storage Gateway volume gateway ARN and target name from target ARN %q: %s", targetARN, err)
		}
		d.Set("gateway_arn", gatewayARN)
		d.Set("target_name", targetName)
	}

	return nil
}

func resourceCachediSCSIVolumeDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).StorageGatewayConn

	input := &storagegateway.DeleteVolumeInput{
		VolumeARN: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Storage Gateway cached iSCSI volume: %s", input)
	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteVolume(input)
		if err != nil {
			if tfawserr.ErrMessageContains(err, storagegateway.ErrorCodeVolumeNotFound, "") {
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
		_, err = conn.DeleteVolume(input)
	}
	if tfawserr.ErrMessageContains(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified volume was not found") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting Storage Gateway cached iSCSI volume %q: %s", d.Id(), err)
	}

	return nil
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
