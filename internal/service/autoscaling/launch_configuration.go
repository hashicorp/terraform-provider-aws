// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscaling

import ( // nosemgrep:ci.semgrep.aws.multiple-service-imports
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	awstypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_launch_configuration", name="Launch Configuration")
func resourceLaunchConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLaunchConfigurationCreate,
		ReadWithoutTimeout:   resourceLaunchConfigurationRead,
		DeleteWithoutTimeout: resourceLaunchConfigurationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"associate_public_ip_address": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
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
							Computed: true,
							ForceNew: true,
						},
						names.AttrIOPS: {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						"no_device": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
						},
						names.AttrSnapshotID: {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
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
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
					},
				},
			},
			"ebs_optimized": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"enable_monitoring": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  true,
			},
			"ephemeral_block_device": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDeviceName: {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"no_device": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
						},
						names.AttrVirtualName: {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},
			"iam_instance_profile": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"image_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrInstanceType: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"key_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"metadata_options": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"http_endpoint": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(enum.Slice(awstypes.InstanceMetadataEndpointStateEnabled, awstypes.InstanceMetadataEndpointStateDisabled), false),
						},
						"http_put_response_hop_limit": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(1, 64),
						},
						"http_tokens": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(enum.Slice(awstypes.InstanceMetadataHttpTokensStateOptional, awstypes.InstanceMetadataHttpTokensStateRequired), false),
						},
					},
				},
			},
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrNamePrefix},
				ValidateFunc:  validation.StringLenBetween(1, 255),
			},
			names.AttrNamePrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrName},
				ValidateFunc:  validation.StringLenBetween(1, 255-id.UniqueIDSuffixLength),
			},
			"placement_tenancy": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"root_block_device": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					// "You can only modify the volume size, volume type, and Delete on
					// Termination flag on the block device mapping entry for the root
					// device volume." - bit.ly/ec2bdmap
					Schema: map[string]*schema.Schema{
						names.AttrDeleteOnTermination: {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
							ForceNew: true,
						},
						names.AttrEncrypted: {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						names.AttrIOPS: {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
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
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
					},
				},
			},
			names.AttrSecurityGroups: {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"spot_price": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"user_data": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"user_data_base64"},
				StateFunc: func(v interface{}) string {
					switch v := v.(type) {
					case string:
						return userDataHashSum(v)
					default:
						return ""
					}
				},
				ValidateFunc: validation.StringLenBetween(1, 16384),
			},
			"user_data_base64": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"user_data"},
				ValidateFunc:  verify.ValidBase64String,
			},
		},
	}
}

func resourceLaunchConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	autoscalingconn := meta.(*conns.AWSClient).AutoScalingClient(ctx)
	ec2conn := meta.(*conns.AWSClient).EC2Client(ctx)

	lcName := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	input := autoscaling.CreateLaunchConfigurationInput{
		EbsOptimized:            aws.Bool(d.Get("ebs_optimized").(bool)),
		ImageId:                 aws.String(d.Get("image_id").(string)),
		InstanceType:            aws.String(d.Get(names.AttrInstanceType).(string)),
		LaunchConfigurationName: aws.String(lcName),
	}

	associatePublicIPAddress := d.GetRawConfig().GetAttr("associate_public_ip_address")
	if associatePublicIPAddress.IsKnown() && !associatePublicIPAddress.IsNull() {
		input.AssociatePublicIpAddress = aws.Bool(associatePublicIPAddress.True())
	}

	if v, ok := d.GetOk("iam_instance_profile"); ok {
		input.IamInstanceProfile = aws.String(v.(string))
	}

	input.InstanceMonitoring = &awstypes.InstanceMonitoring{
		Enabled: aws.Bool(d.Get("enable_monitoring").(bool)),
	}

	if v, ok := d.GetOk("key_name"); ok {
		input.KeyName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("metadata_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.MetadataOptions = expandInstanceMetadataOptions(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("placement_tenancy"); ok {
		input.PlacementTenancy = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrSecurityGroups); ok && v.(*schema.Set).Len() > 0 {
		input.SecurityGroups = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("spot_price"); ok {
		input.SpotPrice = aws.String(v.(string))
	}

	if v, ok := d.GetOk("user_data"); ok {
		input.UserData = flex.StringValueToBase64String(v.(string))
	} else if v, ok := d.GetOk("user_data_base64"); ok {
		input.UserData = aws.String(v.(string))
	}

	// We'll use this to detect if we're declaring it incorrectly as an ebs_block_device.
	rootDeviceName, err := findImageRootDeviceName(ctx, ec2conn, d.Get("image_id").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Auto Scaling Launch Configuration (%s): %s", lcName, err)
	}

	var blockDeviceMappings []awstypes.BlockDeviceMapping

	if v, ok := d.GetOk("ebs_block_device"); ok && v.(*schema.Set).Len() > 0 {
		v := expandBlockDeviceMappings(v.(*schema.Set).List(), expandBlockDeviceMappingForEBSBlockDevice)

		for _, v := range v {
			if aws.ToString(v.DeviceName) == rootDeviceName {
				return sdkdiag.AppendErrorf(diags, "root device (%s) declared as an 'ebs_block_device'. Use 'root_block_device' argument.", rootDeviceName)
			}
		}

		blockDeviceMappings = append(blockDeviceMappings, v...)
	}

	if v, ok := d.GetOk("ephemeral_block_device"); ok && v.(*schema.Set).Len() > 0 {
		v := expandBlockDeviceMappings(v.(*schema.Set).List(), expandBlockDeviceMappingForEphemeralBlockDevice)

		blockDeviceMappings = append(blockDeviceMappings, v...)
	}

	if v, ok := d.GetOk("root_block_device"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		v := expandBlockDeviceMappingForRootBlockDevice(v.([]interface{})[0].(map[string]interface{}))
		v.DeviceName = aws.String(rootDeviceName)

		blockDeviceMappings = append(blockDeviceMappings, v)
	}

	if len(blockDeviceMappings) > 0 {
		input.BlockDeviceMappings = blockDeviceMappings
	}

	// IAM profiles can take ~10 seconds to propagate in AWS:
	// http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/iam-roles-for-amazon-ec2.html#launch-instance-with-role-console
	_, err = tfresource.RetryWhen(ctx, propagationTimeout,
		func() (interface{}, error) {
			return autoscalingconn.CreateLaunchConfiguration(ctx, &input)
		},
		func(err error) (bool, error) {
			if tfawserr.ErrMessageContains(err, errCodeValidationError, "Invalid IamInstanceProfile") ||
				tfawserr.ErrMessageContains(err, errCodeValidationError, "You are not authorized to perform this operation") {
				return true, err
			}

			return false, err
		})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Auto Scaling Launch Configuration (%s): %s", lcName, err)
	}

	d.SetId(lcName)

	return append(diags, resourceLaunchConfigurationRead(ctx, d, meta)...)
}

func resourceLaunchConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	autoscalingconn := meta.(*conns.AWSClient).AutoScalingClient(ctx)
	ec2conn := meta.(*conns.AWSClient).EC2Client(ctx)

	lc, err := findLaunchConfigurationByName(ctx, autoscalingconn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Auto Scaling Launch Configuration %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Auto Scaling Launch Configuration (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, lc.LaunchConfigurationARN)
	d.Set("associate_public_ip_address", lc.AssociatePublicIpAddress)
	d.Set("ebs_optimized", lc.EbsOptimized)
	if lc.InstanceMonitoring != nil {
		d.Set("enable_monitoring", lc.InstanceMonitoring.Enabled)
	} else {
		d.Set("enable_monitoring", false)
	}
	d.Set("iam_instance_profile", lc.IamInstanceProfile)
	d.Set("image_id", lc.ImageId)
	d.Set(names.AttrInstanceType, lc.InstanceType)
	d.Set("key_name", lc.KeyName)
	if lc.MetadataOptions != nil {
		if err := d.Set("metadata_options", []interface{}{flattenInstanceMetadataOptions(lc.MetadataOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting metadata_options: %s", err)
		}
	} else {
		d.Set("metadata_options", nil)
	}
	d.Set(names.AttrName, lc.LaunchConfigurationName)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(lc.LaunchConfigurationName)))
	d.Set("placement_tenancy", lc.PlacementTenancy)
	d.Set(names.AttrSecurityGroups, lc.SecurityGroups)
	d.Set("spot_price", lc.SpotPrice)
	if v := aws.ToString(lc.UserData); v != "" {
		if _, ok := d.GetOk("user_data_base64"); ok {
			d.Set("user_data_base64", v)
		} else {
			d.Set("user_data", userDataHashSum(v))
		}
	}

	rootDeviceName, err := findImageRootDeviceName(ctx, ec2conn, d.Get("image_id").(string))

	if tfresource.NotFound(err) {
		// Don't block a refresh for a bad image.
		rootDeviceName = ""
	} else if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Auto Scaling Launch Configuration (%s): %s", d.Id(), err)
	}

	configuredEBSBlockDevices := make(map[string]map[string]interface{})

	if v, ok := d.GetOk("ebs_block_device"); ok && v.(*schema.Set).Len() > 0 {
		for _, v := range v.(*schema.Set).List() {
			tfMap, ok := v.(map[string]interface{})

			if !ok {
				continue
			}

			configuredEBSBlockDevices[tfMap[names.AttrDeviceName].(string)] = tfMap
		}
	}

	tfListEBSBlockDevice, tfListEphemeralBlockDevice, tfListRootBlockDevice := flattenBlockDeviceMappings(lc.BlockDeviceMappings, rootDeviceName, configuredEBSBlockDevices)

	if err := d.Set("ebs_block_device", tfListEBSBlockDevice); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ebs_block_device: %s", err)
	}
	if err := d.Set("ephemeral_block_device", tfListEphemeralBlockDevice); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ephemeral_block_device: %s", err)
	}
	if err := d.Set("root_block_device", tfListRootBlockDevice); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting root_block_device: %s", err)
	}

	return diags
}

func resourceLaunchConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingClient(ctx)

	log.Printf("[DEBUG] Deleting Auto Scaling Launch Configuration: %s", d.Id())
	_, err := tfresource.RetryWhenIsA[*awstypes.ResourceInUseFault](ctx, propagationTimeout,
		func() (interface{}, error) {
			return conn.DeleteLaunchConfiguration(ctx, &autoscaling.DeleteLaunchConfigurationInput{
				LaunchConfigurationName: aws.String(d.Id()),
			})
		})

	if tfawserr.ErrMessageContains(err, errCodeValidationError, "not found") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Auto Scaling Launch Configuration (%s): %s", d.Id(), err)
	}

	return diags
}

func expandBlockDeviceMappingForEBSBlockDevice(tfMap map[string]interface{}) awstypes.BlockDeviceMapping {
	apiObject := awstypes.BlockDeviceMapping{
		Ebs: &awstypes.Ebs{},
	}

	if v, ok := tfMap[names.AttrDeviceName].(string); ok && v != "" {
		apiObject.DeviceName = aws.String(v)
	}

	if v, ok := tfMap["no_device"].(bool); ok && v {
		apiObject.NoDevice = aws.Bool(v)
	} else if v, ok := tfMap[names.AttrDeleteOnTermination].(bool); ok {
		apiObject.Ebs.DeleteOnTermination = aws.Bool(v)
	}

	if v, ok := tfMap[names.AttrEncrypted].(bool); ok && v {
		apiObject.Ebs.Encrypted = aws.Bool(v)
	}

	if v, ok := tfMap[names.AttrIOPS].(int); ok && v != 0 {
		apiObject.Ebs.Iops = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrSnapshotID].(string); ok && v != "" {
		apiObject.Ebs.SnapshotId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrThroughput].(int); ok && v != 0 {
		apiObject.Ebs.Throughput = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrVolumeSize].(int); ok && v != 0 {
		apiObject.Ebs.VolumeSize = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrVolumeType].(string); ok && v != "" {
		apiObject.Ebs.VolumeType = aws.String(v)
	}

	return apiObject
}

func expandBlockDeviceMappingForEphemeralBlockDevice(tfMap map[string]interface{}) awstypes.BlockDeviceMapping {
	apiObject := awstypes.BlockDeviceMapping{}

	if v, ok := tfMap[names.AttrDeviceName].(string); ok && v != "" {
		apiObject.DeviceName = aws.String(v)
	}

	if v, ok := tfMap["no_device"].(bool); ok && v {
		apiObject.NoDevice = aws.Bool(v)
	}

	if v, ok := tfMap[names.AttrVirtualName].(string); ok && v != "" {
		apiObject.VirtualName = aws.String(v)
	}

	return apiObject
}

