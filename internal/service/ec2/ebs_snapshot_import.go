package ec2

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
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
		Create:        resourceEBSSnapshotImportCreate,
		Read:          resourceEBSSnapshotImportRead,
		Update:        resourceEBSSnapshotUpdate,
		Delete:        resourceEBSSnapshotDelete,
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
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"upload_size": {
							Type:     schema.TypeFloat,
							Optional: true,
							Computed: true,
						},
						"upload_start": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
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
							Type:         schema.TypeList,
							Optional:     true,
							ForceNew:     true,
							ExactlyOneOf: []string{"disk_container.0.user_bucket", "disk_container.0.url"},
							MaxItems:     1,
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
				Default:  "vmimport",
			},
			"storage_tier": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.Any(
					validation.StringInSlice(ec2.TargetStorageTier_Values(), false),
					validation.StringInSlice([]string{"standard"}, false), //Enum slice does not include `standard` type.
				),
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

	request := &ec2.ImportSnapshotInput{
		TagSpecifications: tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeImportSnapshotTask),
	}

	if clientData, ok := d.GetOk("client_data"); ok {
		for _, v := range clientData.([]interface{}) {
			if cdv, ok := v.(map[string]interface{}); ok {

				clientData, err := expandEBSSnapshotClientData(cdv)
				if err != nil {
					return err
				}

				request.ClientData = clientData
			}
		}
	}

	request.ClientToken = aws.String(resource.UniqueId())

	if v, ok := d.GetOk("description"); ok {
		request.Description = aws.String(v.(string))
	}

	diskContainer := d.Get("disk_container")
	for _, v := range diskContainer.([]interface{}) {
		if dcv, ok := v.(map[string]interface{}); ok {

			diskContainer := expandEBSSnapshotDiskContainer(dcv)
			request.DiskContainer = diskContainer
		}
	}

	if v, ok := d.GetOk("encrypted"); ok {
		request.Encrypted = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		request.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("role_name"); ok {
		request.RoleName = aws.String(v.(string))
	}

	err := resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		resp, err := conn.ImportSnapshot(request)

		if tfawserr.ErrMessageContains(err, "InvalidParameter", "provided does not exist or does not have sufficient permissions") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		importTaskId := aws.StringValue(resp.ImportTaskId)

		res, err := WaitEBSSnapshotImportComplete(conn, importTaskId)
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("Error waiting for snapshot (%s) to be imported: %s", importTaskId, err))
		}

		d.SetId(aws.StringValue(res.SnapshotId))

		tags := d.Get("tags").(map[string]interface{})
		if len(tags) > 0 {
			if err := CreateTags(conn, d.Id(), tags); err != nil {
				return resource.NonRetryableError(fmt.Errorf("error setting tags: %s", err))
			}
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		return fmt.Errorf("timeout error importing EBS Snapshot: %s", err)
	}

	if err != nil {
		return fmt.Errorf("error importing EBS Snapshot: %s", err)
	}

	if v, ok := d.GetOk("storage_tier"); ok && v.(string) == ec2.TargetStorageTierArchive {
		_, err = conn.ModifySnapshotTier(&ec2.ModifySnapshotTierInput{
			SnapshotId:  aws.String(d.Id()),
			StorageTier: aws.String(v.(string)),
		})

		if err != nil {
			return fmt.Errorf("error setting EBS Snapshot Import (%s) Storage Tier: %w", d.Id(), err)
		}

		_, err = WaitEBSSnapshotTierArchive(conn, d.Id())
		if err != nil {
			return fmt.Errorf("Error waiting for EBS Snapshot Import (%s) Storage Tier to be archived: %w", d.Id(), err)
		}
	}

	return resourceEBSSnapshotImportRead(d, meta)
}

func resourceEBSSnapshotImportRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	snapshot, err := FindSnapshotById(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EBS Snapshot (%s) Not found - removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EBS Snapshot (%s): %w", d.Id(), err)
	}

	d.Set("description", snapshot.Description)
	d.Set("owner_id", snapshot.OwnerId)
	d.Set("encrypted", snapshot.Encrypted)
	d.Set("owner_alias", snapshot.OwnerAlias)
	d.Set("data_encryption_key_id", snapshot.DataEncryptionKeyId)
	d.Set("kms_key_id", snapshot.KmsKeyId)
	d.Set("volume_size", snapshot.VolumeSize)
	d.Set("storage_tier", snapshot.StorageTier)

	tags := KeyValueTags(snapshot.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	snapshotArn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("snapshot/%s", d.Id()),
		Service:   "ec2",
	}.String()

	d.Set("arn", snapshotArn)

	return nil
}

func expandEBSSnapshotClientData(tfMap map[string]interface{}) (*ec2.ClientData, error) {
	clientData := &ec2.ClientData{}

	if v, ok := tfMap["comment"].(string); ok {
		clientData.Comment = aws.String(v)
	}

	if v, ok := tfMap["upload_end"].(string); ok {
		upload_end, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return nil, fmt.Errorf("error parsing upload_end to timestamp: %s", err)
		}
		clientData.UploadEnd = aws.Time(upload_end)
	}

	if v, ok := tfMap["upload_size"].(float64); ok {
		clientData.UploadSize = aws.Float64(v)
	}

	if v, ok := tfMap["upload_start"].(string); ok {
		upload_start, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return nil, fmt.Errorf("error parsing upload_start to timestamp: %s", err)
		}
		clientData.UploadStart = aws.Time(upload_start)
	}

	return clientData, nil
}

func expandEBSSnapshotDiskContainer(tfMap map[string]interface{}) *ec2.SnapshotDiskContainer {
	diskContainer := &ec2.SnapshotDiskContainer{
		Format: aws.String(tfMap["format"].(string)),
	}

	if v, ok := tfMap["description"].(string); ok {
		diskContainer.Description = aws.String(v)
	}

	if v, ok := tfMap["url"].(string); ok && v != "" {
		diskContainer.Url = aws.String(v)
	}

	if v, ok := tfMap["user_bucket"]; ok {
		vL := v.([]interface{})
		for _, v := range vL {
			ub := v.(map[string]interface{})
			diskContainer.UserBucket = &ec2.UserBucket{
				S3Bucket: aws.String(ub["s3_bucket"].(string)),
				S3Key:    aws.String(ub["s3_key"].(string)),
			}
		}
	}

	return diskContainer
}
