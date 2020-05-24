package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsDataSyncLocationFsxWindowsFileSystem() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDataSyncLocationFsxWindowsFileSystemCreate,
		Read:   resourceAwsDataSyncLocationFsxWindowsFileSystemRead,
		Update: resourceAwsDataSyncLocationFsxWindowsFileSystemUpdate,
		Delete: resourceAwsDataSyncLocationFsxWindowsFileSystemDelete,
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
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(1, 104),
			},
			"user": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 104),
			},
			"domain": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 253),
			},
			"security_group_arns": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				MaxItems: 5,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateArn,
				},
			},
			"subdirectory": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 4096),
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

func resourceAwsDataSyncLocationFsxWindowsFileSystemCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datasyncconn
	fsxArn := d.Get("fsx_filesystem_arn").(string)

	input := &datasync.CreateLocationFsxWindowsInput{
		FsxFilesystemArn:  aws.String(fsxArn),
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

	log.Printf("[DEBUG] Creating DataSync Location Fsx Windows File System: %s", input)
	output, err := conn.CreateLocationFsxWindows(input)
	if err != nil {
		return fmt.Errorf("error creating DataSync Location Fsx Windows File System: %s", err)
	}

	d.SetId(fmt.Sprintf("%s#%s", aws.StringValue(output.LocationArn), fsxArn))

	return resourceAwsDataSyncLocationFsxWindowsFileSystemRead(d, meta)
}

func resourceAwsDataSyncLocationFsxWindowsFileSystemRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datasyncconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	locationArn, fsxArn, err := decodeAwsDataSyncLocationFsxWindowsFileSystemID(d.Id())
	if err != nil {
		return err
	}

	input := &datasync.DescribeLocationFsxWindowsInput{
		LocationArn: aws.String(locationArn),
	}

	log.Printf("[DEBUG] Reading DataSync Location Fsx Windows: %s", input)
	output, err := conn.DescribeLocationFsxWindows(input)

	if isAWSErr(err, datasync.ErrCodeInvalidRequestException, "not found") {
		log.Printf("[WARN] DataSync Location Fsx Windows %q not found - removing from state", locationArn)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading DataSync Location Fsx Windows (%s): %s", locationArn, err)
	}

	subdirectory, err := dataSyncParseLocationURI(aws.StringValue(output.LocationUri))

	if err != nil {
		return fmt.Errorf("error parsing Location Fsx Windows File System (%s) URI (%s): %s", d.Id(), aws.StringValue(output.LocationUri), err)
	}

	d.Set("arn", output.LocationArn)
	d.Set("fsx_filesystem_arn", fsxArn)
	d.Set("security_group_arns", flattenStringSet(output.SecurityGroupArns))
	d.Set("subdirectory", subdirectory)
	d.Set("uri", output.LocationUri)
	d.Set("user", output.User)
	d.Set("domain", output.Domain)

	if err := d.Set("creation_time", output.CreationTime.Format(time.RFC3339)); err != nil {
		return fmt.Errorf("error setting creation_time: %s", err)
	}

	tags, err := keyvaluetags.DatasyncListTags(conn, locationArn)

	if err != nil {
		return fmt.Errorf("error listing tags for DataSync Location Fsx Windows (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsDataSyncLocationFsxWindowsFileSystemUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datasyncconn
	locationArn, _, err := decodeAwsDataSyncLocationFsxWindowsFileSystemID(d.Id())
	if err != nil {
		return err
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.DatasyncUpdateTags(conn, locationArn, o, n); err != nil {
			return fmt.Errorf("error updating DataSync Location Fsx Windows File System (%s) tags: %s", locationArn, err)
		}
	}

	return resourceAwsDataSyncLocationFsxWindowsFileSystemRead(d, meta)
}

func resourceAwsDataSyncLocationFsxWindowsFileSystemDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datasyncconn
	locationArn, _, err := decodeAwsDataSyncLocationFsxWindowsFileSystemID(d.Id())
	if err != nil {
		return err
	}

	input := &datasync.DeleteLocationInput{
		LocationArn: aws.String(locationArn),
	}

	log.Printf("[DEBUG] Deleting DataSync Location Fsx Windows File System: %s", input)
	_, err = conn.DeleteLocation(input)

	if isAWSErr(err, datasync.ErrCodeInvalidRequestException, "not found") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting DataSync Location Fsx Windows (%s): %s", locationArn, err)
	}

	return nil
}

func decodeAwsDataSyncLocationFsxWindowsFileSystemID(id string) (string, string, error) {
	parts := strings.Split(id, "#")

	if len(parts) != 2 {
		return "", "", fmt.Errorf("Unexpected format of ID (%q), expected DataSyncLocationArn:FsxArn", id)
	}

	return parts[0], parts[1], nil
}