func expandBlockDeviceMappingForRootBlockDevice(tfMap map[string]interface{}) awstypes.BlockDeviceMapping {
	apiObject := awstypes.BlockDeviceMapping{
		Ebs: &awstypes.Ebs{},
	}

	if v, ok := tfMap[names.AttrDeleteOnTermination].(bool); ok {
		apiObject.Ebs.DeleteOnTermination = aws.Bool(v)
	}

	if v, ok := tfMap[names.AttrEncrypted].(bool); ok && v {
		apiObject.Ebs.Encrypted = aws.Bool(v)
	}

	if v, ok := tfMap[names.AttrIOPS].(int); ok && v != 0 {
		apiObject.Ebs.Iops = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrThroughput].(int); ok && v != 0 {
		apiObject.Ebs.Throughput = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrVolumeSize].(int); ok && v != 0 {
		apiObject.Ebs.VolumeSize = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrVolumeType].(string); ok && v != "" {
		apiObject.Ebs.VolumeType = aws.String(v)
	}

	return apiObject
}

func expandBlockDeviceMappings(tfList []interface{}, fn func(map[string]interface{}) awstypes.BlockDeviceMapping) []awstypes.BlockDeviceMapping {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.BlockDeviceMapping

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := fn(tfMap)
		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenBlockDeviceMappings(apiObjects []awstypes.BlockDeviceMapping, rootDeviceName string, configuredEBSBlockDevices map[string]map[string]interface{}) ([]interface{}, []interface{}, []interface{}) {
	if len(apiObjects) == 0 {
		return nil, nil, nil
	}

	var tfListEBSBlockDevice, tfListEphemeralBlockDevice, tfListRootBlockDevice []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{}

		if v := apiObject.NoDevice; v != nil {
			if v, ok := configuredEBSBlockDevices[aws.ToString(apiObject.DeviceName)]; ok {
				tfMap[names.AttrDeleteOnTermination] = v[names.AttrDeleteOnTermination].(bool)
			} else {
				// Keep existing value in place to avoid spurious diff.
				tfMap[names.AttrDeleteOnTermination] = true
			}
		} else if v := apiObject.Ebs; v != nil {
			if v := v.DeleteOnTermination; v != nil {
				tfMap[names.AttrDeleteOnTermination] = aws.ToBool(v)
			}
		}

		if v := apiObject.Ebs; v != nil {
			if v := v.Encrypted; v != nil {
				tfMap[names.AttrEncrypted] = aws.ToBool(v)
			}

			if v := v.Iops; v != nil {
				tfMap[names.AttrIOPS] = aws.ToInt32(v)
			}

			if v := v.Throughput; v != nil {
				tfMap[names.AttrThroughput] = aws.ToInt32(v)
			}

			if v := v.VolumeSize; v != nil {
				tfMap[names.AttrVolumeSize] = aws.ToInt32(v)
			}

			if v := v.VolumeType; v != nil {
				tfMap[names.AttrVolumeType] = aws.ToString(v)
			}
		}

		if v := apiObject.DeviceName; aws.ToString(v) == rootDeviceName {
			tfListRootBlockDevice = append(tfListRootBlockDevice, tfMap)

			continue
		}

		if v := apiObject.DeviceName; v != nil {
			tfMap[names.AttrDeviceName] = aws.ToString(v)
		}

		if v := apiObject.VirtualName; v != nil {
			tfMap[names.AttrVirtualName] = aws.ToString(v)

			tfListEphemeralBlockDevice = append(tfListEphemeralBlockDevice, tfMap)

			continue
		}

		if v := apiObject.NoDevice; v != nil {
			tfMap["no_device"] = aws.ToBool(v)
		}

		if v := apiObject.Ebs; v != nil {
			if v := v.SnapshotId; v != nil {
				tfMap[names.AttrSnapshotID] = aws.ToString(v)
			}
		}

		tfListEBSBlockDevice = append(tfListEBSBlockDevice, tfMap)
	}

	return tfListEBSBlockDevice, tfListEphemeralBlockDevice, tfListRootBlockDevice
}

