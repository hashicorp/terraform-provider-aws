package ec2

import (
	"bytes"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceImageImport() *schema.Resource {
	return &schema.Resource{
		Create: resourceImageImportCreate,
		// The remaining operations are shared with the generic aws_ami resource,
		// since the aws_ec2_image_import resource only differs in how it's created.
		Read:   resourceAMIUpdate,
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
			"arn": {
				Type:     schema.TypeString,
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
							Computed: true,
						},
						"device_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"encrypted": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"iops": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"outpost_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"snapshot_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"throughput": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"volume_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"volume_type": {
							Type:     schema.TypeString,
							Computed: true,
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
				Computed: true,
			},
			"encrypted": {
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
							Computed: true,
						},
						"virtual_name": {
							Type:     schema.TypeString,
							Computed: true,
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
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"image_location": {
				Type:     schema.TypeString,
				Computed: true,
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
				Computed: true,
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
			"name": {
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
			"platform": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"platform_details": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"ramdisk_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"role_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  DefaultSnapshotImportRoleName,
			},
			"root_device_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"root_snapshot_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sriov_net_support": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"usage_operation": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"virtualization_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceImageImportCreate(d *schema.ResourceData, meta interface{}) error {
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
	output, err := WaitImageImportComplete(conn, taskID, d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return fmt.Errorf("waiting for EC2 Image Import (%s) create: %w", taskID, err)
	}

	d.SetId(aws.StringValue(output.ImageId))
	d.Set("manage_ebs_snapshots", true)

	if len(tags) > 0 {
		if err := CreateTags(conn, d.Id(), tags); err != nil {
			return fmt.Errorf("setting EC2 Image Import (%s) tags: %w", d.Id(), err)
		}
	}

	if _, err := WaitImageAvailable(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for EC2 AMI (%s) create: %w", d.Id(), err)
	}

	return resourceAMIRead(d, meta)
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
