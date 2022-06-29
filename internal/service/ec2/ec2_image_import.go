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

func ResourceEC2ImageImport() *schema.Resource {
	return &schema.Resource{
		Create: resourceEC2ImageImportCreate,
		Read:   resourceEC2ImageImportRead,
		Update: resourceAMIUpdate,
		Delete: resourceAMIDelete,

		CustomizeDiff: verify.SetTagsDiff,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"architecture": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"boot_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(ec2.BootModeValues_Values(), false),
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
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
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
			"license_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
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
			"platform": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"role_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  DefaultSnapshotImportRoleName,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceEC2ImageImportCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.ImportImageInput{
		ClientToken:       aws.String(resource.UniqueId()),
		TagSpecifications: tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeImportImageTask),
	}

	if v, ok := d.GetOk("client_data"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ClientData = ExpandClientData(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("disk_container"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		var err error
		input.DiskContainers, err = expandImageDiskContainers(v.([]interface{}))
		if err != nil {
			return fmt.Errorf("creating EC2 Image Import: %w", err)
		}
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

	if v, ok := d.GetOk("platform"); ok {
		input.Platform = aws.String(v.(string))
	}

	if v, ok := d.GetOk("license_type"); ok {
		input.LicenseType = aws.String(v.(string))
	}

	outputRaw, err := tfresource.RetryWhenAWSErrMessageContains(propagationTimeout,
		func() (interface{}, error) {
			return conn.ImportImage(input)
		},
		errCodeInvalidParameter, "provided does not exist or does not have sufficient permissions")

	if err != nil {
		return fmt.Errorf("creating EC2 Image Import: %w", err)
	}

	taskID := aws.StringValue(outputRaw.(*ec2.ImportImageOutput).ImportTaskId)
	output, err := WaitEC2ImageImportComplete(conn, taskID, d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return fmt.Errorf("waiting for EC2 Image Import (%s) create: %w", taskID, err)
	}

	d.SetId(aws.StringValue(output.ImageId))

	if len(tags) > 0 {
		if err := CreateTags(conn, d.Id(), tags); err != nil {
			return fmt.Errorf("setting EC2 Image Import (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceEC2ImageImportRead(d, meta)
}

func resourceEC2ImageImportRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	image, err := FindImageByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Image %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading EC2 Image (%s): %w", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("image/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("description", image.Description)
	d.Set("encrypted", image.BlockDeviceMappings[0].Ebs.Encrypted)
	d.Set("kms_key_id", image.BlockDeviceMappings[0].Ebs.KmsKeyId)
	d.Set("owner_alias", image.ImageOwnerAlias)
	d.Set("owner_id", image.OwnerId)
	d.Set("platform", image.Platform)

	tags := KeyValueTags(image.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("setting tags_all: %w", err)
	}

	return nil
}

func expandImageDiskContainers(containers []interface{}) ([]*ec2.ImageDiskContainer, error) {
	if containers == nil {
		return nil, nil
	}

	apiObjects := []*ec2.ImageDiskContainer{}

	for _, raw := range containers {
		apiObject := &ec2.ImageDiskContainer{}

		m := raw.(map[string]interface{})

		if v, ok := m["description"].(string); ok && v != "" {
			apiObject.Description = aws.String(v)
		}

		if v, ok := m["format"].(string); ok && v != "" {
			apiObject.Format = aws.String(v)
		}

		if v, ok := m["url"].(string); ok && v != "" {
			apiObject.Url = aws.String(v)
		}

		// Do a manual one of check here: https://github.com/hashicorp/terraform-plugin-sdk/issues/71
		if v, ok := m["user_bucket"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			if apiObject.Url != nil {
				return nil, fmt.Errorf(
					"url and user_bucket cannot be set on the same disk container")
			}
			apiObject.UserBucket = ExpandUserBucket(v[0].(map[string]interface{}))
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects, nil
}
