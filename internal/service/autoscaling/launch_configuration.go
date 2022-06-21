package autoscaling

import ( // nosemgrep: aws-sdk-go-multiple-service-imports

	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceLaunchConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceLaunchConfigurationCreate,
		Read:   resourceLaunchConfigurationRead,
		Delete: resourceLaunchConfigurationDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
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
						"delete_on_termination": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
							ForceNew: true,
						},
						"device_name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"encrypted": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						"iops": {
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
						"snapshot_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						"throughput": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						"volume_size": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						"volume_type": {
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
						"device_name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"no_device": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
						},
						"virtual_name": {
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
			"instance_type": {
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
							ValidateFunc: validation.StringInSlice([]string{autoscaling.InstanceMetadataEndpointStateEnabled, autoscaling.InstanceMetadataEndpointStateDisabled}, false),
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
							ValidateFunc: validation.StringInSlice([]string{autoscaling.InstanceMetadataHttpTokensStateOptional, autoscaling.InstanceMetadataHttpTokensStateRequired}, false),
						},
					},
				},
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validation.StringLenBetween(1, 255),
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validation.StringLenBetween(1, 255-resource.UniqueIDSuffixLength),
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
						"delete_on_termination": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
							ForceNew: true,
						},
						"encrypted": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						"iops": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						"throughput": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						"volume_size": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						"volume_type": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
					},
				},
			},
			"security_groups": {
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
				ValidateFunc: func(v interface{}, name string) (warns []string, errs []error) {
					s := v.(string)
					if !verify.IsBase64Encoded([]byte(s)) {
						errs = append(errs, fmt.Errorf(
							"%s: must be base64-encoded", name,
						))
					}
					return
				},
			},
			"vpc_classic_link_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"vpc_classic_link_security_groups": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceLaunchConfigurationCreate(d *schema.ResourceData, meta interface{}) error {
	autoscalingconn := meta.(*conns.AWSClient).AutoScalingConn
	ec2conn := meta.(*conns.AWSClient).EC2Conn

	lcName := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	input := autoscaling.CreateLaunchConfigurationInput{
		EbsOptimized:            aws.Bool(d.Get("ebs_optimized").(bool)),
		ImageId:                 aws.String(d.Get("image_id").(string)),
		InstanceType:            aws.String(d.Get("instance_type").(string)),
		LaunchConfigurationName: aws.String(lcName),
	}

	associatePublicIPAddress := d.GetRawConfig().GetAttr("associate_public_ip_address")
	if associatePublicIPAddress.IsKnown() && !associatePublicIPAddress.IsNull() {
		input.AssociatePublicIpAddress = aws.Bool(associatePublicIPAddress.True())
	}

	if v, ok := d.GetOk("vpc_classic_link_id"); ok {
		input.ClassicLinkVPCId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("vpc_classic_link_security_groups"); ok && v.(*schema.Set).Len() > 0 {
		input.ClassicLinkVPCSecurityGroups = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("iam_instance_profile"); ok {
		input.IamInstanceProfile = aws.String(v.(string))
	}

	input.InstanceMonitoring = &autoscaling.InstanceMonitoring{
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

	if v, ok := d.GetOk("security_groups"); ok && v.(*schema.Set).Len() > 0 {
		input.SecurityGroups = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("spot_price"); ok {
		input.SpotPrice = aws.String(v.(string))
	}

	if v, ok := d.GetOk("user_data"); ok {
		input.UserData = aws.String(verify.Base64Encode([]byte(v.(string))))
	} else if v, ok := d.GetOk("user_data_base64"); ok {
		input.UserData = aws.String(v.(string))
	}

	// We'll use this to detect if we're declaring it incorrectly as an ebs_block_device.
	rootDeviceName, err := findImageRootDeviceName(ec2conn, d.Get("image_id").(string))

	if err != nil {
		return err
	}

	var blockDeviceMappings []*autoscaling.BlockDeviceMapping

	if v, ok := d.GetOk("ebs_block_device"); ok && v.(*schema.Set).Len() > 0 {
		v := expandBlockDeviceMappings(v.(*schema.Set).List(), expandBlockDeviceMappingForEBSBlockDevice)

		for _, v := range v {
			if aws.StringValue(v.DeviceName) == rootDeviceName {
				return fmt.Errorf("root device (%s) declared as an 'ebs_block_device'. Use 'root_block_device' argument.", rootDeviceName)
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

	log.Printf("[DEBUG] Creating Auto Scaling Launch Configuration: %s", input)
	// IAM profiles can take ~10 seconds to propagate in AWS:
	// http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/iam-roles-for-amazon-ec2.html#launch-instance-with-role-console
	_, err = tfresource.RetryWhen(propagationTimeout,
		func() (interface{}, error) {
			return autoscalingconn.CreateLaunchConfiguration(&input)
		},
		func(err error) (bool, error) {
			if tfawserr.ErrMessageContains(err, ErrCodeValidationError, "Invalid IamInstanceProfile") ||
				tfawserr.ErrMessageContains(err, ErrCodeValidationError, "You are not authorized to perform this operation") {
				return true, err
			}

			return false, err
		})

	if err != nil {
		return fmt.Errorf("creating Auto Scaling Launch Configuration (%s): %w", lcName, err)
	}

	d.SetId(lcName)

	return resourceLaunchConfigurationRead(d, meta)
}

func resourceLaunchConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	autoscalingconn := meta.(*conns.AWSClient).AutoScalingConn
	ec2conn := meta.(*conns.AWSClient).EC2Conn

	lc, err := FindLaunchConfigurationByName(autoscalingconn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Auto Scaling Launch Configuration %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading Auto Scaling Launch Configuration (%s): %w", d.Id(), err)
	}

	d.Set("arn", lc.LaunchConfigurationARN)
	d.Set("associate_public_ip_address", lc.AssociatePublicIpAddress)
	d.Set("ebs_optimized", lc.EbsOptimized)
	if lc.InstanceMonitoring != nil {
		d.Set("enable_monitoring", lc.InstanceMonitoring.Enabled)
	} else {
		d.Set("enable_monitoring", false)
	}
	d.Set("iam_instance_profile", lc.IamInstanceProfile)
	d.Set("image_id", lc.ImageId)
	d.Set("instance_type", lc.InstanceType)
	d.Set("key_name", lc.KeyName)
	if lc.MetadataOptions != nil {
		if err := d.Set("metadata_options", []interface{}{flattenInstanceMetadataOptions(lc.MetadataOptions)}); err != nil {
			return fmt.Errorf("setting metadata_options: %w", err)
		}
	} else {
		d.Set("metadata_options", nil)
	}
	d.Set("name", lc.LaunchConfigurationName)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(lc.LaunchConfigurationName)))
	d.Set("placement_tenancy", lc.PlacementTenancy)
	d.Set("security_groups", aws.StringValueSlice(lc.SecurityGroups))
	d.Set("spot_price", lc.SpotPrice)
	if v := aws.StringValue(lc.UserData); v != "" {
		if _, ok := d.GetOk("user_data_base64"); ok {
			d.Set("user_data_base64", v)
		} else {
			d.Set("user_data", userDataHashSum(v))
		}
	}
	d.Set("vpc_classic_link_id", lc.ClassicLinkVPCId)
	d.Set("vpc_classic_link_security_groups", aws.StringValueSlice(lc.ClassicLinkVPCSecurityGroups))

	rootDeviceName, err := findImageRootDeviceName(ec2conn, d.Get("image_id").(string))

	if tfresource.NotFound(err) {
		// Don't block a refresh for a bad image.
		rootDeviceName = ""
	} else if err != nil {
		return err
	}

	configuredEBSBlockDevices := make(map[string]map[string]interface{})

	if v, ok := d.GetOk("ebs_block_device"); ok && v.(*schema.Set).Len() > 0 {
		for _, v := range v.(*schema.Set).List() {
			tfMap, ok := v.(map[string]interface{})

			if !ok {
				continue
			}

			configuredEBSBlockDevices[tfMap["device_name"].(string)] = tfMap
		}
	}

	tfListEBSBlockDevice, tfListEphemeralBlockDevice, tfListRootBlockDevice := flattenBlockDeviceMappings(lc.BlockDeviceMappings, rootDeviceName, configuredEBSBlockDevices)

	if err := d.Set("ebs_block_device", tfListEBSBlockDevice); err != nil {
		return fmt.Errorf("setting ebs_block_device: %w", err)
	}
	if err := d.Set("ephemeral_block_device", tfListEphemeralBlockDevice); err != nil {
		return fmt.Errorf("setting ephemeral_block_device: %w", err)
	}
	if err := d.Set("root_block_device", tfListRootBlockDevice); err != nil {
		return fmt.Errorf("setting root_block_device: %w", err)
	}

	return nil
}

func resourceLaunchConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AutoScalingConn

	log.Printf("[DEBUG] Deleting Auto Scaling Launch Configuration: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrCodeEquals(propagationTimeout,
		func() (interface{}, error) {
			return conn.DeleteLaunchConfiguration(&autoscaling.DeleteLaunchConfigurationInput{
				LaunchConfigurationName: aws.String(d.Id()),
			})
		},
		autoscaling.ErrCodeResourceInUseFault)

	if tfawserr.ErrMessageContains(err, ErrCodeValidationError, "not found") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting Auto Scaling Launch Configuration (%s): %w", d.Id(), err)
	}

	return nil
}

func expandBlockDeviceMappingForEBSBlockDevice(tfMap map[string]interface{}) *autoscaling.BlockDeviceMapping {
	if tfMap == nil {
		return nil
	}

	apiObject := &autoscaling.BlockDeviceMapping{
		Ebs: &autoscaling.Ebs{},
	}

	if v, ok := tfMap["device_name"].(string); ok && v != "" {
		apiObject.DeviceName = aws.String(v)
	}

	if v, ok := tfMap["no_device"].(bool); ok && v {
		apiObject.NoDevice = aws.Bool(v)
	} else if v, ok := tfMap["delete_on_termination"].(bool); ok {
		apiObject.Ebs.DeleteOnTermination = aws.Bool(v)
	}

	if v, ok := tfMap["encrypted"].(bool); ok && v {
		apiObject.Ebs.Encrypted = aws.Bool(v)
	}

	if v, ok := tfMap["iops"].(int); ok && v != 0 {
		apiObject.Ebs.Iops = aws.Int64(int64(v))
	}

	if v, ok := tfMap["snapshot_id"].(string); ok && v != "" {
		apiObject.Ebs.SnapshotId = aws.String(v)
	}

	if v, ok := tfMap["throughput"].(int); ok && v != 0 {
		apiObject.Ebs.Throughput = aws.Int64(int64(v))
	}

	if v, ok := tfMap["volume_size"].(int); ok && v != 0 {
		apiObject.Ebs.VolumeSize = aws.Int64(int64(v))
	}

	if v, ok := tfMap["volume_type"].(string); ok && v != "" {
		apiObject.Ebs.VolumeType = aws.String(v)
	}

	return apiObject
}

func expandBlockDeviceMappingForEphemeralBlockDevice(tfMap map[string]interface{}) *autoscaling.BlockDeviceMapping {
	if tfMap == nil {
		return nil
	}

	apiObject := &autoscaling.BlockDeviceMapping{}

	if v, ok := tfMap["device_name"].(string); ok && v != "" {
		apiObject.DeviceName = aws.String(v)
	}

	if v, ok := tfMap["no_device"].(bool); ok && v {
		apiObject.NoDevice = aws.Bool(v)
	}

	if v, ok := tfMap["virtual_name"].(string); ok && v != "" {
		apiObject.VirtualName = aws.String(v)
	}

	return apiObject
}

func expandBlockDeviceMappingForRootBlockDevice(tfMap map[string]interface{}) *autoscaling.BlockDeviceMapping {
	if tfMap == nil {
		return nil
	}

	apiObject := &autoscaling.BlockDeviceMapping{
		Ebs: &autoscaling.Ebs{},
	}

	if v, ok := tfMap["delete_on_termination"].(bool); ok {
		apiObject.Ebs.DeleteOnTermination = aws.Bool(v)
	}

	if v, ok := tfMap["encrypted"].(bool); ok && v {
		apiObject.Ebs.Encrypted = aws.Bool(v)
	}

	if v, ok := tfMap["iops"].(int); ok && v != 0 {
		apiObject.Ebs.Iops = aws.Int64(int64(v))
	}

	if v, ok := tfMap["throughput"].(int); ok && v != 0 {
		apiObject.Ebs.Throughput = aws.Int64(int64(v))
	}

	if v, ok := tfMap["volume_size"].(int); ok && v != 0 {
		apiObject.Ebs.VolumeSize = aws.Int64(int64(v))
	}

	if v, ok := tfMap["volume_type"].(string); ok && v != "" {
		apiObject.Ebs.VolumeType = aws.String(v)
	}

	return apiObject
}

func expandBlockDeviceMappings(tfList []interface{}, fn func(map[string]interface{}) *autoscaling.BlockDeviceMapping) []*autoscaling.BlockDeviceMapping {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*autoscaling.BlockDeviceMapping

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := fn(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenBlockDeviceMappings(apiObjects []*autoscaling.BlockDeviceMapping, rootDeviceName string, configuredEBSBlockDevices map[string]map[string]interface{}) ([]interface{}, []interface{}, []interface{}) {
	if len(apiObjects) == 0 {
		return nil, nil, nil
	}

	var tfListEBSBlockDevice, tfListEphemeralBlockDevice, tfListRootBlockDevice []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfMap := map[string]interface{}{}

		if v := apiObject.NoDevice; v != nil {
			if v, ok := configuredEBSBlockDevices[aws.StringValue(apiObject.DeviceName)]; ok {
				tfMap["delete_on_termination"] = v["delete_on_termination"].(bool)
			} else {
				// Keep existing value in place to avoid spurious diff.
				tfMap["delete_on_termination"] = true
			}
		} else if v := apiObject.Ebs; v != nil {
			if v := v.DeleteOnTermination; v != nil {
				tfMap["delete_on_termination"] = aws.BoolValue(v)
			}
		}

		if v := apiObject.Ebs; v != nil {
			if v := v.Encrypted; v != nil {
				tfMap["encrypted"] = aws.BoolValue(v)
			}

			if v := v.Iops; v != nil {
				tfMap["iops"] = aws.Int64Value(v)
			}

			if v := v.Throughput; v != nil {
				tfMap["throughput"] = aws.Int64Value(v)
			}

			if v := v.VolumeSize; v != nil {
				tfMap["volume_size"] = aws.Int64Value(v)
			}

			if v := v.VolumeType; v != nil {
				tfMap["volume_type"] = aws.StringValue(v)
			}
		}

		if v := apiObject.DeviceName; aws.StringValue(v) == rootDeviceName {
			tfListRootBlockDevice = append(tfListRootBlockDevice, tfMap)

			continue
		}

		if v := apiObject.DeviceName; v != nil {
			tfMap["device_name"] = aws.StringValue(v)
		}

		if v := apiObject.VirtualName; v != nil {
			tfMap["virtual_name"] = aws.StringValue(v)

			tfListEphemeralBlockDevice = append(tfListEphemeralBlockDevice, tfMap)

			continue
		}

		if v := apiObject.NoDevice; v != nil {
			tfMap["no_device"] = aws.BoolValue(v)
		}

		if v := apiObject.Ebs; v != nil {
			if v := v.SnapshotId; v != nil {
				tfMap["snapshot_id"] = aws.StringValue(v)
			}
		}

		tfListEBSBlockDevice = append(tfListEBSBlockDevice, tfMap)
	}

	return tfListEBSBlockDevice, tfListEphemeralBlockDevice, tfListRootBlockDevice
}

func expandInstanceMetadataOptions(tfMap map[string]interface{}) *autoscaling.InstanceMetadataOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &autoscaling.InstanceMetadataOptions{}

	if v, ok := tfMap["http_endpoint"].(string); ok && v != "" {
		apiObject.HttpEndpoint = aws.String(v)

		if v == autoscaling.InstanceMetadataEndpointStateEnabled {
			if v, ok := tfMap["http_tokens"].(string); ok && v != "" {
				apiObject.HttpTokens = aws.String(v)
			}

			if v, ok := tfMap["http_put_response_hop_limit"].(int); ok && v != 0 {
				apiObject.HttpPutResponseHopLimit = aws.Int64(int64(v))
			}
		}
	}

	return apiObject
}

func flattenInstanceMetadataOptions(apiObject *autoscaling.InstanceMetadataOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.HttpEndpoint; v != nil {
		tfMap["http_endpoint"] = aws.StringValue(v)
	}

	if v := apiObject.HttpPutResponseHopLimit; v != nil {
		tfMap["http_put_response_hop_limit"] = aws.Int64Value(v)
	}

	if v := apiObject.HttpTokens; v != nil {
		tfMap["http_tokens"] = aws.StringValue(v)
	}

	return tfMap
}

func userDataHashSum(userData string) string {
	// Check whether the user_data is not Base64 encoded.
	// Always calculate hash of base64 decoded value since we
	// check against double-encoding when setting it.
	v, err := base64.StdEncoding.DecodeString(userData)

	if err != nil {
		v = []byte(userData)
	}

	hash := sha1.Sum(v)
	return hex.EncodeToString(hash[:])
}

func findLaunchConfiguration(conn *autoscaling.AutoScaling, input *autoscaling.DescribeLaunchConfigurationsInput) (*autoscaling.LaunchConfiguration, error) {
	output, err := findLaunchConfigurations(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func findLaunchConfigurations(conn *autoscaling.AutoScaling, input *autoscaling.DescribeLaunchConfigurationsInput) ([]*autoscaling.LaunchConfiguration, error) {
	var output []*autoscaling.LaunchConfiguration

	err := conn.DescribeLaunchConfigurationsPages(input, func(page *autoscaling.DescribeLaunchConfigurationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.LaunchConfigurations {
			if v == nil {
				continue
			}

			output = append(output, v)
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindLaunchConfigurationByName(conn *autoscaling.AutoScaling, name string) (*autoscaling.LaunchConfiguration, error) {
	input := &autoscaling.DescribeLaunchConfigurationsInput{
		LaunchConfigurationNames: aws.StringSlice([]string{name}),
	}

	output, err := findLaunchConfiguration(conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.LaunchConfigurationName) != name {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findImageRootDeviceName(conn *ec2.EC2, imageID string) (string, error) {
	image, err := tfec2.FindImageByID(conn, imageID)

	if err != nil {
		return "", err
	}

	// Instance store backed AMIs do not provide a root device name.
	if aws.StringValue(image.RootDeviceType) == ec2.DeviceTypeInstanceStore {
		return "", nil
	}

	rootDeviceName := aws.StringValue(image.RootDeviceName)

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
		if aws.StringValue(v.DeviceName) == rootDeviceName {
			rootDeviceInBlockDeviceMappings = true
		}
	}

	if !rootDeviceInBlockDeviceMappings && len(image.BlockDeviceMappings) > 0 {
		rootDeviceName = aws.StringValue(image.BlockDeviceMappings[0].DeviceName)
	}

	if rootDeviceName == "" {
		return "", &resource.NotFoundError{
			Message: fmt.Sprintf("finding root device name for EC2 AMI (%s)", imageID),
		}
	}

	return rootDeviceName, nil
}
