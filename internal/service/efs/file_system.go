package efs

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceFileSystem() *schema.Resource {
	return &schema.Resource{
		Create: resourceFileSystemCreate,
		Read:   resourceFileSystemRead,
		Update: resourceFileSystemUpdate,
		Delete: resourceFileSystemDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"availability_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"availability_zone_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"creation_token": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 64),
			},

			"performance_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(efs.PerformanceMode_Values(), false),
			},

			"encrypted": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"kms_key_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},

			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"provisioned_throughput_in_mibps": {
				Type:     schema.TypeFloat,
				Optional: true,
			},
			"number_of_mount_targets": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),

			"throughput_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      efs.ThroughputModeBursting,
				ValidateFunc: validation.StringInSlice(efs.ThroughputMode_Values(), false),
			},

			"lifecycle_policy": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 2,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"transition_to_ia": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(efs.TransitionToIARules_Values(), false),
						},
						"transition_to_primary_storage_class": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(efs.TransitionToPrimaryStorageClassRules_Values(), false),
						},
					},
				},
			},
			"size_in_bytes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"value": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"value_in_ia": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"value_in_standard": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceFileSystemCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EFSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	creationToken := ""
	if v, ok := d.GetOk("creation_token"); ok {
		creationToken = v.(string)
	} else {
		creationToken = resource.UniqueId()
	}
	throughputMode := d.Get("throughput_mode").(string)

	createOpts := &efs.CreateFileSystemInput{
		CreationToken:  aws.String(creationToken),
		ThroughputMode: aws.String(throughputMode),
		Tags:           Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("availability_zone_name"); ok {
		createOpts.AvailabilityZoneName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("performance_mode"); ok {
		createOpts.PerformanceMode = aws.String(v.(string))
	}

	if throughputMode == efs.ThroughputModeProvisioned {
		createOpts.ProvisionedThroughputInMibps = aws.Float64(d.Get("provisioned_throughput_in_mibps").(float64))
	}

	encrypted, hasEncrypted := d.GetOk("encrypted")
	kmsKeyId, hasKmsKeyId := d.GetOk("kms_key_id")

	if hasEncrypted {
		createOpts.Encrypted = aws.Bool(encrypted.(bool))
	}

	if hasKmsKeyId {
		createOpts.KmsKeyId = aws.String(kmsKeyId.(string))
	}

	if encrypted == false && hasKmsKeyId {
		return errors.New("encrypted must be set to true when kms_key_id is specified")
	}

	log.Printf("[DEBUG] Creating EFS file system: %s", createOpts)
	fs, err := conn.CreateFileSystem(createOpts)

	if err != nil {
		return fmt.Errorf("error creating EFS file system: %w", err)
	}

	d.SetId(aws.StringValue(fs.FileSystemId))

	if _, err := waitFileSystemAvailable(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EFS file system (%s) to be available: %w", d.Id(), err)
	}

	_, hasLifecyclePolicy := d.GetOk("lifecycle_policy")
	if hasLifecyclePolicy {
		_, err := conn.PutLifecycleConfiguration(&efs.PutLifecycleConfigurationInput{
			FileSystemId:      aws.String(d.Id()),
			LifecyclePolicies: expandFileSystemLifecyclePolicies(d.Get("lifecycle_policy").([]interface{})),
		})

		if err != nil {
			return fmt.Errorf("error creating EFS file system (%s) lifecycle configuration: %w", d.Id(), err)
		}
	}

	return resourceFileSystemRead(d, meta)
}

func resourceFileSystemUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EFSConn

	if d.HasChanges("provisioned_throughput_in_mibps", "throughput_mode") {
		throughputMode := d.Get("throughput_mode").(string)

		input := &efs.UpdateFileSystemInput{
			FileSystemId:   aws.String(d.Id()),
			ThroughputMode: aws.String(throughputMode),
		}

		if throughputMode == efs.ThroughputModeProvisioned {
			input.ProvisionedThroughputInMibps = aws.Float64(d.Get("provisioned_throughput_in_mibps").(float64))
		}

		_, err := conn.UpdateFileSystem(input)

		if err != nil {
			return fmt.Errorf("error updating EFS file system (%s): %w", d.Id(), err)
		}

		if _, err := waitFileSystemAvailable(conn, d.Id()); err != nil {
			return fmt.Errorf("error waiting for EFS file system (%s) to be available: %w", d.Id(), err)
		}
	}

	if d.HasChange("lifecycle_policy") {
		input := &efs.PutLifecycleConfigurationInput{
			FileSystemId:      aws.String(d.Id()),
			LifecyclePolicies: expandFileSystemLifecyclePolicies(d.Get("lifecycle_policy").([]interface{})),
		}

		// Prevent the following error during removal:
		// InvalidParameter: 1 validation error(s) found.
		// - missing required field, PutLifecycleConfigurationInput.LifecyclePolicies.
		if input.LifecyclePolicies == nil {
			input.LifecyclePolicies = []*efs.LifecyclePolicy{}
		}

		_, err := conn.PutLifecycleConfiguration(input)

		if err != nil {
			return fmt.Errorf("error updating EFS file system (%s) lifecycle configuration: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EFS file system (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceFileSystemRead(d, meta)
}

func resourceFileSystemRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EFSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	fs, err := FindFileSystemByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EFS file system (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EFS file system (%s): %w", d.Id(), err)
	}

	d.Set("arn", fs.FileSystemArn)
	d.Set("availability_zone_id", fs.AvailabilityZoneId)
	d.Set("availability_zone_name", fs.AvailabilityZoneName)
	d.Set("creation_token", fs.CreationToken)
	d.Set("encrypted", fs.Encrypted)
	d.Set("kms_key_id", fs.KmsKeyId)
	d.Set("performance_mode", fs.PerformanceMode)
	d.Set("provisioned_throughput_in_mibps", fs.ProvisionedThroughputInMibps)
	d.Set("throughput_mode", fs.ThroughputMode)
	d.Set("owner_id", fs.OwnerId)
	d.Set("number_of_mount_targets", fs.NumberOfMountTargets)

	tags := KeyValueTags(fs.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	if err := d.Set("size_in_bytes", flattenFileSystemSizeInBytes(fs.SizeInBytes)); err != nil {
		return fmt.Errorf("error setting size_in_bytes: %w", err)
	}

	d.Set("dns_name", meta.(*conns.AWSClient).RegionalHostname(fmt.Sprintf("%s.efs", aws.StringValue(fs.FileSystemId))))

	res, err := conn.DescribeLifecycleConfiguration(&efs.DescribeLifecycleConfigurationInput{
		FileSystemId: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("error reading EFS file system (%s) lifecycle configuration: %w", d.Id(), err)
	}

	if err := d.Set("lifecycle_policy", flattenFileSystemLifecyclePolicies(res.LifecyclePolicies)); err != nil {
		return fmt.Errorf("error setting lifecycle_policy: %w", err)
	}

	return nil
}

func resourceFileSystemDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EFSConn

	log.Printf("[DEBUG] Deleting EFS file system: %s", d.Id())
	_, err := conn.DeleteFileSystem(&efs.DeleteFileSystemInput{
		FileSystemId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, efs.ErrCodeFileSystemNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EFS file system (%s): %w", d.Id(), err)
	}

	if _, err := waitFileSystemDeleted(conn, d.Id()); err != nil {
		if tfawserr.ErrCodeEquals(err, efs.ErrCodeFileSystemNotFound) {
			return nil
		}
		return fmt.Errorf("error waiting for EFS file system (%s) deletion: %w", d.Id(), err)
	}

	return nil
}

func flattenFileSystemLifecyclePolicies(apiObjects []*efs.LifecyclePolicy) []interface{} {
	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfMap := make(map[string]interface{})

		if apiObject.TransitionToIA != nil {
			tfMap["transition_to_ia"] = aws.StringValue(apiObject.TransitionToIA)
		}

		if apiObject.TransitionToPrimaryStorageClass != nil {
			tfMap["transition_to_primary_storage_class"] = aws.StringValue(apiObject.TransitionToPrimaryStorageClass)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandFileSystemLifecyclePolicies(tfList []interface{}) []*efs.LifecyclePolicy {
	var apiObjects []*efs.LifecyclePolicy

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := &efs.LifecyclePolicy{}

		if v, ok := tfMap["transition_to_ia"].(string); ok && v != "" {
			apiObject.TransitionToIA = aws.String(v)
		}

		if v, ok := tfMap["transition_to_primary_storage_class"].(string); ok && v != "" {
			apiObject.TransitionToPrimaryStorageClass = aws.String(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenFileSystemSizeInBytes(sizeInBytes *efs.FileSystemSize) []interface{} {
	if sizeInBytes == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"value": aws.Int64Value(sizeInBytes.Value),
	}

	if sizeInBytes.ValueInIA != nil {
		m["value_in_ia"] = aws.Int64Value(sizeInBytes.ValueInIA)
	}

	if sizeInBytes.ValueInStandard != nil {
		m["value_in_standard"] = aws.Int64Value(sizeInBytes.ValueInStandard)
	}

	return []interface{}{m}
}
