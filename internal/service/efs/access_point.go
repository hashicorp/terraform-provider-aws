package efs

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceAccessPoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceAccessPointCreate,
		Read:   resourceAccessPointRead,
		Update: resourceAccessPointUpdate,
		Delete: resourceAccessPointDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"file_system_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"file_system_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"posix_user": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"gid": {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},
						"uid": {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},
						"secondary_gids": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeInt},
							Set:      schema.HashInt,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},
			"root_directory": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				ForceNew: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"path": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Computed: true,
						},
						"creation_info": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"owner_gid": {
										Type:     schema.TypeInt,
										Required: true,
										ForceNew: true,
									},
									"owner_uid": {
										Type:     schema.TypeInt,
										Required: true,
										ForceNew: true,
									},
									"permissions": {
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceAccessPointCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EFSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	fsId := d.Get("file_system_id").(string)

	input := efs.CreateAccessPointInput{
		FileSystemId: aws.String(fsId),
		Tags:         Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("posix_user"); ok {
		input.PosixUser = expandAccessPointPOSIXUser(v.([]interface{}))
	}

	if v, ok := d.GetOk("root_directory"); ok {
		input.RootDirectory = expandAccessPointRootDirectory(v.([]interface{}))
	}

	log.Printf("[DEBUG] Creating EFS Access Point: %#v", input)

	ap, err := conn.CreateAccessPoint(&input)
	if err != nil {
		return fmt.Errorf("error creating EFS Access Point for File System (%s): %w", fsId, err)
	}

	d.SetId(aws.StringValue(ap.AccessPointId))

	if _, err := waitAccessPointCreated(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EFS access point (%s) to be available: %w", d.Id(), err)
	}

	return resourceAccessPointRead(d, meta)
}

func resourceAccessPointUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EFSConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EFS file system (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceAccessPointRead(d, meta)
}

func resourceAccessPointRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EFSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	resp, err := conn.DescribeAccessPoints(&efs.DescribeAccessPointsInput{
		AccessPointId: aws.String(d.Id()),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, efs.ErrCodeAccessPointNotFound) {
			log.Printf("[WARN] EFS access point %q could not be found.", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading EFS access point %s: %w", d.Id(), err)
	}

	if hasEmptyAccessPoints(resp) {
		return fmt.Errorf("EFS access point %q could not be found.", d.Id())
	}

	ap := resp.AccessPoints[0]

	log.Printf("[DEBUG] Found EFS access point: %#v", ap)

	fsARN := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("file-system/%s", aws.StringValue(ap.FileSystemId)),
		Service:   "elasticfilesystem",
	}.String()

	d.Set("file_system_arn", fsARN)
	d.Set("file_system_id", ap.FileSystemId)
	d.Set("arn", ap.AccessPointArn)
	d.Set("owner_id", ap.OwnerId)

	if err := d.Set("posix_user", flattenAccessPointPOSIXUser(ap.PosixUser)); err != nil {
		return fmt.Errorf("error setting posix user: %w", err)
	}

	if err := d.Set("root_directory", flattenAccessPointRootDirectory(ap.RootDirectory)); err != nil {
		return fmt.Errorf("error setting root directory: %w", err)
	}

	tags := KeyValueTags(ap.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAccessPointDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EFSConn

	log.Printf("[DEBUG] Deleting EFS access point %q", d.Id())
	_, err := conn.DeleteAccessPoint(&efs.DeleteAccessPointInput{
		AccessPointId: aws.String(d.Id()),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, efs.ErrCodeAccessPointNotFound) {
			return nil
		}
		return fmt.Errorf("error deleting EFS Access Point (%s): %w", d.Id(), err)
	}

	if _, err := waitAccessPointDeleted(conn, d.Id()); err != nil {
		if tfawserr.ErrCodeEquals(err, efs.ErrCodeAccessPointNotFound) {
			return nil
		}
		return fmt.Errorf("error waiting for EFS access point (%s) deletion: %w", d.Id(), err)
	}

	log.Printf("[DEBUG] EFS access point %q deleted.", d.Id())

	return nil
}

func hasEmptyAccessPoints(aps *efs.DescribeAccessPointsOutput) bool {
	if aps != nil && len(aps.AccessPoints) > 0 {
		return false
	}
	return true
}

func expandAccessPointPOSIXUser(pUser []interface{}) *efs.PosixUser {
	if len(pUser) < 1 || pUser[0] == nil {
		return nil
	}

	m := pUser[0].(map[string]interface{})

	posixUser := &efs.PosixUser{
		Gid: aws.Int64(int64(m["gid"].(int))),
		Uid: aws.Int64(int64(m["uid"].(int))),
	}

	if v, ok := m["secondary_gids"].(*schema.Set); ok && len(v.List()) > 0 {
		posixUser.SecondaryGids = flex.ExpandInt64Set(v)
	}

	return posixUser
}

func expandAccessPointRootDirectory(rDir []interface{}) *efs.RootDirectory {
	if len(rDir) < 1 || rDir[0] == nil {
		return nil
	}

	m := rDir[0].(map[string]interface{})

	rootDir := &efs.RootDirectory{}

	if v, ok := m["path"]; ok {
		rootDir.Path = aws.String(v.(string))
	}

	if v, ok := m["creation_info"]; ok {
		rootDir.CreationInfo = expandAccessPointRootDirectoryCreationInfo(v.([]interface{}))
	}

	return rootDir
}

func expandAccessPointRootDirectoryCreationInfo(cInfo []interface{}) *efs.CreationInfo {
	if len(cInfo) < 1 || cInfo[0] == nil {
		return nil
	}

	m := cInfo[0].(map[string]interface{})

	creationInfo := &efs.CreationInfo{
		OwnerGid:    aws.Int64(int64(m["owner_gid"].(int))),
		OwnerUid:    aws.Int64(int64(m["owner_uid"].(int))),
		Permissions: aws.String(m["permissions"].(string)),
	}

	return creationInfo
}

func flattenAccessPointPOSIXUser(posixUser *efs.PosixUser) []interface{} {
	if posixUser == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"gid":            aws.Int64Value(posixUser.Gid),
		"uid":            aws.Int64Value(posixUser.Uid),
		"secondary_gids": aws.Int64ValueSlice(posixUser.SecondaryGids),
	}

	return []interface{}{m}
}

func flattenAccessPointRootDirectory(rDir *efs.RootDirectory) []interface{} {
	if rDir == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"path":          aws.StringValue(rDir.Path),
		"creation_info": flattenAccessPointRootDirectoryCreationInfo(rDir.CreationInfo),
	}

	return []interface{}{m}
}

func flattenAccessPointRootDirectoryCreationInfo(cInfo *efs.CreationInfo) []interface{} {
	if cInfo == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"owner_gid":   aws.Int64Value(cInfo.OwnerGid),
		"owner_uid":   aws.Int64Value(cInfo.OwnerUid),
		"permissions": aws.StringValue(cInfo.Permissions),
	}

	return []interface{}{m}
}
