package logs

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceGroupCreate,
		Read:   resourceGroupRead,
		Update: resourceGroupUpdate,
		Delete: resourceGroupDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validLogGroupName,
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validLogGroupNamePrefix,
			},
			"retention_in_days": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntInSlice([]int{0, 1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1827, 3653}),
			},
			"skip_destroy": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LogsConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	input := &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(name),
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	_, err := conn.CreateLogGroup(input)

	if err != nil {
		return fmt.Errorf("creating CloudWatch Log Group (%s): %w", name, err)
	}

	d.SetId(name)

	if v, ok := d.GetOk("retention_in_days"); ok {
		_, err := conn.PutRetentionPolicy(&cloudwatchlogs.PutRetentionPolicyInput{
			LogGroupName:    aws.String(d.Id()),
			RetentionInDays: aws.Int64(int64(v.(int))),
		})

		if err != nil {
			return fmt.Errorf("setting CloudWatch Log Group (%s) retention policy: %w", d.Id(), err)
		}
	}

	return resourceGroupRead(d, meta)
}

func resourceGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LogsConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	log.Printf("[DEBUG] Reading CloudWatch Log Group: %q", d.Get("name").(string))
	lg, err := LookupGroup(conn, d.Id())
	if err != nil {
		return err
	}

	if lg == nil {
		log.Printf("[DEBUG] CloudWatch Group %q Not Found", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", TrimLogGroupARNWildcardSuffix(aws.StringValue(lg.Arn)))
	d.Set("kms_key_id", lg.KmsKeyId)
	d.Set("name", lg.LogGroupName)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(lg.LogGroupName)))
	d.Set("retention_in_days", lg.RetentionInDays)

	tags, err := ListTags(conn, d.Id())

	if err != nil {
		return fmt.Errorf("listing tags for CloudWatch Logs Group (%s): %w", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("setting tags_all: %w", err)
	}

	return nil
}

func resourceGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LogsConn

	if d.HasChange("retention_in_days") {
		if v, ok := d.GetOk("retention_in_days"); ok {
			_, err := conn.PutRetentionPolicy(&cloudwatchlogs.PutRetentionPolicyInput{
				LogGroupName:    aws.String(d.Id()),
				RetentionInDays: aws.Int64(int64(v.(int))),
			})

			if err != nil {
				return fmt.Errorf("setting CloudWatch Log Group (%s) retention policy: %w", d.Id(), err)
			}
		} else {
			_, err := conn.DeleteRetentionPolicy(&cloudwatchlogs.DeleteRetentionPolicyInput{
				LogGroupName: aws.String(d.Id()),
			})

			if err != nil {
				return fmt.Errorf("deleting CloudWatch Log Group (%s) retention policy: %w", d.Id(), err)
			}
		}
	}

	if d.HasChange("kms_key_id") {
		if v, ok := d.GetOk("kms_key_id"); ok {
			_, err := conn.AssociateKmsKey(&cloudwatchlogs.AssociateKmsKeyInput{
				KmsKeyId:     aws.String(v.(string)),
				LogGroupName: aws.String(d.Id()),
			})

			if err != nil {
				return fmt.Errorf("associating CloudWatch Log Group (%s) KMS key: %w", d.Id(), err)
			}
		} else {
			_, err := conn.DisassociateKmsKey(&cloudwatchlogs.DisassociateKmsKeyInput{
				LogGroupName: aws.String(d.Id()),
			})

			if err != nil {
				return fmt.Errorf("disassociating CloudWatch Log Group (%s) KMS key: %w", d.Id(), err)
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("updating CloudWatch Log Group (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceGroupRead(d, meta)
}

func resourceGroupDelete(d *schema.ResourceData, meta interface{}) error {
	if v, ok := d.GetOk("skip_destroy"); ok && v.(bool) {
		log.Printf("[DEBUG] Retaining CloudWatch Log Group: %s", d.Id())
		return nil
	}

	conn := meta.(*conns.AWSClient).LogsConn

	log.Printf("[INFO] Deleting CloudWatch Log Group: %s", d.Id())
	_, err := conn.DeleteLogGroup(&cloudwatchlogs.DeleteLogGroupInput{
		LogGroupName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, cloudwatchlogs.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting CloudWatch Log Group (%s): %w", d.Id(), err)
	}

	return nil
}

func LookupGroup(conn *cloudwatchlogs.CloudWatchLogs, name string) (*cloudwatchlogs.LogGroup, error) {
	input := &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: aws.String(name),
	}
	var logGroup *cloudwatchlogs.LogGroup
	err := conn.DescribeLogGroupsPages(input, func(page *cloudwatchlogs.DescribeLogGroupsOutput, lastPage bool) bool {
		for _, lg := range page.LogGroups {
			if aws.StringValue(lg.LogGroupName) == name {
				logGroup = lg
				return false
			}
		}
		return !lastPage
	})
	if err != nil {
		return nil, err
	}

	return logGroup, nil
}
