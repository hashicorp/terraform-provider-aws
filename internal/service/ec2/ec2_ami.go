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
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	amiRetryTimeout    = 40 * time.Minute
	amiDeleteTimeout   = 90 * time.Minute
	amiRetryDelay      = 5 * time.Second
	amiRetryMinTimeout = 3 * time.Second
)

// @SDKResource("aws_ami", name="AMI")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceAMI() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAMICreate,
		// The Read, Update and Delete operations are shared with aws_ami_copy and aws_ami_from_instance,
		// since they differ only in how the image is created.
		ReadWithoutTimeout:   resourceAMIRead,
		UpdateWithoutTimeout: resourceAMIUpdate,
		DeleteWithoutTimeout: resourceAMIDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(amiRetryTimeout),
			Update: schema.DefaultTimeout(amiRetryTimeout),
			Delete: schema.DefaultTimeout(amiDeleteTimeout),
		},

		// Keep in sync with aws_ami_copy's and aws_ami_from_instance's schemas.
		Schema: map[string]*schema.Schema{
			"architecture": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          awstypes.ArchitectureValuesX8664,
				ValidateDiagFunc: enum.Validate[awstypes.ArchitectureValues](),
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"boot_mode": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.BootModeValues](),
			},
			"deprecation_time": {
				Type:                  schema.TypeString,
				Optional:              true,
				ValidateFunc:          validation.IsRFC3339Time,
				DiffSuppressFunc:      verify.SuppressEquivalentRoundedTime(time.RFC3339, time.Minute),
				DiffSuppressOnRefresh: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			// The following block device attributes intentionally mimick the
			// corresponding attributes on aws_instance, since they have the
			// same meaning.
			// However, we don't use root_block_device here because the constraint
			// on which root device attributes can be overridden for an instance to
			// not apply when registering an AMI.
			"ebs_block_device": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDeleteOnTermination: {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
							ForceNew: true,
						},
						names.AttrDeviceName: {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						names.AttrEncrypted: {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
						},
						names.AttrIOPS: {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
						},
						"outpost_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
						names.AttrSnapshotID: {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						names.AttrThroughput: {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						names.AttrVolumeSize: {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						names.AttrVolumeType: {
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							Default:          awstypes.VolumeTypeStandard,
							ValidateDiagFunc: enum.Validate[awstypes.VolumeType](),
						},
					},
				},
				Set: func(v interface{}) int {
					var buf bytes.Buffer
					m := v.(map[string]interface{})
					buf.WriteString(fmt.Sprintf("%s-", m[names.AttrDeviceName].(string)))
					buf.WriteString(fmt.Sprintf("%s-", m[names.AttrSnapshotID].(string)))
					return create.StringHashcode(buf.String())
				},
			},
			"ena_support": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"ephemeral_block_device": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDeviceName: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrVirtualName: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
				Set: func(v interface{}) int {
					var buf bytes.Buffer
					m := v.(map[string]interface{})
					buf.WriteString(fmt.Sprintf("%s-", m[names.AttrDeviceName].(string)))
					buf.WriteString(fmt.Sprintf("%s-", m[names.AttrVirtualName].(string)))
					return create.StringHashcode(buf.String())
				},
			},
			"hypervisor": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"image_location": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"image_owner_alias": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"image_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"imds_support": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true, // this attribute can only be set at registration time
				ValidateFunc: validation.StringInSlice([]string{"v2.0"}, false),
			},
			"kernel_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			// Not a public attribute; used to let the aws_ami_copy and aws_ami_from_instance
			// resources record that they implicitly created new EBS snapshots that we should
			// now manage. Not set by aws_ami, since the snapshots used there are presumed to
			// be independently managed.
			"manage_ebs_snapshots": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"platform_details": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"platform": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"ramdisk_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"root_device_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"root_snapshot_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sriov_net_support": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  sriovNetSupportSimple,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"tpm_support": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.TpmSupportValues](),
			},
			"usage_operation": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"virtualization_type": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          awstypes.VirtualizationTypeParavirtual,
				ValidateDiagFunc: enum.Validate[awstypes.VirtualizationType](),
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAMICreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	name := d.Get(names.AttrName).(string)
	input := &ec2.RegisterImageInput{
		Architecture:       awstypes.ArchitectureValues(d.Get("architecture").(string)),
		Description:        aws.String(d.Get(names.AttrDescription).(string)),
		EnaSupport:         aws.Bool(d.Get("ena_support").(bool)),
		ImageLocation:      aws.String(d.Get("image_location").(string)),
		Name:               aws.String(name),
		RootDeviceName:     aws.String(d.Get("root_device_name").(string)),
		SriovNetSupport:    aws.String(d.Get("sriov_net_support").(string)),
		VirtualizationType: aws.String(d.Get("virtualization_type").(string)),
	}

	if v := d.Get("boot_mode").(string); v != "" {
		input.BootMode = awstypes.BootModeValues(v)
	}

	if v := d.Get("imds_support").(string); v != "" {
		input.ImdsSupport = awstypes.ImdsSupportValues(v)
	}

	if kernelId := d.Get("kernel_id").(string); kernelId != "" {
		input.KernelId = aws.String(kernelId)
	}

	if ramdiskId := d.Get("ramdisk_id").(string); ramdiskId != "" {
		input.RamdiskId = aws.String(ramdiskId)
	}

	if v := d.Get("tpm_support").(string); v != "" {
		input.TpmSupport = awstypes.TpmSupportValues(v)
	}

	if v, ok := d.GetOk("ebs_block_device"); ok && v.(*schema.Set).Len() > 0 {
		for _, tfMapRaw := range v.(*schema.Set).List() {
			tfMap, ok := tfMapRaw.(map[string]interface{})

			if !ok {
				continue
			}

			var encrypted bool

			if v, ok := tfMap[names.AttrEncrypted].(bool); ok {
				encrypted = v
			}

			var snapshot string

			if v, ok := tfMap[names.AttrSnapshotID].(string); ok && v != "" {
				snapshot = v
			}

			if snapshot != "" && encrypted {
				return sdkdiag.AppendErrorf(diags, "can't set both 'snapshot_id' and 'encrypted'")
			}
		}

		input.BlockDeviceMappings = expandBlockDeviceMappingsForAMIEBSBlockDevice(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("ephemeral_block_device"); ok && v.(*schema.Set).Len() > 0 {
		input.BlockDeviceMappings = append(input.BlockDeviceMappings, expandBlockDeviceMappingsForAMIEphemeralBlockDevice(v.(*schema.Set).List())...)
	}

	output, err := conn.RegisterImage(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 AMI (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.ImageId))

	if err := createTags(ctx, conn, d.Id(), getTagsIn(ctx)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting EC2 AMI (%s) tags: %s", d.Id(), err)
	}

	if _, err := waitImageAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 AMI (%s): waiting for completion: %s", name, err)
	}

	if v, ok := d.GetOk("deprecation_time"); ok {
		if err := enableImageDeprecation(ctx, conn, d.Id(), v.(string)); err != nil {
			return sdkdiag.AppendErrorf(diags, "creating EC2 AMI (%s): %s", name, err)
		}
	}

	return append(diags, resourceAMIRead(ctx, d, meta)...)
}

func resourceAMIRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, ec2PropagationTimeout, func() (interface{}, error) {
		return findImageByID(ctx, conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 AMI %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 AMI (%s): %s", d.Id(), err)
	}

	image := outputRaw.(*awstypes.Image)

	if image.State == awstypes.ImageStatePending {
		// This could happen if a user manually adds an image we didn't create
		// to the state. We'll wait for the image to become available
		// before we continue. We should never take this branch in normal
		// circumstances since we would've waited for availability during
		// the "Create" step.
		image, err = waitImageAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate))

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EC2 AMI (%s) create: %s", d.Id(), err)
		}
	}

	d.Set("architecture", image.Architecture)
	imageArn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("image/%s", d.Id()),
		Service:   names.EC2,
	}.String()
	d.Set(names.AttrARN, imageArn)
	d.Set("boot_mode", image.BootMode)
	d.Set(names.AttrDescription, image.Description)
	d.Set("deprecation_time", image.DeprecationTime)
	d.Set("ena_support", image.EnaSupport)
	d.Set("hypervisor", image.Hypervisor)
	d.Set("image_location", image.ImageLocation)
	d.Set("image_owner_alias", image.ImageOwnerAlias)
	d.Set("image_type", image.ImageType)
	d.Set("imds_support", image.ImdsSupport)
	d.Set("kernel_id", image.KernelId)
	d.Set(names.AttrName, image.Name)
	d.Set(names.AttrOwnerID, image.OwnerId)
	d.Set("platform_details", image.PlatformDetails)
	d.Set("platform", image.Platform)
	d.Set("public", image.Public)
	d.Set("ramdisk_id", image.RamdiskId)
	d.Set("root_device_name", image.RootDeviceName)
	d.Set("root_snapshot_id", amiRootSnapshotId(*image))
	d.Set("sriov_net_support", image.SriovNetSupport)
	d.Set("tpm_support", image.TpmSupport)
	d.Set("usage_operation", image.UsageOperation)
	d.Set("virtualization_type", image.VirtualizationType)

	if err := d.Set("ebs_block_device", flattenBlockDeviceMappingsForAMIEBSBlockDevice(image.BlockDeviceMappings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ebs_block_device: %s", err)
	}

	if err := d.Set("ephemeral_block_device", flattenBlockDeviceMappingsForAMIEphemeralBlockDevice(image.BlockDeviceMappings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ephemeral_block_device: %s", err)
	}

	setTagsOut(ctx, image.Tags)

	return diags
}

func resourceAMIUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if d.HasChange(names.AttrDescription) {
		err := updateDescription(ctx, conn, d.Id(), d.Get(names.AttrDescription).(string))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 AMI (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("deprecation_time") {
		if v := d.Get("deprecation_time").(string); v != "" {
			if err := enableImageDeprecation(ctx, conn, d.Id(), v); err != nil {
				return sdkdiag.AppendErrorf(diags, "updating EC2 AMI (%s): %s", d.Id(), err)
			}
		} else {
			if err := disableImageDeprecation(ctx, conn, d.Id()); err != nil {
				return sdkdiag.AppendErrorf(diags, "updating EC2 AMI (%s):  %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceAMIRead(ctx, d, meta)...)
}

func resourceAMIDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[INFO] Deleting EC2 AMI: %s", d.Id())
	_, err := conn.DeregisterImage(ctx, &ec2.DeregisterImageInput{
		ImageId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidAMIIDNotFound, errCodeInvalidAMIIDUnavailable) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deregistering EC2 AMI (%s): %s", d.Id(), err)
	}

	// If we're managing the EBS snapshots then we need to delete those too.
	if d.Get("manage_ebs_snapshots").(bool) {
		errs := map[string]error{}
		ebsBlockDevsSet := d.Get("ebs_block_device").(*schema.Set)
		req := &ec2.DeleteSnapshotInput{}
		for _, ebsBlockDevI := range ebsBlockDevsSet.List() {
			ebsBlockDev := ebsBlockDevI.(map[string]interface{})
			snapshotId := ebsBlockDev[names.AttrSnapshotID].(string)
			if snapshotId != "" {
				req.SnapshotId = aws.String(snapshotId)
				_, err := conn.DeleteSnapshot(ctx, req)
				if err != nil {
					errs[snapshotId] = err
				}
			}
		}

		if len(errs) > 0 {
			errParts := []string{"Errors while deleting associated EBS snapshots:"}
			for snapshotId, err := range errs {
				errParts = append(errParts, fmt.Sprintf("%s: %s", snapshotId, err))
			}
			errParts = append(errParts, "These are no longer managed by Terraform and must be deleted manually.")
			return sdkdiag.AppendErrorf(diags, strings.Join(errParts, "\n"))
		}
	}

	if _, err := waitImageDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 AMI (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func updateDescription(ctx context.Context, conn *ec2.Client, id string, description string) error {
	input := &ec2.ModifyImageAttributeInput{
		Description: &awstypes.AttributeValue{
			Value: aws.String(description),
		},
		ImageId: aws.String(id),
	}

	_, err := conn.ModifyImageAttribute(ctx, input)
	if err != nil {
		return fmt.Errorf("updating description: %s", err)
	}

	err = waitImageDescriptionUpdated(ctx, conn, id, description)
	if err != nil {
		return fmt.Errorf("updating description: waiting for completion: %s", err)
	}

	return nil
}

func enableImageDeprecation(ctx context.Context, conn *ec2.Client, id string, deprecateAt string) error {
	v, _ := time.Parse(time.RFC3339, deprecateAt)
	input := &ec2.EnableImageDeprecationInput{
		DeprecateAt: aws.Time(v),
		ImageId:     aws.String(id),
	}

	_, err := conn.EnableImageDeprecation(ctx, input)

	if err != nil {
		return fmt.Errorf("enabling deprecation: %w", err)
	}

	err = waitImageDeprecationTimeUpdated(ctx, conn, id, deprecateAt)

	if err != nil {
		return fmt.Errorf("enabling deprecation: waiting for completion: %w", err)
	}

	return nil
}

func disableImageDeprecation(ctx context.Context, conn *ec2.Client, id string) error {
	input := &ec2.DisableImageDeprecationInput{
		ImageId: aws.String(id),
	}

	_, err := conn.DisableImageDeprecation(ctx, input)

	if err != nil {
		return fmt.Errorf("disabling deprecation: %w", err)
	}

	err = waitImageDeprecationTimeDisabled(ctx, conn, id)

	if err != nil {
		return fmt.Errorf("disabling deprecation: waiting for completion: %w", err)
	}

	return nil
}

func expandBlockDeviceMappingForAMIEBSBlockDevice(tfMap map[string]interface{}) awstypes.BlockDeviceMapping {
	apiObject := awstypes.BlockDeviceMapping{
		Ebs: &awstypes.EbsBlockDevice{},
	}

	if v, ok := tfMap[names.AttrDeleteOnTermination].(bool); ok {
		apiObject.Ebs.DeleteOnTermination = aws.Bool(v)
	}

	if v, ok := tfMap[names.AttrDeviceName].(string); ok && v != "" {
		apiObject.DeviceName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrIOPS].(int); ok && v != 0 {
		apiObject.Ebs.Iops = aws.Int32(int32(v))
	}

	// "Parameter encrypted is invalid. You cannot specify the encrypted flag if specifying a snapshot id in a block device mapping."
	if v, ok := tfMap[names.AttrSnapshotID].(string); ok && v != "" {
		apiObject.Ebs.SnapshotId = aws.String(v)
	} else if v, ok := tfMap[names.AttrEncrypted].(bool); ok {
		apiObject.Ebs.Encrypted = aws.Bool(v)
	}

	if v, ok := tfMap[names.AttrThroughput].(int); ok && v != 0 {
		apiObject.Ebs.Throughput = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrVolumeSize].(int); ok && v != 0 {
		apiObject.Ebs.VolumeSize = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrVolumeType].(string); ok && v != "" {
		apiObject.Ebs.VolumeType = awstypes.VolumeType(v)
	}

	if v, ok := tfMap["outpost_arn"].(string); ok && v != "" {
		apiObject.Ebs.OutpostArn = aws.String(v)
	}

	return apiObject
}

func expandBlockDeviceMappingsForAMIEBSBlockDevice(tfList []interface{}) []awstypes.BlockDeviceMapping {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.BlockDeviceMapping

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandBlockDeviceMappingForAMIEBSBlockDevice(tfMap))
	}

	return apiObjects
}

func flattenBlockDeviceMappingForAMIEBSBlockDevice(apiObject awstypes.BlockDeviceMapping) map[string]interface{} {
	if apiObject.Ebs == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		names.AttrVolumeType: apiObject.Ebs.VolumeType,
	}

	if v := apiObject.Ebs.DeleteOnTermination; v != nil {
		tfMap[names.AttrDeleteOnTermination] = aws.ToBool(v)
	}

	if v := apiObject.DeviceName; v != nil {
		tfMap[names.AttrDeviceName] = aws.ToString(v)
	}

	if v := apiObject.Ebs.Encrypted; v != nil {
		tfMap[names.AttrEncrypted] = aws.ToBool(v)
	}

	if v := apiObject.Ebs.Iops; v != nil {
		tfMap[names.AttrIOPS] = aws.ToInt32(v)
	}

	if v := apiObject.Ebs.SnapshotId; v != nil {
		tfMap[names.AttrSnapshotID] = aws.ToString(v)
	}

	if v := apiObject.Ebs.Throughput; v != nil {
		tfMap[names.AttrThroughput] = aws.ToInt32(v)
	}

	if v := apiObject.Ebs.VolumeSize; v != nil {
		tfMap[names.AttrVolumeSize] = aws.ToInt32(v)
	}

	if v := apiObject.Ebs.OutpostArn; v != nil {
		tfMap["outpost_arn"] = aws.ToString(v)
	}

	return tfMap
}

func flattenBlockDeviceMappingsForAMIEBSBlockDevice(apiObjects []awstypes.BlockDeviceMapping) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject.Ebs == nil {
			continue
		}

		tfList = append(tfList, flattenBlockDeviceMappingForAMIEBSBlockDevice(apiObject))
	}

	return tfList
}

func expandBlockDeviceMappingForAMIEphemeralBlockDevice(tfMap map[string]interface{}) awstypes.BlockDeviceMapping {
	apiObject := awstypes.BlockDeviceMapping{}

	if v, ok := tfMap[names.AttrDeviceName].(string); ok && v != "" {
		apiObject.DeviceName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrVirtualName].(string); ok && v != "" {
		apiObject.VirtualName = aws.String(v)
	}

	return apiObject
}

func expandBlockDeviceMappingsForAMIEphemeralBlockDevice(tfList []interface{}) []awstypes.BlockDeviceMapping {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.BlockDeviceMapping

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandBlockDeviceMappingForAMIEphemeralBlockDevice(tfMap))
	}

	return apiObjects
}

