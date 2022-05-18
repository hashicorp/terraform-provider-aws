package autoscaling

import ( // nosemgrep: aws-sdk-go-multiple-service-imports

	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"errors"
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
				ForceNew: true,
				Default:  false,
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
				Set:      schema.HashString,
			},
		},
	}
}

func resourceLaunchConfigurationCreate(d *schema.ResourceData, meta interface{}) error {
	autoscalingconn := meta.(*conns.AWSClient).AutoScalingConn
	ec2conn := meta.(*conns.AWSClient).EC2Conn

	lcName := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))

	createLaunchConfigurationOpts := autoscaling.CreateLaunchConfigurationInput{
		LaunchConfigurationName: aws.String(lcName),
		ImageId:                 aws.String(d.Get("image_id").(string)),
		InstanceType:            aws.String(d.Get("instance_type").(string)),
		EbsOptimized:            aws.Bool(d.Get("ebs_optimized").(bool)),
	}

	if v, ok := d.GetOk("user_data"); ok {
		userData := verify.Base64Encode([]byte(v.(string)))
		createLaunchConfigurationOpts.UserData = aws.String(userData)
	} else if v, ok := d.GetOk("user_data_base64"); ok {
		createLaunchConfigurationOpts.UserData = aws.String(v.(string))
	}

	createLaunchConfigurationOpts.InstanceMonitoring = &autoscaling.InstanceMonitoring{
		Enabled: aws.Bool(d.Get("enable_monitoring").(bool)),
	}

	if v, ok := d.GetOk("iam_instance_profile"); ok {
		createLaunchConfigurationOpts.IamInstanceProfile = aws.String(v.(string))
	}

	if v, ok := d.GetOk("placement_tenancy"); ok {
		createLaunchConfigurationOpts.PlacementTenancy = aws.String(v.(string))
	}

	if v, ok := d.GetOk("associate_public_ip_address"); ok {
		createLaunchConfigurationOpts.AssociatePublicIpAddress = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("key_name"); ok {
		createLaunchConfigurationOpts.KeyName = aws.String(v.(string))
	}
	if v, ok := d.GetOk("spot_price"); ok {
		createLaunchConfigurationOpts.SpotPrice = aws.String(v.(string))
	}

	if v, ok := d.GetOk("security_groups"); ok {
		createLaunchConfigurationOpts.SecurityGroups = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("vpc_classic_link_id"); ok {
		createLaunchConfigurationOpts.ClassicLinkVPCId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("metadata_options"); ok {
		createLaunchConfigurationOpts.MetadataOptions = expandLaunchConfigInstanceMetadataOptions(v.([]interface{}))
	}

	if v, ok := d.GetOk("vpc_classic_link_security_groups"); ok {
		createLaunchConfigurationOpts.ClassicLinkVPCSecurityGroups = flex.ExpandStringSet(v.(*schema.Set))
	}

	var blockDevices []*autoscaling.BlockDeviceMapping

	// We'll use this to detect if we're declaring it incorrectly as an ebs_block_device.
	rootDeviceName, err := fetchRootDeviceName(d.Get("image_id").(string), ec2conn)
	if err != nil {
		return err
	}
	if rootDeviceName == nil {
		// We do this so the value is empty so we don't have to do nil checks later
		var blank string
		rootDeviceName = &blank
	}

	if v, ok := d.GetOk("ebs_block_device"); ok {
		vL := v.(*schema.Set).List()
		for _, v := range vL {
			bd := v.(map[string]interface{})
			ebs := &autoscaling.Ebs{}

			var noDevice *bool
			if v, ok := bd["no_device"].(bool); ok && v {
				noDevice = aws.Bool(v)
			} else {
				ebs.DeleteOnTermination = aws.Bool(bd["delete_on_termination"].(bool))
			}

			if v, ok := bd["snapshot_id"].(string); ok && v != "" {
				ebs.SnapshotId = aws.String(v)
			}

			if v, ok := bd["encrypted"].(bool); ok && v {
				ebs.Encrypted = aws.Bool(v)
			}

			if v, ok := bd["volume_size"].(int); ok && v != 0 {
				ebs.VolumeSize = aws.Int64(int64(v))
			}

			if v, ok := bd["volume_type"].(string); ok && v != "" {
				ebs.VolumeType = aws.String(v)
			}

			if v, ok := bd["iops"].(int); ok && v > 0 {
				ebs.Iops = aws.Int64(int64(v))
			}

			if v, ok := bd["throughput"].(int); ok && v > 0 {
				ebs.Throughput = aws.Int64(int64(v))
			}

			if bd["device_name"].(string) == aws.StringValue(rootDeviceName) {
				return fmt.Errorf("Root device (%s) declared as an 'ebs_block_device'.  Use 'root_block_device' keyword.", *rootDeviceName)
			}

			blockDevices = append(blockDevices, &autoscaling.BlockDeviceMapping{
				DeviceName: aws.String(bd["device_name"].(string)),
				Ebs:        ebs,
				NoDevice:   noDevice,
			})
		}
	}

	if v, ok := d.GetOk("ephemeral_block_device"); ok {
		vL := v.(*schema.Set).List()
		for _, v := range vL {
			bd := v.(map[string]interface{})
			bdm := &autoscaling.BlockDeviceMapping{
				DeviceName: aws.String(bd["device_name"].(string)),
			}

			if v, ok := bd["no_device"].(bool); ok && v {
				bdm.NoDevice = aws.Bool(true)
			}

			if v, ok := bd["virtual_name"].(string); ok && v != "" {
				bdm.VirtualName = aws.String(v)
			}

			blockDevices = append(blockDevices, bdm)
		}
	}

	if v, ok := d.GetOk("root_block_device"); ok {
		vL := v.([]interface{})
		for _, v := range vL {
			bd := v.(map[string]interface{})
			ebs := &autoscaling.Ebs{
				DeleteOnTermination: aws.Bool(bd["delete_on_termination"].(bool)),
			}

			if v, ok := bd["encrypted"].(bool); ok && v {
				ebs.Encrypted = aws.Bool(v)
			}

			if v, ok := bd["volume_size"].(int); ok && v != 0 {
				ebs.VolumeSize = aws.Int64(int64(v))
			}

			if v, ok := bd["volume_type"].(string); ok && v != "" {
				ebs.VolumeType = aws.String(v)
			}

			if v, ok := bd["iops"].(int); ok && v > 0 {
				ebs.Iops = aws.Int64(int64(v))
			}

			if v, ok := bd["throughput"].(int); ok && v > 0 {
				ebs.Throughput = aws.Int64(int64(v))
			}

			if dn, err := fetchRootDeviceName(d.Get("image_id").(string), ec2conn); err == nil {
				if dn == nil {
					return fmt.Errorf(
						"Expected to find a Root Device name for AMI (%s), but got none",
						d.Get("image_id").(string))
				}
				blockDevices = append(blockDevices, &autoscaling.BlockDeviceMapping{
					DeviceName: dn,
					Ebs:        ebs,
				})
			} else {
				return err
			}
		}
	}

	if len(blockDevices) > 0 {
		createLaunchConfigurationOpts.BlockDeviceMappings = blockDevices
	}

	log.Printf("[DEBUG] autoscaling create launch configuration: %s", createLaunchConfigurationOpts)

	// IAM profiles can take ~10 seconds to propagate in AWS:
	// http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/iam-roles-for-amazon-ec2.html#launch-instance-with-role-console
	err = resource.Retry(propagationTimeout, func() *resource.RetryError {
		_, err := autoscalingconn.CreateLaunchConfiguration(&createLaunchConfigurationOpts)
		if err != nil {
			if tfawserr.ErrMessageContains(err, "ValidationError", "Invalid IamInstanceProfile") {
				return resource.RetryableError(err)
			}
			if tfawserr.ErrMessageContains(err, "ValidationError", "You are not authorized to perform this operation") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = autoscalingconn.CreateLaunchConfiguration(&createLaunchConfigurationOpts)
	}
	if err != nil {
		return fmt.Errorf("Error creating launch configuration: %w", err)
	}

	d.SetId(lcName)

	return resourceLaunchConfigurationRead(d, meta)
}

func resourceLaunchConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	autoscalingconn := meta.(*conns.AWSClient).AutoScalingConn
	ec2conn := meta.(*conns.AWSClient).EC2Conn

	lc, err := FindLaunchConfigurationByName(autoscalingconn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Autoscaling Launch Configuration %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Autoscaling Launch Configuration (%s): %w", d.Id(), err)
	}

	d.Set("key_name", lc.KeyName)
	d.Set("image_id", lc.ImageId)
	d.Set("instance_type", lc.InstanceType)
	d.Set("name", lc.LaunchConfigurationName)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(lc.LaunchConfigurationName)))
	d.Set("arn", lc.LaunchConfigurationARN)

	d.Set("iam_instance_profile", lc.IamInstanceProfile)
	d.Set("ebs_optimized", lc.EbsOptimized)
	d.Set("spot_price", lc.SpotPrice)
	d.Set("enable_monitoring", lc.InstanceMonitoring.Enabled)
	d.Set("associate_public_ip_address", lc.AssociatePublicIpAddress)
	if err := d.Set("security_groups", flex.FlattenStringList(lc.SecurityGroups)); err != nil {
		return fmt.Errorf("error setting security_groups: %w", err)
	}
	if v := aws.StringValue(lc.UserData); v != "" {
		_, b64 := d.GetOk("user_data_base64")
		if b64 {
			d.Set("user_data_base64", v)
		} else {
			d.Set("user_data", userDataHashSum(v))
		}
	}

	d.Set("vpc_classic_link_id", lc.ClassicLinkVPCId)
	if err := d.Set("vpc_classic_link_security_groups", flex.FlattenStringList(lc.ClassicLinkVPCSecurityGroups)); err != nil {
		return fmt.Errorf("error setting vpc_classic_link_security_groups: %w", err)
	}

	if err := d.Set("metadata_options", flattenLaunchConfigInstanceMetadataOptions(lc.MetadataOptions)); err != nil {
		return fmt.Errorf("error setting metadata_options: %w", err)
	}

	if err := readLCBlockDevices(d, lc, ec2conn); err != nil {
		return err
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

func readLCBlockDevices(d *schema.ResourceData, lc *autoscaling.LaunchConfiguration, ec2conn *ec2.EC2) error {
	ibds, err := readBlockDevicesFromLaunchConfiguration(d, lc, ec2conn)
	if err != nil {
		return err
	}

	if err := d.Set("ebs_block_device", ibds["ebs"]); err != nil {
		return err
	}
	if err := d.Set("ephemeral_block_device", ibds["ephemeral"]); err != nil {
		return err
	}
	if ibds["root"] != nil {
		if err := d.Set("root_block_device", []interface{}{ibds["root"]}); err != nil {
			return err
		}
	} else {
		d.Set("root_block_device", []interface{}{})
	}

	return nil
}

func expandLaunchConfigInstanceMetadataOptions(l []interface{}) *autoscaling.InstanceMetadataOptions {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	opts := &autoscaling.InstanceMetadataOptions{
		HttpEndpoint: aws.String(m["http_endpoint"].(string)),
	}

	if m["http_endpoint"].(string) == autoscaling.InstanceMetadataEndpointStateEnabled {
		// These parameters are not allowed unless HttpEndpoint is enabled

		if v, ok := m["http_tokens"].(string); ok && v != "" {
			opts.HttpTokens = aws.String(v)
		}

		if v, ok := m["http_put_response_hop_limit"].(int); ok && v != 0 {
			opts.HttpPutResponseHopLimit = aws.Int64(int64(v))
		}
	}

	return opts
}

func flattenLaunchConfigInstanceMetadataOptions(opts *autoscaling.InstanceMetadataOptions) []interface{} {
	if opts == nil {
		return nil
	}

	m := map[string]interface{}{
		"http_endpoint":               aws.StringValue(opts.HttpEndpoint),
		"http_put_response_hop_limit": aws.Int64Value(opts.HttpPutResponseHopLimit),
		"http_tokens":                 aws.StringValue(opts.HttpTokens),
	}

	return []interface{}{m}
}

func readBlockDevicesFromLaunchConfiguration(d *schema.ResourceData, lc *autoscaling.LaunchConfiguration, ec2conn *ec2.EC2) (
	map[string]interface{}, error) {
	blockDevices := make(map[string]interface{})
	blockDevices["ebs"] = make([]map[string]interface{}, 0)
	blockDevices["ephemeral"] = make([]map[string]interface{}, 0)
	blockDevices["root"] = nil
	if len(lc.BlockDeviceMappings) == 0 {
		return nil, nil
	}
	v, err := fetchRootDeviceName(d.Get("image_id").(string), ec2conn)
	if err != nil {
		return nil, err
	}
	rootDeviceName := aws.StringValue(v)

	// Collect existing configured devices, so we can check
	// existing value of delete_on_termination below
	existingEbsBlockDevices := make(map[string]map[string]interface{})
	if v, ok := d.GetOk("ebs_block_device"); ok {
		ebsBlocks := v.(*schema.Set)
		for _, ebd := range ebsBlocks.List() {
			m := ebd.(map[string]interface{})
			deviceName := m["device_name"].(string)
			existingEbsBlockDevices[deviceName] = m
		}
	}

	for _, bdm := range lc.BlockDeviceMappings {
		bd := make(map[string]interface{})

		if bdm.NoDevice != nil {
			// Keep existing value in place to avoid spurious diff
			deleteOnTermination := true
			if device, ok := existingEbsBlockDevices[*bdm.DeviceName]; ok {
				deleteOnTermination = device["delete_on_termination"].(bool)
			}
			bd["delete_on_termination"] = deleteOnTermination
		} else if bdm.Ebs != nil && bdm.Ebs.DeleteOnTermination != nil {
			bd["delete_on_termination"] = aws.BoolValue(bdm.Ebs.DeleteOnTermination)
		}

		if bdm.Ebs != nil && bdm.Ebs.VolumeSize != nil {
			bd["volume_size"] = aws.Int64Value(bdm.Ebs.VolumeSize)
		}
		if bdm.Ebs != nil && bdm.Ebs.VolumeType != nil {
			bd["volume_type"] = aws.StringValue(bdm.Ebs.VolumeType)
		}
		if bdm.Ebs != nil && bdm.Ebs.Iops != nil {
			bd["iops"] = aws.Int64Value(bdm.Ebs.Iops)
		}
		if bdm.Ebs != nil && bdm.Ebs.Throughput != nil {
			bd["throughput"] = aws.Int64Value(bdm.Ebs.Throughput)
		}
		if bdm.Ebs != nil && bdm.Ebs.Encrypted != nil {
			bd["encrypted"] = aws.BoolValue(bdm.Ebs.Encrypted)
		}

		if bdm.DeviceName != nil && aws.StringValue(bdm.DeviceName) == rootDeviceName {
			blockDevices["root"] = bd
		} else {
			if bdm.DeviceName != nil {
				bd["device_name"] = aws.StringValue(bdm.DeviceName)
			}

			if bdm.VirtualName != nil {
				bd["virtual_name"] = aws.StringValue(bdm.VirtualName)
				blockDevices["ephemeral"] = append(blockDevices["ephemeral"].([]map[string]interface{}), bd)
			} else {
				if bdm.Ebs != nil && bdm.Ebs.SnapshotId != nil {
					bd["snapshot_id"] = aws.StringValue(bdm.Ebs.SnapshotId)
				}
				if bdm.NoDevice != nil {
					bd["no_device"] = aws.BoolValue(bdm.NoDevice)
				}
				blockDevices["ebs"] = append(blockDevices["ebs"].([]map[string]interface{}), bd)
			}
		}
	}
	return blockDevices, nil
}

func userDataHashSum(user_data string) string {
	// Check whether the user_data is not Base64 encoded.
	// Always calculate hash of base64 decoded value since we
	// check against double-encoding when setting it
	v, base64DecodeError := base64.StdEncoding.DecodeString(user_data)
	if base64DecodeError != nil {
		v = []byte(user_data)
	}

	hash := sha1.Sum(v)
	return hex.EncodeToString(hash[:])
}

func fetchRootDeviceName(ami string, conn *ec2.EC2) (*string, error) {
	if ami == "" {
		return nil, errors.New("Cannot fetch root device name for blank AMI ID.")
	}

	log.Printf("[DEBUG] Describing AMI %q to get root block device name", ami)
	res, err := conn.DescribeImages(&ec2.DescribeImagesInput{
		ImageIds: []*string{aws.String(ami)},
	})
	if err != nil {
		return nil, err
	}

	// For a bad image, we just return nil so we don't block a refresh
	if len(res.Images) == 0 {
		return nil, nil
	}

	image := res.Images[0]
	rootDeviceName := image.RootDeviceName

	// Instance store backed AMIs do not provide a root device name.
	if aws.StringValue(image.RootDeviceType) == ec2.DeviceTypeInstanceStore {
		return nil, nil
	}

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
	rootDeviceNameInMapping := false
	for _, bdm := range image.BlockDeviceMappings {
		if aws.StringValue(bdm.DeviceName) == aws.StringValue(image.RootDeviceName) {
			rootDeviceNameInMapping = true
		}
	}

	if !rootDeviceNameInMapping && len(image.BlockDeviceMappings) > 0 {
		rootDeviceName = image.BlockDeviceMappings[0].DeviceName
	}

	if rootDeviceName == nil {
		return nil, fmt.Errorf("Error finding Root Device Name for AMI (%s)", ami)
	}

	return rootDeviceName, nil
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
