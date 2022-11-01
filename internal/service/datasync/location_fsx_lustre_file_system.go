package datasync

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceLocationFSxLustreFileSystem() *schema.Resource {
	return &schema.Resource{
		Create: resourceLocationFSxLustreFileSystemCreate,
		Read:   resourceLocationFSxLustreFileSystemRead,
		Update: resourceLocationFSxLustreFileSystemUpdate,
		Delete: resourceLocationFSxLustreFileSystemDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "#")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected DataSyncLocationArn#FsxArn", d.Id())
				}

				DSArn := idParts[0]
				FSxArn := idParts[1]

				d.Set("fsx_filesystem_arn", FSxArn)
				d.SetId(DSArn)

				return []*schema.ResourceData{d}, nil
			},
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
				ValidateFunc: verify.ValidARN,
			},
			"security_group_arns": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				MaxItems: 5,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
			"subdirectory": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 4096),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"uri": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceLocationFSxLustreFileSystemCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataSyncConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	fsxArn := d.Get("fsx_filesystem_arn").(string)

	input := &datasync.CreateLocationFsxLustreInput{
		FsxFilesystemArn:  aws.String(fsxArn),
		SecurityGroupArns: flex.ExpandStringSet(d.Get("security_group_arns").(*schema.Set)),
		Tags:              Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("subdirectory"); ok {
		input.Subdirectory = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating DataSync Location Fsx Lustre File System: %#v", input)
	output, err := conn.CreateLocationFsxLustre(input)
	if err != nil {
		return fmt.Errorf("error creating DataSync Location Fsx Lustre File System: %w", err)
	}

	d.SetId(aws.StringValue(output.LocationArn))

	return resourceLocationFSxLustreFileSystemRead(d, meta)
}

func resourceLocationFSxLustreFileSystemRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataSyncConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	output, err := FindFSxLustreLocationByARN(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DataSync Location Fsx Lustre (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading DataSync Location Fsx Lustre (%s): %w", d.Id(), err)
	}

	subdirectory, err := SubdirectoryFromLocationURI(aws.StringValue(output.LocationUri))

	if err != nil {
		return err
	}

	d.Set("arn", output.LocationArn)
	d.Set("subdirectory", subdirectory)
	d.Set("uri", output.LocationUri)

	if err := d.Set("security_group_arns", flex.FlattenStringSet(output.SecurityGroupArns)); err != nil {
		return fmt.Errorf("error setting security_group_arns: %w", err)
	}

	if err := d.Set("creation_time", output.CreationTime.Format(time.RFC3339)); err != nil {
		return fmt.Errorf("error setting creation_time: %w", err)
	}

	tags, err := ListTags(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing tags for DataSync Location Fsx Lustre (%s): %w", d.Id(), err)
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

func resourceLocationFSxLustreFileSystemUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataSyncConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating DataSync Location Fsx Lustre File System (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceLocationFSxLustreFileSystemRead(d, meta)
}

func resourceLocationFSxLustreFileSystemDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataSyncConn

	input := &datasync.DeleteLocationInput{
		LocationArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting DataSync Location Fsx Lustre File System: %#v", input)
	_, err := conn.DeleteLocation(input)

	if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "not found") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting DataSync Location Fsx Lustre (%s): %w", d.Id(), err)
	}

	return nil
}