func expandInstanceMetadataOptions(tfMap map[string]interface{}) *awstypes.InstanceMetadataOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.InstanceMetadataOptions{}

	if v, ok := tfMap["http_endpoint"].(string); ok && v != "" {
		apiObject.HttpEndpoint = awstypes.InstanceMetadataEndpointState(v)

		if v := awstypes.InstanceMetadataEndpointState(v); v == awstypes.InstanceMetadataEndpointStateEnabled {
			if v, ok := tfMap["http_tokens"].(string); ok && v != "" {
				apiObject.HttpTokens = awstypes.InstanceMetadataHttpTokensState(v)
			}

			if v, ok := tfMap["http_put_response_hop_limit"].(int); ok && v != 0 {
				apiObject.HttpPutResponseHopLimit = aws.Int32(int32(v))
			}
		}
	}

	return apiObject
}

func flattenInstanceMetadataOptions(apiObject *awstypes.InstanceMetadataOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["http_endpoint"] = string(apiObject.HttpEndpoint)

	if v := apiObject.HttpPutResponseHopLimit; v != nil {
		tfMap["http_put_response_hop_limit"] = aws.ToInt32(v)
	}

	tfMap["http_tokens"] = string(apiObject.HttpTokens)

	return tfMap
}

func userDataHashSum(userData string) string {
	// Check whether the user_data is not Base64 encoded.
	// Always calculate hash of base64 decoded value since we
	// check against double-encoding when setting it.
	v, err := itypes.Base64Decode(userData)
	if err != nil {
		v = []byte(userData)
	}

	hash := sha1.Sum(v)
	return hex.EncodeToString(hash[:])
}

func findLaunchConfiguration(ctx context.Context, conn *autoscaling.Client, input *autoscaling.DescribeLaunchConfigurationsInput) (*awstypes.LaunchConfiguration, error) {
	output, err := findLaunchConfigurations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findLaunchConfigurations(ctx context.Context, conn *autoscaling.Client, input *autoscaling.DescribeLaunchConfigurationsInput) ([]awstypes.LaunchConfiguration, error) {
	var output []awstypes.LaunchConfiguration

	pages := autoscaling.NewDescribeLaunchConfigurationsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.LaunchConfigurations...)
	}

	return output, nil
}

func findLaunchConfigurationByName(ctx context.Context, conn *autoscaling.Client, name string) (*awstypes.LaunchConfiguration, error) {
	input := &autoscaling.DescribeLaunchConfigurationsInput{
		LaunchConfigurationNames: []string{name},
	}

	output, err := findLaunchConfiguration(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.LaunchConfigurationName) != name {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findImageRootDeviceName(ctx context.Context, conn *ec2.Client, imageID string) (string, error) {
	image, err := tfec2.FindImageByID(ctx, conn, imageID)

	if err != nil {
		return "", err
	}

	// Instance store backed AMIs do not provide a root device name.
	if image.RootDeviceType == ec2awstypes.DeviceTypeInstanceStore {
		return "", nil
	}

	rootDeviceName := aws.ToString(image.RootDeviceName)

	// Some AMIs have a RootDeviceName like "/dev/sda1" that does not appear as a
	// DeviceName in the BlockDeviceMapping list (which will instead have
	// something like "/dev/sda")
	//
	// While this seems like it breaks an invariant of AMIs, it ends up working
	// on the AWS side, and AMIs like this are common enough that we need to
	// special case it so Terraform does the right thing.
	//
	// Our heuristic is: if the RootDeviceName does not appear in the
	// BlockDeviceMapping, assume that the DeviceName of the first
	// BlockDeviceMapping entry serves as the root device.
	rootDeviceInBlockDeviceMappings := false

	for _, v := range image.BlockDeviceMappings {
		if aws.ToString(v.DeviceName) == rootDeviceName {
			rootDeviceInBlockDeviceMappings = true
		}
	}

	if !rootDeviceInBlockDeviceMappings && len(image.BlockDeviceMappings) > 0 {
		rootDeviceName = aws.ToString(image.BlockDeviceMappings[0].DeviceName)
	}

	if rootDeviceName == "" {
		return "", &retry.NotFoundError{
			Message: fmt.Sprintf("finding root device name for EC2 AMI (%s)", imageID),
		}
	}

	return rootDeviceName, nil
}