func flattenBlockDeviceMappingForAMIEphemeralBlockDevice(apiObject awstypes.BlockDeviceMapping) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.DeviceName; v != nil {
		tfMap[names.AttrDeviceName] = aws.ToString(v)
	}

	if v := apiObject.VirtualName; v != nil {
		tfMap[names.AttrVirtualName] = aws.ToString(v)
	}

	return tfMap
}

func flattenBlockDeviceMappingsForAMIEphemeralBlockDevice(apiObjects []awstypes.BlockDeviceMapping) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject.Ebs != nil {
			continue
		}

		tfList = append(tfList, flattenBlockDeviceMappingForAMIEphemeralBlockDevice(apiObject))
	}

	return tfList
}

const imageDeprecationPropagationTimeout = 2 * time.Minute

func waitImageDescriptionUpdated(ctx context.Context, conn *ec2.Client, imageID, expectedValue string) error {
	return tfresource.WaitUntil(ctx, imageDeprecationPropagationTimeout, func() (bool, error) {
		output, err := findImageByID(ctx, conn, imageID)

		if tfresource.NotFound(err) {
			return false, nil
		}

		if err != nil {
			return false, err
		}

		return aws.ToString(output.Description) == expectedValue, nil
	},
		tfresource.WaitOpts{
			Delay:      amiRetryDelay,
			MinTimeout: amiRetryMinTimeout,
		},
	)
}

