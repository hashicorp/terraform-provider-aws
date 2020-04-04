package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsDataSyncLocationFsxWindows() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDataSyncLocationFsxWindowsCreate,
		Read:   resourceAwsDataSyncLocationFsxWindowsRead,
		Update: resourceAwsDataSyncLocationFsxWindowsUpdate,
		Delete: resourceAwsDataSyncLocationFsxWindowsDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"fsx_filesystem_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"password": {
				Type:      schema.TypeString,
				Required:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"user": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"domain": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"security_group_arns": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateArn,
				},
			},
			"subdirectory": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"tags": tagsSchema(),
			"uri": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsDataSyncLocationFsxWindowsCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datasyncconn

	input := &datasync.CreateLocationFsxWindowsInput{
		FsxFilesystemArn:  aws.String(d.Get("fsx_filesystem_arn").(string)),
		User:              aws.String(d.Get("user").(string)),
		Password:          aws.String(d.Get("password").(string)),
		SecurityGroupArns: expandStringSet(d.Get("security_group_arns").(*schema.Set)),
		Tags:              keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().DatasyncTags(),
	}

	if v, ok := d.GetOk("subdirectory"); ok {
		input.Subdirectory = aws.String(v.(string))
	}

	if v, ok := d.GetOk("domain"); ok {
		input.Domain = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating DataSync Location FsxWindows: %s", input)
	output, err := conn.CreateLocationFsxWindows(input)
	if err != nil {
		return fmt.Errorf("error creating DataSync Location FsxWindows: %s", err)
	}

	d.SetId(aws.StringValue(output.LocationArn))

	return resourceAwsDataSyncLocationFsxWindowsRead(d, meta)
}

func resourceAwsDataSyncLocationFsxWindowsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datasyncconn

	input := &datasync.DescribeLocationFsxWindowsInput{
		LocationArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading DataSync Location FsxWindows: %s", input)
	output, err := conn.DescribeLocationFsxWindows(input)

	if isAWSErr(err, datasync.ErrCodeInvalidRequestException, "not found") {
		log.Printf("[WARN] DataSync Location FsxWindows %q not found - removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading DataSync Location FsxWindows (%s): %s", d.Id(), err)
	}

	subdirectory, err := dataSyncParseLocationURI(aws.StringValue(output.LocationUri))

	if err != nil {
		return fmt.Errorf("error parsing Location FsxWindows (%s) URI (%s): %s", d.Id(), aws.StringValue(output.LocationUri), err)
	}

	d.Set("arn", output.LocationArn)
	d.Set("security_group_arns", flattenStringSet(output.SecurityGroupArns))
	d.Set("subdirectory", subdirectory)
	d.Set("uri", output.LocationUri)
	d.Set("user", output.User)
	d.Set("domain", output.Domain)

	if err := d.Set("creation_time", output.CreationTime.Format(time.RFC3339)); err != nil {
		return fmt.Errorf("error setting creation_time: %s", err)
	}

	tags, err := keyvaluetags.DatasyncListTags(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing tags for DataSync Location FsxWindows (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsDataSyncLocationFsxWindowsUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datasyncconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.DatasyncUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating DataSync Location FsxWindows (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsDataSyncLocationFsxWindowsRead(d, meta)
}

func resourceAwsDataSyncLocationFsxWindowsDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datasyncconn

	input := &datasync.DeleteLocationInput{
		LocationArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting DataSync Location FsxWindows: %s", input)
	_, err := conn.DeleteLocation(input)

	if isAWSErr(err, datasync.ErrCodeInvalidRequestException, "not found") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting DataSync Location FsxWindows (%s): %s", d.Id(), err)
	}

	return nil
}
