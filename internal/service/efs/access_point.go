// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package efs

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/efs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_efs_access_point", name="Access Point")
// @Tags(identifierAttribute="id")
func ResourceAccessPoint() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAccessPointCreate,
		ReadWithoutTimeout:   resourceAccessPointRead,
		UpdateWithoutTimeout: resourceAccessPointUpdate,
		DeleteWithoutTimeout: resourceAccessPointDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"file_system_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrFileSystemID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrOwnerID: {
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
						names.AttrPath: {
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
									names.AttrPermissions: {
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceAccessPointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSClient(ctx)

	fsId := d.Get(names.AttrFileSystemID).(string)
	input := efs.CreateAccessPointInput{
		FileSystemId: aws.String(fsId),
		Tags:         getTagsIn(ctx),
	}

	if v, ok := d.GetOk("posix_user"); ok {
		input.PosixUser = expandAccessPointPOSIXUser(v.([]interface{}))
	}

	if v, ok := d.GetOk("root_directory"); ok {
		input.RootDirectory = expandAccessPointRootDirectory(v.([]interface{}))
	}

	log.Printf("[DEBUG] Creating EFS Access Point: %#v", input)

	ap, err := conn.CreateAccessPoint(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EFS Access Point for File System (%s): %s", fsId, err)
	}

	d.SetId(aws.ToString(ap.AccessPointId))

	if _, err := waitAccessPointCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EFS access point (%s) to be available: %s", d.Id(), err)
	}

	return append(diags, resourceAccessPointRead(ctx, d, meta)...)
}

func resourceAccessPointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceAccessPointRead(ctx, d, meta)...)
}

func resourceAccessPointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSClient(ctx)

	resp, err := conn.DescribeAccessPoints(ctx, &efs.DescribeAccessPointsInput{
		AccessPointId: aws.String(d.Id()),
	})
	if err != nil {
		if errs.IsA[*awstypes.AccessPointNotFound](err) {
			log.Printf("[WARN] EFS access point %q could not be found.", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading EFS access point %s: %s", d.Id(), err)
	}

	if hasEmptyAccessPoints(resp) {
		return sdkdiag.AppendErrorf(diags, "EFS access point %q could not be found.", d.Id())
	}

	ap := resp.AccessPoints[0]

	d.Set(names.AttrARN, ap.AccessPointArn)
	fsID := aws.ToString(ap.FileSystemId)
	fsARN := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  "file-system/" + fsID,
		Service:   "elasticfilesystem",
	}.String()
	d.Set("file_system_arn", fsARN)
	d.Set(names.AttrFileSystemID, fsID)
	d.Set(names.AttrOwnerID, ap.OwnerId)
	if err := d.Set("posix_user", flattenAccessPointPOSIXUser(ap.PosixUser)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting posix_user: %s", err)
	}
	if err := d.Set("root_directory", flattenAccessPointRootDirectory(ap.RootDirectory)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting root_directory: %s", err)
	}

	setTagsOut(ctx, ap.Tags)

	return diags
}

func resourceAccessPointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSClient(ctx)

	log.Printf("[DEBUG] Deleting EFS access point %q", d.Id())
	_, err := conn.DeleteAccessPoint(ctx, &efs.DeleteAccessPointInput{
		AccessPointId: aws.String(d.Id()),
	})
	if err != nil {
		if errs.IsA[*awstypes.AccessPointNotFound](err) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting EFS Access Point (%s): %s", d.Id(), err)
	}

	if _, err := waitAccessPointDeleted(ctx, conn, d.Id()); err != nil {
		if errs.IsA[*awstypes.AccessPointNotFound](err) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "waiting for EFS access point (%s) deletion: %s", d.Id(), err)
	}

	log.Printf("[DEBUG] EFS access point %q deleted.", d.Id())

	return diags
}

func hasEmptyAccessPoints(aps *efs.DescribeAccessPointsOutput) bool {
	if aps != nil && len(aps.AccessPoints) > 0 {
		return false
	}
	return true
}

func expandAccessPointPOSIXUser(pUser []interface{}) *awstypes.PosixUser {
	if len(pUser) < 1 || pUser[0] == nil {
		return nil
	}

	m := pUser[0].(map[string]interface{})

	posixUser := &awstypes.PosixUser{
		Gid: aws.Int64(int64(m["gid"].(int))),
		Uid: aws.Int64(int64(m["uid"].(int))),
	}

	if v, ok := m["secondary_gids"].(*schema.Set); ok && len(v.List()) > 0 {
		posixUser.SecondaryGids = flex.ExpandInt64ValueSet(v)
	}

	return posixUser
}

func expandAccessPointRootDirectory(rDir []interface{}) *awstypes.RootDirectory {
	if len(rDir) < 1 || rDir[0] == nil {
		return nil
	}

	m := rDir[0].(map[string]interface{})

	rootDir := &awstypes.RootDirectory{}

	if v, ok := m[names.AttrPath]; ok {
		rootDir.Path = aws.String(v.(string))
	}

	if v, ok := m["creation_info"]; ok {
		rootDir.CreationInfo = expandAccessPointRootDirectoryCreationInfo(v.([]interface{}))
	}

	return rootDir
}

func expandAccessPointRootDirectoryCreationInfo(cInfo []interface{}) *awstypes.CreationInfo {
	if len(cInfo) < 1 || cInfo[0] == nil {
		return nil
	}

	m := cInfo[0].(map[string]interface{})

	creationInfo := &awstypes.CreationInfo{
		OwnerGid:    aws.Int64(int64(m["owner_gid"].(int))),
		OwnerUid:    aws.Int64(int64(m["owner_uid"].(int))),
		Permissions: aws.String(m[names.AttrPermissions].(string)),
	}

	return creationInfo
}

func flattenAccessPointPOSIXUser(posixUser *awstypes.PosixUser) []interface{} {
	if posixUser == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"gid":            aws.ToInt64(posixUser.Gid),
		"uid":            aws.ToInt64(posixUser.Uid),
		"secondary_gids": posixUser.SecondaryGids,
	}

	return []interface{}{m}
}

func flattenAccessPointRootDirectory(rDir *awstypes.RootDirectory) []interface{} {
	if rDir == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrPath:  aws.ToString(rDir.Path),
		"creation_info": flattenAccessPointRootDirectoryCreationInfo(rDir.CreationInfo),
	}

	return []interface{}{m}
}

func flattenAccessPointRootDirectoryCreationInfo(cInfo *awstypes.CreationInfo) []interface{} {
	if cInfo == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"owner_gid":           aws.ToInt64(cInfo.OwnerGid),
		"owner_uid":           aws.ToInt64(cInfo.OwnerUid),
		names.AttrPermissions: aws.ToString(cInfo.Permissions),
	}

	return []interface{}{m}
}
