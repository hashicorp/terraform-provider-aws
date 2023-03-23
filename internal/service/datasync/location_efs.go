package datasync

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceLocationEFS() *schema.Resource {
	return &schema.Resource{
		Create: resourceLocationEFSCreate,
		Read:   resourceLocationEFSRead,
		Update: resourceLocationEFSUpdate,
		Delete: resourceLocationEFSDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ec2_config": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"security_group_arns": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subnet_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.NoZeroValues,
						},
					},
				},
			},
			"efs_file_system_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"subdirectory": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "/",
				// Ignore missing trailing slash
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if new == "/" {
						return false
					}
					if strings.TrimSuffix(old, "/") == strings.TrimSuffix(new, "/") {
						return true
					}
					return false
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"uri": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceLocationEFSCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataSyncConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &datasync.CreateLocationEfsInput{
		Ec2Config:        expandEC2Config(d.Get("ec2_config").([]interface{})),
		EfsFilesystemArn: aws.String(d.Get("efs_file_system_arn").(string)),
		Subdirectory:     aws.String(d.Get("subdirectory").(string)),
		Tags:             Tags(tags.IgnoreAWS()),
	}

	log.Printf("[DEBUG] Creating DataSync Location EFS: %s", input)
	output, err := conn.CreateLocationEfs(input)
	if err != nil {
		return fmt.Errorf("error creating DataSync Location EFS: %s", err)
	}

	d.SetId(aws.StringValue(output.LocationArn))

	return resourceLocationEFSRead(d, meta)
}

func resourceLocationEFSRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataSyncConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &datasync.DescribeLocationEfsInput{
		LocationArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading DataSync Location EFS: %s", input)
	output, err := conn.DescribeLocationEfs(input)

	if tfawserr.ErrMessageContains(err, "InvalidRequestException", "not found") {
		log.Printf("[WARN] DataSync Location EFS %q not found - removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading DataSync Location EFS (%s): %s", d.Id(), err)
	}

	subdirectory, err := SubdirectoryFromLocationURI(aws.StringValue(output.LocationUri))

	if err != nil {
		return err
	}

	d.Set("arn", output.LocationArn)

	if err := d.Set("ec2_config", flattenEC2Config(output.Ec2Config)); err != nil {
		return fmt.Errorf("error setting ec2_config: %s", err)
	}

	d.Set("subdirectory", subdirectory)
	d.Set("uri", output.LocationUri)

	tags, err := ListTags(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing tags for DataSync Location EFS (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceLocationEFSUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataSyncConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating DataSync Location EFS (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceLocationEFSRead(d, meta)
}

func resourceLocationEFSDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataSyncConn

	input := &datasync.DeleteLocationInput{
		LocationArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting DataSync Location EFS: %s", input)
	_, err := conn.DeleteLocation(input)

	if tfawserr.ErrMessageContains(err, "InvalidRequestException", "not found") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting DataSync Location EFS (%s): %s", d.Id(), err)
	}

	return nil
}

func flattenEC2Config(ec2Config *datasync.Ec2Config) []interface{} {
	if ec2Config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"security_group_arns": flex.FlattenStringSet(ec2Config.SecurityGroupArns),
		"subnet_arn":          aws.StringValue(ec2Config.SubnetArn),
	}

	return []interface{}{m}
}

func expandEC2Config(l []interface{}) *datasync.Ec2Config {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	ec2Config := &datasync.Ec2Config{
		SecurityGroupArns: flex.ExpandStringSet(m["security_group_arns"].(*schema.Set)),
		SubnetArn:         aws.String(m["subnet_arn"].(string)),
	}

	return ec2Config
}
