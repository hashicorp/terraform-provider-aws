package ec2

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceEBSSnapshotImport() *schema.Resource {
	return &schema.Resource{
		Create: resourceEBSSnapshotImportCreate,
		Read:   resourceEBSSnapshotImportRead,
		Update: resourceEBSSnapshotUpdate,
		Delete: resourceEBSSnapshotDelete,

		CustomizeDiff: verify.SetTagsDiff,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"client_data": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"comment": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"upload_end": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IsRFC3339Time,
						},
						"upload_size": {
							Type:     schema.TypeFloat,
							Optional: true,
							Computed: true,
						},
						"upload_start": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IsRFC3339Time,
						},
					},
				},
			},
			"data_encryption_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"disk_container": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"format": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(ec2.DiskImageFormat_Values(), false),
						},
						"url": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ExactlyOneOf: []string{"disk_container.0.user_bucket", "disk_container.0.url"},
						},
						"user_bucket": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"s3_bucket": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									"s3_key": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
								},
							},
							ExactlyOneOf: []string{"disk_container.0.user_bucket", "disk_container.0.url"},
						},
					},
				},
			},
			"encrypted": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"outpost_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_alias": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"permanent_restore": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"role_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  DefaultSnapshotImportRoleName,
			},
			"storage_tier": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(append(ec2.TargetStorageTier_Values(), TargetStorageTierStandard), false),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"temporary_restore_days": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"volume_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"volume_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceEBSSnapshotImportCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.ImportSnapshotInput{
		ClientToken:       aws.String(resource.UniqueId()),
		TagSpecifications: tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeImportSnapshotTask),
	}

	if v, ok := d.GetOk("client_data"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ClientData = expandClientData(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("disk_container"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.DiskContainer = expandSnapshotDiskContainer(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("encrypted"); ok {
		input.Encrypted = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("role_name"); ok {
		input.RoleName = aws.String(v.(string))
	}

	outputRaw, err := tfresource.RetryWhenAWSErrMessageContains(propagationTimeout,
		func() (interface{}, error) {
			return conn.ImportSnapshot(input)
		},
		errCodeInvalidParameter, "provided does not exist or does not have sufficient permissions")

	if err != nil {
		return fmt.Errorf("creating EBS Snapshot Import: %w", err)
	}

	taskID := aws.StringValue(outputRaw.(*ec2.ImportSnapshotOutput).ImportTaskId)
	output, err := WaitEBSSnapshotImportComplete(conn, taskID, d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return fmt.Errorf("waiting for EBS Snapshot Import (%s) create: %w", taskID, err)
	}

	d.SetId(aws.StringValue(output.SnapshotId))

	if len(tags) > 0 {
		if err := CreateTags(conn, d.Id(), tags); err != nil {
			return fmt.Errorf("setting EBS Snapshot Import (%s) tags: %w", d.Id(), err)
		}
	}

	if v, ok := d.GetOk("storage_tier"); ok && v.(string) == ec2.TargetStorageTierArchive {
		_, err = conn.ModifySnapshotTier(&ec2.ModifySnapshotTierInput{
			SnapshotId:  aws.String(d.Id()),
			StorageTier: aws.String(v.(string)),
		})

		if err != nil {
			return fmt.Errorf("setting EBS Snapshot Import (%s) Storage Tier: %w", d.Id(), err)
		}

		_, err = waitEBSSnapshotTierArchive(conn, d.Id(), ebsSnapshotArchivedTimeout)

		if err != nil {
			return fmt.Errorf("waiting for EBS Snapshot Import (%s) Storage Tier archive: %w", d.Id(), err)
		}
	}

	return resourceEBSSnapshotImportRead(d, meta)
}

func resourceEBSSnapshotImportRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	snapshot, err := FindSnapshotByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EBS Snapshot %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading EBS Snapshot (%s): %w", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("snapshot/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("data_encryption_key_id", snapshot.DataEncryptionKeyId)
	d.Set("description", snapshot.Description)
	d.Set("encrypted", snapshot.Encrypted)
	d.Set("kms_key_id", snapshot.KmsKeyId)
	d.Set("owner_alias", snapshot.OwnerAlias)
	d.Set("owner_id", snapshot.OwnerId)
	d.Set("storage_tier", snapshot.StorageTier)
	d.Set("volume_size", snapshot.VolumeSize)

	tags := KeyValueTags(snapshot.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("setting tags_all: %w", err)
	}

	return nil
}

func expandClientData(tfMap map[string]interface{}) *ec2.ClientData {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.ClientData{}

	if v, ok := tfMap["comment"].(string); ok && v != "" {
		apiObject.Comment = aws.String(v)
	}

	if v, ok := tfMap["upload_end"].(string); ok && v != "" {
		v, _ := time.Parse(time.RFC3339, v)

		apiObject.UploadEnd = aws.Time(v)
	}

	if v, ok := tfMap["upload_size"].(float64); ok && v != 0.0 {
		apiObject.UploadSize = aws.Float64(v)
	}

	if v, ok := tfMap["upload_start"].(string); ok {
		v, _ := time.Parse(time.RFC3339, v)

		apiObject.UploadStart = aws.Time(v)
	}

	return apiObject
}

func expandSnapshotDiskContainer(tfMap map[string]interface{}) *ec2.SnapshotDiskContainer {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.SnapshotDiskContainer{}

	if v, ok := tfMap["description"].(string); ok && v != "" {
		apiObject.Description = aws.String(v)
	}

	if v, ok := tfMap["format"].(string); ok && v != "" {
		apiObject.Format = aws.String(v)
	}

	if v, ok := tfMap["url"].(string); ok && v != "" {
		apiObject.Url = aws.String(v)
	}

	if v, ok := tfMap["user_bucket"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.UserBucket = expandUserBucket(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandUserBucket(tfMap map[string]interface{}) *ec2.UserBucket {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.UserBucket{}

	if v, ok := tfMap["s3_bucket"].(string); ok && v != "" {
		apiObject.S3Bucket = aws.String(v)
	}

	if v, ok := tfMap["s3_key"].(string); ok && v != "" {
		apiObject.S3Key = aws.String(v)
	}

	return apiObject
}
