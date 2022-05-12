package ec2

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	AWSAMIRetryTimeout    = 40 * time.Minute
	AMIDeleteRetryTimeout = 90 * time.Minute
	AWSAMIRetryDelay      = 5 * time.Second
	AMIRetryMinTimeout    = 3 * time.Second
)

func ResourceAMI() *schema.Resource {
	return &schema.Resource{
		Create: resourceAMICreate,
		// The Read, Update and Delete operations are shared with aws_ami_copy
		// and aws_ami_from_instance, since they differ only in how the image
		// is created.
		Read:   resourceAMIRead,
		Update: resourceAMIUpdate,
		Delete: resourceAMIDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(AWSAMIRetryTimeout),
			Update: schema.DefaultTimeout(AWSAMIRetryTimeout),
			Delete: schema.DefaultTimeout(AMIDeleteRetryTimeout),
		},

		Schema: map[string]*schema.Schema{
			"architecture": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      ec2.ArchitectureValuesX8664,
				ValidateFunc: validation.StringInSlice(ec2.ArchitectureValues_Values(), false),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"boot_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ec2.BootModeValues_Values(), false),
			},
			"deprecation_time": {
				Type:                  schema.TypeString,
				Optional:              true,
				ValidateFunc:          validation.IsRFC3339Time,
				DiffSuppressFunc:      verify.SuppressEquivalentRoundedTime(time.RFC3339, time.Minute),
				DiffSuppressOnRefresh: true,
			},
			"description": {
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
							ForceNew: true,
						},
						"iops": {
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
						"snapshot_id": {
							Type:     schema.TypeString,
							Optional: true,
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
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							Default:      ec2.VolumeTypeStandard,
							ValidateFunc: validation.StringInSlice(ec2.VolumeType_Values(), false),
						},
					},
				},
				Set: func(v interface{}) int {
					var buf bytes.Buffer
					m := v.(map[string]interface{})
					buf.WriteString(fmt.Sprintf("%s-", m["device_name"].(string)))
					buf.WriteString(fmt.Sprintf("%s-", m["snapshot_id"].(string)))
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
						"device_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"virtual_name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
				Set: func(v interface{}) int {
					var buf bytes.Buffer
					m := v.(map[string]interface{})
					buf.WriteString(fmt.Sprintf("%s-", m["device_name"].(string)))
					buf.WriteString(fmt.Sprintf("%s-", m["virtual_name"].(string)))
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
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"owner_id": {
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
				Default:  "simple",
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"usage_operation": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"virtualization_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      ec2.VirtualizationTypeParavirtual,
				ValidateFunc: validation.StringInSlice(ec2.VirtualizationType_Values(), false),
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAMICreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &ec2.RegisterImageInput{
		Architecture:       aws.String(d.Get("architecture").(string)),
		Description:        aws.String(d.Get("description").(string)),
		EnaSupport:         aws.Bool(d.Get("ena_support").(bool)),
		ImageLocation:      aws.String(d.Get("image_location").(string)),
		Name:               aws.String(name),
		RootDeviceName:     aws.String(d.Get("root_device_name").(string)),
		SriovNetSupport:    aws.String(d.Get("sriov_net_support").(string)),
		VirtualizationType: aws.String(d.Get("virtualization_type").(string)),
	}

	if v := d.Get("boot_mode").(string); v != "" {
		input.BootMode = aws.String(v)
	}

	if kernelId := d.Get("kernel_id").(string); kernelId != "" {
		input.KernelId = aws.String(kernelId)
	}

	if ramdiskId := d.Get("ramdisk_id").(string); ramdiskId != "" {
		input.RamdiskId = aws.String(ramdiskId)
	}

	if v, ok := d.GetOk("ebs_block_device"); ok && v.(*schema.Set).Len() > 0 {
		for _, tfMapRaw := range v.(*schema.Set).List() {
			tfMap, ok := tfMapRaw.(map[string]interface{})

			if !ok {
				continue
			}

			var encrypted bool

			if v, ok := tfMap["encrypted"].(bool); ok {
				encrypted = v
			}

			var snapshot string

			if v, ok := tfMap["snapshot_id"].(string); ok && v != "" {
				snapshot = v
			}

			if snapshot != "" && encrypted {
				return errors.New("can't set both 'snapshot_id' and 'encrypted'")
			}
		}

		input.BlockDeviceMappings = expandEc2BlockDeviceMappingsForAmiEbsBlockDevice(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("ephemeral_block_device"); ok && v.(*schema.Set).Len() > 0 {
		input.BlockDeviceMappings = append(input.BlockDeviceMappings, expandEc2BlockDeviceMappingsForAmiEphemeralBlockDevice(v.(*schema.Set).List())...)
	}

	output, err := conn.RegisterImage(input)

	if err != nil {
		return fmt.Errorf("error creating EC2 AMI (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(output.ImageId))

	if len(tags) > 0 {
		if err := CreateTags(conn, d.Id(), tags); err != nil {
			return fmt.Errorf("error adding tags: %s", err)
		}
	}

	if _, err := WaitImageAvailable(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for EC2 AMI (%s) create: %w", d.Id(), err)
	}

	if v, ok := d.GetOk("deprecation_time"); ok {
		if err := enableImageDeprecation(conn, d.Id(), v.(string)); err != nil {
			return err
		}
	}

	return resourceAMIRead(d, meta)
}

func resourceAMIRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(propagationTimeout, func() (interface{}, error) {
		return FindImageByID(conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 AMI %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 AMI (%s): %w", d.Id(), err)
	}

	image := outputRaw.(*ec2.Image)

	if aws.StringValue(image.State) == ec2.ImageStatePending {
		// This could happen if a user manually adds an image we didn't create
		// to the state. We'll wait for the image to become available
		// before we continue. We should never take this branch in normal
		// circumstances since we would've waited for availability during
		// the "Create" step.
		image, err = WaitImageAvailable(conn, d.Id(), d.Timeout(schema.TimeoutCreate))

		if err != nil {
			return fmt.Errorf("error waiting for EC2 AMI (%s) create: %w", d.Id(), err)
		}
	}

	d.Set("architecture", image.Architecture)
	imageArn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("image/%s", d.Id()),
		Service:   ec2.ServiceName,
	}.String()
	d.Set("arn", imageArn)
	d.Set("boot_mode", image.BootMode)
	d.Set("description", image.Description)
	d.Set("deprecation_time", image.DeprecationTime)
	d.Set("ena_support", image.EnaSupport)
	d.Set("hypervisor", image.Hypervisor)
	d.Set("image_location", image.ImageLocation)
	d.Set("image_owner_alias", image.ImageOwnerAlias)
	d.Set("image_type", image.ImageType)
	d.Set("kernel_id", image.KernelId)
	d.Set("name", image.Name)
	d.Set("owner_id", image.OwnerId)
	d.Set("platform_details", image.PlatformDetails)
	d.Set("platform", image.Platform)
	d.Set("public", image.Public)
	d.Set("ramdisk_id", image.RamdiskId)
	d.Set("root_device_name", image.RootDeviceName)
	d.Set("root_snapshot_id", amiRootSnapshotId(image))
	d.Set("sriov_net_support", image.SriovNetSupport)
	d.Set("usage_operation", image.UsageOperation)
	d.Set("virtualization_type", image.VirtualizationType)

	if err := d.Set("ebs_block_device", flattenEc2BlockDeviceMappingsForAmiEbsBlockDevice(image.BlockDeviceMappings)); err != nil {
		return fmt.Errorf("error setting ebs_block_device: %w", err)
	}

	if err := d.Set("ephemeral_block_device", flattenEc2BlockDeviceMappingsForAmiEphemeralBlockDevice(image.BlockDeviceMappings)); err != nil {
		return fmt.Errorf("error setting ephemeral_block_device: %w", err)
	}

	tags := KeyValueTags(image.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAMIUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating AMI (%s) tags: %s", d.Id(), err)
		}
	}

	if d.Get("description").(string) != "" {
		_, err := conn.ModifyImageAttribute(&ec2.ModifyImageAttributeInput{
			Description: &ec2.AttributeValue{
				Value: aws.String(d.Get("description").(string)),
			},
			ImageId: aws.String(d.Id()),
		})

		if err != nil {
			return fmt.Errorf("error updating EC2 AMI (%s) description: %w", d.Id(), err)
		}
	}

	if d.HasChange("deprecation_time") {
		if err := enableImageDeprecation(conn, d.Id(), d.Get("deprecation_time").(string)); err != nil {
			return err
		}
	}

	return resourceAMIRead(d, meta)
}

func resourceAMIDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[INFO] Deleting EC2 AMI: %s", d.Id())
	_, err := conn.DeregisterImage(&ec2.DeregisterImageInput{
		ImageId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidAMIIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deregistering EC2 AMI (%s): %w", d.Id(), err)
	}

	// If we're managing the EBS snapshots then we need to delete those too.
	if d.Get("manage_ebs_snapshots").(bool) {
		errs := map[string]error{}
		ebsBlockDevsSet := d.Get("ebs_block_device").(*schema.Set)
		req := &ec2.DeleteSnapshotInput{}
		for _, ebsBlockDevI := range ebsBlockDevsSet.List() {
			ebsBlockDev := ebsBlockDevI.(map[string]interface{})
			snapshotId := ebsBlockDev["snapshot_id"].(string)
			if snapshotId != "" {
				req.SnapshotId = aws.String(snapshotId)
				_, err := conn.DeleteSnapshot(req)
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
			return errors.New(strings.Join(errParts, "\n"))
		}
	}

	if _, err := WaitImageDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for EC2 AMI (%s) delete: %w", d.Id(), err)
	}

	return nil
}

func enableImageDeprecation(conn *ec2.EC2, id string, deprecateAt string) error {
	v, _ := time.Parse(time.RFC3339, deprecateAt)
	input := &ec2.EnableImageDeprecationInput{
		DeprecateAt: aws.Time(v),
		ImageId:     aws.String(id),
	}

	_, err := conn.EnableImageDeprecation(input)

	if err != nil {
		return fmt.Errorf("error enabling EC2 AMI (%s) image deprecation: %w", id, err)
	}

	return nil
}

func expandEc2BlockDeviceMappingForAmiEbsBlockDevice(tfMap map[string]interface{}) *ec2.BlockDeviceMapping {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.BlockDeviceMapping{
		Ebs: &ec2.EbsBlockDevice{},
	}

	if v, ok := tfMap["delete_on_termination"].(bool); ok {
		apiObject.Ebs.DeleteOnTermination = aws.Bool(v)
	}

	if v, ok := tfMap["device_name"].(string); ok && v != "" {
		apiObject.DeviceName = aws.String(v)
	}

	if v, ok := tfMap["iops"].(int); ok && v != 0 {
		apiObject.Ebs.Iops = aws.Int64(int64(v))
	}

	// "Parameter encrypted is invalid. You cannot specify the encrypted flag if specifying a snapshot id in a block device mapping."
	if v, ok := tfMap["snapshot_id"].(string); ok && v != "" {
		apiObject.Ebs.SnapshotId = aws.String(v)
	} else if v, ok := tfMap["encrypted"].(bool); ok {
		apiObject.Ebs.Encrypted = aws.Bool(v)
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

	if v, ok := tfMap["outpost_arn"].(string); ok && v != "" {
		apiObject.Ebs.OutpostArn = aws.String(v)
	}

	return apiObject
}

func expandEc2BlockDeviceMappingsForAmiEbsBlockDevice(tfList []interface{}) []*ec2.BlockDeviceMapping {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*ec2.BlockDeviceMapping

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandEc2BlockDeviceMappingForAmiEbsBlockDevice(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenEc2BlockDeviceMappingForAmiEbsBlockDevice(apiObject *ec2.BlockDeviceMapping) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	if apiObject.Ebs == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Ebs.DeleteOnTermination; v != nil {
		tfMap["delete_on_termination"] = aws.BoolValue(v)
	}

	if v := apiObject.DeviceName; v != nil {
		tfMap["device_name"] = aws.StringValue(v)
	}

	if v := apiObject.Ebs.Encrypted; v != nil {
		tfMap["encrypted"] = aws.BoolValue(v)
	}

	if v := apiObject.Ebs.Iops; v != nil {
		tfMap["iops"] = aws.Int64Value(v)
	}

	if v := apiObject.Ebs.SnapshotId; v != nil {
		tfMap["snapshot_id"] = aws.StringValue(v)
	}

	if v := apiObject.Ebs.Throughput; v != nil {
		tfMap["throughput"] = aws.Int64Value(v)
	}

	if v := apiObject.Ebs.VolumeSize; v != nil {
		tfMap["volume_size"] = aws.Int64Value(v)
	}

	if v := apiObject.Ebs.VolumeType; v != nil {
		tfMap["volume_type"] = aws.StringValue(v)
	}

	if v := apiObject.Ebs.OutpostArn; v != nil {
		tfMap["outpost_arn"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenEc2BlockDeviceMappingsForAmiEbsBlockDevice(apiObjects []*ec2.BlockDeviceMapping) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		if apiObject.Ebs == nil {
			continue
		}

		tfList = append(tfList, flattenEc2BlockDeviceMappingForAmiEbsBlockDevice(apiObject))
	}

	return tfList
}

func expandEc2BlockDeviceMappingForAmiEphemeralBlockDevice(tfMap map[string]interface{}) *ec2.BlockDeviceMapping {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.BlockDeviceMapping{}

	if v, ok := tfMap["device_name"].(string); ok && v != "" {
		apiObject.DeviceName = aws.String(v)
	}

	if v, ok := tfMap["virtual_name"].(string); ok && v != "" {
		apiObject.VirtualName = aws.String(v)
	}

	return apiObject
}

func expandEc2BlockDeviceMappingsForAmiEphemeralBlockDevice(tfList []interface{}) []*ec2.BlockDeviceMapping {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*ec2.BlockDeviceMapping

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandEc2BlockDeviceMappingForAmiEphemeralBlockDevice(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenEc2BlockDeviceMappingForAmiEphemeralBlockDevice(apiObject *ec2.BlockDeviceMapping) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.DeviceName; v != nil {
		tfMap["device_name"] = aws.StringValue(v)
	}

	if v := apiObject.VirtualName; v != nil {
		tfMap["virtual_name"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenEc2BlockDeviceMappingsForAmiEphemeralBlockDevice(apiObjects []*ec2.BlockDeviceMapping) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		if apiObject.Ebs != nil {
			continue
		}

		tfList = append(tfList, flattenEc2BlockDeviceMappingForAmiEphemeralBlockDevice(apiObject))
	}

	return tfList
}