func waitImageDeprecationTimeUpdated(ctx context.Context, conn *ec2.Client, imageID, expectedValue string) error {
	expected, err := time.Parse(time.RFC3339, expectedValue)
	if err != nil {
		return err
	}
	expected = expected.Round(time.Minute)

	return tfresource.WaitUntil(ctx, imageDeprecationPropagationTimeout, func() (bool, error) {
		output, err := findImageByID(ctx, conn, imageID)

		if tfresource.NotFound(err) {
			return false, nil
		}

		if err != nil {
			return false, err
		}

		if output.DeprecationTime == nil {
			return false, nil
		}

		dt, err := time.Parse(time.RFC3339, *output.DeprecationTime)
		if err != nil {
			return false, err
		}
		dt = dt.Round(time.Minute)

		return expected.Equal(dt), nil
	},
		tfresource.WaitOpts{
			Delay:      amiRetryDelay,
			MinTimeout: amiRetryMinTimeout,
		},
	)
}

func waitImageDeprecationTimeDisabled(ctx context.Context, conn *ec2.Client, imageID string) error {
	return tfresource.WaitUntil(ctx, imageDeprecationPropagationTimeout, func() (bool, error) {
		output, err := findImageByID(ctx, conn, imageID)

		if tfresource.NotFound(err) {
			return false, nil
		}

		if err != nil {
			return false, err
		}

		return output.DeprecationTime == nil, nil
	},
		tfresource.WaitOpts{
			Delay:      amiRetryDelay,
			MinTimeout: amiRetryMinTimeout,
		},
	)
}
